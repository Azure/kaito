# Copyright (c) KAITO authors.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Benchmark concurrent Secrets scanning and streaming re-verification."""

import argparse
import json
import multiprocessing
import os
import platform
import resource
import sys
import time
from concurrent.futures import ThreadPoolExecutor
from dataclasses import asdict, dataclass
from datetime import UTC, datetime
from pathlib import Path
from typing import Any

from guardrail_core.secrets import NativeSecretsScanner

DEFAULT_CONCURRENCIES = (10, 50, 100)
SCHEMA_VERSION = 1


@dataclass(frozen=True)
class LatencyPercentiles:
    p50: float
    p95: float
    p99: float


@dataclass(frozen=True)
class BenchmarkConfig:
    concurrency: int
    windows_per_stream: int
    payload_bytes: int


@dataclass(frozen=True)
class BenchmarkResult:
    concurrency: int
    windows_per_stream: int
    operations: int
    wall_seconds: float
    throughput_operations_per_second: float
    cpu_seconds: float
    cpu_utilization_percent: float
    peak_rss_bytes: int
    peak_rss_growth_bytes: int
    io_read_bytes: int | None
    io_write_bytes: int | None
    scan_latency_ms: LatencyPercentiles
    reverification_latency_ms: LatencyPercentiles


def run_scenario(config: BenchmarkConfig) -> BenchmarkResult:
    """Run one benchmark scenario in the current process."""
    _validate_config(config)
    payload = _build_payload(config.payload_bytes)
    io_before = _read_proc_io()
    rss_before = _peak_rss_bytes()
    cpu_before = time.process_time()
    wall_before = time.perf_counter()

    with ThreadPoolExecutor(max_workers=config.concurrency) as executor:
        stream_results = list(
            executor.map(
                lambda _: _run_stream(payload, config.windows_per_stream),
                range(config.concurrency),
            )
        )

    wall_seconds = time.perf_counter() - wall_before
    cpu_seconds = time.process_time() - cpu_before
    rss_after = _peak_rss_bytes()
    io_after = _read_proc_io()
    scan_latencies = [
        latency for scan_samples, _ in stream_results for latency in scan_samples
    ]
    reverification_latencies = [
        latency
        for _, verification_samples in stream_results
        for latency in verification_samples
    ]
    operations = len(scan_latencies) + len(reverification_latencies)

    return BenchmarkResult(
        concurrency=config.concurrency,
        windows_per_stream=config.windows_per_stream,
        operations=operations,
        wall_seconds=wall_seconds,
        throughput_operations_per_second=operations / wall_seconds,
        cpu_seconds=cpu_seconds,
        cpu_utilization_percent=(cpu_seconds / wall_seconds) * 100,
        peak_rss_bytes=rss_after,
        peak_rss_growth_bytes=max(0, rss_after - rss_before),
        io_read_bytes=_io_delta(io_before, io_after, "read_bytes"),
        io_write_bytes=_io_delta(io_before, io_after, "write_bytes"),
        scan_latency_ms=_latency_percentiles(scan_latencies),
        reverification_latency_ms=_latency_percentiles(reverification_latencies),
    )


def run_isolated_scenario(config: BenchmarkConfig) -> BenchmarkResult:
    """Run a scenario in a fresh process so resource metrics are isolated."""
    context = multiprocessing.get_context("spawn")
    parent_connection, child_connection = context.Pipe(duplex=False)
    process = context.Process(
        target=_scenario_worker,
        args=(config, child_connection),
    )
    process.start()
    child_connection.close()
    try:
        result = parent_connection.recv()
    finally:
        parent_connection.close()
        process.join()

    if process.exitcode != 0:
        raise RuntimeError(f"benchmark worker failed with exit code {process.exitcode}")
    if isinstance(result, Exception):
        raise result
    return result


def build_report(
    results: list[BenchmarkResult],
    *,
    baseline: dict[str, Any] | None = None,
) -> dict[str, Any]:
    """Build a machine-readable benchmark report."""
    report: dict[str, Any] = {
        "schema_version": SCHEMA_VERSION,
        "generated_at": datetime.now(UTC).isoformat(),
        "environment": {
            "python": platform.python_version(),
            "platform": platform.platform(),
            "cpu_count": os.cpu_count(),
            "implementation": str(
                Path(sys.modules[NativeSecretsScanner.__module__].__file__)
            ),
        },
        "results": [asdict(result) for result in results],
    }
    if baseline is not None:
        report["comparison"] = compare_reports(report, baseline)
    return report


def compare_reports(
    current: dict[str, Any],
    baseline: dict[str, Any],
) -> list[dict[str, float | int]]:
    """Compare current results with a baseline by concurrency."""
    baseline_by_concurrency = {
        result["concurrency"]: result for result in baseline["results"]
    }
    comparisons = []
    for result in current["results"]:
        previous = baseline_by_concurrency.get(result["concurrency"])
        if previous is None:
            continue
        comparisons.append(
            {
                "concurrency": result["concurrency"],
                "throughput_change_percent": _percent_change(
                    previous["throughput_operations_per_second"],
                    result["throughput_operations_per_second"],
                ),
                "scan_p95_change_percent": _percent_change(
                    previous["scan_latency_ms"]["p95"],
                    result["scan_latency_ms"]["p95"],
                ),
                "reverification_p95_change_percent": _percent_change(
                    previous["reverification_latency_ms"]["p95"],
                    result["reverification_latency_ms"]["p95"],
                ),
                "cpu_per_operation_change_percent": _percent_change(
                    previous["cpu_seconds"] / previous["operations"],
                    result["cpu_seconds"] / result["operations"],
                ),
                "peak_rss_change_bytes": (
                    result["peak_rss_bytes"] - previous["peak_rss_bytes"]
                ),
                "io_write_change_bytes": _optional_delta(
                    previous["io_write_bytes"],
                    result["io_write_bytes"],
                ),
            }
        )
    return comparisons


def _scenario_worker(config: BenchmarkConfig, connection: Any) -> None:
    try:
        connection.send(run_scenario(config))
    except Exception as error:
        connection.send(error)
    finally:
        connection.close()


def _run_stream(
    payload: str,
    windows_per_stream: int,
) -> tuple[list[float], list[float]]:
    scanner = NativeSecretsScanner(redact_mode="all")
    scan_latencies = []
    reverification_latencies = []

    for _ in range(windows_per_stream):
        start = time.perf_counter_ns()
        result = scanner.scan("", payload)
        scan_latencies.append(_elapsed_ms(start))
        if result.valid or "******" not in result.output:
            raise RuntimeError("Secrets scanner did not redact benchmark payload")

        start = time.perf_counter_ns()
        verified = scanner.scan("", result.output)
        reverification_latencies.append(_elapsed_ms(start))
        if not verified.valid or verified.output != result.output:
            raise RuntimeError("Secrets scanner failed clean-output re-verification")

    return scan_latencies, reverification_latencies


def _build_payload(payload_bytes: int) -> str:
    secret = "AKIA" + "A" * 16
    prefix = ("safe model output " * ((payload_bytes // 18) + 1))[:payload_bytes]
    return f"{prefix} key={secret}"


def _latency_percentiles(samples: list[float]) -> LatencyPercentiles:
    return LatencyPercentiles(
        p50=_percentile(samples, 0.50),
        p95=_percentile(samples, 0.95),
        p99=_percentile(samples, 0.99),
    )


def _percentile(samples: list[float], percentile: float) -> float:
    if not samples:
        raise ValueError("percentile requires at least one sample")
    ordered = sorted(samples)
    position = (len(ordered) - 1) * percentile
    lower = int(position)
    upper = min(lower + 1, len(ordered) - 1)
    fraction = position - lower
    return ordered[lower] + (ordered[upper] - ordered[lower]) * fraction


def _elapsed_ms(start_ns: int) -> float:
    return (time.perf_counter_ns() - start_ns) / 1_000_000


def _peak_rss_bytes() -> int:
    peak_rss = resource.getrusage(resource.RUSAGE_SELF).ru_maxrss
    return peak_rss if sys.platform == "darwin" else peak_rss * 1024


def _read_proc_io() -> dict[str, int] | None:
    try:
        lines = Path("/proc/self/io").read_text(encoding="utf-8").splitlines()
    except OSError:
        return None
    return {
        key: int(value.strip())
        for line in lines
        for key, value in [line.split(":", maxsplit=1)]
    }


def _io_delta(
    before: dict[str, int] | None,
    after: dict[str, int] | None,
    key: str,
) -> int | None:
    if before is None or after is None:
        return None
    return after[key] - before[key]


def _optional_delta(previous: int | None, current: int | None) -> int:
    if previous is None or current is None:
        return 0
    return current - previous


def _percent_change(previous: float, current: float) -> float:
    if previous == 0:
        return 0.0
    return ((current - previous) / previous) * 100


def _validate_config(config: BenchmarkConfig) -> None:
    if config.concurrency <= 0:
        raise ValueError("concurrency must be positive")
    if config.windows_per_stream <= 0:
        raise ValueError("windows_per_stream must be positive")
    if config.payload_bytes <= 0:
        raise ValueError("payload_bytes must be positive")


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--concurrency",
        nargs="+",
        type=int,
        default=DEFAULT_CONCURRENCIES,
    )
    parser.add_argument("--windows-per-stream", type=int, default=20)
    parser.add_argument("--payload-bytes", type=int, default=1024)
    parser.add_argument("--baseline", type=Path)
    parser.add_argument("--output", type=Path)
    args = parser.parse_args()

    results = [
        run_isolated_scenario(
            BenchmarkConfig(
                concurrency=concurrency,
                windows_per_stream=args.windows_per_stream,
                payload_bytes=args.payload_bytes,
            )
        )
        for concurrency in args.concurrency
    ]
    baseline = (
        json.loads(args.baseline.read_text(encoding="utf-8")) if args.baseline else None
    )
    report = build_report(results, baseline=baseline)
    serialized = json.dumps(report, indent=2, sort_keys=True)
    if args.output:
        args.output.write_text(serialized + "\n", encoding="utf-8")
    print(serialized)


if __name__ == "__main__":
    main()

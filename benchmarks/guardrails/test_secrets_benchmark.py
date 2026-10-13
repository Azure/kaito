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

"""Tests for the Secrets benchmark."""

import pytest

from benchmarks.guardrails.secrets_benchmark import (
    BenchmarkConfig,
    _percentile,
    build_report,
    compare_reports,
    run_scenario,
)


def test_percentile_interpolates_samples():
    assert _percentile([1.0, 2.0, 3.0, 4.0], 0.50) == 2.5
    assert _percentile([1.0, 2.0, 3.0, 4.0], 0.95) == pytest.approx(3.85)


def test_run_scenario_measures_scan_and_reverification():
    result = run_scenario(
        BenchmarkConfig(
            concurrency=2,
            windows_per_stream=2,
            payload_bytes=128,
        )
    )

    assert result.operations == 8
    assert result.throughput_operations_per_second > 0
    assert result.scan_latency_ms.p50 > 0
    assert result.reverification_latency_ms.p50 > 0
    assert result.peak_rss_bytes > 0


def test_compare_reports_matches_concurrency():
    result = run_scenario(
        BenchmarkConfig(
            concurrency=1,
            windows_per_stream=1,
            payload_bytes=64,
        )
    )
    baseline = build_report([result])
    current = build_report([result])

    assert compare_reports(current, baseline) == [
        {
            "concurrency": 1,
            "throughput_change_percent": 0.0,
            "scan_p95_change_percent": 0.0,
            "reverification_p95_change_percent": 0.0,
            "cpu_per_operation_change_percent": 0.0,
            "peak_rss_change_bytes": 0,
            "io_write_change_bytes": 0,
        }
    ]

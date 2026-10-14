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

"""Tests for KAITO-owned scanner orchestration."""

from guardrail_core.scanning import ScanResult, scan_output


class _ReplacingScanner:
    def __init__(self, old: str, new: str, valid: bool, score: float) -> None:
        self._old = old
        self._new = new
        self._valid = valid
        self._score = score

    def scan(self, prompt: str, output: str) -> ScanResult:
        del prompt
        return ScanResult(
            output.replace(self._old, self._new),
            self._valid,
            self._score,
        )


class _FailingScanner:
    def scan(self, prompt: str, output: str) -> ScanResult:
        del prompt
        return ScanResult(f"{output}-failed", False, 1.0)


class _PassingScanner:
    def scan(self, prompt: str, output: str) -> ScanResult:
        del prompt
        return ScanResult(f"{output}-passed", True, -1.0)


def test_scan_output_runs_scanners_sequentially():
    scanners = [
        _ReplacingScanner("secret", "redacted", False, 1.0),
        _ReplacingScanner("redacted", "safe", True, -1.0),
    ]

    assert scan_output(scanners, "prompt", "a secret") == (
        "a safe",
        {"_ReplacingScanner": True},
        {"_ReplacingScanner": -1.0},
    )


def test_scan_output_stops_after_invalid_scanner_when_fail_fast():
    assert scan_output(
        [_FailingScanner(), _PassingScanner()],
        "prompt",
        "output",
        fail_fast=True,
    ) == (
        "output-failed",
        {"_FailingScanner": False},
        {"_FailingScanner": 1.0},
    )


def test_scan_output_returns_empty_results_without_scanning_empty_output():
    assert scan_output([_FailingScanner()], "prompt", "  ") == ("  ", {}, {})


def test_scan_output_returns_empty_results_without_scanning_none_output():
    assert scan_output([_FailingScanner()], "prompt", None) == (None, {}, {})

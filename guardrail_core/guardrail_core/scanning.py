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

"""KAITO-owned guardrail scanner orchestration."""

from collections.abc import Sequence
from typing import NamedTuple, Protocol


class ScanResult(NamedTuple):
    """Result returned by a single guardrail scanner."""

    output: str
    valid: bool
    score: float


class Scanner(Protocol):
    def scan(self, prompt: str, output: str) -> ScanResult: ...


def scan_output(
    scanners: Sequence[Scanner],
    prompt: str,
    output: str | None,
    fail_fast: bool = False,
) -> tuple[str | None, dict[str, bool], dict[str, float]]:
    """Run output scanners sequentially and collect their validity and scores."""
    results_valid: dict[str, bool] = {}
    results_score: dict[str, float] = {}

    if output is None:
        return output, results_valid, results_score

    sanitized_output = output
    if not scanners or output.strip() == "":
        return output, results_valid, results_score

    for scanner in scanners:
        result = scanner.scan(
            prompt,
            sanitized_output,
        )
        sanitized_output = result.output
        scanner_name = type(scanner).__name__
        results_valid[scanner_name] = result.valid
        results_score[scanner_name] = result.score
        if fail_fast and not result.valid:
            break

    return sanitized_output, results_valid, results_score

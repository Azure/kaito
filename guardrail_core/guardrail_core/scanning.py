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

from typing import Protocol


class Scanner(Protocol):
    def scan(self, prompt: str, output: str) -> tuple[str, bool, float]: ...


def scan_output(
    scanners: list[Scanner],
    prompt: str,
    output: str,
    fail_fast: bool = False,
) -> tuple[str, dict[str, bool], dict[str, float]]:
    """Run output scanners sequentially and collect their validity and scores."""
    sanitized_output = output
    results_valid: dict[str, bool] = {}
    results_score: dict[str, float] = {}

    if not scanners or output.strip() == "":
        return sanitized_output, results_valid, results_score

    for scanner in scanners:
        sanitized_output, is_valid, risk_score = scanner.scan(
            prompt,
            sanitized_output,
        )
        scanner_name = type(scanner).__name__
        results_valid[scanner_name] = is_valid
        results_score[scanner_name] = risk_score
        if fail_fast and not is_valid:
            break

    return sanitized_output, results_valid, results_score

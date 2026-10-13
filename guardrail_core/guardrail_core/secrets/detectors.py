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

"""In-memory detect-secrets execution."""

from detect_secrets.core.plugins.util import get_mapping_from_secret_type_to_class
from detect_secrets.filters.allowlist import is_line_allowlisted
from detect_secrets.filters.heuristic import (
    is_indirect_reference,
    is_likely_id_string,
    is_not_alphanumeric_string,
    is_potential_uuid,
    is_prefixed_with_dollar_sign,
    is_sequential_string,
    is_templated_secret,
)
from detect_secrets.plugins.base import BasePlugin
from detect_secrets.plugins.private_key import PrivateKeyDetector
from detect_secrets.util.code_snippet import CodeSnippet, get_code_snippet

_MODEL_OUTPUT_FILENAME = "model-output.txt"
_MIN_LINE_LENGTH = 5
_MAX_LINE_LENGTH = 100_000
_BUILT_IN_PLUGIN_TYPES = tuple(get_mapping_from_secret_type_to_class().values())


def detect_builtin_secrets(text: str) -> set[str]:
    """Run isolated detect-secrets plugin instances against in-memory text."""
    plugins = tuple(plugin_type() for plugin_type in _BUILT_IN_PLUGIN_TYPES)
    lines = text.splitlines() or [text]
    secret_values: set[str] = set()

    for line_number, raw_line in enumerate(lines, start=1):
        line = raw_line.strip()
        if not _MIN_LINE_LENGTH <= len(line) <= _MAX_LINE_LENGTH:
            continue

        context = get_code_snippet(lines=lines, line_number=line_number)
        if _is_line_filtered(line, context):
            continue

        for plugin in plugins:
            if isinstance(plugin, PrivateKeyDetector):
                continue

            findings = plugin.analyze_line(
                filename=_MODEL_OUTPUT_FILENAME,
                line=line,
                line_number=line_number,
                context=context,
                raw_context=context,
            )
            secret_values.update(
                secret_value
                for finding in findings
                if (secret_value := _normalize_secret(finding.secret_value))
                and not _is_secret_filtered(secret_value, line, plugin)
            )

    secret_values.update(_detect_private_keys(text, plugins))
    return secret_values


def _detect_private_keys(text: str, plugins: tuple[BasePlugin, ...]) -> set[str]:
    for plugin in plugins:
        if isinstance(plugin, PrivateKeyDetector):
            return {
                secret_value
                for value in plugin.analyze_string(text)
                if (secret_value := _normalize_secret(value))
                and not _is_secret_filtered(secret_value, text, plugin)
            }
    return set()


def _normalize_secret(value: str | None) -> str:
    return value.strip() if value else ""


def _is_line_filtered(line: str, context: CodeSnippet) -> bool:
    return is_line_allowlisted(_MODEL_OUTPUT_FILENAME, line, context) or (
        is_indirect_reference(line)
    )


def _is_secret_filtered(secret: str, line: str, plugin: BasePlugin) -> bool:
    return (
        is_sequential_string(secret)
        or is_potential_uuid(secret)
        or is_likely_id_string(secret, line, plugin)
        or is_templated_secret(secret)
        or is_prefixed_with_dollar_sign(secret)
        or is_not_alphanumeric_string(secret)
    )

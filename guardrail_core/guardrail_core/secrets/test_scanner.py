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

"""Tests for the KAITO-owned secrets scanner."""

import tempfile
import threading
from concurrent.futures import ThreadPoolExecutor
from copy import deepcopy

import pytest
from detect_secrets.settings import get_settings

import guardrail_core.secrets.detectors as detectors_module
from guardrail_core.secrets import NativeSecretsScanner

_AWS_KEY = "AKIA" + "A" * 16
_AZURE_KEY = "A" * 86 + "=="
_GITHUB_TOKEN = "ghp_" + "A" * 36
_GITHUB_FINE_GRAINED_TOKEN = "github_pat_" + "A" * 82
_GCP_KEY = "AIza" + "A" * 35
_OPENAI_KEY = "sk-" + "A" * 20 + "T3BlbkFJ" + "B" * 20
_OPENAI_PROJECT_KEY = "sk-proj-" + "A" * 80
_SLACK_TOKEN = "-".join(
    ("xoxb", "1234567890", "1234567890", "abcdefghijklmnopqrstuvwx")
)
_JWT = "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.c2lnbmF0dXJl"
_PRIVATE_KEY_BODY = "A" * 32
_PRIVATE_KEY = (
    f"-----BEGIN PRIVATE KEY-----\n{_PRIVATE_KEY_BODY}\n-----END PRIVATE KEY-----"
)
_BASIC_AUTH_PASSWORD = "password"
_BASIC_AUTH_URL = f"https://user:{_BASIC_AUTH_PASSWORD}@internal.test/path"


def test_native_scanners_keeps_compatibility_export():
    from guardrail_core.native_scanners import (
        NativeSecretsScanner as CompatibilitySecretsScanner,
    )

    assert CompatibilitySecretsScanner is NativeSecretsScanner


@pytest.mark.parametrize(
    ("provider", "text", "expected_secret"),
    [
        ("AWS", _AWS_KEY, _AWS_KEY),
        ("Azure", f"AccountKey={_AZURE_KEY}", _AZURE_KEY),
        ("GitHub", _GITHUB_TOKEN, _GITHUB_TOKEN),
        ("GitHub fine-grained", _GITHUB_FINE_GRAINED_TOKEN, _GITHUB_FINE_GRAINED_TOKEN),
        ("GCP", _GCP_KEY, _GCP_KEY),
        ("OpenAI", _OPENAI_KEY, _OPENAI_KEY),
        ("OpenAI project", _OPENAI_PROJECT_KEY, _OPENAI_PROJECT_KEY),
        ("Slack", _SLACK_TOKEN, _SLACK_TOKEN),
        ("JWT", _JWT, _JWT),
        ("Private Key", _PRIVATE_KEY, _PRIVATE_KEY_BODY),
        ("Basic Auth", _BASIC_AUTH_URL, _BASIC_AUTH_PASSWORD),
    ],
)
def test_detects_supported_provider_secrets(provider, text, expected_secret):
    del provider
    scanner = NativeSecretsScanner()

    assert expected_secret in scanner._detect_secret_values(text)


@pytest.mark.parametrize(
    ("provider", "near_match"),
    [
        ("AWS", "AKIA" + "A" * 15),
        ("Azure", "azure " + "A" * 83 + "=="),
        ("GitHub", "ghp_" + "A" * 35),
        ("GitHub fine-grained", "github_pat_" + "A" * 81),
        ("GCP", "AIza" + "A" * 34),
        ("OpenAI", "sk-" + "A" * 19 + "T3BlbkFJ" + "B" * 20),
        ("OpenAI project", "sk-proj-" + "A" * 79),
        ("Slack", "xoxz-1234567890-abcdefghijklmnopqrstuvwx"),
        ("JWT", "eyJnot-json.not-json.signature"),
        ("Private Key", "-----BEGIN PUBLIC KEY-----\nAAAA\n-----END PUBLIC KEY-----"),
        ("Basic Auth", "https://user:password@example.com/path"),
    ],
)
def test_ignores_provider_near_matches(provider, near_match):
    del provider
    scanner = NativeSecretsScanner()

    assert scanner._detect_secret_values(near_match) == set()


def test_allows_text_without_secrets():
    scanner = NativeSecretsScanner()

    assert scanner.scan("", "ordinary model output") == (
        "ordinary model output",
        True,
        -1.0,
    )


@pytest.mark.parametrize(
    ("secret", "redact_mode", "expected"),
    [
        ("secret-token", "all", "******"),
        ("secret-token", "partial", "se..en"),
        ("abc", "partial", "ab..bc"),
        ("secret-token", "hash", "3c28d89b80f70302b04fce2a1451f6ea"),
    ],
)
def test_matches_legacy_redaction(monkeypatch, secret, redact_mode, expected):
    scanner = NativeSecretsScanner(redact_mode=redact_mode)
    monkeypatch.setattr(scanner, "_detect_secret_values", lambda output: {secret})

    assert scanner.scan("", secret) == (expected, False, 1.0)


def test_rejects_unknown_redaction_mode(monkeypatch):
    scanner = NativeSecretsScanner(redact_mode="unknown")
    monkeypatch.setattr(scanner, "_detect_secret_values", lambda output: {"secret"})

    with pytest.raises(ValueError, match="redact mode wasn't recognized unknown"):
        scanner.scan("", "secret")


def test_redacts_multiple_provider_secrets():
    scanner = NativeSecretsScanner(redact_mode="all")

    assert scanner.scan("", f"{_AWS_KEY} {_GCP_KEY} {_OPENAI_KEY}") == (
        "****** ****** ******",
        False,
        1.0,
    )


def test_redacts_repeated_secrets():
    scanner = NativeSecretsScanner(redact_mode="all")

    assert scanner.scan("", f"{_GITHUB_TOKEN} and {_GITHUB_TOKEN}") == (
        "****** and ******",
        False,
        1.0,
    )


def test_concurrent_scans_preserve_detection():
    scanner = NativeSecretsScanner(redact_mode="all")
    samples = [
        (_AWS_KEY, ("******", False, 1.0)),
        (_GITHUB_TOKEN, ("******", False, 1.0)),
        ("plain text", ("plain text", True, -1.0)),
    ]

    def scan(index):
        output, expected = samples[index % len(samples)]
        return scanner.scan("", output), expected

    with ThreadPoolExecutor(max_workers=20) as executor:
        results = executor.map(scan, range(100))

    for actual, expected in results:
        assert actual == expected


def test_scan_does_not_write_model_output_to_disk(monkeypatch):
    def fail_if_called(*args, **kwargs):
        del args, kwargs
        raise AssertionError("scanner must not create temporary files")

    monkeypatch.setattr(tempfile, "NamedTemporaryFile", fail_if_called)

    assert NativeSecretsScanner().scan("", _AWS_KEY) == ("******", False, 1.0)


def test_scan_does_not_mutate_global_detect_secrets_settings():
    settings = get_settings()
    plugins_before = deepcopy(settings.plugins)
    filters_before = deepcopy(settings.filters)

    NativeSecretsScanner().scan("", _AWS_KEY)

    assert settings.plugins == plugins_before
    assert settings.filters == filters_before


def test_concurrent_scans_are_not_serialized(monkeypatch):
    barrier = threading.Barrier(2)
    instance_ids = set()
    instance_lock = threading.Lock()

    class ConcurrentPlugin:
        def __init__(self):
            with instance_lock:
                instance_ids.add(id(self))

        def analyze_line(self, **kwargs):
            del kwargs
            barrier.wait(timeout=2)
            return set()

    monkeypatch.setattr(
        detectors_module,
        "_BUILT_IN_PLUGIN_TYPES",
        (ConcurrentPlugin,),
    )

    scanner = NativeSecretsScanner()
    with ThreadPoolExecutor(max_workers=2) as executor:
        results = list(executor.map(lambda _: scanner.scan("", "plain text"), range(2)))

    assert results == [
        ("plain text", True, -1.0),
        ("plain text", True, -1.0),
    ]
    assert len(instance_ids) == 2

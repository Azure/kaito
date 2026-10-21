# KAITO guardrail core

`guardrail-core` contains shared scanner implementations, scanner orchestration,
and a policy-independent streaming text window. Install it from the repository
root with `python -m pip install ./guardrail_core`.

## Streaming window

The `guardrail_core.streaming_window` module uses only the Python standard
library. It can be imported without loading scanner dependencies, FastAPI,
RAGEngine configuration, or Envoy.

Its public types are:

- `StreamingBufferWindow(scanner, *, holdback_len)`
- `WindowScanner`, a structural protocol with
  `scan(text, *, flush=False, preceding_char="") -> WindowScanResult`
- `WindowScanResult(sanitized_text=None, blocked=False)`
- `WindowEmitResult(chunks, blocked=False)`
- `WindowScanError`, a scanner failure with the original exception as `__cause__`

### Units and release rules

`holdback_len` is a non-negative count of **Python string characters (Unicode
code points)**, not UTF-8 bytes, tokens, or grapheme clusters. It uses the same
units as `len(text)` and Python string slicing, including for Chinese text.

`feed(text)` appends to the pending buffer. If the buffer exceeds the holdback,
the scanner sees the **whole** pending buffer, including the retained tail,
before any prefix is released. If it does not exceed the holdback, no scanning
or release is needed until a later feed or flush. Zero holdback scans and
releases each non-empty feed immediately.

The scanner may:

- Return `WindowScanResult()` to allow the pending text unchanged.
- Return `sanitized_text` to replace the entire pending buffer. Release length
  is recalculated from the sanitized text, so shortening, expanding, or removing
  text does not release an unchecked tail. `redacted` becomes true.
- Return `blocked=True` to discard pending text and permanently stop release.
  Blocking takes precedence over `sanitized_text`.

`preceding_char` is the last character actually released, or `""` before the
first release. The scanner can use it for left word-boundary checks.
`flush()` scans the remaining buffer with `flush=True`, allowing the scanner to
resolve a right boundary at the end of the stream; only then does it release
the remaining sanitized text. Flushing an empty buffer emits nothing.

`blocked` and `redacted` are readable window state. After blocking, every
`feed()` and `flush()` returns an empty, blocked result without calling the
scanner again. A normal flush does not close the window; the caller determines
the stream lifecycle.

### Exception contract

An operational exception raised by `scanner.scan()` becomes `WindowScanError`,
with the original exception chained. The catch is limited to that scanner call,
not the caller's stream processing. No failed scan is treated as an allow:
the window discards pending text, becomes blocked, and raises without returning
an emission result. Do not retry a failed window; create a new window for a new
stream.

Cancellation and other `BaseException` control signals are not converted into
scan errors. Callers must terminate the stream on cancellation.

The window emits no refusal messages, does no logging of model text, and chooses
no fail-open policy. The caller must handle `WindowScanError` explicitly.
RAGEngine records the failure, uses its existing block-message / `content_filter`
/ `[DONE]` refusal flow, and closes the upstream iterator.

### Policy responsibilities

The window only buffers, invokes the scanner, and releases scanner-confirmed
text. A caller must choose sufficient holdback for its scanners; zero or
insufficient holdback does not guarantee detection of cross-feed patterns.

RAGEngine continues to own scanner construction, required holdback calculation,
redact-before-block ordering, word-match semantics, Secrets post-redaction
verification, response formatting, and upstream closure. Its old window module
only re-exports these shared types for compatibility.

Gateway consumers should import this window rather than copy it. Any future
shared policy-execution capability belongs in this core package, not in a
separate gateway implementation.

## Validation

```bash
python -m pip install -e ./guardrail_core
python -m pytest guardrail_core/guardrail_core/test_streaming_window.py presets/ragengine/tests/streaming
python -m build --wheel ./guardrail_core
```

The existing setuptools package discovery automatically includes the new module
in the wheel. Moved window tests are also included in `make rag-service-test`
so the normal RAGEngine CI continues to exercise the shared implementation.

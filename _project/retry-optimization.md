# Retry Policy Optimization — Cover All Transient Server Errors (5xx)

## Motivation

The client's `isRetryableError` (`client/retry.go`) currently retries only HTTP **429, 502, 503,
504** (plus network/DNS errors). It does **not** retry **500 Internal Server Error**.

Azure OpenAI (and OpenAI) return transient `500`s under load — a single such response on one call
is not recoverable today. A downstream consumer (Herald) hit this: one page in a 27-page
classification received a transient `500`, which propagated as a hard failure and aborted the
entire document (the per-page calls run in a fail-fast `errgroup`, so one non-retryable error kills
the batch).

This is a gap relative to standard LLM SDK behavior. The official **OpenAI** and **Anthropic** SDKs
both retry `408`, `429`, and **all `5xx`** responses. tau enumerating only `429/502/503/504` is the
anomaly. The fix belongs in the client (retry is a transport concern), keeping retry a single owned
mechanism that consumers tune purely via `agent.client.retry` config — no per-consumer retry code.

## Change 1 — `client/retry.go`: broaden the retryable status set

In `isRetryableError`, replace the enumerated status check with the standard transient set:
`408 Request Timeout`, `429 Too Many Requests`, and **all `5xx`** server errors.

```go
// before
var httpErr *HTTPStatusError
if errors.As(err, &httpErr) {
    return httpErr.StatusCode == 429 ||
        httpErr.StatusCode == 502 ||
        httpErr.StatusCode == 503 ||
        httpErr.StatusCode == 504
}

// after
var httpErr *HTTPStatusError
if errors.As(err, &httpErr) {
    return httpErr.StatusCode == http.StatusRequestTimeout || // 408
        httpErr.StatusCode == http.StatusTooManyRequests || // 429
        httpErr.StatusCode >= 500 // 500, 502, 503, 504, … all server errors
}
```

Add `"net/http"` to the import block (currently imports `net` and `net/url`, not `net/http`).

Rationale for `>= 500` rather than re-enumerating: future-proof (covers any 5xx, e.g., `500`, `529`),
matches OpenAI/Anthropic SDK policy, and is simpler than maintaining an explicit list. `4xx` other
than `408`/`429` remain non-retryable (deterministic client errors — including `400` content-filter
rejections, which must NOT be retried). Context-cancellation and network/DNS handling are unchanged.

## Change 2 — `client/doc.go`: update the documented policy

The `# Retry Logic` section lists:

```
//   - HTTP 429 (rate limit), 502, 503, 504 (server errors)
```

Update it to reflect the new policy, e.g.:

```
//   - HTTP 408 (request timeout), 429 (rate limit), and all 5xx server errors
//   - Network operation errors (connection failures, timeouts) and temporary DNS errors
```

Also review the `Execute` interface doc comment in `client/client.go`, which currently says
"Automatically retries on transient failures (HTTP 429/502/503/504, network errors)" — update to
"HTTP 408/429 and 5xx, network errors".

## Change 3 — test coverage (`tests/client/`, black-box `package client_test`)

`isRetryableError` is unexported, so verify behavior through `Execute` against an `httptest` server.
Add table-driven cases asserting retry vs no-retry by counting received requests:

- **Retries to exhaustion then succeeds:** server returns `500` for the first N calls, then `200`;
  assert the final result is the success and the request count == N+1 (within `MaxRetries`).
- **Retryable codes:** `408`, `429`, `500`, `502`, `503`, `504` each trigger > 1 attempt.
- **Non-retryable codes:** `400`, `401`, `404` each result in exactly 1 attempt (no retry) — this
  guards the content-filter `400` case so it fails fast, not after retries.
- Use a tiny `RetryConfig` (e.g., `InitialBackoff: "1ms"`, `MaxRetries: 2`) so the test is fast.

(If tau prefers a white-box unit test of `isRetryableError` directly, add a `package client` test
file in the module asserting the boolean per status code; the black-box `Execute` test above is the
behavior-level guard and is recommended regardless.)

## Change 4 — CHANGELOG + version

Add to `CHANGELOG.md` under a new `## [v0.1.2]` heading:

```
### Changed
- Broaden client retry policy to cover all transient server errors: retry HTTP 408, 429, and all
  5xx responses (previously only 429/502/503/504), aligning with OpenAI/Anthropic SDK behavior.
  Transient 500s are now retried with the existing exponential backoff + jitter.
```

Tag/release `v0.1.2` per the tau release workflow.

## Downstream integration (Herald)

After `agent v0.1.2` is published, Herald upgrades with:

```
go get github.com/tailored-agentic-units/agent@v0.1.2
mise run vet && mise run test
```

No Herald code change is required — Herald already uses the client's default retry (3×, jitter), so
it inherits the broader policy automatically. Retry parameters remain tunable via Herald's
`agent.client.retry` config if a different posture is wanted (e.g., higher `max_retries` for the
27+ page documents).

## Scope notes

- **Do not** make `4xx` (other than 408/429) retryable — Azure content-filter rejections are `400`
  and must fail fast.
- Retry remains entirely client-side; consumers carry no retry logic. This was a deliberate design
  decision (see Herald's `prompt-infrastructure-review` / retry discussion): retry is a transport
  concern owned by the client and tuned via agent configuration.

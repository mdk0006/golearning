# Day 19 — Structured Logging in Go

---

> **بِسْمِ اللهِ الرَّحْمَنِ الرَّحِيم**
>
> **Al-Khabir — الخَبِير — The All-Aware**
>
> _He knows every detail of every thing — nothing is hidden from Him. In your systems, structured logging is how you become aware of what's happening inside. Begin with His name._

---

## Blog of the Day

[Structured Logging in Go — go.dev/blog](https://go.dev/blog/slog)

Read this after the session. Covers the design decisions behind `log/slog`, how handlers work, and how to build custom handlers for advanced use cases.

---

## Concept 1: Why `log.Printf` Breaks in Production

`log.Printf` and `fmt.Println` produce unstructured strings:

```
2026/07/11 10:01:23 health check failed for web-01
2026/07/11 10:01:23 retrying request
2026/07/11 10:01:24 timeout after 3s
```

With 50 pods logging this, you can't answer:
- Which pod logged this?
- What region is web-01 in?
- How many times did this happen in the last 10 minutes?
- Can Loki or CloudWatch query this?

Unstructured logs are readable by humans — not queryable by machines.

---

## Concept 2: Structured Logging — Key-Value Pairs

Structured logging attaches key-value pairs to every log line:

```json
{"time":"2026-07-11T10:01:23Z","level":"ERROR","msg":"health check failed","host":"web-01","region":"us-east-1","service":"health-checker"}
```

Now a machine can:
- Filter: `level=ERROR`
- Group: `by host`
- Count: `where msg="health check failed" last 10m`
- Alert: `count > 5`

This is what Loki, Datadog, CloudWatch Logs Insights, and Splunk are built for.

---

## Concept 3: `log/slog` — Built Into Go 1.21+

No third-party package needed. Three key pieces:

**Log levels:** `Debug`, `Info`, `Warn`, `Error`

**Handlers — control output format:**
- `slog.NewTextHandler` → human-readable `key=value` (local dev)
- `slog.NewJSONHandler` → JSON objects (production)

**Attributes — key-value pairs after the message:**
```go
logger.Info("health check failed", "host", "web-01", "region", "us-east-1")
//                                  key      value     key        value
```

---

## Concept 4: Creating a Logger

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
```

- `slog.NewJSONHandler(os.Stdout, nil)` — writes JSON to stdout, default options
- `slog.New(handler)` — creates the logger

---

## Concept 5: Default Fields with `.With()`

```go
logger = logger.With("service", "health-checker", "version", "1.0")
```

Creates a new logger that automatically appends `service` and `version` to **every** log line. You write it once — it appears everywhere. No need to repeat common fields on every call.

---

## Concept 6: Text vs JSON Handler

| Handler | Output | Use case |
|---------|--------|----------|
| `NewTextHandler` | `time=... level=INFO msg="check passed" host=web-01` | Local development — human readable |
| `NewJSONHandler` | `{"time":"...","level":"INFO","msg":"check passed","host":"web-01"}` | Production — machine queryable |

**Rule:** JSON in production. Text locally. Some teams switch via env var:

```go
if os.Getenv("ENV") == "production" {
    logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
} else {
    logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}
```

---

## Mistakes Made Today

### Mistake 1 — Missing key before value in `logger.Info`

```go
// ❌ Wrong — slog doesn't know what "web-01" means without a key
logger.Info("starting health check", host)
```

```go
// ✅ Correct — key first, then value, alternating
logger.Info("starting health check", "host", host, "region", region)
```

`slog` expects pairs: `key, value, key, value`. A bare value with no key produces unexpected output.

---

### Mistake 2 — Logging the error but not returning it

```go
// ❌ Wrong — error is logged but function returns nil (caller thinks it passed)
if host == "db-01" {
    err := fmt.Errorf("connection refused")
    logger.Error("check failed", "host", host, "err", err)
}
return nil
```

```go
// ✅ Correct — return the error so the caller knows the check failed
if host == "db-01" {
    err := fmt.Errorf("connection refused")
    logger.Error("check failed", "host", host, "err", err)
    return err
}
```

Logging an error and swallowing it is a common SRE bug — your logs say something failed but your code acts like everything is fine.

---

### Mistake 3 — `else` after early return (un-idiomatic)

```go
// ❌ Un-idiomatic — else is unnecessary after return
if host == "db-01" {
    return err
} else {
    logger.Info("check passed", "host", host)
}
return nil
```

```go
// ✅ Idiomatic — early return removes the need for else
if host == "db-01" {
    return err
}
logger.Info("check passed", "host", host)
return nil
```

When an `if` block always returns, the `else` is dead weight. Drop it.

---

## Final Code

```go
package main

import (
	"fmt"
	"log/slog"
	"os"
)

func checkServer(logger *slog.Logger, host, region string) error {
	logger.Info("starting health check ", "host", host, "region", region)
	if host == "db-01" {
		err := fmt.Errorf("connection refused")
		logger.Error("check failed", "host", host, "err", err)
		return err
	}
	logger.Info("check passed", "host", host)
	return nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger = logger.With("service", "health-checker", "version", "1.0")
	checkServer(logger, "web-01", "us-east1")
	checkServer(logger, "web-02", "eu-west1")
	checkServer(logger, "db-01", "us-east-1")
	checkServer(logger, "web-03", "us-east-4")
}
```

Output (`go run main.go`):
```json
{"time":"2026-07-11T...","level":"INFO","msg":"starting health check ","service":"health-checker","version":"1.0","host":"web-01","region":"us-east1"}
{"time":"2026-07-11T...","level":"INFO","msg":"check passed","service":"health-checker","version":"1.0","host":"web-01"}
{"time":"2026-07-11T...","level":"INFO","msg":"starting health check ","service":"health-checker","version":"1.0","host":"web-02","region":"eu-west1"}
{"time":"2026-07-11T...","level":"INFO","msg":"check passed","service":"health-checker","version":"1.0","host":"web-02"}
{"time":"2026-07-11T...","level":"INFO","msg":"starting health check ","service":"health-checker","version":"1.0","host":"db-01","region":"us-east-1"}
{"time":"2026-07-11T...","level":"ERROR","msg":"check failed","service":"health-checker","version":"1.0","host":"db-01","err":"connection refused"}
{"time":"2026-07-11T...","level":"INFO","msg":"starting health check ","service":"health-checker","version":"1.0","host":"web-03","region":"us-east-4"}
{"time":"2026-07-11T...","level":"INFO","msg":"check passed","service":"health-checker","version":"1.0","host":"web-03"}
```

Notice `service` and `version` appear on every line — set once with `.With()`, applied everywhere.

---

## System Design: Distributed Tracing — Spans, Traces, Context

### The Problem Structured Logging Doesn't Solve

Structured logging tells you what happened on **one service**. But in a microservices system, a single user request touches multiple services:

```
User request
  → API Gateway
    → Auth Service
      → User Service
        → Database
```

If the request is slow, which service is responsible? Logs from each service are separate — you can't connect them into one picture.

---

### What Is Distributed Tracing?

Distributed tracing tracks a single request as it flows through multiple services. It gives you a **trace** — the full journey of one request end-to-end.

**Key terms:**

**Trace** — the entire journey of one request across all services. Has a unique `trace_id`.

**Span** — one unit of work within a trace. Each service call creates a span. A trace is a tree of spans.

```
Trace: abc-123
├── Span: API Gateway        (0ms → 150ms)
│   ├── Span: Auth Service   (5ms → 30ms)
│   └── Span: User Service   (35ms → 140ms)
│       └── Span: Database   (40ms → 135ms)  ← slow!
```

**Context propagation** — the `trace_id` and `span_id` are passed in HTTP headers between services so they can be linked:

```
X-Trace-Id: abc-123
X-Span-Id: def-456
```

---

### Trace vs Log

| | Structured Log | Distributed Trace |
|--|---------------|-------------------|
| Scope | One event on one service | One request across all services |
| Key identifier | `host`, `level`, `msg` | `trace_id`, `span_id` |
| Question answered | "What happened on this pod?" | "Where did this request spend its time?" |
| Tool | Loki, CloudWatch | Jaeger, Zipkin, Tempo, Datadog APM |

They complement each other — attach the `trace_id` to your logs and you can jump from a log line directly to the trace.

---

### The OpenTelemetry Standard

OpenTelemetry (OTel) is the industry standard for distributed tracing (and metrics and logs). One SDK, works with any backend — Jaeger, Tempo, Datadog, Honeycomb.

In Go:
```go
import "go.opentelemetry.io/otel"

tracer := otel.Tracer("health-checker")
ctx, span := tracer.Start(ctx, "checkServer")
defer span.End()
```

- `tracer.Start` creates a span
- `defer span.End()` records when the work finished
- `ctx` carries the trace context — pass it through every function call

---

### SRE Relevance

Without tracing, debugging a slow request in microservices means:
- Grep logs across 10 services
- Guess which one was slow
- No visibility into call order or timing

With tracing:
- Open Jaeger/Tempo, search by `trace_id`
- See the full waterfall — which span took 95% of the time
- Click into the span, see the logs attached to it

**Stack in practice:**
- **Jaeger / Grafana Tempo** — open source trace backends
- **OpenTelemetry** — the collection standard
- **Loki** — logs (attach `trace_id` to correlate)
- **Grafana** — unified view of traces + logs + metrics

---

## Key Takeaways

1. `log.Printf` produces unstructured strings — machines can't query them
2. Structured logging attaches key-value pairs — machines can filter, group, and alert on them
3. `log/slog` is built into Go 1.21+ — no third-party package needed
4. `slog.NewJSONHandler` → JSON for production; `slog.NewTextHandler` → human-readable for local dev
5. `logger.With(key, value)` — bakes default fields into every log line from that logger
6. Key-value pairs alternate after the message: `logger.Info("msg", "key1", val1, "key2", val2)`
7. Log an error AND return it — logging without returning hides failures from callers
8. Drop `else` after an early return — idiomatic Go
9. Distributed tracing tracks one request across multiple services — structured logging tracks one event on one service
10. A **trace** is the full journey; a **span** is one unit of work within that journey
11. `trace_id` propagates via HTTP headers between services — links spans into one trace
12. OpenTelemetry is the industry standard SDK — works with Jaeger, Tempo, Datadog, Honeycomb
13. Attach `trace_id` to your structured logs — jump from a log line directly to the full trace

---

> **Al-Basir — البَصِير — The All-Seeing**
>
> _He sees what no eye can reach — the hidden and the apparent. Distributed tracing gives your system that same sight: nothing that passes through is unseen. See you on Day 20._

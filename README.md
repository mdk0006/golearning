# Go Learning Journey

3-month plan: Golang (beginner → advanced) + System Design + AIOps  
Background: SRE with expertise in Kubernetes, Terraform, AWS, GCP  
Started: April 2026

---

## Progress

| Day | Topic | Concepts | Status |
|-----|-------|----------|--------|
| [Day 01](day01/README.md) | Variables, Types & Zero Values | `var`, `:=`, zero values, `fmt.Printf`, format verbs | ✅ |
| [Day 02](day02/README.md) | Functions | Multiple return values, early return, `fmt.Sprintf`, `_` discard | ✅ |
| Day 03 | Control Flow | `if/else`, `for`, `switch`, no `while` in Go | |
| Day 04 | Structs | Defining structs, fields, methods on structs | |
| Day 05 | Pointers | `*` and `&`, pass by value vs reference, when to use pointers | |
| Day 06 | Arrays & Slices | Fixed arrays, dynamic slices, `append`, `len`, `cap` | |
| Day 07 | Maps | Key-value store, iteration, checking key existence | |
| Day 08 | Error Handling | `error` type, `errors.New`, `fmt.Errorf`, `if err != nil` pattern | |
| Day 09 | Interfaces | Defining interfaces, implicit implementation, `any` type | |
| Day 10 | Goroutines & Channels | `go` keyword, channels, basic concurrency | |
| Day 11 | Packages & Modules | Creating packages, `go.mod`, imports, exported vs unexported | |
| Day 12 | File I/O | Reading/writing files, `os`, `bufio`, config file parsing | |
| Day 13 | HTTP Client | `net/http`, GET/POST requests, reading responses, timeouts | |
| Day 14 | HTTP Server | Building an HTTP server, handlers, routing, health check endpoint | |
| Day 15 | JSON | `encoding/json`, marshal/unmarshal, struct tags | |
| Day 16 | Closures & Variadic Functions | Functions as values, closures, `...args` pattern | |
| Day 17 | Defer, Panic & Recover | `defer` for cleanup, `panic`, `recover`, real-world use cases | |
| Day 18 | Testing in Go | `testing` package, `go test`, writing table-driven tests | |
| Day 19 | Logging | `log` package, structured logging with `slog`, log levels | |
| Day 20 | CLI Tool | `os.Args`, `flag` package, building a real CLI tool | |

---

## Month 1 — Go Foundations (Days 01–10)

Core language features. Every day uses SRE-relevant examples: health checks, servers, alerts, metrics.

- [x] Day 01 — Variables, Types, Zero Values
- [x] Day 02 — Functions, Multiple Return Values
- [ ] Day 03 — Control Flow: `if`, `for`, `switch`
- [ ] Day 04 — Structs: modeling a Server, Pod, Alert
- [ ] Day 05 — Pointers: value vs reference, mutation
- [ ] Day 06 — Slices: list of servers, filtering unhealthy nodes
- [ ] Day 07 — Maps: label maps, config key-value pairs
- [ ] Day 08 — Error Handling: the `error` type, wrapping errors
- [ ] Day 09 — Interfaces: abstraction, mock-able health checkers
- [ ] Day 10 — Goroutines & Channels: concurrent health checks

---

## Month 2 — Intermediate Go + System Design (Days 11–20)

Building real tools. HTTP servers, file I/O, JSON, testing, CLI tools.

- [ ] Day 11 — Packages & Modules
- [ ] Day 12 — File I/O: read configs, write logs
- [ ] Day 13 — HTTP Client: call an API, handle timeouts
- [ ] Day 14 — HTTP Server: build a `/healthz` endpoint
- [ ] Day 15 — JSON: parse Kubernetes-style payloads
- [ ] Day 16 — Closures & Variadic Functions
- [ ] Day 17 — Defer, Panic & Recover
- [ ] Day 18 — Testing: unit tests, table-driven tests
- [ ] Day 19 — Structured Logging with `slog`
- [ ] Day 20 — CLI Tool: build a server health checker CLI

---

## Month 3 — Advanced Go + AIOps (Days 21–30)

Concurrency patterns, Kubernetes operators, AI integrations, production-grade code.

- [ ] Day 21 — Advanced Concurrency: `sync.WaitGroup`, `sync.Mutex`
- [ ] Day 22 — Context: `context.WithTimeout`, cancellation, propagation
- [ ] Day 23 — Channels Deep Dive: select, fan-out, fan-in patterns
- [ ] Day 24 — Kubernetes Client-Go: list pods, watch events
- [ ] Day 25 — Writing a Kubernetes Controller (basic)
- [ ] Day 26 — Prometheus Metrics: expose `/metrics` from Go
- [ ] Day 27 — gRPC: define a proto, build client/server
- [ ] Day 28 — AIOps: call Claude/OpenAI API from Go, parse responses
- [ ] Day 29 — AIOps: build an alert summarizer using LLM
- [ ] Day 30 — Capstone: production-ready SRE tool in Go

---

## Running Code

```bash
cd dayXX
go run main.go
```

Auto-format before committing:

```bash
gofmt -w main.go
```

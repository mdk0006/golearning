# Go Learning Journey

3-month plan: Golang (beginner → advanced) + System Design + AIOps  
Background: SRE with expertise in Kubernetes, Terraform, AWS, GCP  
Started: April 2026

**Each day has 3 parts:**
- Go coding session
- System design concept
- Article / blog to read

---

## Month 1 — Go Foundations + System Design Fundamentals

| Day | Go Topic | System Design | Article to Read | Status |
|-----|----------|---------------|-----------------|--------|
| [Day 01](day01/README.md) | Variables, Types, Zero Values | What is Scalability? | [A Word on Scalability — All Things Distributed](https://www.allthingsdistributed.com/2006/03/a_word_on_scalability.html) | ✅ |
| [Day 02](day02/README.md) | Functions, Multiple Return Values | Availability vs Reliability | [Errors are values — go.dev/blog](https://go.dev/blog/errors-are-values) | ✅ |
| [Day 03](day03/README.md) | Control Flow: `if`, `for`, `switch` | Load Balancers — how traffic is distributed | [Go Tour: Flow Control](https://go.dev/tour/flowcontrol/1) | ✅ |
| Day 04 | Structs | DNS & How the internet resolves names | "The Go Blog: JSON and Go" — go.dev/blog |  |
| Day 05 | Pointers | CDN — Content Delivery Networks | "Pointers vs Values in Go" — go.dev/doc/faq |  |
| Day 06 | Slices | Caching — Redis, in-memory, eviction policies | "Go Slices: usage and internals" — go.dev/blog |  |
| Day 07 | Maps | Databases — SQL vs NoSQL, when to use which | "Go Maps in action" — go.dev/blog |  |
| Day 08 | Error Handling | CAP Theorem — Consistency, Availability, Partition | "Error handling and Go" — go.dev/blog |  |
| Day 09 | Interfaces | Message Queues — Kafka, SQS, async communication | "The Go Blog: Laws of Reflection" — go.dev/blog |  |
| Day 10 | Goroutines & Channels | Rate Limiting — token bucket, leaky bucket | "Concurrency is not parallelism" — go.dev/blog |  |

---

## Month 2 — Intermediate Go + Distributed Systems

| Day | Go Topic | System Design | Article to Read | Status |
|-----|----------|---------------|-----------------|--------|
| Day 11 | Packages & Modules | API Gateway — single entry point pattern | "Organizing a Go module" — go.dev/doc |  |
| Day 12 | File I/O | Monolith vs Microservices — trade-offs | "Microservices" — Martin Fowler |  |
| Day 13 | HTTP Client | Service Discovery — how services find each other | "HTTP/2 in Go" — go.dev/blog |  |
| Day 14 | HTTP Server | Load Balancer Deep Dive — L4 vs L7 | "Writing Web Apps in Go" — go.dev/doc |  |
| Day 15 | JSON | Consistent Hashing — distributed data routing | "JSON and Go" — go.dev/blog |  |
| Day 16 | Closures & Variadic Functions | Replication — primary/replica, sync vs async | "Functional options in Go" — Dave Cheney |  |
| Day 17 | Defer, Panic & Recover | Sharding — horizontal partitioning strategies | "Defer, Panic and Recover" — go.dev/blog |  |
| Day 18 | Testing | Circuit Breaker — fail fast, prevent cascade | "The Go Blog: Table-driven tests" — go.dev/blog |  |
| Day 19 | Structured Logging | Distributed Tracing — spans, traces, context | "Structured logging in Go" — go.dev/blog |  |
| Day 20 | CLI Tool | Design a URL Shortener (end-to-end) | "How I built a CLI in Go" — Carolyn Van Slyck |  |

---

## Month 3 — Advanced Go + AIOps + Real System Designs

| Day | Go Topic | System Design | Article to Read | Status |
|-----|----------|---------------|-----------------|--------|
| Day 21 | `sync.WaitGroup`, `sync.Mutex` | Leader Election — Raft, ZooKeeper, etcd | "Share memory by communicating" — go.dev/blog |  |
| Day 22 | Context — timeout, cancellation | Event Sourcing & CQRS | "Go Concurrency Patterns: Context" — go.dev/blog |  |
| Day 23 | Channels — select, fan-out, fan-in | Design a Notification System | "Go Concurrency Patterns: Pipelines" — go.dev/blog |  |
| Day 24 | Kubernetes client-go | Design a Monitoring System | "Kubernetes client-go overview" — kubernetes.io/docs |  |
| Day 25 | Kubernetes Controller | Design a CI/CD Pipeline | "Writing a Kubernetes Controller" — kubebuilder.io |  |
| Day 26 | Prometheus Metrics from Go | Design a Logging System (ELK/Loki) | "Instrumenting a Go app" — prometheus.io/docs |  |
| Day 27 | gRPC — proto, client, server | Design a Distributed Job Scheduler | "gRPC basics in Go" — grpc.io/docs |  |
| Day 28 | Call LLM API from Go | AIOps — AI for incident detection | "Anthropic API docs" — docs.anthropic.com |  |
| Day 29 | Build alert summarizer with LLM | AIOps — automated root cause analysis | "How Slack uses ML for alert fatigue" — slack.engineering |  |
| Day 30 | Capstone: production SRE tool | Full system design review + trade-offs | "Production Go" — Peter Bourgon |  |

---

## Checklist

### Month 1
- [x] Day 01 — Variables, Types, Zero Values
- [x] Day 02 — Functions, Multiple Return Values
- [x] Day 03 — Control Flow
- [ ] Day 04 — Structs
- [ ] Day 05 — Pointers
- [ ] Day 06 — Slices
- [ ] Day 07 — Maps
- [ ] Day 08 — Error Handling
- [ ] Day 09 — Interfaces
- [ ] Day 10 — Goroutines & Channels

### Month 2
- [ ] Day 11 — Packages & Modules
- [ ] Day 12 — File I/O
- [ ] Day 13 — HTTP Client
- [ ] Day 14 — HTTP Server
- [ ] Day 15 — JSON
- [ ] Day 16 — Closures & Variadic Functions
- [ ] Day 17 — Defer, Panic & Recover
- [ ] Day 18 — Testing
- [ ] Day 19 — Structured Logging
- [ ] Day 20 — CLI Tool

### Month 3
- [ ] Day 21 — Advanced Concurrency
- [ ] Day 22 — Context
- [ ] Day 23 — Channels Deep Dive
- [ ] Day 24 — Kubernetes client-go
- [ ] Day 25 — Kubernetes Controller
- [ ] Day 26 — Prometheus Metrics
- [ ] Day 27 — gRPC
- [ ] Day 28 — AIOps: Call LLM API
- [ ] Day 29 — AIOps: Alert Summarizer
- [ ] Day 30 — Capstone

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

# Day 09 — Interfaces in Go

---

> **بِسْمِ اللهِ الرَّحْمَنِ الرَّحِيم**
>
> **Al-Adl — العَدْل — The Just**
>
> _He judges by what you do, not by what you are. Interfaces work the same way — Go does not ask what type you are, only what you can do. Begin with His name._

---

## Blog of the Day

[The Go Blog: Laws of Reflection](https://go.dev/blog/laws-of-reflection)

Read this after the session. Reflection in Go is built on top of interfaces — every `interface{}` value carries a type and a value inside it. This article will make the `(type, value)` pair click permanently.

---

## Concept: What Is an Interface?

An interface defines a **contract** — a set of method signatures a type must have. It describes behaviour, not data.

```go
type HealthChecker interface {
    CheckHealth() string
}
```

This says: "any type that has a `CheckHealth() string` method is a `HealthChecker`."

No `implements` keyword. No registration. If the method exists on the type, the interface is satisfied — **automatically and implicitly** at compile time. This is called **structural typing**.

---

## Why Interfaces Matter

Without an interface, a function is locked to one concrete type:

```go
func runChecks(s *WebServer)  // only works for WebServers
```

With an interface, it works for any type that satisfies the contract:

```go
func runChecks(targets []HealthChecker)  // works for WebServer, Database, KubernetesNode — anything
```

You write the logic once. New types plug in without changing the function. This is the open/closed principle in practice: open for extension, closed for modification.

---

## Implicit Satisfaction — The Key Insight

```go
type WebServer struct { Name string; URL string }

func (w WebServer) CheckHealth() string {
    return fmt.Sprintf("Webserver %s: checking endpoint %s", w.Name, w.URL)
}
```

`WebServer` never mentions `HealthChecker`. It just has the method. Go sees that the method signature matches the interface, and `WebServer` automatically satisfies `HealthChecker`. The compiler verifies this — if the method is missing or has the wrong signature, it's a compile error, not a runtime panic.

---

## The Interface Value — Two Fields Under the Hood

An interface variable holds two things internally:

```
interface value = (concrete type, pointer to data)

Example:
  var h HealthChecker = WebServer{Name: "web-01", URL: "http://10.0.0.1"}

  h = (WebServer, → {Name:"web-01", URL:"http://10.0.0.1"})
```

When you call `h.CheckHealth()`, Go uses the `type` field to look up the right method and the `value` field to pass the receiver. This is **dynamic dispatch** — the method called depends on the concrete type stored, resolved at runtime.

This is why a `[]HealthChecker` can hold both `WebServer` and `Database` — the slice stores interface values, each wrapping a different concrete type.

---

## The Empty Interface

```go
interface{}   // or: any  (alias introduced in Go 1.18)
```

Has zero methods — every type satisfies it automatically. Used when the type is genuinely unknown at compile time (e.g. `fmt.Println` accepts `any`).

**Avoid it when you can.** If you use `any`, you lose all compile-time type safety. Prefer a specific interface with the methods you actually need.

---

## Interfaces vs Structs — When to Use Which

| Use a struct when... | Use an interface when... |
|----------------------|--------------------------|
| You're defining data and its methods | You want to accept multiple different types |
| Only one type will ever do this job | The caller shouldn't care about the concrete type |
| You control all the types | You want to swap implementations (e.g. for testing) |

---

## Mistakes Made Today

### Mistake 1 — `fmt.Printf` instead of `fmt.Sprintf`

```go
// ❌ Wrong — fmt.Printf returns (int, error), not string
func (w WebServer) CheckHealth() string {
    return fmt.Printf("Webserver %s: checking endpoint %s", w.Name, w.URL)
}
```

`fmt.Printf` prints to stdout and returns `(int, error)` — the number of bytes written and any error. The method signature promises a `string`, so the compiler rejects it.

```go
// ✅ Correct — fmt.Sprintf builds and returns a string
func (w WebServer) CheckHealth() string {
    return fmt.Sprintf("Webserver %s: checking endpoint %s", w.Name, w.URL)
}
```

| Function | Does what | Returns |
|----------|-----------|---------|
| `fmt.Printf` | prints to stdout | `(int, error)` |
| `fmt.Sprintf` | builds a string | `string` |
| `fmt.Println` | prints + newline | `(int, error)` |

---

### Mistake 2 — Comparing a bool to `true`

```go
// ❌ Redundant
if k.Healthy == true {
```

`k.Healthy` is already a `bool`. Comparing it to `true` adds no information.

```go
// ✅ Idiomatic
if k.Healthy {
```

For the negative: `if !k.Healthy` not `if k.Healthy == false`.

---

### Mistake 3 — Struct named `Kubernetes` instead of `KubernetesNode`

Minor but matters: type names should be descriptive and precise. `Kubernetes` is the system; `KubernetesNode` is the thing being modelled. In real SRE tooling, vague names cause confusion when you later add `KubernetesCluster`, `KubernetesPod`, etc.

---

## Final Code

```go
package main

import "fmt"

type HealthChecker interface {
	CheckHealth() string
}

type WebServer struct {
	Name string
	URL  string
}

type Database struct {
	Name string
	Port int
}

type KubernetesNode struct {
	Name    string
	Healthy bool
}

func (w WebServer) CheckHealth() string {
	return fmt.Sprintf("Webserver %v: checking endpoint %v", w.Name, w.URL)
}

func (d Database) CheckHealth() string {
	return fmt.Sprintf("Database %v: checking port %v", d.Name, d.Port)
}

func (k KubernetesNode) CheckHealth() string {
	if k.Healthy {
		return fmt.Sprintf("KubernetesNode %v: node is ready", k.Name)
	}
	return fmt.Sprintf("KubernetesNode %v: node is not ready", k.Name)
}

func runChecks(targets []HealthChecker) {
	for _, t := range targets {
		fmt.Println(t.CheckHealth())
	}
}

func main() {
	targets := []HealthChecker{
		WebServer{Name: "web-01", URL: "http://10.0.0.1"},
		Database{Name: "postgres-01", Port: 5432},
		KubernetesNode{Name: "node-01", Healthy: true},
		KubernetesNode{Name: "node-02", Healthy: false},
	}
	runChecks(targets)
}
```

Output:
```
Webserver web-01: checking endpoint http://10.0.0.1
Database postgres-01: checking port 5432
KubernetesNode node-01: node is ready
KubernetesNode node-02: node is not ready
```

---

## System Design: Message Queues — Kafka, SQS, Async Communication

So far every system design topic has been about **synchronous** communication — a client calls a service and waits for the response. Message queues introduce **asynchronous** communication.

---

### The Problem Synchronous Communication Solves — And Doesn't

Synchronous:
```
Alert fired → AlertManager → calls PagerDuty API → waits → PagerDuty responds → done
```

This works fine until:
- PagerDuty is slow — your AlertManager blocks waiting
- PagerDuty is down — your alert is lost
- 10,000 alerts fire at once — you hammer PagerDuty with 10,000 simultaneous requests

---

### What a Message Queue Does

A message queue sits between producer and consumer:

```
Alert fired → AlertManager → [Queue] → PagerDuty worker reads at its own pace
```

- **Producer** — writes a message to the queue and moves on immediately (no waiting)
- **Queue** — holds messages durably until a consumer reads them
- **Consumer** — reads messages at its own pace, processes them, acknowledges

The producer and consumer are **decoupled** — they don't need to be running at the same time or at the same speed.

---

### Kafka

Apache Kafka is a **distributed log**. Messages are appended to a topic (a named log). Consumers read from any position in the log. Messages are retained for a configurable period (e.g. 7 days) — not deleted after being read.

```
Topic: "alerts"
  offset 0: {alert: "HighCPU", host: "web-01", time: 09:00}
  offset 1: {alert: "DiskFull", host: "db-01",  time: 09:01}
  offset 2: {alert: "HighCPU", host: "web-02", time: 09:02}
             ▲
             consumer reads here, tracks its own offset
```

**Key properties:**
- **High throughput** — millions of messages/second
- **Durable** — messages persisted to disk, replicated across brokers
- **Multiple consumers** — different services can read the same topic independently (each tracks its own offset)
- **Replay** — a consumer can re-read old messages (useful for debugging or reprocessing)
- **Ordered within a partition** — messages in a partition are strictly ordered

**SRE use cases:** audit logs, metrics pipelines (Prometheus → Kafka → long-term storage), alert event streams, change events for distributed systems

---

### SQS (Simple Queue Service)

AWS SQS is a **managed queue**. A producer sends a message; a consumer receives and deletes it. Once acknowledged, the message is gone.

```
Producer → SQS Queue → Consumer reads → Consumer deletes message
```

**Key properties:**
- **At-least-once delivery** — a message may be delivered more than once (design consumers to be idempotent)
- **Visibility timeout** — when a consumer reads a message, it becomes invisible to others for a period; if not deleted in time, it reappears (retry)
- **Dead letter queue (DLQ)** — messages that fail processing N times get moved to a DLQ for investigation
- **No replay** — once deleted, the message is gone
- **Fully managed** — no brokers to operate

**SRE use cases:** job queues, async task processing, decoupling microservices, triggering Lambda functions

---

### Kafka vs SQS

| Property | Kafka | SQS |
|----------|-------|-----|
| Retention | Days/weeks — messages stay after read | Deleted after consumer acknowledges |
| Replay | Yes — re-read old messages | No |
| Throughput | Very high — millions/sec | High — thousands/sec |
| Multiple consumers | Yes — each tracks own offset | No — one consumer per message |
| Ordering | Guaranteed within partition | Best-effort (FIFO queue available) |
| Ops burden | High — you manage brokers | None — fully managed |
| Best for | Event streaming, audit logs, pipelines | Job queues, task dispatch, microservice decoupling |

---

### Async Patterns in SRE

**Fan-out:** one event triggers multiple independent consumers

```
Alert fired → Kafka topic "alerts"
    → Consumer A: page on-call via PagerDuty
    → Consumer B: write to incident database
    → Consumer C: send Slack notification
```

All three run independently. If Slack is down, PagerDuty still gets paged.

**Buffering bursts:** 10,000 alerts fire simultaneously

```
Without queue: AlertManager hits PagerDuty with 10,000 requests → PagerDuty rate-limits → alerts lost
With queue:    AlertManager writes 10,000 messages → worker reads at 100/sec → PagerDuty never overwhelmed
```

**Retry with DLQ:**

```
Message → Consumer fails processing → message reappears (visibility timeout)
→ fails 3 more times → moves to Dead Letter Queue
→ SRE investigates DLQ — finds malformed message, fixes producer
```

---

## Key Takeaways

1. An interface is a contract — a set of method signatures a type must implement
2. Satisfaction is implicit — no `implements` keyword, Go checks at compile time
3. An interface value holds `(concrete type, pointer to data)` — dynamic dispatch resolves the method at runtime
4. A `[]HealthChecker` can hold any type that satisfies the interface — the slice stores interface values
5. `fmt.Sprintf` returns a `string`; `fmt.Printf` returns `(int, error)` and prints to stdout
6. `if k.Healthy` is idiomatic for bool checks — never `== true`
7. The empty interface (`any`) accepts everything but loses type safety — use specific interfaces when possible
8. Message queues decouple producers from consumers — producer doesn't wait, consumer processes at its own pace
9. Kafka is a durable distributed log — messages persist, multiple consumers, replayable
10. SQS is a managed job queue — messages deleted after ack, simpler ops, no replay
11. Use queues to buffer bursts, enable fan-out, and add retry/DLQ for reliability

---

> **Al-Wakeel — الوَكِيل — The Trustee, The Disposer of Affairs**
>
> _He handles what you cannot — the things you dispatch and release. A message queue works the same way: hand off the work, move on, trust it will be handled. You learned to write code that does the same. See you on Day 10._

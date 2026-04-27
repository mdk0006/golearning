# Day 08 — Error Handling in Go

---

> **بِسْمِ اللهِ الرَّحْمَنِ الرَّحِيم**
>
> **Al-Khabeer — الخَبِير — The All-Aware**
>
> _He is aware of every failure, every fault, every hidden flaw. A good engineer is also aware — they do not ignore errors, they face them. Begin with His name and build systems that tell the truth._

---

## Blog of the Day

[Error Handling and Go — The Go Blog](https://go.dev/blog/error-handling-and-go)

Read this after the session. It explains why Go chose explicit error values over exceptions, and shows the idiomatic patterns used in the standard library.

---

## Concept: Errors Are Values

Most languages use exceptions — something fails, an exception is thrown, it bubbles up the call stack until something catches it. If nobody catches it, the program crashes.

**Go has no exceptions.** Errors are ordinary values returned from functions. The caller receives the error and decides what to do. Nothing is thrown, nothing bubbles silently.

```go
func checkHealth(s Server) error
```

A function that can fail returns `error` as its last return value. `nil` means success. A non-nil value means something went wrong.

---

## The `error` Interface

`error` is a built-in interface:

```go
type error interface {
    Error() string
}
```

Any type that has an `Error() string` method satisfies this interface. `nil` means the interface holds nothing — no error occurred.

This is why both `errors.New` and `fmt.Errorf` work wherever an `error` is expected — they both return types that implement `Error() string`.

---

## Creating Errors

### `errors.New` — plain string, no formatting

```go
import "errors"

return errors.New("host is not configured")
```

Use when the message is fixed and has no variable data.

### `fmt.Errorf` — formatted string with context

```go
return fmt.Errorf("server %s: CPU usage %.1f exceeds threshold", s.Name, s.CPUUsage)
```

Use when you need to embed values into the message. The `%s` verb is for strings, `%.1f` formats a float to one decimal place.

**Convention — error strings must be:**
- Lowercase (`"host is not configured"`, not `"Host is not configured"`)
- No trailing punctuation (no `.` or `!`)

Why: errors are often wrapped into larger messages — `"health check failed: host is not configured"`. A capital letter or period in the middle of a sentence looks wrong.

---

## The Pattern — Always Check Errors

```go
err := checkHealth(server)
if err != nil {
    fmt.Println("ALERT:", err)
} else {
    fmt.Println("OK:", server.Name, "is healthy")
}
```

Check immediately after the call. Handle it, log it, or return it — but never ignore it.

**Never do this:**

```go
result, _ := checkHealth(server)   // silently swallowed
```

Silently ignoring errors is how production incidents happen. In SRE tooling where errors represent real infrastructure failures, a swallowed error means a dead server that looks healthy.

---

## Wrapping Errors — `%w`

When you call another function and want to add context before returning the error upward:

```go
err := checkHealth(s)
if err != nil {
    return fmt.Errorf("runbook check for %s failed: %w", s.Name, err)
}
```

`%w` **wraps** the original error inside the new one. The original is preserved and can be inspected later with `errors.Is()` or `errors.As()`. Use `%w` when wrapping, `%v` or `%s` when just formatting into a string.

---

## What Is an Interface? (Preview of Day 09)

`error` is Go's most common interface. An interface defines **what a type can do**, not what it is:

```go
type error interface {
    Error() string   // any type with this method is an error
}
```

There's no `implements` keyword. If your type has the required method, it satisfies the interface automatically. This is called **implicit satisfaction** — one of the most powerful ideas in Go.

Full deep dive on interfaces is Day 09.

---

## Mistakes Made Today

### Mistake 1 — `errors.New` result discarded (not returned)

```go
// ❌ Wrong — creates an error value then throws it away
if s.Host == "" {
    errors.New("Host is not present")
}
return nil   // always reaches here, error is lost
```

`errors.New` creates and returns an error value — it does not automatically stop the function. You must `return` it.

```go
// ✅ Correct
if s.Host == "" {
    return errors.New("host is not configured")
}
```

---

### Mistake 2 — Using `errors.New` with format verbs

```go
// ❌ Wrong — errors.New takes a plain string, not a format string
errors.New("server %v: CPU usage %v exceeds threshold", s.Name, s.CPUUsage)
```

`errors.New` has signature `func New(text string) error` — one argument only. `%v` is treated as a literal character, not a format verb.

```go
// ✅ Correct — use fmt.Errorf for formatted messages
return fmt.Errorf("server %s: CPU usage %.1f exceeds threshold", s.Name, s.CPUUsage)
```

---

### Mistake 3 — Uppercase local variable names

```go
// ❌ Wrong
HealthServer1 := checkHealth(server1)
```

Local variables are always camelCase in Go. Capital letters signal exported (public) identifiers.

```go
// ✅ Correct
err1 := checkHealth(server1)
```

---

### Mistake 4 — Uppercase and missing context in error message

```go
// ❌ Wrong — capital letter, no server name
return fmt.Errorf("Host is not present")
```

```go
// ✅ Correct — lowercase, includes server name for context
return fmt.Errorf("server %s: host is not configured", s.Name)
```

---

### Mistake 5 — `%v` for a string argument

```go
// ❌ imprecise — %v works but is the catch-all verb
return fmt.Errorf("server %v: CPU usage %.1f exceeds threshold", s.Name, s.CPUUsage)
```

```go
// ✅ precise — %s for strings
return fmt.Errorf("server %s: CPU usage %.1f exceeds threshold", s.Name, s.CPUUsage)
```

`%v` is the default format for any type. For strings, `%s` is the correct and explicit verb. Use `%v` for structs or when you don't know the type; use `%s`, `%d`, `%f` when you do.

---

## Final Code

```go
package main

import (
	"fmt"
)

type Server struct {
	Name     string
	Host     string
	CPUUsage float64
}

func checkHealth(s Server) error {
	if s.Host == "" {
		return fmt.Errorf("server %s: host is not configured", s.Name)
	}
	if s.CPUUsage > 90.0 {
		return fmt.Errorf("server %s: CPU usage %.1f exceeds threshold", s.Name, s.CPUUsage)
	}
	return nil
}

func main() {
	server1 := Server{Name: "web-01", Host: "10.0.0.1", CPUUsage: 45.0}
	server2 := Server{Name: "web-02", Host: "", CPUUsage: 30.0}
	server3 := Server{Name: "web-03", Host: "10.0.0.3", CPUUsage: 95.5}
	err1 := checkHealth(server1)
	err2 := checkHealth(server2)
	err3 := checkHealth(server3)
	if err1 != nil {
		fmt.Println("ALERT:", err1)
	} else {
		fmt.Printf("OK: %v is healthy \n", server1.Name)
	}
	if err2 != nil {
		fmt.Println("ALERT:", err2)
	} else {
		fmt.Printf("OK: %v is healthy \n", server2.Name)
	}
	if err3 != nil {
		fmt.Println("ALERT:", err3)
	} else {
		fmt.Printf("OK: %v is healthy \n", server3.Name)
	}
}
```

Output:
```
OK: web-01 is healthy
ALERT: server web-02: host is not configured
ALERT: server web-03: CPU usage 95.5 exceeds threshold
```

---

## System Design: CAP Theorem

Every distributed system that stores data makes a promise about three properties. The CAP theorem says **you can only fully guarantee two of the three at the same time**.

```
        Consistency
            /\
           /  \
          /    \
         /      \
Availability ── Partition Tolerance
```

---

### The Three Properties

**Consistency (C)**
Every read receives the most recent write — or an error. All nodes see the same data at the same time.

```
Node A writes: cpu=95%
Node B reads:  cpu=95%   ← consistent
```

If the system is consistent, you will never read stale data. If it can't guarantee that, it returns an error instead of a wrong answer.

**Availability (A)**
Every request receives a response — not an error. The system is always up and always answers.

Even if some nodes are down or out of sync, the system keeps serving requests. It may return stale data, but it never says "I can't answer right now."

**Partition Tolerance (P)**
The system continues operating even when network messages between nodes are dropped or delayed — a **network partition**.

```
Node A ──✗✗✗── Node B    ← network partition
```

In any real distributed system running over a network (cloud, multi-region, multi-AZ), partitions happen. They are not optional — hardware fails, cables get cut, AWS AZs go dark. **Partition tolerance is not negotiable in practice.** You must design for it.

---

### The Real Choice: CP vs AP

Since P is mandatory, the real trade-off is between **C and A during a partition**:

#### CP — Consistent + Partition Tolerant

During a partition, the system **refuses to answer** rather than risk returning stale data.

```
Partition occurs:
  Node B can't reach Node A
  Node B stops serving reads → returns error
  Stale data is never returned
```

**Examples:** HBase, Zookeeper, etcd, CockroachDB  
**SRE use case:** etcd (Kubernetes' backing store) is CP. If etcd loses quorum, the API server stops accepting writes. You can't create pods. The cluster is degraded but it will never tell you a pod exists when it doesn't — correctness over availability.

#### AP — Available + Partition Tolerant

During a partition, the system **keeps serving requests** with potentially stale data.

```
Partition occurs:
  Node B can't reach Node A
  Node B keeps serving from its local copy
  Clients get a response — possibly stale
```

**Examples:** Cassandra, DynamoDB (default), CouchDB, DNS  
**SRE use case:** Cassandra is AP. During a partition, replicas serve reads from local data. Two nodes might return different values for the same key. Eventual consistency — all nodes converge once the partition heals.

---

### Real-World Examples

| System | CAP choice | Why |
|--------|-----------|-----|
| etcd | CP | Kubernetes correctness — wrong state is worse than no state |
| Cassandra | AP | High availability matters more than perfect consistency for metrics/logs |
| DynamoDB (default) | AP | Eventually consistent reads for scale; strongly consistent reads available at cost |
| PostgreSQL (single node) | CA | No partition tolerance — one machine, no network split possible |
| DNS | AP | Always answers, may return stale records until TTL expires |
| Zookeeper | CP | Leader election — two nodes can't both think they're leader |

---

### PACELC — The Extension

CAP only talks about behaviour during partitions. **PACELC** extends it:

> Even when there is no partition (**E**lse), there is a trade-off between **L**atency and **C**onsistency.

A CP system that requires quorum writes (e.g. writing to 3 of 5 replicas before acknowledging) is slower than one that acknowledges after writing to 1. Stronger consistency = higher latency. This trade-off exists even in normal operation, not just during failures.

---

### For an SRE Alerting System

| Component | Choice | Reason |
|-----------|--------|--------|
| Alert rule definitions | **CP (PostgreSQL / etcd)** | Wrong rules = wrong pages. Correctness critical. |
| Active alert dedup state | **AP (Redis)** | Occasionally firing a duplicate alert is acceptable; downtime is not |
| Metrics time-series | **AP (Cassandra / Prometheus)** | A few seconds of stale metrics is fine; availability is not optional |
| On-call schedule | **CP (PostgreSQL)** | Two people can't both think they're primary on-call |

---

## Key Takeaways

1. Go has no exceptions — errors are values returned from functions
2. `nil` error = success; non-nil error = failure
3. Always check errors with `if err != nil` — never silently discard with `_`
4. `errors.New` for plain strings; `fmt.Errorf` for formatted messages with variable data
5. Error strings are lowercase with no trailing punctuation — they get embedded in larger messages
6. Always include context in errors: which server, which operation, what value
7. `%w` wraps an error (preserves it for `errors.Is`/`errors.As`); `%v` just formats it as a string
8. CAP theorem: Consistency, Availability, Partition Tolerance — only two fully guaranteed at once
9. Partition tolerance is mandatory in real distributed systems — the real choice is CP vs AP
10. CP systems (etcd, Zookeeper) refuse to answer during a partition — correctness over availability
11. AP systems (Cassandra, DNS) always answer during a partition — availability over correctness
12. PACELC extends CAP: even without a partition, stronger consistency means higher latency

---

> **As-Samee' — السَّمِيع — The All-Hearing**
>
> _He hears every error, every call for help, every moment you pushed through confusion. You learned today to never ignore a signal — in code and in life. Every `err != nil` is a moment of honesty. See you on Day 09._

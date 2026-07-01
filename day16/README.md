# Day 16 — Closures & Variadic Functions in Go

---

> **بِسْمِ اللهِ الرَّحْمَنِ الرَّحِيم**
>
> **Al-Hafiz — الحَفِيظ — The Preserver, The Guardian**
>
> _He preserves everything — no action forgotten, no count lost. A closure preserves its captured variable across every call, just as He preserves what matters across all time. Begin with His name._

---

## Blog of the Day

[Functional Options in Go — Dave Cheney](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis)

Read this after the session. Dave Cheney's classic post on using closures to build clean, extensible APIs — a pattern used heavily in production Go libraries including `net/http`, `grpc-go`, and Kubernetes client-go.

---

## Concept 1: Closures

A **closure** is a function bundled with the variables it captured from the surrounding scope — those variables stay alive as long as the closure does, and every call reads and modifies the same captured variables.

### Building up from scratch

**Normal function — variable dies when function returns:**

```go
func greet() {
    message := "hello"
    fmt.Println(message)
}
// message is gone when greet() returns
```

**Function returning a function — variable lives on:**

```go
func makeGreeter() func() {
    message := "hello"
    return func() {
        fmt.Println(message)  // captures message
    }
}

greet := makeGreeter()
greet()   // "hello" — works even though makeGreeter already returned
```

Go moved `message` from the stack to the heap because the inner function captured it. `message` now lives as long as `greet` lives.

---

### The captured variable is shared and mutable

```go
func makeCounter() func() int {
    count := 0
    return func() int {
        count++    // same count every call
        return count
    }
}

counter := makeCounter()
fmt.Println(counter())  // 1
fmt.Println(counter())  // 2
fmt.Println(counter())  // 3
```

Every call to `counter()` reads and modifies the **same** `count`. It's not reset between calls — the closure holds a reference, not a copy.

---

### Two closures — two independent captured variables

```go
counterA := makeCounter()
counterB := makeCounter()

fmt.Println(counterA())  // 1
fmt.Println(counterA())  // 2
fmt.Println(counterB())  // 1  ← fresh count, independent of counterA
fmt.Println(counterA())  // 3
```

Each call to `makeCounter()` creates a **new** `count`. `counterA` and `counterB` have completely separate state.

---

### Mental model — the backpack

```
counterA's backpack: { count: 2 }
counterB's backpack: { count: 1 }
```

Each closure carries its own private backpack. Callers can't reach in directly — they can only call the function, which reaches into its own backpack.

---

### SRE use case — rate limiter

```go
func makeRateLimiter(limit int) func() bool {
    count := 0
    return func() bool {
        count++
        return count <= limit
    }
}

dbLimiter := makeRateLimiter(5)   // 5 calls allowed
for i := 0; i < 7; i++ {
    fmt.Println("allowed:", dbLimiter())
}
// true true true true true false false
```

Private `count` per limiter, no shared state, no global variables, no locks needed.

---

## Concept 2: Variadic Functions

A **variadic function** accepts any number of arguments of the same type using `...`:

```go
func maxCPU(usages ...float64) float64 {
    max := usages[0]
    for _, usage := range usages {
        if usage > max {
            max = usage
        }
    }
    return max
}

maxCPU(45.2, 88.1, 23.5, 91.0, 67.3)  // any number of args
```

Inside the function, `usages` is a plain `[]float64` slice — you loop over it with `range`. The `...` only appears in the function signature.

`fmt.Println(a, b, c, ...)` is itself variadic — that's why it accepts any number of arguments.

---

### Passing a slice to a variadic function

To pass an existing slice, unpack it with `...`:

```go
nums := []float64{45.2, 88.1, 91.0}
maxCPU(nums...)   // unpacks the slice as individual arguments
```

Without `...`, you'd get a compile error — a `[]float64` is not a `float64`.

---

### Common pitfall — empty variadic

If you call a variadic function with no arguments, the slice inside is empty. Accessing `usages[0]` without checking `len(usages)` first will panic:

```go
// ❌ Panics if called with no arguments
func maxCPU(usages ...float64) float64 {
    max := usages[0]   // panic: index out of range
    ...
}

// ✅ Safe
func maxCPU(usages ...float64) float64 {
    if len(usages) == 0 {
        return 0
    }
    max := usages[0]
    ...
}
```

---

## Printing Functions vs Calling Functions

```go
dbLimiter := makeRateLimiter(5)

fmt.Println(dbLimiter)    // ❌ prints memory address: 0x10212cde0
fmt.Println(dbLimiter())  // ✅ calls the function, prints: true
```

`dbLimiter` is a variable holding a function value. Without `()` you're printing the function itself — Go shows it as a hex address. With `()` you execute it and get the result.

**Rule:** a function does nothing until you call it with `()`.

---

## Mistakes Made Today

### Mistake 1 — `service` variable captured but not used in return string

```go
// ❌ Wrong — "service:" is a literal string, not the variable
func makeAlertCounter(service string) func() string {
    calls := 0
    return func() string {
        calls++
        return fmt.Sprintf("service: alert #%d", calls)
    }
}
// output: "service: alert #1"  — same for every service
```

```go
// ✅ Correct — use %s to inject the captured service variable
return fmt.Sprintf("%s: alert #%d", service, calls)
// output: "web-01: alert #1", "db-01: alert #1" — each distinct
```

The closure captures both `calls` and `service`. `service` doesn't change (it's a parameter), but it's still captured and available inside the inner function.

---

### Mistake 2 — Printing function value instead of calling it

```go
// ❌ Wrong — dbLimiter is a function, not a bool
fmt.Println("dbLimter allowed", dbLimiter)
// output: dbLimter allowed 0x10212cde0
```

```go
// ✅ Correct — call it with ()
fmt.Println("dbLimter allowed", dbLimiter())
// output: dbLimter allowed true
```

---

### Mistake 3 — Commented out `maxCPU` before completing it

```go
// func maxCPU(usages ...float64) float64 {
// }
```

A commented-out function doesn't compile or run. Write the body first, then uncomment. Or write it directly without commenting — Go won't compile partial functions anyway, so there's no benefit to commenting them out mid-exercise.

---

## Final Code

```go
package main

import "fmt"

func makeAlertCounter(service string) func() string {
	calls := 0
	return func() string {
		calls++
		return fmt.Sprintf("%s: alert #%d", service, calls)
	}
}

func makeRateLimiter(limit int) func() bool {
	count := 0
	return func() bool {
		count++
		return count <= limit
	}
}

func maxCPU(usages ...float64) float64 {
	max := usages[0]
	for _, usage := range usages {
		if usage > max {
			max = usage
		}
	}
	return max
}

func main() {
	webAlert := makeAlertCounter("web-01")
	dbAlert := makeAlertCounter("db-01")
	fmt.Println(webAlert())
	fmt.Println(webAlert())
	fmt.Println(dbAlert())

	dbLimiter := makeRateLimiter(5)
	for i := 0; i < 7; i++ {
		fmt.Println("dbLimter allowed", dbLimiter())
	}

	fmt.Printf("%.1f\n", maxCPU(45.2, 88.1, 23.5, 91.0, 67.3))
}
```

Output:
```
web-01: alert #1
web-01: alert #2
db-01: alert #1
dbLimter allowed true
dbLimter allowed true
dbLimter allowed true
dbLimter allowed true
dbLimter allowed true
dbLimter allowed false
dbLimter allowed false
91.0
```

---

## System Design: Replication — Primary/Replica, Sync vs Async

Every production database uses replication — keeping copies of data on multiple nodes so the system survives node failures and can serve read traffic at scale.

---

### Why Replication?

```
Single node database:
  node-A goes down → all reads and writes fail → full outage
```

With replication:
```
  Primary (node-A) goes down → replica (node-B) promoted → writes resume
                              → reads continue from other replicas
```

Replication gives you **durability** (data survives node failure), **availability** (reads continue even during primary failure), and **read scalability** (spread reads across replicas).

---

### Primary / Replica Architecture

```
Writes → Primary
           ↓ replication
         Replica-1   ← Reads
         Replica-2   ← Reads
         Replica-3   ← Reads
```

All writes go to the **primary**. Replicas receive a copy of the write log and apply it. Reads can be served from any replica — the primary handles writes and may also serve reads.

---

### Synchronous Replication

The primary waits for at least one replica to confirm it has received and stored the write before acknowledging success to the client.

```
Client writes → Primary writes → sends to Replica → Replica confirms
                                                    ↓
                                Primary ACKs client ←
```

**Guarantee:** if the primary crashes right after acknowledging, the data is already on at least one replica — no data loss.

**Cost:** every write waits for network round-trip to the replica. Higher latency, lower throughput.

**Used when:** data loss is unacceptable — financial transactions, configuration changes, Kubernetes etcd writes.

---

### Asynchronous Replication

The primary acknowledges the client immediately after writing locally. Replication happens in the background.

```
Client writes → Primary writes → ACKs client immediately
                    ↓ (background)
               Replica applies write later
```

**Guarantee:** none. If the primary crashes between the ACK and the replica receiving the write, that write is lost permanently.

**Benefit:** primary never waits for replicas — lower latency, higher throughput.

**Used when:** some data loss is acceptable — analytics, metrics, read replicas for reporting, Prometheus remote write.

---

### Sync vs Async — The Trade-off

| Property | Synchronous | Asynchronous |
|----------|-------------|-------------|
| Data loss on primary failure | ❌ None | ✅ Possible (replication lag) |
| Write latency | Higher (waits for replica) | Lower (no waiting) |
| Write throughput | Lower | Higher |
| Replica lag | None — always in sync | Can be seconds to minutes behind |
| Use case | Financial, config, k8s etcd | Metrics, analytics, read replicas |

---

### Replication Lag

With async replication, replicas are always slightly behind the primary — this is called **replication lag**.

```
Client writes user record → Primary: user exists
                          → Read from replica 50ms later: user not found yet (lag)
```

This is **read-after-write inconsistency** — you write to the primary but immediately read from a replica that hasn't caught up yet. Common bugs:

- User signs up → immediately redirected to profile page → reads from replica → "user not found"
- Config updated → replica still serving old config for 2 seconds

**Fix:** for sensitive reads right after a write, route to the primary or wait for replica lag to clear.

---

### Failover

When the primary fails, a replica must be promoted:

**Manual failover:** ops team detects the failure, picks a replica, promotes it. Slow (minutes), but controlled.

**Automatic failover:** a system (Sentinel for Redis, Patroni for PostgreSQL, etcd for Kubernetes) detects failure and promotes a replica automatically. Fast (seconds), but can promote wrong replica if split-brain.

**Split-brain:** two nodes both think they are primary — both accept writes, data diverges, impossible to reconcile. Prevented with quorum: a node can only become primary if a majority of nodes agree.

---

### SRE Relevance

| System | Replication type |
|--------|-----------------|
| PostgreSQL (streaming replication) | Sync or async (configurable) |
| MySQL (binlog replication) | Async by default |
| Redis Sentinel | Async |
| etcd (Raft) | Sync — quorum write required |
| Cassandra | Async, tunable per query (`QUORUM`, `ONE`, `ALL`) |
| Prometheus remote write | Async |

When you're on-call and a replica is lagging, ask: is this async replication (expected, monitor the lag) or sync (something is wrong, replica is falling behind)?

---

## Key Takeaways

1. A closure is a function bundled with the variables it captured — those variables live on the heap, shared across calls
2. Each call to the outer function creates a new independent captured variable — separate state per closure
3. The closure holds a reference to the captured variable, not a copy — mutations persist across calls
4. `func() string` as a return type means "this function returns a function that returns a string"
5. Variadic functions use `...T` in the signature — inside the function it's a `[]T` slice
6. To pass a slice to a variadic function, unpack it with `slice...`
7. Always check `len()` before accessing `[0]` on a variadic argument — it may be empty
8. `dbLimiter` prints the function address; `dbLimiter()` calls it and returns the result
9. Sync replication: primary waits for replica ACK — no data loss, higher latency
10. Async replication: primary ACKs immediately, replication happens in background — possible data loss, lower latency
11. Replication lag = replica is behind primary — causes read-after-write inconsistency
12. Failover promotes a replica to primary — quorum prevents split-brain in auto-failover

---

> **Al-Qayyum — القَيُّوم — The Self-Subsisting, The Sustainer**
>
> _He sustains all things — replicas sync, counters increment, state is preserved without forgetting. You built functions that remember today. See you on Day 17._

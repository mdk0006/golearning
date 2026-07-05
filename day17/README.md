# Day 17 — Defer, Panic & Recover in Go

---

> **بِسْمِ اللهِ الرَّحْمَنِ الرَّحِيم**
>
> **Al-Ghaffar — الغَفَّار — The Ever-Forgiving**
>
> _He recovers what is lost and restores what is broken. `recover` does the same — it catches what would have crashed and gives the program a chance to continue gracefully. Begin with His name._

---

## Blog of the Day

[Defer, Panic and Recover — The Go Blog](https://go.dev/blog/defer-panic-and-recover)

Read this after the session. The official Go blog post that explains all three mechanisms with examples. Short, clear, essential reading.

---

## Concept 1: `defer` — Deep Dive

`defer` schedules a function call to run when the surrounding function returns — no matter how it exits (normal return, early return, or panic).

```go
func process() {
    defer fmt.Println("3 - cleanup")
    fmt.Println("1 - start")
    fmt.Println("2 - work")
}
// output:
// 1 - start
// 2 - work
// 3 - cleanup   ← defer runs last
```

### Multiple defers — LIFO order (last in, first out)

```go
defer fmt.Println("first")    // registered first — runs third
defer fmt.Println("second")   // registered second — runs second
defer fmt.Println("third")    // registered third — runs first
```

Think of it as a stack — last pushed, first popped.

### Defers run even on early return

```go
func process() error {
    defer fmt.Println("cleanup")
    if err != nil {
        return err   // cleanup still runs here
    }
    return nil       // cleanup runs here too
}
```

This is why `defer file.Close()` works safely — no matter what path the function takes, the file always closes.

### Arguments evaluated at defer time, not run time

```go
x := 10
defer fmt.Println(x)   // x is captured as 10 right now
x = 99
// defer prints: 10 — not 99
```

The value of `x` is locked in at the point `defer` is registered.

---

## Concept 2: `panic` — Unrecoverable Failure

`panic` stops normal execution immediately, runs all deferred functions, then crashes the program with a message and stack trace.

```go
panic("something went terribly wrong")
```

Go itself panics automatically on:
- Nil pointer dereference
- Index out of bounds
- Type assertion failure

### When to `panic` vs return `error`

| Situation | Use |
|-----------|-----|
| File not found | `error` |
| Network timeout | `error` |
| Invalid JSON input | `error` |
| Function called with nil that must not be nil | `panic` |
| Required config missing at startup | `panic` |
| Index out of bounds | `panic` (Go does it automatically) |
| Impossible state — program logic is broken | `panic` |

**The rule:**
> `error` is for things that **can** go wrong during normal operation — external failures, user input, network. `panic` is for things that **should never happen** — programmer mistakes, broken invariants.

If your function can fail due to external factors → return `error`. If it was called incorrectly by the programmer → `panic`.

---

## Concept 3: `recover` — Catching a Panic

`recover` stops a panic from crashing the program. **It only works inside a deferred function.**

```go
func safeRun() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("recovered:", r)
        }
    }()
    panic("something went wrong")
}
```

### Why `r := recover()` and not just `recover()`?

```go
// ❌ stops the panic but you lose the message
defer func() {
    recover()
}()

// ✅ captures the panic value so you can use it
defer func() {
    if r := recover(); r != nil {
        fmt.Println("panic was:", r)   // r = "something went wrong"
        err = fmt.Errorf("caught: %v", r)
    }
}()
```

`r` holds exactly what was passed to `panic(...)`. Without capturing it, you can stop the crash but can't report what went wrong or return a meaningful error.

### Why `defer` is required for `recover`

```go
// ❌ Wrong — recover() runs before any panic, sees nothing
func badRecover() {
    recover()            // returns nil — no panic yet
    panic("crash")       // still crashes
}

// ✅ Correct — deferred function runs at the moment of panic
func goodRecover() {
    defer func() {
        recover()        // runs DURING the panic — catches it
    }()
    panic("crash")       // caught
}
```

You can also use a named function with defer:

```go
func catchPanic() {
    if r := recover(); r != nil {
        fmt.Println("recovered:", r)
    }
}

func riskyWork() {
    defer catchPanic()   // ✅ deferred named function — works perfectly
    panic("crash")
}
```

**Where you define the function doesn't matter. What matters is that it's called with `defer`.**

---

## Concept 4: Named Return Values + `recover`

To return an error from inside a deferred function, you need named return values:

```go
func safeDivide(a, b int) (result int, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("caught panic: %v", r)  // assigns to named return
        }
    }()
    if b == 0 {
        panic("division by zero")
    }
    return a / b, nil
}
```

`result` and `err` are declared in the function signature as real variables. The deferred function can assign to `err` because it's in scope for the entire function. When `safeDivide` returns, whatever `err` holds is what the caller receives.

Without named returns, there's no way to pass an error back from inside a deferred function.

---

## Anonymous Functions

An anonymous function is a function with no name — defined inline:

```go
// named function
func add(a, b int) int { return a + b }

// anonymous function — same, no name
func(a, b int) int { return a + b }
```

Three ways to use one:

```go
// 1. Call immediately
func(a, b int) int { return a + b }(10, 5)

// 2. Assign to variable
add := func(a, b int) int { return a + b }
add(10, 5)

// 3. Defer — schedule for later
defer func() {
    recover()
}()
```

The `}()` at the end closes the function body `}` and calls it `()`. With `defer` in front, the call is scheduled rather than immediate.

---

## Line by Line

```go
func safeDivide(a, b int) (result int, err error) {
```
Named return values — `result = 0`, `err = nil` at start. Deferred functions can modify them.

---

```go
defer func() {
    if r := recover(); r != nil {
        err = fmt.Errorf("caught panic: %v", r)
    }
}()
```
Registers anonymous function to run on return. `r := recover()` captures the panic value. If `r != nil` a panic happened — set `err` with the panic message. `}()` closes and schedules the call.

---

```go
if b == 0 {
    panic("cannot divide by zero")
}
```
Fires panic. Execution stops. Deferred function runs. `recover()` catches `"cannot divide by zero"`. `err` is set. `safeDivide` returns `0, error`.

---

```go
return a / b, nil
```
Normal path — `b != 0`. Deferred function still runs but `recover()` returns `nil`, nothing changes.

---

```go
func connectDB(host string) {
    defer fmt.Printf("closing DB connection to %s \n", host)
    fmt.Printf("connecting to: %s\n", host)
    fmt.Println("running queries")
}
```
`defer` schedules the closing message. Body prints connect and queries first. On return, defer fires last. Demonstrates defer without panic — just guaranteed cleanup.

---

## Mistakes Made Today

### Mistake 1 — `defer recover()` placed inside the `if b == 0` block

```go
// ❌ Wrong — defer only fires when function returns, panic happens first
if b == 0 {
    panic("division by zero")
    defer recover() { ... }   // never reached — panic already fired
}
```

```go
// ✅ Correct — defer registered at the TOP, before any panic
func safeDivide(a, b int) (result int, err error) {
    defer func() {
        if r := recover(); r != nil { ... }
    }()
    if b == 0 {
        panic("division by zero")
    }
}
```

`defer` must be registered before the panic can happen. Always put it at the top of the function.

---

### Mistake 2 — Commented out the entire function

When the structure was unclear, the whole `safeDivide` was commented out. In Go, you can't partially define a function — write the full skeleton (even with an empty body), then fill it in. Commenting out prevents you from seeing compile errors that would guide you.

---

## Final Code

```go
package main

import "fmt"

func safeDivide(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("caught panic: %v", r)
		}
	}()
	if b == 0 {
		panic("cannot divide by zero")
	}
	return a / b, nil
}

func connectDB(host string) {
	defer fmt.Printf("closing DB connection to %s \n", host)
	fmt.Printf("connecting to: %s\n", host)
	fmt.Println("running queries")
}

func main() {
	connectDB("postgres-01")
	result, err := safeDivide(10, 2)
	fmt.Println(result, err)
	result, err = safeDivide(10, 0)
	fmt.Println(result, err)
}
```

Output:
```
connecting to: postgres-01
running queries
closing DB connection to postgres-01
5 <nil>
0 caught panic: cannot divide by zero
```

---

## System Design: Sharding — Horizontal Partitioning

Sharding splits a large dataset across multiple database nodes so no single node holds all the data. Each shard holds a subset.

---

### Why Sharding?

A single database node has limits — CPU, RAM, disk I/O, connection count. When your data grows beyond what one machine can handle efficiently, you shard.

```
Without sharding:
  All 500M users → one PostgreSQL node → slow queries, disk full, one point of failure

With sharding:
  Users A–H → shard-1
  Users I–P → shard-2
  Users Q–Z → shard-3
```

Each shard handles a fraction of the load. Add more shards as data grows.

---

### Shard Key — The Most Important Decision

The **shard key** determines which shard a record goes to. Choose it wrong and you get **hot spots** — one shard doing 90% of the work while others sit idle.

```go
// Bad shard key — all new users go to same shard
shardID = hash(created_at) % N   // time-based — hot shard during peak signup

// Good shard key — evenly distributed
shardID = hash(user_id) % N      // random-looking IDs distribute evenly
```

**Good shard key properties:**
- High cardinality — many distinct values
- Even distribution — no single value dominates
- Immutable — never changes after creation (changing shard key = moving data)
- Appears in most queries — so you can always route to the right shard

---

### Sharding Strategies

**Range sharding** — each shard owns a range of values:
```
shard-1: user_id 1 – 1,000,000
shard-2: user_id 1,000,001 – 2,000,000
shard-3: user_id 2,000,001 – 3,000,000
```
Simple but prone to hot spots — all new users in the highest range shard.

**Hash sharding** — hash the key, modulo by shard count:
```
shard = hash(user_id) % 3
```
Even distribution but resharding is expensive (consistent hashing from Day 15 solves this).

**Directory sharding** — a lookup table maps keys to shards:
```
user_id → shard mapping stored in a separate metadata store
```
Most flexible — move data between shards without formula changes. But the directory itself becomes a bottleneck.

---

### The Cross-Shard Problem

Sharding breaks joins and transactions across shards:

```sql
-- Easy on one node
SELECT u.name, o.total FROM users u JOIN orders o ON u.id = o.user_id

-- Hard when users on shard-1, orders on shard-2
-- You must query both shards and join in application code
```

**Rule:** design your shard key so that related data lives on the same shard. Users and their orders should share the same shard key (`user_id`) so the join stays local.

---

### Sharding vs Replication

| | Sharding | Replication |
|--|----------|-------------|
| Purpose | Scale **writes** and **storage** | Scale **reads** and **availability** |
| Data | Split across nodes | Copied to all nodes |
| Each node holds | A subset of the data | All the data |
| Failure impact | Lose access to that shard's data | Failover to replica |

Real systems use both — each shard has its own replicas.

---

### SRE Relevance

| System | Sharding |
|--------|---------|
| MongoDB | Auto-sharding with configurable shard key |
| Cassandra | Consistent hash sharding built-in |
| MySQL (Vitess) | Horizontal sharding layer on top of MySQL |
| Redis Cluster | 16,384 hash slots distributed across nodes |
| Elasticsearch | Indices split into shards, distributed across nodes |

When a shard goes down, you lose access to that portion of data — unlike replication where another node has a copy. Shards need replicas too.

---

## Key Takeaways

1. `defer` schedules a call to run when the function returns — always, no matter how it exits
2. Multiple defers run in LIFO order — last registered, first executed
3. Defer arguments are evaluated at registration time, not at run time
4. `panic` stops execution immediately, runs deferred functions, then crashes
5. Use `error` for expected failures (network, disk, user input); use `panic` for programmer mistakes
6. `recover()` only works inside a deferred function — called elsewhere it returns nil and does nothing
7. `r := recover()` captures the panic value so you can use it in the error message
8. Named return values let deferred functions return errors to the caller
9. Anonymous function = function with no name; `}()` defines and calls it; `defer func(){}()` schedules it
10. Where you define the recover function doesn't matter — it must be called with `defer`
11. Sharding splits data across nodes to scale writes and storage — replication copies data to scale reads
12. Shard key choice is critical — bad key causes hot spots, good key distributes evenly
13. Cross-shard joins are expensive — design shard keys so related data lives on the same shard
14. Real systems combine sharding + replication — each shard has replicas for availability

---

> **As-Sabur — الصَّبُور — The Patient, The Enduring**
>
> _He endures all things without haste. Debugging defer, panic, and recover requires patience — the execution order is subtle. You stayed with it. See you on Day 18._

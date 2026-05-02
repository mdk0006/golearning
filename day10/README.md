# Day 10 — Goroutines & Channels in Go

---

> **بِسْمِ اللهِ الرَّحْمَنِ الرَّحِيم**
>
> **Al-Qayyum — القَيُّوم — The Self-Subsisting, The Sustainer of All**
>
> _He sustains everything simultaneously — every star, every heartbeat, every moment — without pause or fatigue. Today you learned to run things simultaneously. Begin with His name and complete Month 1._

---

## Blog of the Day

[Concurrency is not parallelism — The Go Blog](https://go.dev/blog/waza-talk)

Read this after the session. Rob Pike explains the difference between concurrency (structuring a program to handle many things) and parallelism (actually running things at the same time). Goroutines are a concurrency tool — parallelism is a side effect.

---

## Concept: Goroutines

A **goroutine** is a lightweight thread managed by the Go runtime — not by the OS. Starting one costs almost nothing (2–4 KB of stack vs 1–8 MB for an OS thread).

```go
go someFunction()   // launches in a new goroutine
```

The `go` keyword is all it takes. The calling function continues immediately — it does not wait. If `main` returns, all goroutines are killed instantly even if they haven't finished.

### Why goroutines are cheap — stack, heap, and threads

**Stack** — small, fast memory for function calls. Each function call pushes a frame (local variables, arguments). When the function returns, the frame is popped instantly. Fixed cost, automatic.

**Heap** — large memory pool for data that needs to outlive the function that created it. Managed by Go's garbage collector. Slower to allocate but lives as long as needed.

**OS Thread** — an independent execution unit the OS kernel schedules. Each gets a fixed stack of 1–8 MB. Creating 10,000 OS threads = ~80 GB of stack alone. Not practical.

**Goroutine** — the Go runtime multiplexes thousands of goroutines onto a small number of OS threads (usually one per CPU core). Each goroutine starts with ~2 KB of stack that grows automatically as needed. When a goroutine blocks (channel wait, I/O), the runtime parks it and runs another on the same OS thread — no OS context switch needed.

```
Go runtime:
  OS Thread 1  ←→  goroutine A, goroutine C, goroutine F ...
  OS Thread 2  ←→  goroutine B, goroutine D, goroutine G ...
```

| | OS Thread | Goroutine |
|--|-----------|-----------|
| Stack size | 1–8 MB, fixed | 2–4 KB, grows as needed |
| Created by | OS kernel | Go runtime |
| Scheduling | OS kernel | Go runtime |
| Cost to create | Heavy | Cheap |
| Practical limit | Hundreds | Hundreds of thousands |

---

## Concept: Channels

A **channel** is a typed pipe between goroutines. One goroutine sends, another receives.

```go
ch := make(chan string)      // unbuffered
ch := make(chan string, 5)   // buffered, capacity 5
```

### The arrow operator `<-`

The direction of `<-` always shows which way data flows:

```go
ch <- value      // SEND — arrow points INTO the channel
value := <-ch    // RECEIVE — arrow points OUT of the channel
```

### Unbuffered channel

Send blocks until someone receives. Receive blocks until someone sends. Both sides must be ready at the same time — like a direct handoff.

```
goroutine A: ch <- "data"   ← blocks here
goroutine B: v := <-ch      ← unblocks A, receives "data"
```

### Buffered channel

Has a waiting room of capacity N. Send only blocks when the buffer is full. Receive only blocks when the buffer is empty.

```go
ch := make(chan string, 5)
ch <- "web-01: OK"   // doesn't block — buffer has room
ch <- "web-02: OK"   // doesn't block
...
fmt.Println(<-ch)    // receives "web-01: OK"
```

Sizing the buffer to exactly the number of goroutines means no goroutine ever has to wait to send — fire and forget.

---

## Anonymous Functions and the IIFE Pattern

An **anonymous function** is a function defined inline with no name — it's a value, like a string or int.

```go
func(name string) {
    fmt.Println(name)
}
```

Putting `(argument)` immediately after the closing `}` calls it right now. This is called an **IIFE** — Immediately Invoked Function Expression.

```go
func(name string) {
    fmt.Println(name)
}("web-01")   // defines it and calls it immediately
```

With `go` in front — launches it as a goroutine:

```go
go func(name string) {
    serversChannel <- name + ": OK"
}(v)
```

Broken down:

```
go    func  (name string)   { ... }   (v)
↑      ↑         ↑             ↑       ↑
run   define  parameter     body    call now,
in    anon    called name            pass v as name
goroutine function
```

---

## The Closure Bug — Why `(v)` Is Not Optional

If you use the loop variable directly inside the goroutine without passing it as an argument:

```go
// ❌ Closure bug
for _, v := range servers {
    go func() {
        serversChannel <- v + ": OK"   // v is shared — changes each iteration
    }()
}
```

`v` is one variable that changes each iteration. By the time the goroutines actually run, the loop may have finished and `v` is the last value — `"db-02"`. All goroutines see the same final value.

```
// Likely output — all the same
db-02: OK
db-02: OK
db-02: OK
db-02: OK
db-02: OK
```

**The fix — pass `v` as an argument:**

```go
// ✅ Safe — v is copied into name at the moment the goroutine launches
for _, v := range servers {
    go func(name string) {
        serversChannel <- name + ": OK"
    }(v)
}
```

`(v)` copies `v`'s current value into `name` immediately. Each goroutine gets its own private copy. Even as `v` changes in subsequent iterations, `name` in each goroutine is unaffected.

---

## The Rule: Don't Share Memory — Communicate

> **Don't communicate by sharing memory. Share memory by communicating.**

Instead of two goroutines writing to the same variable (requires locks, causes data races), pass data through channels. The channel owns the transfer — only one goroutine holds the data at a time.

---

## Mistakes Made Today

### Mistake 1 — Receive loop before goroutines are launched

```go
// ❌ Wrong — tries to receive before anything has been sent — deadlock
for i := 0; i < len(servers); i++ {
    fmt.Println(<-serversChannel)   // blocks forever, goroutines never launch
}
for _, v := range servers {
    go func(name string) { ... }(v)
}
```

The receive loop blocks on the first `<-` waiting for a message. The goroutines that would send those messages come after — they never get to run. The program hangs forever.

```go
// ✅ Correct — launch first, collect second
for _, v := range servers {
    go func(name string) {
        serversChannel <- name + ": OK"
    }(v)
}
for i := 0; i < len(servers); i++ {
    fmt.Println(<-serversChannel)
}
```

---

### Mistake 2 — Wrong channel variable name

```go
// ❌ Wrong — ch is undefined
ch <- v
```

The channel was created as `serversChannel`. Using a different name is a compile error — Go is strict about undefined identifiers.

```go
// ✅ Correct
serversChannel <- name + ": OK"
```

---

### Mistake 3 — `go` with no function body

```go
// ❌ Wrong — incomplete, won't compile
go
```

`go` must be followed by a complete function call — either a named function call or an anonymous function with a body and argument list.

---

## Final Code

```go
package main

import "fmt"

func main() {
	servers := []string{
		"web-01",
		"web-02",
		"web-03",
		"db-01",
		"db-02",
	}
	serversChannel := make(chan string, len(servers))

	// 1. Launch all goroutines
	for _, v := range servers {
		go func(name string) {
			serversChannel <- name + ": OK"
		}(v)
	}

	// 2. Collect all results
	for i := 0; i < len(servers); i++ {
		fmt.Println(<-serversChannel)
	}
	fmt.Println("All checks complete")
}
```

Output (order varies each run — goroutines race):
```
db-02: OK
web-01: OK
web-03: OK
web-02: OK
db-01: OK
All checks complete
```

---

## System Design: Rate Limiting — Token Bucket & Leaky Bucket

An SRE system that accepts external requests must protect its downstream services from being overwhelmed. Rate limiting controls how many requests are allowed through per unit of time.

---

### Why Rate Limiting?

Without it:
- A single misbehaving client sends 100,000 requests/sec → your service crashes
- A sudden traffic spike → cascades into downstream database failures
- A scraper hammers your API → legitimate users get 503s

Rate limiting is one of the first lines of defence in SRE — alongside circuit breakers and bulkheads.

---

### Token Bucket

Imagine a bucket that holds tokens. Tokens are added at a fixed rate. Each request consumes one token. If the bucket is empty, the request is rejected (or queued).

```
Token refill rate: 10 tokens/sec
Bucket capacity:   20 tokens (max burst)

t=0s:  bucket = 20 tokens (full)
       request comes in → consume 1 → bucket = 19
       request comes in → consume 1 → bucket = 18
       ...20 requests in one burst → bucket = 0

t=0.1s: +1 token added → bucket = 1
        request comes in → consume 1 → bucket = 0

t=0.1s: request comes in → bucket empty → REJECTED
```

**Key properties:**
- Allows **bursts** up to the bucket capacity
- Steady-state throughput is capped at the refill rate
- Smooth — doesn't punish normal traffic, only sustained excess

**SRE use cases:** API rate limiting per client, Alertmanager's `rate` and `burst` config, AWS API Gateway throttling, Kubernetes API server rate limits

---

### Leaky Bucket

Imagine a bucket with a hole in the bottom. Requests pour in at any rate. They drain out at a fixed rate. If the bucket overflows, excess requests are dropped.

```
Drain rate:     10 requests/sec (fixed)
Bucket size:    20 requests (queue depth)

Burst of 50 requests arrives:
  → 20 fit in the bucket
  → 30 are dropped immediately
  → bucket drains at exactly 10/sec
```

**Key properties:**
- Output is perfectly smooth — exactly N requests/sec, always
- Bursts are absorbed up to queue depth, excess dropped immediately
- No burst allowed through to downstream — output rate is constant

**SRE use cases:** smoothing traffic before a database that can't handle spikes, shaping outbound API calls to a third-party with strict rate limits

---

### Token Bucket vs Leaky Bucket

| Property | Token Bucket | Leaky Bucket |
|----------|-------------|--------------|
| Burst handling | Allowed up to capacity | Absorbed up to queue, rest dropped |
| Output rate | Variable (up to refill rate) | Perfectly constant |
| Rejected requests | When bucket empty | When queue full |
| Best for | API clients — allow short bursts | Protecting a downstream that needs smooth input |

In practice **token bucket is more common** — it's friendlier to clients that have bursty but low average traffic (a monitoring system that fires 10 alerts at once every hour).

---

### Sliding Window

A third algorithm worth knowing. Instead of tokens, count requests in a rolling time window.

```
Window: 60 seconds, limit: 100 requests

At t=61s, check: how many requests in the last 60s?
  → 87 → allow
  → 103 → reject
```

More accurate than fixed windows (which allow double the rate at window boundaries) but more expensive to implement. Used in Redis rate limiters (`INCR` + `EXPIRE`).

---

### In Go — The `time/rate` Package

Go's standard library has a token bucket implementation:

```go
import "golang.org/x/time/rate"

limiter := rate.NewLimiter(rate.Limit(10), 20)
// 10 tokens/sec refill rate, 20 token burst capacity

if !limiter.Allow() {
    // reject the request
}
```

Every Kubernetes controller uses this internally — the client-go library wraps `rate.NewLimiter` to throttle API server requests.

---

## Month 1 Complete

| Day | Topic | Status |
|-----|-------|--------|
| 01 | Variables, Types, Zero Values | ✅ |
| 02 | Functions, Multiple Return Values | ✅ |
| 03 | Control Flow | ✅ |
| 04 | Structs | ✅ |
| 05 | Pointers | ✅ |
| 06 | Slices | ✅ |
| 07 | Maps | ✅ |
| 08 | Error Handling | ✅ |
| 09 | Interfaces | ✅ |
| 10 | Goroutines & Channels | ✅ |

Month 2 starts with Packages & Modules — building real multi-file Go programs.

---

## Key Takeaways

1. A goroutine is a lightweight thread managed by the Go runtime — starts with ~2 KB stack, grows as needed
2. `go f()` launches `f` in a new goroutine and returns immediately
3. A channel is a typed pipe — `ch <- v` sends, `v := <-ch` receives, arrow shows data direction
4. Unbuffered channel — both sides block until the other is ready (direct handoff)
5. Buffered channel — send only blocks when full, receive only blocks when empty
6. Size buffer to number of goroutines to ensure no goroutine ever blocks on send
7. Anonymous function — `func(args) { body }` — a function as a value, no name needed
8. IIFE — `func(args) { body }(values)` — define and call immediately
9. Closure bug — goroutines sharing a loop variable see its final value; pass it as an argument to copy it
10. `(v)` in `go func(name string){ ... }(v)` copies `v` into `name` at launch time — each goroutine gets its own copy
11. Launch goroutines first, receive results second — reversed order causes deadlock
12. Token bucket — allows bursts, rejects when empty, smooth for bursty clients
13. Leaky bucket — constant output rate, drops excess, protects smooth-input downstreams
14. Random output order from goroutines is expected — they race and finish whenever they finish

---

> **Al-Fattah — الفَتَّاح — The Opener of Gates**
>
> _He opens the gates of understanding for those who seek with sincerity. Month 1 is complete — ten concepts, ten days, ten doors opened. The foundation is built. Rest, reflect, and return for Month 2 with the same sincerity._
>
> **بَارَكَ اللهُ فِيكَ**

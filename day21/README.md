# Day 21 — `sync.WaitGroup` and `sync.Mutex`

---

> **بِسْمِ اللهِ الرَّحْمَنِ الرَّحِيم**
>
> **Al-Muhaymin — المُهَيْمِن — The Guardian, The Overseer**
>
> _He watches over all things and keeps them in order. Today you learn how to coordinate goroutines so they don't step on each other — order in concurrency. Begin with His name._

---

## Blog of the Day

[Share Memory By Communicating — go.dev/blog](https://go.dev/blog/codelab-share)

Read this after the session. The Go team's philosophy on concurrency — when to use channels vs shared memory with mutexes, and why Go favors communication over locking.

---

## Concept 1: The Problem from Day 10

On Day 10 you used channels to pass data between goroutines. Channels are great for communication. But sometimes you need two simpler things:

1. **Wait for all goroutines to finish** — without a channel
2. **Protect shared data** from concurrent writes

The `sync` package gives you both.

---

## Concept 2: `sync.WaitGroup` — Wait for Goroutines

A WaitGroup is a counter:

- `wg.Add(1)` — "one more goroutine starting"
- `wg.Done()` — "one goroutine finished" (decrements)
- `wg.Wait()` — "block here until counter hits zero"

```go
var wg sync.WaitGroup

wg.Add(1)
go func() {
    defer wg.Done()
    // do work
}()

wg.Wait()  // blocks until Done() is called
```

`defer wg.Done()` — always at the top of the goroutine. Even if it panics, Done gets called.

---

## Concept 3: `sync.Mutex` — Protect Shared Data

**The Google Doc analogy:**

Two people edit a shared doc at the same time:

```
Person A reads: errors = 5
Person B reads: errors = 5
Person A writes: errors = 6
Person B writes: errors = 6   ← wrong, should be 7
```

One update was lost. This is a **race condition** — two goroutines write to the same variable simultaneously, the result is unpredictable.

**With a Mutex — only one person holds the pen at a time:**

```
Person A: Lock → reads 5 → writes 6 → Unlock
Person B: Lock → reads 6 → writes 7 → Unlock  ← correct
```

In Go:
```go
mu.Lock()
defer mu.Unlock()
count++   // only one goroutine here at a time
```

---

## Concept 4: `sync.RWMutex` — Reads Don't Block Each Other

A regular `Mutex` locks for both reads and writes. `RWMutex` is smarter:

- Multiple goroutines can **read** simultaneously — `RLock()` / `RUnlock()`
- Only one goroutine can **write** — `Lock()` / `Unlock()`

```go
// reading — many goroutines can do this at the same time
mu.RLock()
val := data["key"]
mu.RUnlock()

// writing — only one at a time
mu.Lock()
data["key"] = "new"
mu.Unlock()
```

For a health registry — many readers, occasional writer — `RWMutex` is the right tool.

---

## Concept 5: Stack vs Heap

Two places Go stores data in memory:

**Stack** — fast, automatic, short-lived. Local variables live here. Gone when the function returns.

**Heap** — slower, long-lived. Data that must outlive the function that created it. Garbage collector cleans it up when nothing points to it.

```go
reg := &HealthRegistry{status: make(map[string]string)}
```

The `&` puts `HealthRegistry` on the heap. It needs to be shared across goroutines and outlive the loop — stack would disappear too soon.

Go decides stack vs heap automatically (escape analysis). When you take the address with `&`, it usually escapes to the heap.

**SRE analogy:**
- Stack = local server RAM — fast, local, gone when the process ends
- Heap = shared storage (Redis, S3) — slower, accessible by anyone, persists longer

---

## Concept 6: Always Use Pointer Receivers With Mutex

```go
// ✅ Correct — pointer receiver
func (r *HealthRegistry) Set(host, status string) {

// ❌ Wrong — value receiver copies the struct, including the mutex
func (r HealthRegistry) Set(host, status string) {
```

A copied mutex is broken — the copy doesn't share lock state with the original. Always use `*T` (pointer receiver) on structs that contain a mutex.

---

## Mistakes Made Today

### Mistake 1 — Wrong method signature syntax

```go
// ❌ Wrong — receiver and parameters mixed up
func (host, status string) Lock(s status) {
    s.host
}
```

```go
// ✅ Correct — receiver first, then method name, then parameters
func (r *HealthRegistry) Set(host, status string) {
    r.status[host] = status
}
```

The receiver `(r *HealthRegistry)` comes before the method name. `r` is how you access the struct inside the method.

---

### Mistake 2 — Capital `R` instead of lowercase `r`

```go
// ❌ Wrong — Go is case-sensitive, R is undefined
defer R.mutex.Unlock()
```

```go
// ✅ Correct
defer r.mutex.Unlock()
```

---

### Mistake 3 — Wrong unlock for RWMutex

```go
// ❌ Wrong — overly complex, not the right pattern
defer r.mutex.RLocker().Unlock()
```

```go
// ✅ Correct — direct method
defer r.mutex.RUnlock()
```

---

### Mistake 4 — Extra parameter in `Get`

```go
// ❌ Wrong — Get reads by host, doesn't need a status parameter
func (r *HealthRegistry) Get(host, status string) string {
```

```go
// ✅ Correct — only host needed to look up the value
func (r *HealthRegistry) Get(host string) string {
```

---

## Final Code

```go
package main

import (
	"fmt"
	"sync"
)

type HealthRegistry struct {
	status map[string]string
	mutex  sync.RWMutex
}

func (r *HealthRegistry) Set(host, status string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.status[host] = status
}

func (r *HealthRegistry) Get(host string) string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.status[host]
}

func main() {
	reg := &HealthRegistry{status: make(map[string]string)}
	var wg sync.WaitGroup

	servers := []string{"web01", "web-02", "db-01", "cache-01"}
	for _, server := range servers {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			if s == "db-01" {
				reg.Set(s, "CRITICAL")
			} else {
				reg.Set(s, "OK")
			}
		}(server)
	}

	wg.Wait()

	for _, server := range servers {
		fmt.Printf("%s: %s\n", server, reg.Get(server))
	}
}
```

**Output:**
```
web01: OK
web-02: OK
db-01: CRITICAL
cache-01: OK
```

---

## System Design: Leader Election — Raft, ZooKeeper, etcd

### The Problem

In a distributed system, you often need exactly one node to be in charge — one scheduler running a job, one service owning a lock, one pod writing to a database. If two nodes both think they're the leader, you get split-brain — duplicate writes, corrupted state.

**How do you pick one leader and keep it consistent?**

---

### What Is Leader Election?

A process where a group of nodes agree on exactly one leader. When the leader fails, the remaining nodes elect a new one automatically.

```
node-1 ──┐
node-2 ──┼── [election] → node-1 is leader
node-3 ──┘

node-1 crashes →

node-2 ──┐
node-3 ──┘── [re-election] → node-2 is new leader
```

---

### The Raft Algorithm

Raft is the most widely understood leader election algorithm. It powers etcd (Kubernetes' backing store).

**Three roles:**
- **Leader** — handles all writes, sends heartbeats to followers
- **Follower** — receives heartbeats, forwards client requests to leader
- **Candidate** — in the middle of an election

**How election works:**
1. Followers expect a heartbeat from the leader every X ms (election timeout)
2. If no heartbeat arrives — the leader is assumed dead
3. A follower becomes a **Candidate**, increments the term number, votes for itself
4. Candidate asks all other nodes for a vote
5. First candidate to get a **majority** (n/2 + 1) becomes the new leader
6. New leader starts sending heartbeats — others revert to Follower

```
Term 1: node-1 is leader
         ↓ node-1 crashes
Term 2: node-2 becomes candidate → gets 2/3 votes → new leader
```

**Why majority?** With 3 nodes, you need 2 votes. Two separate partitions can't both reach majority — prevents split-brain.

---

### etcd — Leader Election in Kubernetes

etcd uses Raft. Every write to the Kubernetes API goes through the etcd leader. If the leader fails, Raft elects a new one in seconds — API server reconnects automatically.

```
kubectl apply -f pod.yaml
  → kube-apiserver
    → etcd leader (via Raft)
      → replicated to etcd followers
```

etcd also exposes distributed locks — used by the Kubernetes scheduler to ensure only one scheduler instance runs at a time.

---

### ZooKeeper

Older than Raft, used by Kafka and Hadoop. Uses a similar consensus protocol (ZAB). Nodes create ephemeral znodes — when a node dies, its znode disappears and triggers re-election.

Kafka is moving away from ZooKeeper to its own Raft-based consensus (KRaft).

---

### Practical SRE: etcd Health

```bash
etcdctl endpoint health --cluster
etcdctl endpoint status --cluster
```

If etcd loses quorum (majority of nodes down), the Kubernetes API becomes read-only — no new pods can be scheduled. This is one of the most severe cluster failures an SRE faces.

**3-node etcd cluster:** tolerates 1 failure (2/3 majority still reachable)
**5-node etcd cluster:** tolerates 2 failures (3/5 majority still reachable)

Always run etcd with an odd number of nodes.

---

### Mutex vs Leader Election

| | `sync.Mutex` | Leader Election (Raft/etcd) |
|--|-------------|----------------------------|
| Scope | Single process, multiple goroutines | Multiple machines across a network |
| Failure handling | No — process crash releases the lock | Yes — new leader elected automatically |
| Tool | Go standard library | etcd, ZooKeeper, Consul |
| Use case | Protect a map, counter, shared struct | Scheduler ownership, distributed lock |

Today's `sync.Mutex` is the single-process version of what Raft does across machines.

---

## Key Takeaways

1. `sync.WaitGroup` — counter for goroutines: `Add(1)` before launch, `Done()` inside goroutine, `Wait()` to block until all finish
2. Always `defer wg.Done()` at the top of a goroutine — ensures it's called even on panic
3. `sync.Mutex` — exclusive lock: `Lock()` before write, `Unlock()` after. Only one goroutine at a time
4. `sync.RWMutex` — multiple readers, one writer: `RLock()`/`RUnlock()` for reads, `Lock()`/`Unlock()` for writes
5. Always use `defer mu.Unlock()` immediately after `mu.Lock()` — prevents forgetting to unlock
6. Always use pointer receivers `*T` on structs with a mutex — copying a mutex breaks it
7. `make(map[string]string)` required before writing to a map — nil map panics on write
8. `&StructLiteral{}` puts the struct on the heap — needed when sharing across goroutines
9. Go has a built-in race detector: `go run -race main.go` — use it to catch race conditions
10. Raft is the consensus algorithm behind etcd — leader elected by majority vote
11. etcd loss of quorum = Kubernetes API goes read-only — one of the most critical SRE incidents
12. Always run etcd with an odd number of nodes — tolerates (n-1)/2 failures
13. `sync.Mutex` = single-process locking; Raft/etcd = distributed locking across machines

---

> **Al-Qayyum — القَيُّوم — The Self-Subsisting, The Sustainer of All**
>
> _He sustains all existence — nothing stands without Him. A leader node sustains the cluster the same way: when it falls, another rises so the system never stops. See you on Day 22._

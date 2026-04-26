# Day 06 — Slices in Go

---

> **بِسْمِ اللهِ الرَّحْمَنِ الرَّحِيم**
>
> **Al-Aleem — العَلِيم — The All-Knowing**
>
> _He knows every concept before you learn it and every mistake before you make it. Come to the table with full attention — the knowledge you build today is a trust He placed in your hands. Begin with His name._

---

## Blog of the Day

[Go Slices: usage and internals — The Go Blog](https://go.dev/blog/go-slices-usage-and-internals)

Read this after the session. It explains the `{pointer, len, cap}` header with diagrams — the exact model you worked with today. The shared-backing-array footgun will click permanently when you see the pictures.

---

## Concept: What Is a Slice?

A Go array has a **fixed size**. `[3]string` is always 3 elements — you can't grow or shrink it. That's almost never what you want in real code.

A **slice** is a dynamic view into an array. Under the hood it is a three-field header:

```
type sliceHeader struct {
    ptr uintptr  // address of the backing array
    len int      // how many elements are visible right now
    cap int      // how far the backing array extends from ptr
}
```

This header is a **value** — 24 bytes on 64-bit systems. When you pass a slice to a function, Go copies this header. The key consequence: **the copy still points at the same backing array.** Two headers, one array.

---

## Length vs Capacity

These two are different and both matter.

| Field | Meaning | What changes it |
|-------|---------|-----------------|
| `len` | how many elements you can read/write with `s[i]` | `append`, re-slice, `make` |
| `cap` | how many elements the backing array can hold from `ptr` | only `make`, `append` with realloc, or `copy` to a larger slice |

```
onCallEngineers := []string{"alice", "bob", "carol"}
```

After this line: `len=3`, `cap=3`. The runtime allocated exactly 3 slots.

```
onCallEngineers = append(onCallEngineers, "danish")
```

After append: `len=4`, `cap=6`. Why 6? The old cap (3) was full, so Go allocated a new backing array with **doubled capacity** (3 × 2 = 6), copied the 3 existing elements, added "danish", and returned a new header pointing at the new array.

> **The doubling rule:** Go roughly doubles capacity each time it needs to grow. This amortizes the cost of repeated appends — instead of one allocation per append, you pay once every N appends.

---

## The Slice Header in Detail

```
After []string{"alice", "bob", "carol"}:

  onCallEngineers
  ┌──────────────────────────────────┐
  │ ptr → ["alice"]["bob"]["carol"]  │
  │ len = 3                          │
  │ cap = 3                          │
  └──────────────────────────────────┘

After append(..., "danish"):

  onCallEngineers
  ┌────────────────────────────────────────────────┐
  │ ptr → ["alice"]["bob"]["carol"]["danish"][_][_] │  ← new backing array, cap=6
  │ len = 4                                         │
  │ cap = 6                                         │
  └────────────────────────────────────────────────┘
```

The two empty slots `[_][_]` are allocated but invisible until a future `append` fills them.

---

## Re-slicing — Shared Backing Array

```go
firstTwoEng := onCallEngineers[0:2]
```

This does **not** allocate a new array. It creates a new header that looks into the same backing array:

```
  onCallEngineers
  ptr → ["alice"]["bob"]["carol"]["danish"][_][_]
         ▲
  firstTwoEng
  ptr ───┘   (same address)
  len = 2
  cap = 6   (from ptr to end of backing array)
```

**Consequence:** mutating `firstTwoEng[0]` changes `onCallEngineers[0]`. Same memory.

Re-slicing is cheap — no allocation, no copy. But it is dangerous if you mutate through it. If you need an independent copy, use `copy`.

---

## Removing an Element — The Idiomatic Pattern

To remove the element at index `i` without leaving a gap:

```go
s = append(s[:i], s[i+1:]...)
```

- `s[:i]` — everything before the element
- `s[i+1:]...` — everything after, unpacked as individual arguments to append
- The `...` is required — `append` takes variadic elements, not a slice, as its second argument

For removing `"bob"` at index 1:

```go
s = append(s[:1], s[2:]...)
```

**Warning:** this mutates the backing array in place. Any other slice sharing that array (like `firstTwoEng`) will silently see the mutation.

---

## The Safe Remove — copy First

If you need to remove an element without corrupting slices that share your backing array, make an independent copy first:

```go
safe := make([]string, len(onCallEngineers))
copy(safe, onCallEngineers)
newOnCallSchedule := append(safe[:1], safe[2:]...)
```

`copy(dst, src)` — **destination first, source second**. It copies `min(len(dst), len(src))` elements. Because `safe` has its own backing array, the append on `safe` cannot reach `onCallEngineers` or `firstTwoEng`.

---

## Slice Operations Quick Reference

| Operation | Shares array? | Notes |
|-----------|--------------|-------|
| `s[i:j]` | **Yes** | new header, same backing array |
| `[]T{...}` literal | No | new array |
| `make([]T, n)` | No | new array, zero-valued |
| `copy(dst, src)` | No | copies into dst's array |
| `append(s, x)` with spare cap | **Yes** | writes into existing array |
| `append(s, x)` full cap | No | allocates new array, copies |

---

## Mistakes Made Today

### Mistake 1 — Thought re-slicing creates a new array

```go
// ❌ Wrong assumption
firstTwoEng := onCallEngineers[0:2]
// "firstTwoEng has its own array"
```

Re-slicing creates a new **header** only. The pointer field still holds the same address. Mutating through `firstTwoEng` changes `onCallEngineers` and vice versa.

```go
// ✅ Correct mental model
// firstTwoEng and onCallEngineers share the same backing array
// Use copy() if you need independence
```

---

### Mistake 2 — copy arguments reversed

```go
// ❌ Wrong — copies FROM empty newOncallSchedule INTO onCallEngineers — does nothing
copy(onCallEngineers, newOncallSchedule)
```

`copy` signature is `copy(dst, src)`. Destination first, source second. And the destination must already have length (a slice with `len=0` receives 0 elements no matter what).

```go
// ✅ Correct — allocate dst first, then copy src into it
safe := make([]string, len(onCallEngineers))
copy(safe, onCallEngineers)
```

---

### Mistake 3 — Created safe copy but still appended on the original

```go
// ❌ Wrong — safe copy was created but ignored
safe := make([]string, len(onCallEngineers))
copy(safe, onCallEngineers)
newOncallSchedule := append(onCallEngineers[:1], onCallEngineers[2:]...)  // still using original!
```

Making a copy is pointless if you then continue to operate on the original. The append must run on `safe`, not on `onCallEngineers`.

```go
// ✅ Correct — operate on safe, not on the original
newOncallSchedule := append(safe[:1], safe[2:]...)
```

---

### Mistake 4 — Uppercase local variable name

```go
// ❌ Wrong — capital N looks like an exported identifier
NewOncallSchedule := ...
```

Local variables are always camelCase in Go. Only exported (package-level) identifiers start with a capital letter.

```go
// ✅ Correct
newOnCallSchedule := ...
```

---

## Final Code

```go
package main

import "fmt"

func main() {
	onCallEngineers := []string{"alice", "bob", "carol"}
	fmt.Println("length of oncallEngineers slice", len(onCallEngineers))
	fmt.Println("cap of oncallEngineers slice", cap(onCallEngineers))
	onCallEngineers = append(onCallEngineers, "danish")
	fmt.Println("updated length of oncallEngineers slice", len(onCallEngineers))
	fmt.Println("updated cap of oncallEngineers slice", cap(onCallEngineers))
	firstTwoEng := onCallEngineers[0:2]
	fmt.Println("First two engineers in rotation", firstTwoEng)
	safe := make([]string, len(onCallEngineers))
	copy(safe, onCallEngineers)
	newOnCallSchedule := append(safe[:1], safe[2:]...)
	fmt.Println("firstTwoEng after remove:", firstTwoEng)
	fmt.Println("New Roster", newOnCallSchedule)
}
```

Output:
```
length of oncallEngineers slice 3
cap of oncallEngineers slice 3
updated length of oncallEngineers slice 4
updated cap of oncallEngineers slice 6
First two engineers in rotation [alice bob]
firstTwoEng after remove: [alice bob]
New Roster [alice carol danish]
```

---

## System Design: Caching — Eviction Policies

A cache holds a limited amount of data in fast storage (memory). When it fills up, it must evict something to make room. The policy you choose determines **what gets thrown out**.

---

### LRU — Least Recently Used

**Rule:** evict the item that was accessed least recently (oldest last-access timestamp).

**How it works internally:** a doubly-linked list + a hash map. Every access moves the item to the front. When full, evict from the back.

```
Access pattern:  web-01 → web-02 → web-01 → web-03 → (cache full, evict one)

State before eviction:
  [web-01 (most recent)] → [web-03] → [web-02 (least recent)]

Evict: web-02
```

**SRE use case — Alertmanager deduplication:**
Alertmanager holds a fingerprint of every active alert to suppress duplicates. Recent alerts are still firing — they need to stay in cache. Alerts from 3 hours ago are probably resolved. LRU naturally keeps the active window hot.

**Drawback:** doesn't consider frequency. A report that runs once a week is accessed once — it looks "stale" right after it runs and gets evicted even though it'll be needed again in 7 days.

---

### LFU — Least Frequently Used

**Rule:** evict the item with the lowest total access count.

**How it works internally:** a frequency counter per item. Every access increments the counter. When full, evict the lowest-count item.

```
Access counts after 1 hour:
  us-east-1  → 9,400 hits  (stays)
  eu-west-1  → 3,200 hits  (stays)
  ap-south-1 → 12 hits     (evict first)
```

**SRE use case — metric label cardinality cache:**
When Prometheus scrapes a target, it needs to look up the label set for each metric. Popular label combinations (region, env, service) are accessed thousands of times per minute. Rare one-off label combos have count 1 and can be safely evicted. LFU keeps the hot label sets in memory.

**Drawback — cache pollution on new items:** a brand new popular item starts with count=1 and gets evicted immediately, before it has a chance to accumulate hits. Old items with high historical counts crowd out new hot items. Some implementations use a decay factor to age down old counts.

---

### TTL — Time To Live

**Rule:** every item expires after a fixed duration, regardless of how often it's accessed.

**How it works:** each cached entry has an expiry timestamp. A background goroutine (or lazy check on access) removes expired entries.

```
Entry cached at 14:00 with TTL=5min → expires at 14:05
Access at 14:04 → hit, returns value
Access at 14:06 → miss, must re-fetch
```

**SRE use case — health check result cache:**
You don't want to hit every downstream service on every health check request. Cache the result for 10 seconds. But you also can't serve a 5-minute-old "healthy" for a node that died 4 minutes ago. TTL forces a guaranteed refresh regardless of access patterns. DNS records work the same way — the TTL on an A record is exactly this.

**Drawback:** if TTL is too short, you're constantly re-fetching (cache misses, high backend load). If too long, you serve stale data. There's no self-tuning — you pick the number upfront and live with it.

---

### Which One for an SRE Alerting System?

| Cache layer | Best policy | Why |
|-------------|-------------|-----|
| Alert deduplication state | **LRU** | Active alerts are recent; resolved ones age out naturally |
| Kubernetes pod metadata | **TTL** | Pod state changes; staleness is dangerous |
| Service label/metric name lookup | **LFU** | Popular labels (env=prod) are always hot |
| Auth token validation | **TTL** | Tokens expire on a known schedule — policy matches reality |

In practice Redis uses **TTL as the baseline** and adds LRU or LFU as the secondary eviction policy via `maxmemory-policy: allkeys-lru`. You almost always combine them.

---

## Key Takeaways

1. A slice is a `{ptr, len, cap}` header — not an array
2. `len` = how many elements you can access; `cap` = how many the backing array holds from ptr
3. Re-slicing (`s[i:j]`) creates a new header but shares the backing array — mutations go both ways
4. `append` never modifies the original header — always assign the return value
5. When `append` exceeds capacity, Go allocates a new (roughly doubled) array and copies
6. `copy(dst, src)` — destination first, source second — gives you a fully independent copy
7. Idiomatic element removal: `s = append(s[:i], s[i+1:]...)` — but it mutates the shared array
8. Safe remove: `make` a new slice, `copy` into it, then do the append on the copy
9. LRU evicts the oldest-accessed item — good for temporal working sets (active alerts)
10. LFU evicts the least-accessed item — good for stable popularity distributions (metric labels)
11. TTL evicts by age — good for data with known staleness (DNS, health checks, auth tokens)
12. Production systems (Redis) combine TTL + LRU/LFU — use TTL for correctness, LRU/LFU for memory pressure

---

> **Al-Hafiz — الحَفِيظ — The Preserver, The Guardian**
>
> _He preserves what matters and lets go of what has served its purpose — the perfect cache. You learned today that not all memory is equal: what you share, you can corrupt; what you copy, you protect. Carry that lesson forward. See you on Day 07._

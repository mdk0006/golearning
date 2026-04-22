# Day 05 — Pointers vs Values in Go

---

> **بِسْمِ اللهِ الرَّحْمَنِ الرَّحِيم**
>
> **Al-Baseer — البَصِير — The All-Seeing**
>
> _He sees every concept you wrestled with, every bug that confused you, every moment you pushed through. Nothing you learn is hidden from Him. Begin with His name and go deep today._

---

## Blog of the Day

[Arrays, slices (and strings): The mechanics of 'append' — The Go Blog](https://go.dev/blog/slices-intro)

The slice footgun you hit today — ghost data in the backing array — is explained in full here. Read it after this session. It will make the `{pointer, len, cap}` header click permanently.

---

## Concept: Go Always Passes by Value

This is the one rule that explains everything about pointers in Go:

> **Go always copies arguments.** There is no pass-by-reference. Ever.

When you call a function, every argument gets copied into the function's own stack frame. The function works on that copy. When it returns, the copy is gone.

This rule applies to all types — `int`, `string`, `struct`, even `*Schedule` (a pointer is just a value that happens to hold a memory address).

---

## The Three Pointer Symbols

Before anything else, these three symbols need to be second nature:

| Symbol | Where you write it | What it means |
|---|---|---|
| `*T` | in a **type** | "pointer to a T" — e.g. `*Schedule` means "pointer to a Schedule" |
| `&x` | on a **value** | "address of x" — turns `x` into a pointer |
| `*p` | on a **pointer variable** | "dereference p" — follow the pointer, give me the actual thing |

`&` and `*` (the operator) are inverses: `*(&x)` is just `x`.

### SRE Analogy

- `sched` is the actual oncall schedule sitting in RAM — like a physical server.
- `&sched` is the **IP address** of that server — not the server itself, just where to find it.
- `*p` is SSH-ing to that IP and getting your hands on the actual machine.

When you write `AddRotationPtr(&sched, r)` you're saying: "Don't give the function a copy of my schedule. Give it the *address* so it can write directly to the original."

---

## Value Receiver vs Pointer Receiver

### Value receiver — `func (s Schedule) AddRotation(r Rotation)`

```
Caller's RAM:
  sched ──► [Schedule{Name:"primary-oncall", Rotations:[...]}]

Function's stack frame (separate copy):
  s     ──► [Schedule{Name:"primary-oncall", Rotations:[...]}]  ← copy
```

`s` is a full copy. You can mutate `s` all you want inside the function. When the function returns, the copy is thrown away. The caller's `sched` is untouched.

### Pointer receiver — `func (s *Schedule) AddRotationPtr(r Rotation)`

```
Caller's RAM:
  sched ──► [Schedule{Name:"primary-oncall", Rotations:[...]}]
                ▲
  s ────────────┘    ← s holds the address, both point at the same struct
```

`s` holds the address of the caller's struct. When you write `s.Rotations = append(...)`, Go auto-dereferences: `(*s).Rotations = append(...)`. You are writing directly into the caller's memory. The mutation survives the function return.

### The key insight

The method bodies are **identical**:
```go
func (s Schedule) AddRotation(r Rotation) {
    s.Rotations = append(s.Rotations, r)   // copy — caller unchanged
}

func (s *Schedule) AddRotationPtr(r Rotation) {
    s.Rotations = append(s.Rotations, r)   // pointer — caller updated
}
```

Same code. Opposite outcomes. The only difference is the receiver type.

---

## The Slice Header — Why It Gets Complicated

A Go slice is not a single thing. It is a **three-field header** in memory:

```
type sliceHeader struct {
    ptr uintptr  // address of the backing array
    len int      // how many elements are visible
    cap int      // how many the backing array can hold
}
```

When you copy a `Schedule` (value receiver), you copy the slice header. That means:
- A new `ptr` field? No — same value, same address. **The backing array is shared.**
- A new `len` field? Yes — the copy has its own `len`. Changes don't propagate.
- A new `cap` field? Yes — the copy has its own `cap`.

This is why value receivers + `append` are not just "no mutation" — they are a **silent data corruption footgun**.

---

## The Ghost Concept — The Footgun Nobody Warns You About

You ran this experiment:

```go
sched := Schedule{
    Name:      "primary-oncall",
    Rotations: make([]Rotation, 0, 4),   // len=0, cap=4
}
sched.AddRotation(Rotation{Engineer: "danish"})
fmt.Println("len:", len(sched.Rotations))   // 0
fmt.Println("ghost:", sched.Rotations[:1])  // [{danish  }]
```

Here is what actually happened inside `AddRotation`:

```
1. s = copy of sched
   s.Rotations.ptr ──► [_  _  _  _]   ← shared backing array (cap=4)
   s.Rotations.len = 0
   s.Rotations.cap = 4

2. append(s.Rotations, r) runs
   cap > len, so NO reallocation
   backing array slot [0] is written: {Engineer:"danish"}
   a new slice header is returned: {same ptr, len=1, cap=4}

3. s.Rotations = that new header
   Now s.Rotations.len = 1

4. Function returns. s is discarded.
   sched.Rotations.len is still 0  ← nobody updated it
   sched.Rotations.ptr still points at the same backing array
```

The result in RAM after the function returns:

```
sched.Rotations:  {ptr → [DANISH  _  _  _], len=0, cap=4}
                           ▲
                           written, but len=0 means caller can't see it
```

Danish is in the array. Invisible. A ghost.

You can see the ghost with:
```go
fmt.Println("ghost:", sched.Rotations[:1])   // [{danish  }]
```

`[:1]` extends the visible window beyond `len` (allowed as long as you stay within `cap`). The data is there.

### Why this is dangerous in production

Suppose two goroutines both take value copies of the same slice (or two functions are called back-to-back). Both have spare capacity. Both `append`. Both write to overlapping positions in the shared backing array. No locks. No errors. Silent data corruption.

This is why the rule exists:
> **If a method mutates the receiver, use a pointer receiver. No exceptions.**

---

## Methods vs Functions — The Idiomatic Difference

Your code used functions-taking-structs first, then methods. Know the difference:

```go
// function — Schedule is just a parameter
func AddRotation(s *Schedule, r Rotation) { ... }
// called as:
AddRotation(&sched, r)

// method — Schedule is the receiver
func (s *Schedule) AddRotation(r Rotation) { ... }
// called as:
sched.AddRotation(r)    ← no & needed, Go takes the address for you
```

Method calls on pointer receivers do **not** require `&` at the call site. Go sees that `AddRotation` wants a `*Schedule`, sees that `sched` is an addressable `Schedule`, and quietly takes `&sched` for you. This is called auto-address-taking and is methods-only.

---

## When to Use Pointer vs Value Receiver — The Rule

Once you write methods on your types, you have to choose one or the other. The Go community rule:

| Situation | Use |
|---|---|
| Method mutates the receiver | **pointer** |
| Struct is large (copying is wasteful) | **pointer** |
| Struct contains `sync.Mutex` or other non-copyable fields | **pointer** |
| Method is a pure read, struct is small and value-like | value |
| **Be consistent** — once one method uses pointer, all methods should | pointer |

For `Schedule` — it contains a slice, it's meant to be mutated — **pointer receivers everywhere**.

---

## Mistakes Made Today

### Mistake 1 — Argument order and wrong variable

```go
// ❌ Wrong — r is a single Rotation, not a slice. And args are swapped.
append(r, s.Rotation)
```

`append` signature: `append(slice, element)` — slice first, element second. And `r` is a `Rotation` (one item), not a `[]Rotation`. You can only append to a slice.

```go
// ✅ Correct
s.Rotations = append(s.Rotations, r)
```

---

### Mistake 2 — Field name typo

```go
// ❌ Wrong — field doesn't exist
s.Rotation   // missing the 's'
```

```go
// ✅ Correct
s.Rotations   // the actual field name from the struct definition
```

---

### Mistake 3 — Second typo in the same field name

```go
// ❌ Wrong — extra 'r' crept in
s.Rotrations
```

```go
// ✅ Correct
s.Rotations
```

Lesson: when the compiler says `s.Rotrations undefined`, read the error literally — it's telling you the field name. Count the characters.

---

### Mistake 4 — Discarding the return value of `append`

```go
// ❌ Wrong — result thrown away, nothing happens
append(s.Rotations, r)
```

```go
// ✅ Correct — assign the returned slice back
s.Rotations = append(s.Rotations, r)
```

`append` does not mutate the original slice. It returns a new slice header (potentially pointing at a new backing array if reallocation happened). If you don't assign the return value, the append is lost. This is the same mistake from Day 04 — repeated here because it's important enough to hit twice.

---

## Final Code

```go
// day05/main.go
package main

import "fmt"

type Rotation struct {
	Engineer string
	Start    string
	End      string
}

type Schedule struct {
	Name      string
	Rotations []Rotation
}

// Value receiver — s is a copy. Mutation does not reach the caller.
func (s Schedule) AddRotation(r Rotation) {
	s.Rotations = append(s.Rotations, r)
}

// Pointer receiver — s is an address. Mutation writes through to caller.
func (s *Schedule) AddRotationPtr(r Rotation) {
	s.Rotations = append(s.Rotations, r)
}

func main() {
	sched := Schedule{
		Name:      "primary-oncall",
		Rotations: make([]Rotation, 0, 4),
	}

	fmt.Println("before - len:", len(sched.Rotations), "cap:", cap(sched.Rotations))

	// Value receiver — caller unchanged
	sched.AddRotation(Rotation{Engineer: "danish"})
	fmt.Println("after AddRotation  - len:", len(sched.Rotations)) // 0

	// Ghost: the write happened into the backing array, caller can't see it
	fmt.Println("ghost:", sched.Rotations[:1]) // [{danish  }]

	// Pointer receiver — caller updated
	sched.AddRotationPtr(Rotation{Engineer: "danish", Start: "09:00", End: "17:00"})
	fmt.Println("after AddRotationPtr - len:", len(sched.Rotations)) // 1
}
```

Output:
```
before - len: 0 cap: 4
after AddRotation  - len: 0
ghost: [{danish  }]
after AddRotationPtr - len: 1
```

---

## System Design: How Kubernetes Controllers Use Pointers (and What Happens When They Don't)

The pointer-vs-value lesson you learned today is not just a Go language detail. It is built into how large-scale distributed systems are designed.

### The Kubernetes Informer Cache

Every Kubernetes controller (the thing that watches Deployments, reconciles state, triggers rollouts) uses an **informer** — a local in-memory cache of all objects in the cluster, synced from the API server.

The informer stores objects as pointers:
```
cache[pod/"web-01"] ──► *Pod{Name:"web-01", Status:"Running", ...}
```

When your controller code asks for a Pod, the informer hands you back that **pointer** — the same pointer that lives in the cache. Not a copy. The address.

### The Bug That Corrupts the Cache

A common mistake Go beginners make when writing controllers:

```go
pod, _ := informer.Lister().Pods("default").Get("web-01")
// pod is *Pod — pointer to the cache entry

pod.Labels["oncall"] = "danish"   // ← NEVER do this
```

You just mutated the live cache entry. The cache now shows a label that was never applied to the real cluster. The next reconcile loop reads the cache, thinks the label is already there, skips the API call. The real pod in Kubernetes has no label. Your controller is lying to itself. Silent corruption — exactly like the ghost in the backing array.

### The Correct Pattern — DeepCopy Before Mutating

```go
pod, _ := informer.Lister().Pods("default").Get("web-01")
// Get a copy — never mutate the cache object directly
podCopy := pod.DeepCopy()
podCopy.Labels["oncall"] = "danish"

// Now apply the copy to the real API server
client.CoreV1().Pods("default").Update(ctx, podCopy, metav1.UpdateOptions{})
```

`DeepCopy()` is generated code in the Kubernetes API machinery. It does what you wish value receivers did — makes a complete independent copy, backing arrays included, so your mutation can't reach the cache.

### The Pattern in Real SRE Work

| Real system | Pointer/value lesson |
|---|---|
| Kubernetes informer cache | Never mutate pointers from the cache — DeepCopy first |
| Prometheus metric registry | Counters are pointers — concurrent increments go to the right place |
| Go HTTP handler `ResponseWriter` | Interface backed by a pointer — your handler and the server share the same response buffer |
| `sync.Mutex` in a struct | Always use pointer receivers — copying a mutex breaks it |

The moment you understand that Go copies values and shares pointers, you start reading production bugs differently. Half of "why did this work but the metric is wrong?" bugs in Go SRE tooling trace back to an accidental copy.

---

## Key Takeaways

1. Go always passes by value — every argument is a copy
2. `&x` = address of x (gives you a pointer) — `*p` = dereference p (gives you the thing)
3. `*T` in a type means "pointer to T" — separate from the dereference operator
4. Value receiver = function gets a copy, caller never sees mutations
5. Pointer receiver = function gets the address, mutations write through to caller
6. Same method body, different receiver type = completely different behaviour
7. A slice header has three fields: `{ptr, len, cap}` — copying the header shares the backing array
8. Value receiver + append with spare capacity = ghost write — data is in the backing array but caller's `len` is still 0
9. `sched.AddRotationPtr(r)` works without `&` — Go auto-takes the address for pointer-receiver methods
10. If any method on a type uses a pointer receiver, all methods should — consistency matters
11. In Kubernetes controllers, objects from the informer are pointers to the cache — DeepCopy before mutating

---

> **Al-Qadir — القَادِر — The All-Powerful**
>
> _With His power, the confusing becomes clear and the complex becomes second nature. You untangled pointers, slice headers, and ghost writes today — not because it was easy, but because you stayed with it. That discipline is His gift to you. See you on Day 06._

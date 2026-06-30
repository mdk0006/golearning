# Day 15 — JSON in Go

---

> **بِسْمِ اللهِ الرَّحْمَنِ الرَّحِيم**
>
> **Al-Mubeen — المُبِين — The Manifest, The One Who Makes Clear**
>
> _He makes the truth clear and unambiguous. JSON is the same — a clear, structured way for systems to speak to each other. Begin with His name._

---

## Blog of the Day

[JSON and Go — The Go Blog](https://go.dev/blog/json-and-go)

Read this after the session. The official explanation of `encoding/json` internals — how Marshal walks a struct using reflection, how struct tags are parsed, and edge cases with nested types.

---

## Concept 1: Marshal and Unmarshal — Two Directions

```
Go struct   →  JSON bytes   =  json.Marshal / json.MarshalIndent
JSON bytes  →  Go struct    =  json.Unmarshal
```

JSON is everywhere in SRE tooling — Kubernetes API responses, Prometheus HTTP API, PagerDuty webhooks, config files. `encoding/json` is the standard library package that handles both directions.

---

## Concept 2: Struct Tags

```go
type Server struct {
    Name     string  `json:"name"`
    CPUUsage float64 `json:"cpu_usage"`
}
```

The backtick string after each field is a **struct tag**. `json:"name"` tells the encoder/decoder to use `"name"` as the JSON key instead of the Go field name `"Name"`.

Without tags:
```go
type Server struct {
    Name string   // no tag
}
// json.Marshal produces: {"Name":"web-01"}  ← capital N, matches Go field name
```

With tags:
```go
type Server struct {
    Name string `json:"name"`
}
// json.Marshal produces: {"name":"web-01"}  ← matches API convention
```

Tags are required whenever the JSON convention (usually `snake_case` or `camelCase`) differs from Go's exported field naming (`PascalCase`) — which is almost always.

---

## Concept 3: `omitempty` — Skipping Zero Values

```go
type Server struct {
    Name     string  `json:"name"`
    Region   string  `json:"region,omitempty"`
    CPUUsage float64 `json:"cpu_usage,omitempty"`
}
```

`omitempty` tells Marshal: **if this field holds its zero value, leave it out of the JSON entirely.**

```go
s := Server{Name: "web-01"}   // Region and CPUUsage left at zero value

data, _ := json.Marshal(s)
fmt.Println(string(data))
// without omitempty: {"name":"web-01","region":"","cpu_usage":0}
// with omitempty:    {"name":"web-01"}
```

**Zero values per type:**

| Type | Zero value | Omitted when |
|------|-----------|-------------|
| `string` | `""` | empty string |
| `int`, `float64` | `0` | exactly zero |
| `bool` | `false` | false |
| pointer, slice, map | `nil` | nil |

**Why this matters for SRE APIs:** optional fields in a health check response — a server with no `Region` set shouldn't send `"region":""` to clients. `omitempty` keeps payloads clean and lets clients distinguish "not provided" from "provided as empty string" (with pointer types).

**Careful with `bool` and `omitempty`:** a `Healthy bool` field with `omitempty` will be **omitted when `false`** — which is rarely what you want for a health flag, since "unhealthy" is meaningful information you don't want to silently drop. Use `omitempty` mainly on strings, numbers, and optional pointer fields — not on booleans that matter both ways.

---

## Concept 4: Other Tag Options

```go
json:"-"                  // never include this field in JSON, ever
json:"name,omitempty"     // rename + omit if zero value
json:"-,"                 // literal field name "-" (rare edge case)
```

```go
type Server struct {
    Name     string `json:"name"`
    Password string `json:"-"`   // never serialize secrets
}
```

`json:"-"` is the pattern for fields that should never leave the process — credentials, internal state, anything sensitive.

---

## Concept 5: Marshal — Go → JSON

```go
server := Server{Name: "web-01", CPUUsage: 45.2, Healthy: true}

data, err := json.Marshal(server)
// data is []byte: {"name":"web-01","cpu_usage":45.2,"healthy":true}
fmt.Println(string(data))
```

`json.MarshalIndent(server, "", "  ")` does the same but pretty-prints with indentation — useful for human-readable output, logs, or debugging. Use plain `Marshal` for wire transmission (smaller payload, no whitespace).

Always check the error — `Marshal` fails on unsupported types like channels or functions, though structs of basic types essentially never fail.

---

## Concept 6: Unmarshal — JSON → Go

```go
var server Server
err := json.Unmarshal([]byte(jsonStr), &server)
```

- `[]byte(jsonStr)` — `Unmarshal` requires `[]byte`, not `string`. Convert with a type conversion.
- `&server` — a **pointer** to the destination struct. `Unmarshal` needs to write into your struct, so it needs the address, not a copy.

```go
var server Server              // empty: Name="", CPUUsage=0, Healthy=false
json.Unmarshal(data, &server)  // writes directly into server's memory
fmt.Println(server.Name)       // now populated
```

If you forget the `&`, you'd pass a copy of the (empty) struct — `Unmarshal` would write into the copy, which is discarded immediately, and your original `server` stays empty with no compile error.

---

## Mistakes Made Today

### Mistake 1 — Wrong variable name in Unmarshal call

```go
// ❌ Wrong — jsonStr was never declared
err := json.Unmarshal([]byte(jsonStr), &DBserver)
```

```go
// ✅ Correct — use the actual variable name
err := json.Unmarshal([]byte(input), &DBserver)
```

---

### Mistake 2 — Manually assigning fields after Unmarshal

```go
// ❌ Wrong — invalid syntax, and defeats the purpose of Unmarshal
DBserver.Name = "db-01", DBserver.Region = "eu-west1"
```

`Unmarshal` already populates every field from the JSON. Manually reassigning afterward is redundant and the comma-chained assignment isn't even valid Go syntax — you can't chain two separate assignment statements with a comma like that.

```go
// ✅ Correct — Unmarshal does the work, just check the error
err := json.Unmarshal([]byte(input), &DBserver)
if err != nil {
    fmt.Println("unmarshal failed:", err)
    return
}
```

---

### Mistake 3 — Missing error check and missing output

The original unmarshal call had no `if err != nil` check and didn't print any of the resulting fields — so there was no way to verify the unmarshal actually worked.

```go
// ✅ Correct — verify and display
if err != nil {
    fmt.Println("unmarshal failed:", err)
    return
}
fmt.Println("Name:", DBserver.Name)
fmt.Println("CPUUsage:", DBserver.CPUUsage)
fmt.Println("Region:", DBserver.Region)
fmt.Println("Healthy:", DBserver.Healthy)
```

---

## Final Code

```go
package main

import (
	"encoding/json"
	"fmt"
)

type Server struct {
	Name     string  `json:"name"`
	CPUUsage float64 `json:"cpu_usage"`
	Region   string  `json:"region"`
	Healthy  bool    `json:"healthy"`
}

func main() {
	servers := []Server{
		{Name: "web-01", CPUUsage: 90, Region: "US-EAST1", Healthy: true},
		{Name: "web-02", CPUUsage: 20, Region: "US-EAST3", Healthy: true},
		{Name: "web-03", CPUUsage: 30, Region: "US-EAST2", Healthy: false},
	}
	input := `{"name":"db-01","region":"eu-west-1","cpu_usage":88.5,"healthy":false}`

	for _, server := range servers {
		data, err := json.MarshalIndent(server, "", "  ")
		if err != nil {
			fmt.Println("marshal failed:", err)
			continue
		}
		fmt.Println(string(data))
	}

	var DBserver Server
	err := json.Unmarshal([]byte(input), &DBserver)
	if err != nil {
		fmt.Println("unmarshal failed:", err)
		return
	}
	fmt.Println("Name:", DBserver.Name)
	fmt.Println("CPUUsage:", DBserver.CPUUsage)
	fmt.Println("Region:", DBserver.Region)
	fmt.Println("Healthy:", DBserver.Healthy)
}
```

Output:
```
{
  "name": "web-01",
  "cpu_usage": 90,
  "region": "US-EAST1",
  "healthy": true
}
{
  "name": "web-02",
  "cpu_usage": 20,
  "region": "US-EAST3",
  "healthy": true
}
{
  "name": "web-03",
  "cpu_usage": 30,
  "region": "US-EAST2",
  "healthy": false
}
Name: db-01
CPUUsage: 88.5
Region: eu-west-1
Healthy: false
```

---

## System Design: Consistent Hashing — Distributed Data Routing

When you have multiple cache or database nodes, you need a way to decide **which node owns which piece of data**. Naive approaches break badly when nodes are added or removed. Consistent hashing solves this.

---

### The Naive Approach — Modulo Hashing

```
node = hash(key) % N    // N = number of nodes
```

```
3 nodes: hash("web-01") % 3 = node 1
         hash("web-02") % 3 = node 0
         hash("web-03") % 3 = node 2
```

**The problem:** when `N` changes (a node is added or removed), almost **every** key remaps to a different node.

```
3 nodes → 4 nodes:
  hash("web-01") % 3 = 1   →   hash("web-01") % 4 = 3   (moved!)
  hash("web-02") % 3 = 0   →   hash("web-02") % 4 = 2   (moved!)
```

Adding one node to a 3-node Redis cluster can invalidate ~75% of your cache. Every client suddenly misses cache for almost every key — a thundering herd hits your database simultaneously.

---

### Consistent Hashing — The Fix

Instead of `hash(key) % N`, place both **nodes** and **keys** on a conceptual ring (0 to 2^32 − 1, often visualized as a circle).

```
                    0
                    │
        node-C ─────┼───── node-A
                    │
              key-X │
                    │
        node-B ─────┴───── 
```

**Algorithm:**
1. Hash each node to a position on the ring
2. Hash each key to a position on the ring
3. A key belongs to the **first node clockwise** from its position

```
Ring positions (simplified):
  node-A: 10
  node-B: 90
  node-C: 200

key "web-01" hashes to 50  →  next node clockwise from 50 is node-B (90)
key "web-02" hashes to 150 →  next node clockwise from 150 is node-C (200)
```

**When you add a node:**

```
Add node-D at position 70

key "web-01" (50) → previously went to node-B (90)
                   → now goes to node-D (70) — only keys between node-A and node-D move
```

Only the keys between the new node and its counter-clockwise neighbor are remapped. Everything else stays exactly where it was. Adding a node only disturbs `~1/N` of the keys, not nearly all of them.

---

### Virtual Nodes — Solving Uneven Distribution

With only a few real nodes on the ring, distribution can be lumpy — one node might own 60% of the ring just by chance of hash placement.

**Fix:** each physical node gets multiple **virtual nodes** scattered around the ring.

```
node-A → virtual positions: 10, 340, 800, 1500, ...
node-B → virtual positions: 90, 410, 920, 1700, ...
```

More virtual nodes per physical node = smoother, more even distribution. Real systems use 100–500 virtual nodes per physical node.

---

### Where Consistent Hashing Is Used

| System | Use |
|--------|-----|
| **Memcached / Redis Cluster** | Determines which shard owns which key |
| **DynamoDB / Cassandra** | Partition placement across nodes — uses consistent hashing internally |
| **CDNs (Akamai, Cloudflare)** | Routes requests to the nearest/correct edge cache node |
| **Load balancers (some configs)** | Session affinity — same client always hits the same backend |
| **Kubernetes (kube-proxy IPVS mode)** | Can use consistent hashing for backend selection |

---

### Why This Matters for SREs

When you scale a Redis cluster from 3 to 4 nodes:
- **Modulo hashing:** ~75% cache miss storm, database gets hammered
- **Consistent hashing:** ~25% of keys remap, manageable load increase

When designing or operating any sharded system — cache, database, message queue partition assignment — ask: **"What happens to existing data when I add or remove a node?"** If the answer is "almost everything moves," you have a modulo-hashing problem hiding in your architecture.

---

## Key Takeaways

1. `json.Marshal` converts a Go struct to JSON bytes; `json.Unmarshal` converts JSON bytes back to a struct
2. Struct tags (`` `json:"name"` ``) control the JSON key name — without them, Go uses the field name as-is
3. `omitempty` skips a field in the output JSON if it holds its zero value — careful with `bool`, since `false` is often meaningful
4. `json:"-"` excludes a field from JSON entirely — used for secrets and internal state
5. `Unmarshal` requires `[]byte`, not `string` — convert with `[]byte(str)`
6. `Unmarshal` requires a pointer (`&struct`) — it writes directly into your struct's memory
7. Always check the error from both Marshal and Unmarshal
8. `MarshalIndent` pretty-prints for humans; `Marshal` is compact for wire transmission
9. Modulo hashing (`hash(key) % N`) remaps almost all keys when `N` changes — causes cache stampedes
10. Consistent hashing places nodes and keys on a ring — only ~1/N of keys remap when a node is added or removed
11. Virtual nodes smooth out uneven distribution on the hash ring
12. Used in Redis Cluster, DynamoDB, Cassandra, CDNs — anywhere data is sharded across nodes that can scale

---

> **Al-Hakeem — الحَكِيم — The All-Wise**
>
> _His design never wastes — every system, every structure, placed with perfect wisdom. You learned today how wise systems route data without chaos when nodes change. See you on Day 16._

# Day 07 — Maps in Go

---

> **بِسْمِ اللهِ الرَّحْمَنِ الرَّحِيم**
>
> **Al-Muhsee — المُحْصِي — The Counter, The Enumerator**
>
> _He keeps count of everything — every action, every atom, every moment. Today you built a counter. Learn this tool well; it will be in every SRE program you write. Begin with His name._

---

## Blog of the Day

[Go Maps in Action — The Go Blog](https://go.dev/blog/maps)

Read this after the session. It covers initialization, zero values, iteration order, and the concurrent-access gotcha — all the things that bite people in production Go.

---

## Concept: What Is a Map?

A map is Go's built-in key-value store. You look something up by key and get the value back in O(1) average time — regardless of how many entries it holds.

```
map[KeyType]ValueType
```

Both types are fixed at declaration. A `map[string]int` only holds `string` keys and `int` values — no mixing.

Under the hood a map is a hash table: Go hashes the key to find the bucket, stores the value there. Lookup, insert, and delete are all O(1) average.

---

## Declaring and Initializing

Two ways:

```go
// literal — ready to use immediately
alertsMap := map[string]int{"HighCPU": 3, "DiskFull": 1}

// make — ready to use, zero entries
alertsMap := make(map[string]int)
```

**Never use `var m map[string]int` without initializing.** That gives you a `nil` map. Reading from a nil map is fine (returns zero value), but writing to it panics:

```go
var m map[string]int
m["HighCPU"] = 3   // panic: assignment to entry in nil map
```

Always use a literal or `make`.

---

## Reading from a Map

### Single-value form

```go
count := alertsMap["HighCPU"]
```

If the key exists, you get the value. If the key **doesn't exist**, you get the **zero value** for the value type — `0` for `int`, `""` for `string`, `false` for `bool`. No panic, no error.

This is the most common footgun: you can't tell the difference between "key missing" and "key present with value 0".

### Two-value form — always prefer this

```go
count, ok := alertsMap["HighCPU"]
if !ok {
    // key does not exist
}
```

`ok` is a `bool`:
- `true` — key was present
- `false` — key was not present, `count` holds the zero value

Use the two-value form any time the key's existence matters — health checks, config lookups, dedup state.

---

## Writing, Deleting, Incrementing

```go
alertsMap["HighCPU"] = 5          // set
alertsMap["HighCPU"]++            // increment — idiomatic, works directly on map values
delete(alertsMap, "DiskFull")     // delete — silent no-op if key doesn't exist
```

**`delete` is case-sensitive and silent.** `delete(m, "diskfull")` when the key is `"DiskFull"` does nothing — no error, no panic. In SRE tooling where alert names are strings, a case typo is an invisible bug.

---

## Iterating

```go
for key, value := range alertsMap {
    fmt.Printf("alert: %s fired %d times\n", key, value)
}
```

**Map iteration order is intentionally random in Go.** Every run may print keys in a different order. This is by design — Go randomizes it to prevent code from accidentally depending on insertion order. If you need sorted output, collect the keys into a slice and sort it first.

---

## Mistakes Made Today

### Mistake 1 — Typos in string keys

```go
// ❌ Wrong
alertsMap := map[string]int{"HighCPU": 3, "Diskfull": 1, "ProdCrashLoop": 7}
```

`"Diskfull"` instead of `"DiskFull"`, `"ProdCrashLoop"` instead of `"PodCrashLoop"`. In a map, the key is an exact string match — wrong case = different key. A lookup for `"DiskFull"` against a map that only has `"Diskfull"` returns `ok=false`. Silent miss.

```go
// ✅ Correct
alertsMap := map[string]int{"HighCPU": 3, "DiskFull": 1, "PodCrashLoop": 7}
```

---

### Mistake 2 — Uppercase local variable names

```go
// ❌ Wrong — capital letter = exported identifier in Go
DiskFull, ok := alertsMap["DiskFull"]
MemoryLeak, exists := alertsMap["MemoryLeak"]
```

In Go, an identifier starting with a capital letter is exported (public). Local variables are always camelCase.

```go
// ✅ Correct
diskFull, ok := alertsMap["DiskFull"]
memoryLeak, exists := alertsMap["MemoryLeak"]
```

---

### Mistake 3 — Typo in delete key (silent no-op)

```go
// ❌ Wrong — "Diskfull" doesn't exist in the map, delete does nothing
delete(alertsMap, "Diskfull")
```

`delete` never errors or panics on a missing key. This is useful but dangerous — a typo gives you zero feedback. `DiskFull` stayed in the final map for two iterations before being caught by reading the output.

```go
// ✅ Correct
delete(alertsMap, "DiskFull")
```

---

### Mistake 4 — Keeping commented-out code

```go
//alertsMap["HighCPU"] = alertsMap["HighCPU"] + 1
alertsMap["HighCPU"]++
```

Once you've replaced a line with the better version, remove the old one. Commented code is noise — it isn't documentation and it isn't running. Delete it.

---

## Final Code

```go
package main

import "fmt"

func main() {
	alertsMap := map[string]int{"HighCPU": 3, "DiskFull": 1, "PodCrashLoop": 7}
	fmt.Println(alertsMap)
	for alertType, alertCount := range alertsMap {
		fmt.Printf("The alert %s occurs %v times \n", alertType, alertCount)
	}
	diskFull, ok := alertsMap["DiskFull"]
	if ok {
		fmt.Printf("DiskFull exists %v times \n", diskFull)
	}
	memoryLeak, exists := alertsMap["MemoryLeak"]
	if !exists {
		fmt.Printf("MemoryLeak does not exist its value is %v \n", memoryLeak)
	}
	alertsMap["HighCPU"]++
	delete(alertsMap, "DiskFull")
	fmt.Println("Printing the final Map")
	for alertType, alertCount := range alertsMap {
		fmt.Printf("The alert %s occurs %v times \n", alertType, alertCount)
	}
}
```

Output:
```
map[DiskFull:1 HighCPU:3 PodCrashLoop:7]
The alert HighCPU occurs 3 times
The alert DiskFull occurs 1 times
The alert PodCrashLoop occurs 7 times
DiskFull exists 1 times
MemoryLeak does not exist its value is 0
Printing the final Map
The alert HighCPU occurs 4 times
The alert PodCrashLoop occurs 7 times
```

---

## System Design: SQL vs NoSQL — When to Use Which

Every SRE eventually has to choose a storage backend. SQL and NoSQL are not competitors — they are tools with different shapes. The skill is knowing which shape fits your problem.

---

### SQL — Relational Databases

Examples: PostgreSQL, MySQL, CockroachDB, Cloud Spanner

Data lives in **tables** with a fixed schema. Rows relate to other rows via foreign keys. The database enforces those relationships.

```
alerts table:
  id | name       | team_id | severity | created_at
  1  | HighCPU    | 3       | critical | 2026-04-25 09:00
  2  | DiskFull   | 3       | warning  | 2026-04-25 10:00

teams table:
  id | name    | oncall
  3  | infra   | danish
```

You can ask: "give me all critical alerts for the infra team this week" — a join across two tables. The database handles it in one query.

**Strengths:**
- **ACID transactions** — Atomicity, Consistency, Isolation, Durability. Either all of a write succeeds or none of it does. Critical for financial data, config changes, anything where partial writes corrupt state.
- **Joins** — query across related tables without duplicating data
- **Schema enforcement** — the database rejects bad data before it's written
- **Flexible queries** — ad-hoc SQL for any question you need to ask

**Weaknesses:**
- **Vertical scaling only** (mostly) — adding more machines is hard. You scale up (bigger server), not out (more servers)
- **Schema changes are painful** at scale — `ALTER TABLE` on a 500M-row table can lock it for hours
- **Not great for unstructured data** — if every alert has a different shape of metadata, SQL fights you

---

### NoSQL — Non-Relational Databases

"NoSQL" covers several very different database types. The common thread: no fixed schema, no joins, designed to scale horizontally.

#### Key-Value (Redis, DynamoDB)

Simplest model: key maps to a value. Like a Go map, but persistent and distributed.

```
"alert:HighCPU:count"  → "4"
"alert:HighCPU:owner"  → "infra-team"
"session:abc123"       → "{user: danish, expires: ...}"
```

- O(1) reads and writes
- Scales to millions of operations/second
- No relationships, no joins, no complex queries — just get/set/delete
- **SRE use cases:** alert dedup state, rate limiting counters, session caching, feature flags

#### Document (MongoDB, Firestore)

Each record is a JSON-like document. No fixed schema — documents in the same collection can have different fields.

```json
{
  "alert_id": "a1",
  "name": "HighCPU",
  "metadata": {
    "node": "web-01",
    "region": "us-east-1",
    "custom_tags": {"team": "infra", "priority": "p1"}
  }
}
```

- Flexible schema — great when each record has a different shape
- No joins — you denormalize (store related data inside the document)
- **SRE use cases:** incident reports, audit logs, config documents with variable structure

#### Wide-Column (Cassandra, Bigtable)

Optimized for massive write throughput and time-series data. Each row has a partition key; data is sorted within partitions.

```
partition: "node:web-01"
  2026-04-25 09:00  → cpu=82%, mem=60%
  2026-04-25 09:01  → cpu=91%, mem=61%
  2026-04-25 09:02  → cpu=88%, mem=63%
```

- Handles millions of writes/second
- Range queries within a partition are fast
- Cross-partition queries are expensive
- **SRE use cases:** metrics storage (Prometheus's long-term backend), distributed tracing spans, event logs

---

### The Decision Framework

| Question | Points toward |
|----------|--------------|
| Do you need transactions across multiple records? | SQL |
| Do records relate to each other (teams own alerts)? | SQL |
| Do you need ad-hoc queries you haven't thought of yet? | SQL |
| Do you need horizontal scale to millions of writes/sec? | NoSQL |
| Is every record a different shape? | Document NoSQL |
| Is it time-series or append-only? | Wide-column NoSQL |
| Is it pure key lookups, counters, or cache? | Key-value NoSQL |

---

### For the SRE Alerting System

| Data | Storage choice | Why |
|------|---------------|-----|
| Alert definitions (name, team, severity, runbook) | **PostgreSQL** | Structured, relational (alerts belong to teams), needs joins |
| Active alert dedup state | **Redis (key-value)** | Fast lookups, TTL-based expiry, no joins needed |
| Alert fire history / audit log | **Cassandra or BigQuery** | Append-only, time-series, high write volume |
| Incident metadata (freeform notes, tags) | **MongoDB** | Variable structure per incident |

In real SRE platforms (PagerDuty, Datadog, Grafana OnCall) you see exactly this pattern: PostgreSQL for the source-of-truth definitions, Redis for hot operational state, a time-series store for history.

---

## Key Takeaways

1. A map is a hash table — O(1) average lookup, insert, delete
2. Always initialize with a literal or `make` — a `nil` map panics on write
3. Reading a missing key returns the zero value silently — never panics
4. Always use the two-value form `v, ok := m[key]` when existence matters
5. `delete(m, key)` on a missing key is a silent no-op — typos give no feedback
6. `m[key]++` is idiomatic for incrementing a map counter
7. Map iteration order is random — by design, don't depend on it
8. Map keys are case-sensitive strings — `"DiskFull"` and `"Diskfull"` are different keys
9. SQL for structured relational data with transactions; NoSQL for scale, flexibility, or simple key lookups
10. In SRE systems, SQL and NoSQL coexist — use each where its shape fits the data

---

> **Al-Hadi — الهَادِي — The Guide**
>
> _He guides to the right path those who seek with sincerity. You chose the right tool for each problem today — that judgment, knowing which shape fits which problem, is what separates a good engineer from a great one. He guided your hands. See you on Day 08._

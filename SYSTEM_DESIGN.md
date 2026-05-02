# System Design Journal

Running notes from the 3-month Go + System Design learning plan.  
One topic per day — concept, how it works, SRE relevance, trade-offs.

---

| Day | Topic | Status |
|-----|-------|--------|
| [Day 01](#day-01--scalability) | Scalability | ✅ |
| [Day 02](#day-02--availability-vs-reliability) | Availability vs Reliability | ✅ |
| [Day 03](#day-03--load-balancers) | Load Balancers | ✅ |
| [Day 04](#day-04--dns) | DNS — How the Internet Resolves Names | ✅ |
| [Day 05](#day-05--kubernetes-controllers--the-informer-cache) | Kubernetes Controllers & the Informer Cache | ✅ |
| [Day 06](#day-06--caching--eviction-policies) | Caching — Redis, In-Memory, Eviction Policies | ✅ |
| [Day 07](#day-07--sql-vs-nosql) | Databases — SQL vs NoSQL, When to Use Which | ✅ |
| [Day 08](#day-08--cap-theorem) | CAP Theorem — Consistency, Availability, Partition Tolerance | ✅ |
| [Day 09](#day-09--message-queues) | Message Queues — Kafka, SQS, Async Communication | ✅ |
| [Day 10](#day-10--rate-limiting) | Rate Limiting — Token Bucket, Leaky Bucket | ✅ |

---

## Day 01 — Scalability

**Covered in:** [day01/README.md](day01/README.md)  
**Reference:** [A Word on Scalability — All Things Distributed](https://www.allthingsdistributed.com/2006/03/a_word_on_scalability.html)

Scalability is the ability of a system to handle more load by adding resources, without requiring a redesign.

- **Vertical scaling** — bigger machine (more CPU, RAM). Simple, but has a hard ceiling and is a single point of failure.
- **Horizontal scaling** — more machines. Harder to build (stateless design, load balancing needed), but no ceiling.

A system is scalable if adding resources produces a proportional increase in throughput. If adding 2× servers only gives you 1.2× throughput, you have a bottleneck that scaling can't fix.

---

## Day 02 — Availability vs Reliability

**Covered in:** [day02/README.md](day02/README.md)

| Term | Definition |
|------|-----------|
| **Availability** | What % of the time is the system responding to requests? (uptime) |
| **Reliability** | When it responds, is the answer correct and consistent? |

A system can be available but unreliable (returns 200 with wrong data). It can be reliable but unavailable (correct when up, but down 20% of the time). You need both.

**SRE framing:** SLIs measure availability (error rate, latency). SLOs set the target. A reliability bug (silent data corruption) is often worse than an availability bug (5xx) because it's harder to detect.

---

## Day 03 — Load Balancers

**Covered in:** [day03/README.md](day03/README.md)

A load balancer sits in front of a fleet of servers and distributes incoming requests so no single server becomes the bottleneck.

**Algorithms:**

| Algorithm | How it works | Best for |
|-----------|-------------|---------|
| Round Robin | Each server gets a request in turn | Homogeneous servers, short requests |
| Least Connections | Route to the server with fewest active connections | Long-lived connections (WebSockets, streaming) |
| IP Hash | Hash the client IP, always route to the same server | Session affinity (stateful apps) |
| Weighted | Some servers get more traffic based on capacity | Mixed hardware fleets |

**L4 vs L7:**
- **L4 (TCP/UDP)** — routes based on IP + port only. Fast, no payload inspection.
- **L7 (HTTP)** — routes based on URL path, headers, cookies. Slower but smarter (can route `/api` to one fleet and `/static` to another).

**SRE relevance:** AWS ALB (L7), NLB (L4), k8s Service (L4 within the cluster), Nginx/HAProxy (L7 in front of it).

---

## Day 04 — DNS

**Covered in:** [day04/README.md](day04/README.md)

DNS translates `google.com` into an IP address like `142.250.80.46`. It is the phone book of the internet.

### The Lookup Chain

```
Your Browser
    → DNS Resolver (your ISP or 8.8.8.8)
        → Root Nameserver (who handles .com?)
            → TLD Nameserver (who handles google.com?)
                → Authoritative Nameserver (what's the IP for google.com?)
                    → returns 142.250.80.46
```

**1. DNS Resolver** — first stop, usually your ISP or a public resolver (Cloudflare `1.1.1.1`, Google `8.8.8.8`). Checks its cache first.  
**2. Root Nameserver** — knows nothing about `google.com` but knows who handles `.com`. 13 root server clusters globally.  
**3. TLD Nameserver** — handles `.com`, `.io`, `.org`. Knows the authoritative nameserver for `google.com`.  
**4. Authoritative Nameserver** — the source of truth. Owned by the domain owner. Returns the actual IP.

### Record Types

| Record | Purpose | Example |
|--------|---------|---------|
| `A` | hostname → IPv4 | `web-01 → 10.0.0.1` |
| `AAAA` | hostname → IPv6 | |
| `CNAME` | alias to another hostname | `www → google.com` |
| `MX` | mail server for a domain | |
| `TXT` | arbitrary text, used for verification | |
| `NS` | which nameservers are authoritative | |

### TTL & Incident Response

Every DNS response includes a **TTL** — how long clients should cache the answer.

**The problem:** TTL is 24h. You need to failover `api.company.com` to a new IP immediately. Clients who cached it won't see the change for up to 24h.

**The fix (proactive):**
1. Keep TTL low on critical records — 300s (5 min) is common
2. Before a planned failover, drop TTL to 60s
3. Wait out the old TTL so all clients pick up the new low TTL
4. Then change the IP — propagates in 60s

**Rule:** lower TTL before you need it, not during the crisis.

**Route 53 health checks** bypass TTL entirely — when a health check fails, Route 53 stops returning that IP immediately regardless of cached TTL.

### SRE Relevance

- Route 53, Cloud DNS — managed authoritative DNS with health-check-based routing
- Split-horizon DNS — same name resolves differently inside vs outside (inside k8s vs public)
- DNS-based load balancing — return multiple IPs, client picks one
- Service discovery in k8s — `my-service.my-namespace.svc.cluster.local` is just DNS

---

## Day 05 — Kubernetes Controllers & the Informer Cache

**Covered in:** [day05/README.md](day05/README.md)

Every Kubernetes controller uses an **informer** — a local in-memory cache of all cluster objects, synced from the API server. The informer stores objects as pointers.

### The Bug That Corrupts the Cache

```go
pod, _ := informer.Lister().Pods("default").Get("web-01")
// pod is *Pod — a pointer to the live cache entry

pod.Labels["oncall"] = "danish"   // ← NEVER do this
```

You just mutated the live cache. The cache now shows a label that was never applied to the real cluster. The next reconcile loop reads the cache, thinks the label is already there, skips the API call. The real pod has no label. Silent corruption.

### The Correct Pattern — DeepCopy Before Mutating

```go
pod, _ := informer.Lister().Pods("default").Get("web-01")
podCopy := pod.DeepCopy()       // independent copy — safe to mutate
podCopy.Labels["oncall"] = "danish"

client.CoreV1().Pods("default").Update(ctx, podCopy, metav1.UpdateOptions{})
```

`DeepCopy()` is generated code in the Kubernetes API machinery. It produces a fully independent copy — no shared backing arrays, no shared maps.

### SRE Relevance

| System | Lesson |
|--------|--------|
| Kubernetes informer cache | Never mutate cache pointers — DeepCopy first |
| Prometheus metric registry | Counters are pointers — concurrent increments go to the right place |
| Go HTTP `ResponseWriter` | Interface backed by a pointer — handler and server share the response buffer |
| `sync.Mutex` in a struct | Always use pointer receivers — copying a mutex breaks it |

---

## Day 06 — Caching & Eviction Policies

**Covered in:** [day06/README.md](day06/README.md)

A cache holds a limited amount of data in fast storage (memory). When it fills up, it must evict something to make room. The eviction policy determines what gets thrown out.

---

### LRU — Least Recently Used

**Rule:** evict the item that was accessed least recently.

**How it works internally:** a doubly-linked list + a hash map. Every access moves the item to the front. When full, evict from the back.

```
Access pattern:  web-01 → web-02 → web-01 → web-03 → (cache full)

State:
  [web-01 (most recent)] → [web-03] → [web-02 (least recent)]

Evict: web-02
```

**Best for:** working sets where recency = relevance. Active alerts, session state, recent request dedup.

**SRE use case:** Alertmanager deduplication cache — active alerts are recent; resolved alerts age out naturally.

**Drawback:** doesn't consider frequency. An item accessed once a week looks stale right after access and gets evicted even though it'll be needed again in 7 days.

---

### LFU — Least Frequently Used

**Rule:** evict the item with the lowest total access count.

```
Access counts after 1 hour:
  us-east-1  → 9,400 hits  (stays)
  eu-west-1  → 3,200 hits  (stays)
  ap-south-1 → 12 hits     (evict first)
```

**Best for:** stable popularity distributions where popular items stay popular.

**SRE use case:** metric label cardinality cache — popular label combinations (`region=us-east-1`, `env=prod`) are accessed thousands of times per minute. One-off label combos have count 1 and can be safely evicted.

**Drawback — cache pollution on new items:** a brand new popular item starts with count=1 and gets evicted before it can accumulate hits. Some implementations add a decay factor to age down historical counts.

---

### TTL — Time To Live

**Rule:** every item expires after a fixed duration, regardless of access pattern.

```
Entry cached at 14:00 with TTL=5min → expires at 14:05
Access at 14:04 → hit
Access at 14:06 → miss, re-fetch required
```

**Best for:** data with known staleness where serving old data is dangerous.

**SRE use case:** health check result cache — you can't serve a 5-minute-old "healthy" for a node that died 4 minutes ago. TTL forces a guaranteed refresh. DNS records work the same way — the TTL on an A record is exactly this mechanism.

**Drawback:** no self-tuning. Too short = constant re-fetches, high backend load. Too long = stale data. You pick the number upfront and live with it.

---

### Which Policy for What

| Cache layer | Best policy | Why |
|-------------|-------------|-----|
| Alert deduplication state | **LRU** | Active alerts are recent; resolved ones age out naturally |
| Kubernetes pod metadata | **TTL** | Pod state changes; staleness is dangerous |
| Service label / metric name lookup | **LFU** | Popular labels are always hot |
| Auth token validation | **TTL** | Tokens expire on a known schedule |
| Recent user sessions | **LRU** | Active sessions are recent |

**In practice:** Redis uses TTL as the baseline and adds LRU or LFU as the secondary eviction policy when memory is full (`maxmemory-policy: allkeys-lru`). Almost all production systems combine them — TTL for correctness, LRU/LFU for memory pressure.

---

## Day 07 — SQL vs NoSQL

**Covered in:** [day07/README.md](day07/README.md)  
**Reference:** [Go Maps in Action — The Go Blog](https://go.dev/blog/maps)

Every SRE eventually has to choose a storage backend. SQL and NoSQL are not competitors — they are tools with different shapes.

---

### SQL — Relational Databases

Examples: PostgreSQL, MySQL, CockroachDB, Cloud Spanner

Data lives in tables with a fixed schema. Rows relate to other rows via foreign keys. The database enforces those relationships.

**Strengths:**
- **ACID transactions** — either all of a write succeeds or none of it does
- **Joins** — query across related tables without duplicating data
- **Schema enforcement** — the database rejects bad data before it's written
- **Flexible ad-hoc queries** — any question you need to ask

**Weaknesses:**
- Vertical scaling only (mostly) — hard to add more machines
- Schema changes at scale are painful — `ALTER TABLE` on 500M rows can lock for hours
- Not great for unstructured or variable-shape data

---

### NoSQL — Non-Relational Databases

"NoSQL" covers several very different types. Common thread: no fixed schema, designed to scale horizontally.

| Type | Examples | Best for |
|------|----------|---------|
| Key-Value | Redis, DynamoDB | Fast O(1) lookups, counters, caching, session state |
| Document | MongoDB, Firestore | Variable-shape records, freeform metadata |
| Wide-Column | Cassandra, Bigtable | Time-series, append-only, millions of writes/sec |

---

### Decision Framework

| Question | Points toward |
|----------|--------------|
| Need transactions across multiple records? | SQL |
| Records relate to each other? | SQL |
| Need ad-hoc queries not defined upfront? | SQL |
| Need horizontal scale to millions of writes/sec? | NoSQL |
| Every record has a different shape? | Document NoSQL |
| Time-series or append-only? | Wide-column NoSQL |
| Pure key lookups, counters, cache? | Key-value NoSQL |

---

### For an SRE Alerting System

| Data | Storage | Why |
|------|---------|-----|
| Alert definitions (name, team, severity, runbook) | **PostgreSQL** | Structured, relational, needs joins |
| Active alert dedup state | **Redis** | Fast lookups, TTL expiry, no joins needed |
| Alert fire history / audit log | **Cassandra / BigQuery** | Append-only, time-series, high write volume |
| Incident metadata (freeform notes, tags) | **MongoDB** | Variable structure per incident |

In real SRE platforms (PagerDuty, Datadog, Grafana OnCall) you see exactly this pattern: PostgreSQL for source-of-truth definitions, Redis for hot operational state, a time-series store for history.

---

## Day 08 — CAP Theorem

**Covered in:** [day08/README.md](day08/README.md)  
**Reference:** [Error Handling and Go — The Go Blog](https://go.dev/blog/error-handling-and-go)

Every distributed system that stores data makes a promise about three properties. CAP says you can only fully guarantee two at the same time.

---

### The Three Properties

| Property | Guarantee |
|----------|-----------|
| **Consistency (C)** | Every read returns the most recent write — or an error. All nodes see the same data. |
| **Availability (A)** | Every request gets a response — never an error. May return stale data. |
| **Partition Tolerance (P)** | The system keeps operating even when network messages between nodes are lost. |

**Partition tolerance is not optional.** In any real distributed system (cloud, multi-AZ, multi-region), network partitions happen. The real choice is between C and A when a partition occurs.

---

### CP vs AP

**CP — Consistent + Partition Tolerant**
During a partition, the system refuses to answer rather than risk returning stale data.

```
Partition occurs → Node B stops serving reads → returns error
Stale data is never returned
```

Examples: etcd, Zookeeper, CockroachDB, HBase

**AP — Available + Partition Tolerant**
During a partition, the system keeps serving requests with potentially stale data.

```
Partition occurs → Node B serves from local copy → client gets a response (possibly stale)
```

Examples: Cassandra, DynamoDB (default), DNS, CouchDB

---

### Real-World Examples

| System | CAP | Why |
|--------|-----|-----|
| etcd | CP | Kubernetes correctness — wrong state is worse than no state |
| Cassandra | AP | Availability matters more than perfect consistency for metrics/logs |
| PostgreSQL (single node) | CA | No partition — one machine |
| DNS | AP | Always answers, may return stale until TTL expires |
| Zookeeper | CP | Leader election — two nodes can't both think they're leader |

---

### PACELC Extension

CAP only covers partition behaviour. PACELC adds: even with no partition (**E**lse), there's a trade-off between **L**atency and **C**onsistency. Quorum writes (3 of 5 replicas) are more consistent but slower than single-replica writes. Stronger consistency = higher latency, always.

---

### For an SRE Alerting System

| Component | Choice | Reason |
|-----------|--------|--------|
| Alert rule definitions | **CP (PostgreSQL)** | Wrong rules = wrong pages |
| Active alert dedup state | **AP (Redis)** | Duplicate alert acceptable; downtime is not |
| Metrics time-series | **AP (Cassandra)** | A few seconds stale is fine |
| On-call schedule | **CP (PostgreSQL)** | Two people can't both be primary on-call |

---

## Day 09 — Message Queues

**Covered in:** [day09/README.md](day09/README.md)  
**Reference:** [The Go Blog: Laws of Reflection](https://go.dev/blog/laws-of-reflection)

Message queues introduce **asynchronous** communication — producer writes a message and moves on immediately; consumer reads at its own pace.

```
Synchronous:  AlertManager → calls PagerDuty API → waits → response → done
Async:        AlertManager → [Queue] → PagerDuty worker reads when ready
```

---

### Kafka

A **distributed log**. Messages are appended to a topic and retained for days/weeks. Consumers track their own offset — different services can read the same topic independently.

**Key properties:**
- Messages persist after being read — replayable
- Multiple independent consumers per topic
- Guaranteed ordering within a partition
- Very high throughput — millions of messages/sec
- You manage the brokers (ops burden)

**SRE use cases:** audit logs, metrics pipelines, alert event streams, change event feeds

---

### SQS

A **managed job queue**. Messages are deleted after a consumer acknowledges them. No replay.

**Key properties:**
- At-least-once delivery — design consumers to be idempotent
- Visibility timeout — unacknowledged messages reappear for retry
- Dead letter queue (DLQ) — messages that fail N times are moved for investigation
- Fully managed — no brokers to operate

**SRE use cases:** async task dispatch, decoupling microservices, triggering Lambda functions

---

### Kafka vs SQS

| Property | Kafka | SQS |
|----------|-------|-----|
| Retention after read | Yes — days/weeks | No — deleted on ack |
| Replay | Yes | No |
| Multiple consumers | Yes — independent | No — one consumer per message |
| Ordering | Guaranteed within partition | Best-effort |
| Ops burden | High | None (managed) |

---

### Key Async Patterns

**Fan-out:** one alert fires → Kafka topic → PagerDuty consumer + Slack consumer + DB consumer all run independently

**Burst buffering:** 10,000 alerts → queue absorbs the spike → worker processes at 100/sec → downstream never overwhelmed

**Retry + DLQ:** failed message reappears after visibility timeout → after N failures moves to DLQ → SRE investigates

---

## Day 10 — Rate Limiting

**Covered in:** [day10/README.md](day10/README.md)  
**Reference:** [Concurrency is not parallelism — The Go Blog](https://go.dev/blog/waza-talk)

Rate limiting controls how many requests are allowed through per unit of time — protecting downstream services from being overwhelmed.

---

### Token Bucket

A bucket holds tokens, refilled at a fixed rate. Each request consumes one token. Requests are rejected when the bucket is empty.

```
Refill rate: 10 tokens/sec
Capacity:    20 tokens (max burst)

Burst of 20 requests → all allowed (consume bucket)
Next request → rejected (bucket empty)
After 0.1s → 1 token refilled → next request allowed
```

- Allows bursts up to capacity
- Steady-state capped at refill rate
- Friendly to bursty-but-low-average traffic

**SRE use cases:** API rate limiting per client, Alertmanager `rate`/`burst` config, Kubernetes API server throttling (client-go uses `rate.NewLimiter`)

---

### Leaky Bucket

Requests pour in at any rate. They drain out at a fixed constant rate. Overflow is dropped immediately.

```
Drain rate: 10 req/sec
Queue size: 20

Burst of 50 arrives → 20 queued, 30 dropped
Queue drains at exactly 10/sec regardless
```

- Output is perfectly constant — no bursts reach downstream
- Best for protecting services that need smooth input

---

### Token Bucket vs Leaky Bucket

| Property | Token Bucket | Leaky Bucket |
|----------|-------------|--------------|
| Burst handling | Allowed up to capacity | Absorbed up to queue, rest dropped |
| Output rate | Variable | Perfectly constant |
| Best for | API clients with bursty traffic | Protecting smooth-input downstreams |

---

### Sliding Window

Count requests in a rolling time window. More accurate than fixed windows (which allow double the rate at boundaries). Used in Redis rate limiters (`INCR` + `EXPIRE`).

---

### In Go

```go
import "golang.org/x/time/rate"

limiter := rate.NewLimiter(rate.Limit(10), 20)
// 10 tokens/sec refill, 20 burst capacity

if !limiter.Allow() {
    // reject
}
```

---

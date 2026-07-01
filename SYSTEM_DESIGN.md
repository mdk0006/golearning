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
| [Day 11](#day-11--api-gateway) | API Gateway — Single Entry Point Pattern | ✅ |
| [Day 12](#day-12--monolith-vs-microservices) | Monolith vs Microservices — Trade-offs | ✅ |
| [Day 13](#day-13--service-discovery) | Service Discovery — Kubernetes DNS, Consul, Service Mesh | ✅ |
| [Day 14](#day-14--load-balancer-deep-dive) | Load Balancer Deep Dive — L4 vs L7 | ✅ |
| [Day 15](#day-15--consistent-hashing) | Consistent Hashing — Distributed Data Routing | ✅ |
| [Day 16](#day-16--replication) | Replication — Primary/Replica, Sync vs Async | ✅ |

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

## Day 11 — API Gateway

**Covered in:** [day11/README.md](day11/README.md)  
**Reference:** [Organizing a Go module — go.dev/doc](https://go.dev/doc/modules/layout)

An API Gateway is a single entry point that sits in front of all backend services. It handles cross-cutting concerns once so every service gets them for free.

---

### The Problem Without a Gateway

Every service must implement auth, rate limiting, TLS, and logging independently. When auth logic changes, every service updates. No single place to observe all traffic.

---

### What a Gateway Does

```
Client → API Gateway → alerting-service
                    → metrics-service
                    → config-service
```

| Concern | Without gateway | With gateway |
|---------|----------------|-------------|
| Auth | Every service validates tokens | Gateway validates, backends trust |
| Rate limiting | Every service implements it | One policy, one place |
| TLS termination | Every service manages certs | Gateway only |
| Routing | Clients know every service URL | Clients know one URL |
| Logging | Scattered | Single access log |

### Request Lifecycle

```
Client → Gateway checks rate limit → validates auth token
       → routes /alerts to alerting-service
       → alerting-service responds
       → Gateway returns response (strips internal headers, adds correlation ID)
```

---

### AWS ALB as an API Gateway

ALB listener rules provide routing:
```
/api/alerts*  → alerting-service target group
/api/metrics* → metrics-service target group
```

ALB + WAF (rate limiting, IP filtering) + Cognito/Lambda authorizer (auth) = full API Gateway without running your own software.

---

### Dedicated Gateways

| Gateway | Used for |
|---------|---------|
| Kong | Plugin-based, on-prem or cloud |
| Envoy | Service mesh sidecar + edge, used by Istio |
| AWS API Gateway | Fully managed, serverless-native |
| Traefik | Kubernetes-native, auto-discovers via labels |
| Nginx | Simple reverse proxy for smaller setups |

In Kubernetes — Ingress + Ingress Controller is the API Gateway for north-south traffic.

---

### When to Use

- Multiple backend services behind one public URL → ✅ Gateway
- Consistent auth across all services → ✅ Gateway
- Single service, simple setup → ❌ Overkill
- Service-to-service internal calls → ❌ Use service mesh instead

---

## Day 12 — Monolith vs Microservices

**Covered in:** [day12/README.md](day12/README.md)  
**Reference:** [Microservices — Martin Fowler](https://martinfowler.com/articles/microservices.html)

Neither architecture is universally right. The choice depends on team size, domain maturity, and operational capacity.

---

### Monolith

All features in one codebase, one binary, one deploy.

**Strengths:** simple to develop and test, no network calls between components, easy cross-cutting changes, low operational overhead.

**Weaknesses:** all-or-nothing scaling, one bad deploy affects everything, codebase grows into a "big ball of mud", technology lock-in.

---

### Microservices

Each feature is a separate service deployed independently, communicating over the network.

**Strengths:** scale components independently, isolated failures, independent deploys per team, technology freedom per service.

**Weaknesses:** network latency and partial failures between services, distributed tracing required, high operational complexity (10 services = 10 pipelines, 10 log streams, 10 health checks).

---

### Trade-off Table

| Factor | Monolith | Microservices |
|--------|----------|--------------|
| Development speed (early) | ✅ Fast | ❌ Slow setup |
| Operational complexity | ✅ Low | ❌ High |
| Independent scaling | ❌ No | ✅ Yes |
| Fault isolation | ❌ One bug kills all | ✅ Contained |
| Debugging | ✅ Simple | ❌ Needs distributed tracing |

---

### The Real Answer — Start Monolith, Split When It Hurts

> "Don't start with microservices. Start with a monolith, understand the boundaries, then extract services where it actually helps." — Martin Fowler

**Split when:** one component has very different scaling needs, teams are blocking each other's deploys, or one component needs a different technology.

**Stay monolith when:** team is small, domain isn't well understood, or you can't afford the operational overhead.

**SRE perspective:** microservices move complexity from code to infrastructure — more health checks, more tracing, more deployment pipelines. The SRE team feels this cost most.

---

## Day 13 — Service Discovery

**Covered in:** [day13/README.md](day13/README.md)  
**Reference:** [HTTP/2 in Go — go.dev/blog](https://go.dev/blog/h2push)

Service Discovery answers: **"where is this service right now?"** Pod IPs change on every restart — callers can never hardcode them.

---

### Kubernetes DNS (CoreDNS)

When you create a Kubernetes Service, CoreDNS automatically registers it:

```
config-service.default.svc.cluster.local → ClusterIP (stable)
```

The ClusterIP never changes. kube-proxy routes traffic from it to healthy pods. Callers just use the service name:

```
http://config-service/config   ← same namespace, short name works
http://config-service.other-namespace.svc.cluster.local  ← cross-namespace
```

---

### Other Patterns

| Mechanism | Where used | How |
|-----------|-----------|-----|
| Kubernetes DNS | Kubernetes | DNS record per Service, kube-proxy routes to pods |
| Consul | VM/bare metal | Service registry, client queries for IPs |
| Service Mesh (Istio) | Kubernetes | Sidecar proxy handles discovery + routing + mTLS |
| Environment variables | Simple k8s | Injected at pod start — stale if service changes |

---

### Common SRE Failure Modes

- **DNS caching too aggressively** — app caches stale IP, pod behind it replaced
- **Missing readiness probe** — pod registered in DNS before app is ready, first requests fail
- **Cross-namespace short name** — resolves to wrong service or NXDOMAIN
- **CoreDNS overload** — high pod churn overwhelms DNS, all service calls slow

---

## Day 14 — Load Balancer Deep Dive: L4 vs L7

**Covered in:** [day14/README.md](day14/README.md)  
**Reference:** [Writing Web Applications in Go — go.dev/doc](https://go.dev/doc/articles/wiki/)

---

### L4 — Transport Layer (TCP/UDP)

Routes by IP + port only. Never opens the packet to read content.

- No SSL termination — encrypted bytes pass through
- Works for any TCP protocol (HTTP, PostgreSQL, Redis, custom binary)
- Extremely fast — no parsing overhead
- Can't route by URL path, headers, or cookies

**AWS:** Network Load Balancer (NLB)
**Kubernetes:** `Service` of type `LoadBalancer` / `NodePort`

---

### L7 — Application Layer (HTTP)

Terminates the client's TLS connection, reads the HTTP request, makes a new connection to the backend.

- Path-based routing — `/api` → service A, `/static` → CDN
- Host-based routing — `api.co` → backend A, `admin.co` → backend B
- SSL termination — backends speak plain HTTP internally
- Auth at the LB — validate tokens before requests reach services
- Header manipulation, sticky sessions via cookies

**AWS:** Application Load Balancer (ALB)
**Kubernetes:** Ingress + Ingress Controller (Nginx, Traefik)

---

### Comparison

| Feature | L4 (NLB) | L7 (ALB) |
|---------|----------|----------|
| Routes by | IP + port | URL, host, headers, cookies |
| SSL termination | ❌ | ✅ |
| Protocol support | Any TCP/UDP | HTTP, HTTPS, gRPC, WebSocket |
| Speed | Faster | Slower (parsing overhead) |
| Auth at LB | ❌ | ✅ |

---

### In Kubernetes

```
External traffic
      ↓
Ingress (L7) — path routing, TLS, host routing
      ↓
Service (L4) — distributes across pods
      ↓
Pod — your Go HTTP server on :8080
```

L7 for smart routing, L4 for pod-level distribution. They stack.

### Health Checks

L4 probe: TCP connect — accepts connection = healthy (doesn't check app logic).
L7 probe: HTTP GET `/health` → must return 200. Your `handleHealth` handler is exactly what ALB, k8s liveness probes, and Consul call. Every service needs `/health`.

---

## Day 15 — Consistent Hashing

**Covered in:** [day15/README.md](day15/README.md)  
**Reference:** [JSON and Go — go.dev/blog](https://go.dev/blog/json-and-go)

When data is sharded across multiple nodes, you need a way to decide which node owns which key — and what happens to that mapping when nodes are added or removed.

---

### The Problem with Modulo Hashing

```
node = hash(key) % N
```

When `N` changes, almost every key remaps to a different node:

```
3 nodes → 4 nodes:
  hash("web-01") % 3 = 1  →  hash("web-01") % 4 = 3  (moved!)
```

Adding one node to a 3-node Redis cluster can invalidate ~75% of the cache — a thundering herd hits the database simultaneously.

---

### Consistent Hashing

Place both nodes and keys on a conceptual ring (0 to 2^32 − 1). A key belongs to the first node clockwise from its position.

```
Add a new node → only keys between the new node and its
counter-clockwise neighbor remap. Everything else stays put.
```

Adding a node disturbs only `~1/N` of keys, not nearly all of them.

---

### Virtual Nodes

Each physical node gets multiple virtual positions on the ring (100–500 typical) to smooth out uneven distribution that would otherwise occur with only a few hash points.

---

### Where It's Used

| System | Use |
|--------|-----|
| Redis Cluster / Memcached | Shard ownership for keys |
| DynamoDB / Cassandra | Partition placement |
| CDNs (Akamai, Cloudflare) | Routes to correct edge cache node |
| kube-proxy (IPVS mode) | Backend selection |

**SRE question to ask:** when scaling a sharded system, "what happens to existing data when I add or remove a node?" If the answer is "almost everything moves," that's a modulo-hashing problem.

---

## Day 16 — Replication

**Covered in:** [day16/README.md](day16/README.md)  
**Reference:** [Functional Options in Go — Dave Cheney](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis)

Replication keeps copies of data on multiple nodes for durability, availability, and read scalability.

---

### Primary / Replica

All writes go to the primary. Replicas receive a copy of the write log and apply it. Reads can be served from any replica.

```
Writes → Primary → replicates → Replica-1 (reads)
                             → Replica-2 (reads)
```

---

### Synchronous Replication

Primary waits for replica to confirm receipt before acknowledging the client.

- No data loss if primary crashes — data already on replica
- Higher write latency — waits for network round trip
- **Used for:** financial data, config changes, etcd

---

### Asynchronous Replication

Primary acknowledges client immediately. Replication happens in background.

- Possible data loss if primary crashes before replica receives write
- Lower write latency — no waiting
- **Used for:** metrics, analytics, read replicas, Prometheus remote write

---

### Replication Lag

With async replication, replicas are always slightly behind — **replication lag**. Causes read-after-write inconsistency: write to primary, read from replica before it's caught up — stale data returned.

**Fix:** route sensitive reads right after a write to the primary.

---

### Comparison

| Property | Synchronous | Asynchronous |
|----------|-------------|-------------|
| Data loss on failure | None | Possible |
| Write latency | Higher | Lower |
| Replica lag | None | Seconds to minutes |

---

### Systems

| System | Type |
|--------|------|
| etcd (Raft) | Sync — quorum required |
| PostgreSQL | Configurable (sync or async) |
| Redis Sentinel | Async |
| Cassandra | Async, tunable per query |
| Prometheus remote write | Async |

**Failover:** when primary fails, a replica is promoted. Quorum prevents split-brain — a node can only become primary if the majority of nodes agree.

---

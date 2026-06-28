# Day 13 — HTTP Client in Go

---

> **بِسْمِ اللهِ الرَّحْمَنِ الرَّحِيم**
>
> **Al-Khabeer — الخَبِير — The All-Aware, The Fully Informed**
>
> _He knows the state of every system — what is healthy, what is failing, what is unreachable. Today you built a tool that asks that same question. Begin with His name._

---

## Blog of the Day

[HTTP/2 in Go — The Go Blog](https://go.dev/blog/h2push)

Read this after the session. Go's `net/http` package supports HTTP/2 transparently — the same client code you wrote today automatically upgrades to HTTP/2 when the server supports it.

---

## Concept 1: Making HTTP Requests

Go's `net/http` package has a built-in HTTP client. No third-party library needed.

```go
resp, err := http.Get("https://example.com/health")
```

But this uses the **default client which has no timeout**. In SRE tooling a hung request blocks the goroutine forever — a single unresponsive server hangs your entire checker. Always create a client with a timeout:

```go
client := &http.Client{
    Timeout: 3 * time.Second,
}
resp, err := client.Get(url)
```

The `Timeout` covers the entire request — DNS lookup, TCP connect, TLS handshake, waiting for response headers, reading the body. If any of these exceed the timeout, Go cancels the request and returns a `context deadline exceeded` error.

---

## Concept 2: The Response — Two Separate Checks

A successful `client.Get` call means you got a response — it does **not** mean the server is healthy.

```go
resp, err := client.Get(url)
if err != nil {
    // network-level failure: DNS failure, TCP refused, timeout
    // resp is nil here — do not touch it
}
defer resp.Body.Close()

if resp.StatusCode == 200 {
    // server is healthy
}
```

**`err != nil`** — something went wrong at the network level before a response arrived. `resp` is `nil`.

**`resp.StatusCode`** — the HTTP status code from the server. `err` can be `nil` (request succeeded) but `StatusCode` can be 500 (server is unhealthy). Check both.

| Status | Meaning |
|--------|---------|
| 200 | OK — healthy |
| 404 | Not found — wrong path |
| 500 | Internal server error — unhealthy |
| 503 | Service unavailable — up but overloaded |

---

## Concept 3: Always Close the Response Body

```go
defer resp.Body.Close()
```

Even if you don't read the body content, you must close it. The body is backed by a live TCP connection. If you don't close it:
- The connection is never returned to the pool
- Go opens a new TCP connection for every request
- Eventually you exhaust file descriptors and the program crashes

Write `defer resp.Body.Close()` immediately after confirming `err == nil` — before doing anything else with the response.

---

## Concept 4: Passing `*http.Client`

```go
func checkURL(client *http.Client, url string) string {
```

`http.Client` is passed as a **pointer** (`*http.Client`). Two reasons:

1. **Shared connection pool** — the client maintains a pool of reusable TCP connections internally. If you passed by value (copying the struct), each call would get a copy without the pool state. All requests share one pool through the pointer.

2. **No unnecessary copy** — `http.Client` contains internal state (mutexes, maps). Copying it would be wrong.

---

## Concept 5: Early Return Pattern

When both branches of an `if/else` return, the `else` is redundant:

```go
// ❌ Unnecessary else
if resp.StatusCode == 200 {
    return fmt.Sprintf("OK: %s", url)
} else {
    return fmt.Sprintf("FAIL: %s - status %d", url, resp.StatusCode)
}

// ✅ Idiomatic — early return, no else needed
if resp.StatusCode == 200 {
    return fmt.Sprintf("OK: %s", url)
}
return fmt.Sprintf("FAIL: %s - status %d", url, resp.StatusCode)
```

If the `if` fires, the function already returned. The line after the `if` block only runs when the condition was false. The `else` adds noise without adding meaning.

---

## Mistakes Made Today

### Mistake 1 — `:=` outside a function

```go
// ❌ Wrong — := only works inside functions
client := &http.Client{Timeout: 3 * time.Second}

func main() { ... }
```

`:=` is a short variable declaration — only valid inside function bodies. Package-level variables use `var`.

```go
// ✅ Correct — move it inside main
func main() {
    client := &http.Client{Timeout: 3 * time.Second}
}
```

---

### Mistake 2 — Wrong slice literal syntax

```go
// ❌ Wrong — square brackets are for indexing, not slice literals
urls := ["url1", "url2", "url3"]
```

```go
// ✅ Correct — Go slice literals use {}
urls := []string{"url1", "url2", "url3"}
```

---

### Mistake 3 — `else` on a new line

```go
// ❌ Wrong — Go requires } and else on the same line
if err != nil {
    return ...
}
else {
```

```go
// ✅ Correct
if err != nil {
    return ...
} else {
```

Go's automatic semicolon insertion puts a `;` after the `}` on its own line, making `else` a syntax error. Always `} else {` on one line.

---

### Mistake 4 — Touching `resp` when `err != nil`

```go
// ❌ Wrong — resp is nil when err != nil, this panics
resp, err := client.Get(url)
defer resp.Body.Close()   // panic: nil pointer dereference
if err != nil { ... }
```

```go
// ✅ Correct — check err first, only touch resp when it's safe
resp, err := client.Get(url)
if err != nil {
    return fmt.Sprintf("ERROR: %s - %s", url, err)
}
defer resp.Body.Close()   // safe — resp is guaranteed non-nil here
```

---

### Mistake 5 — `retrun` typo and `fmt.Srintf` typo

```go
retrun fmt.Sprintf(...)   // ← typo: retrun instead of return
fmt.Srintf(...)           // ← typo: Srintf instead of Sprintf
```

The compiler catches both immediately — Go won't run code with undefined identifiers or syntax errors. Read the error message literally: it tells you the exact line and what it expected.

---

### Mistake 6 — `%s` for an integer status code

```go
// ❌ Wrong — StatusCode is int, %s expects a string
fmt.Sprintf("FAIL: %s - status %s", url, resp.StatusCode)
```

```go
// ✅ Correct — %d for integers
fmt.Sprintf("FAIL: %s - status %d", url, resp.StatusCode)
```

---

## Final Code

```go
package main

import (
	"fmt"
	"net/http"
	"time"
)

func checkURL(client *http.Client, url string) string {
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Sprintf("ERROR: %s - %s", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return fmt.Sprintf("OK: %s", url)
	}
	return fmt.Sprintf("FAIL: %s - status %d", url, resp.StatusCode)
}

func main() {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	urls := []string{
		"https://httpbin.org/status/200",
		"https://httpbin.org/status/500",
		"https://httpbin.org/delay/5",
	}
	for _, url := range urls {
		fmt.Println(checkURL(client, url))
	}
}
```

Output:
```
OK: https://httpbin.org/status/200
FAIL: https://httpbin.org/status/500 - status 500
ERROR: https://httpbin.org/delay/5 - Get "https://httpbin.org/delay/5": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
```

---

## System Design: Service Discovery

When `alerting-service` needs to call `config-service`, it can't hardcode an IP. Pod IPs change every time a pod restarts. Service Discovery is the mechanism that answers: **"where is this service right now?"**

---

### The Problem

```
alerting-service hardcodes: http://10.0.4.21:8080/config
→ config-service pod restarts
→ new IP: 10.0.4.87
→ alerting-service is now broken
```

Manual IP management doesn't scale. You need something that tracks where services are and stays current.

---

### Kubernetes DNS — Service Discovery You Already Use

When you create a Kubernetes Service, CoreDNS automatically registers it:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: config-service
  namespace: default
```

CoreDNS creates:
```
config-service.default.svc.cluster.local → ClusterIP (e.g. 10.96.45.12)
```

The ClusterIP is stable — it never changes even when pods restart. kube-proxy routes traffic from the ClusterIP to healthy pods automatically.

`alerting-service` just calls:
```
http://config-service/config
```

DNS resolves `config-service` → `10.96.45.12` → kube-proxy → healthy pod. The caller never knows or cares about pod IPs.

---

### DNS Record Format in Kubernetes

```
<service-name>.<namespace>.svc.cluster.local
```

| Call | Resolves to |
|------|-------------|
| `http://config-service` | Same namespace — short name works |
| `http://config-service.default` | Explicit namespace |
| `http://config-service.default.svc.cluster.local` | Fully qualified — always works |

Within the same namespace, the short name is enough. Cross-namespace calls need the full path.

---

### Other Service Discovery Patterns

Kubernetes DNS is **server-side discovery** — a load balancer (kube-proxy) sits between caller and callee. Other patterns exist:

**Client-side discovery (Consul, Eureka)**
Each service registers itself in a registry with its IP and port. Callers query the registry directly and pick an instance themselves.

```
config-service starts → registers in Consul: {ip: 10.0.4.87, port: 8080}
alerting-service → queries Consul for "config-service" → gets list of IPs → picks one
```

Used heavily in non-Kubernetes environments (bare metal, VMs). Consul is common in HashiCorp stacks.

**Environment variables (simple, but fragile)**
Kubernetes injects service IPs as env vars into every pod:
```
CONFIG_SERVICE_HOST=10.96.45.12
CONFIG_SERVICE_PORT=8080
```
Works but doesn't update when services change. DNS is always preferred.

**Service Mesh (Istio, Linkerd)**
A sidecar proxy (Envoy) runs alongside every pod. All traffic goes through the sidecar, which handles discovery, load balancing, retries, and mTLS automatically. The app code calls `http://config-service` — the sidecar intercepts and handles the rest.

---

### Comparison

| Mechanism | Where used | How it works |
|-----------|-----------|-------------|
| Kubernetes DNS (CoreDNS) | Kubernetes | DNS record per Service, kube-proxy routes to pods |
| Consul | VM/bare metal, multi-cloud | Service registry, client queries for IPs |
| Eureka | Java/Spring ecosystem | Client-side registry, Netflix-origin |
| Service Mesh (Istio) | Kubernetes | Sidecar proxy handles all discovery + traffic management |
| Environment variables | Simple k8s | Injected at pod start, stale if service changes |

---

### SRE Relevance

Service discovery failures are a common source of incidents:

- **DNS caching too long** — app caches the IP from DNS and doesn't refresh; the pod behind it was replaced
- **Missing readiness probe** — pod is registered in DNS before the app is ready, first requests fail
- **Cross-namespace calls** — using short name across namespaces fails silently (resolves to wrong service or NXDOMAIN)
- **CoreDNS overload** — high-churn environments (many pod restarts) can overwhelm CoreDNS; response times spike, all service-to-service calls slow

---

## Key Takeaways

1. Always create `http.Client` with a `Timeout` — the default client has no timeout
2. `Timeout` covers the entire request — DNS, TCP, TLS, headers, body
3. `err != nil` = network failure, `resp` is nil — never touch `resp` before checking `err`
4. `resp.StatusCode` must be checked separately — `err == nil` just means you got a response, not that it was healthy
5. Always `defer resp.Body.Close()` immediately after a successful request — TCP connection leaks without it
6. Pass `*http.Client` as a pointer — shared connection pool, correct internal state
7. Early return pattern — when both `if` and `else` return, drop the `else`
8. `%d` for integers, `%s` for strings, `%.1f` for floats — always match verb to type
9. Kubernetes DNS is service discovery — CoreDNS registers a record per Service, ClusterIP is stable
10. Pod IPs change on restart — always call services by DNS name, never by IP
11. Within a namespace, short name works (`config-service`); cross-namespace needs full path
12. Service mesh (Istio) takes discovery further — sidecar handles routing, retries, mTLS automatically
13. CoreDNS overload and DNS caching are common service discovery failure modes in production

---

> **Al-Hadi — الهَادِي — The Guide**
>
> _He guides every call to its destination — no request is lost with Him. You built a tool today that finds the truth about every server it reaches. See you on Day 14._

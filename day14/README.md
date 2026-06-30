# Day 14 — HTTP Server in Go

---

> **بِسْمِ اللهِ الرَّحْمَنِ الرَّحِيم**
>
> **Al-Mujib — المُجِيب — The Responsive, The Answerer**
>
> _Every call finds its answer with Him. Today you built a server that answers every call — health checks, metrics, unknown paths — each one handled and responded to. Begin with His name._

---

## Blog of the Day

[Writing Web Applications in Go — go.dev/doc](https://go.dev/doc/articles/wiki/)

Read this after the session. Go's official guide to building web apps — covers handlers, templates, and persistent state. The patterns you used today are the same ones used in production Go services.

---

## Concept 1: Building an HTTP Server

Go's `net/http` package makes building a server two steps:

```go
http.HandleFunc("/health", handleHealth)   // register a route
log.Fatal(http.ListenAndServe(":8080", nil))  // start listening
```

That's it. No framework required. The standard library handles TCP connections, HTTP parsing, concurrent request handling, and routing.

---

## Concept 2: Handler Functions

A handler is any function with this exact signature:

```go
func handlerName(w http.ResponseWriter, r *http.Request)
```

- **`w http.ResponseWriter`** — what you write your response into. Status code, headers, body all go here. It implements `io.Writer` so `fmt.Fprintf(w, ...)` works exactly like writing to a file.
- **`r *http.Request`** — the incoming request. Contains the method (`r.Method`), URL (`r.URL.Path`), headers (`r.Header`), and body (`r.Body`).

---

## Concept 3: Writing a Response

```go
w.WriteHeader(200)             // set the HTTP status code
fmt.Fprintf(w, "status: healthy")  // write the body
```

If you call `fmt.Fprintf` without `WriteHeader`, Go automatically sends `200`. Always set it explicitly — it's more intentional and readable.

`fmt.Fprintf` works on `w` because `http.ResponseWriter` satisfies `io.Writer` — the same interface used when writing to files on Day 12. Go's interfaces mean one function (`fmt.Fprintf`) works for files, HTTP responses, buffers, and anything else that can be written to.

---

## Concept 4: Why `*http.Request` but not `*http.ResponseWriter`?

```go
func handleHealth(w http.ResponseWriter, r *http.Request)
```

`w` has no star. `r` has a star. This trips everyone up — here's why.

### `r *http.Request` — pointer to a struct

`http.Request` is a **struct** — a large one. It holds the URL, all headers, the body, cookies, form data, query params, context, and more. Passing it by value would copy the entire thing for every single request. That's wasteful.

By using `*http.Request` (a pointer), Go passes just the **memory address** — 8 bytes. The handler gets direct access to the original request data, no copy made.

```go
// What's inside http.Request (simplified):
type Request struct {
    Method     string         // "GET", "POST"
    URL        *url.URL       // parsed URL
    Header     Header         // all headers — a map
    Body       io.ReadCloser  // request body
    // ... many more fields
}
```

Copying all of that per request = wasteful. Pointer = just the address.

### `w http.ResponseWriter` — no star because it's an interface

`http.ResponseWriter` is an **interface**, not a struct:

```go
type ResponseWriter interface {
    Header() Header
    Write([]byte) (int, error)
    WriteHeader(statusCode int)
}
```

Interfaces in Go are already reference-like under the hood — they hold a pointer to the concrete type internally. Adding `*` would give you a pointer to an interface, which is almost never correct in Go.

**Rule:**
- Struct that's large or should not be copied → use `*StructName`
- Interface → never add `*`, it's already reference-like

```go
// Summary
w http.ResponseWriter   // interface — already reference-like, no * needed
r *http.Request         // struct — large, pass by pointer to avoid copying
```

---

## Concept 5: Routing — `http.HandleFunc`

`http.HandleFunc` registers a path and its handler:

```go
http.HandleFunc("/health", handleHealth)    // GET /health → handleHealth
http.HandleFunc("/metrics", handleMetrics)  // GET /metrics → handleMetrics
http.HandleFunc("/", handle404)             // everything else → handle404
```

Go's default router (ServeMux) matches paths in order of specificity — `/health` matches before `/`. The `"/"` pattern is the catch-all — it matches any path not claimed by a more specific route.

**The router does the path matching — you never write that logic yourself.** There is no switch on `r.URL.Path` in `main`. Register the route, write the handler, done.

---

## Concept 5: `log.Fatal` for Server Errors

```go
log.Fatal(http.ListenAndServe(":8080", nil))
```

`http.ListenAndServe` blocks forever while the server runs. It only returns if something goes wrong — port already in use, permission denied, etc.

`log.Fatal` prints the error with a timestamp and calls `os.Exit(1)`. Without it, a startup failure would be completely silent — the program would just exit with no message.

---

## Concept 6: The Routing Decision Belongs in `main`, Not the Handler

A common mistake is putting path-switching logic inside the handler:

```go
// ❌ Wrong — handler doing the router's job
func handle(w http.ResponseWriter, r *http.Request) {
    switch r.URL.Path {
    case "/health":
        ...
    case "/metrics":
        ...
    }
}
```

Each handler should have one responsibility. `main` wires paths to handlers. Handlers process requests and write responses. Keep them separate.

---

## Line by Line

```go
import ("fmt" "log" "net/http")
```
- `fmt` — `Fprintf` to write response bodies
- `log` — `log.Fatal` to handle server startup errors
- `net/http` — HTTP server, handler types, `HandleFunc`, `ListenAndServe`

---

```go
func handleHealth(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(200)
    fmt.Fprintf(w, "status: healthy")
}
```
Handler for `/health`. Sets status 200, writes body. `r` is available but not used here — the path and method are already handled by the router.

---

```go
func handleMetrics(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(200)
    fmt.Fprintf(w, "cpu_usage 42.5 \nmemory_usage 61.2")
}
```
Handler for `/metrics`. Returns fake Prometheus-style metrics. In a real system this would read from actual gauges.

---

```go
func handle404(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(404)
    fmt.Fprintf(w, "not found")
}
```
Catch-all handler. Any path not registered elsewhere lands here. Returns 404 explicitly.

---

```go
http.HandleFunc("/health", handleHealth)
http.HandleFunc("/metrics", handleMetrics)
http.HandleFunc("/", handle404)
```
Register three routes. The router matches by path — `/` catches everything not matched above.

---

```go
log.Fatal(http.ListenAndServe(":8080", nil))
```
Start the server on port 8080. Blocks forever. `nil` means use the default router (the one you registered routes on above). `log.Fatal` exits with an error if startup fails.

---

## Mistakes Made Today

### Mistake 1 — Switch in `main` trying to do routing

```go
// ❌ Wrong — r doesn't exist in main, and routing is not your job
func main() {
    switch {
    case r.Request == "/health":
        http.HandleFunc("/health", handleHealth)
    }
}
```

`r` (the request) only exists inside handler functions — it's created per request as clients connect. In `main` there are no requests yet. And `http.HandleFunc` already handles routing — you just declare which path goes to which handler.

```go
// ✅ Correct — register once, router handles the rest
func main() {
    http.HandleFunc("/health", handleHealth)
    http.HandleFunc("/metrics", handleMetrics)
    http.HandleFunc("/", handle404)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

---

### Mistake 2 — `return 404` in `main`

```go
// ❌ Wrong — can't return a value from main, and 404 is a response not a return value
default:
    return 404
```

`main` returns nothing. HTTP status codes are written via `w.WriteHeader(404)` inside a handler function, not returned from `main`.

---

### Mistake 3 — `http.ListenAndServe` error not handled

```go
// ❌ Wrong — silent failure if port is in use
http.ListenAndServe(":8080", nil)
```

```go
// ✅ Correct — prints error and exits if server can't start
log.Fatal(http.ListenAndServe(":8080", nil))
```

---

## Final Code

```go
package main

import (
	"fmt"
	"log"
	"net/http"
)

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	fmt.Fprintf(w, "status: healthy")
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	fmt.Fprintf(w, "cpu_usage 42.5 \nmemory_usage 61.2")
}

func handle404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	fmt.Fprintf(w, "not found")
}

func main() {
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/metrics", handleMetrics)
	http.HandleFunc("/", handle404)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

Test with curl:
```bash
curl localhost:8080/health    # → status: healthy (200)
curl localhost:8080/metrics   # → cpu_usage 42.5 \nmemory_usage 61.2 (200)
curl localhost:8080/unknown   # → not found (404)
```

---

## System Design: Load Balancer Deep Dive — L4 vs L7

You covered Load Balancers briefly on Day 03. Today go deeper — the difference between L4 and L7 is something every SRE gets asked in interviews and deals with daily.

---

### Where Load Balancers Sit in the OSI Model

```
Layer 7 — Application  (HTTP, gRPC, headers, cookies, URLs)
Layer 4 — Transport    (TCP, UDP — IP + port only)
Layer 3 — Network      (IP addresses)
Layer 2 — Data Link    (MAC addresses)
```

A **Layer 4 load balancer** sees TCP packets — IP addresses and ports only. It has no idea what's inside the packet.

A **Layer 7 load balancer** terminates the HTTP connection, reads the actual request, then makes a new connection to the backend. It can see everything — method, path, headers, cookies.

---

### L4 Load Balancer

Routes based on **IP + port** only. Never opens the TCP packet to look inside.

```
Client → L4 LB (sees: src IP, dst port 443) → backend-01:8080
                                             → backend-02:8080
                                             → backend-03:8080
```

**How it works:**
- Receives a TCP SYN packet
- Picks a backend (round-robin, least connections, IP hash)
- Forwards all packets for this connection to the same backend (connection stickiness)
- Never reads the HTTP content

**Strengths:**
- Extremely fast — no HTTP parsing, no TLS termination
- Works for any TCP protocol — HTTP, gRPC, PostgreSQL, Redis, custom binary protocols
- Very low latency overhead

**Weaknesses:**
- Can't route based on URL path, headers, or cookies
- Can't do SSL termination (the encrypted bytes pass through)
- Can't inspect or modify requests

**AWS equivalent:** Network Load Balancer (NLB)
**Kubernetes equivalent:** `Service` of type `LoadBalancer` or `NodePort`

---

### L7 Load Balancer

**Terminates** the client's TCP/TLS connection, reads the HTTP request, then makes a **new** connection to the backend.

```
Client ──TLS──► L7 LB  ──HTTP──► alerting-service
               reads:            (new TCP connection)
               GET /alerts
               Host: api.co
               Auth: Bearer xyz
```

**How it works:**
1. Client connects to the L7 LB — TLS handshake happens here
2. L7 LB decrypts the request, reads the HTTP headers and path
3. L7 LB applies rules — auth check, rate limit, path routing
4. L7 LB makes a new HTTP connection to the selected backend
5. L7 LB forwards the response back to the client

**Strengths:**
- Path-based routing — `/api` → service A, `/static` → CDN
- Host-based routing — `api.co` → backend, `admin.co` → different backend
- SSL termination — backends speak plain HTTP internally, one cert at the LB
- Auth at the gateway — validate tokens before requests reach backends
- Header manipulation — add/strip/rewrite headers
- Sticky sessions — route same user to same backend via cookie

**Weaknesses:**
- Slower than L4 — TLS termination + HTTP parsing adds latency
- Doesn't work for non-HTTP protocols without special config
- More complex — more attack surface, more config

**AWS equivalent:** Application Load Balancer (ALB)
**Kubernetes equivalent:** Ingress + Ingress Controller (Nginx, Traefik, Contour)

---

### Side by Side

| Feature | L4 (NLB) | L7 (ALB) |
|---------|----------|----------|
| Routing by | IP + port | URL path, host, headers, cookies |
| SSL termination | ❌ Passes through | ✅ Terminates at LB |
| Protocol support | Any TCP/UDP | HTTP, HTTPS, gRPC, WebSocket |
| Latency overhead | Minimal | Higher (parsing + new connection) |
| Auth at LB | ❌ No | ✅ Yes (Cognito, Lambda authorizer) |
| Sticky sessions | IP-based only | Cookie-based |
| Speed | Faster | Slower |
| Use case | Database, Redis, custom protocols | REST APIs, web apps, microservices |

---

### Health Checks at the Load Balancer

Both L4 and L7 LBs probe backends to detect unhealthy instances:

**L4 health check:** TCP connect to `backend:8080` — if it accepts the connection, it's healthy. Doesn't care if the app is actually working.

**L7 health check:** HTTP GET to `/health` — checks the response is `200`. Your `handleHealth` endpoint from today is exactly what an L7 health check calls.

This is why every service you build needs a `/health` endpoint. ALB, Kubernetes liveness/readiness probes, and Consul all call it.

---

### In Kubernetes

```
External traffic
       ↓
  Ingress (L7) — path routing, TLS, host-based routing
       ↓
  Service (L4) — ClusterIP, load balances across pods
       ↓
  Pod — your Go HTTP server on :8080
```

The Ingress is your L7 LB. The Service is your L4 LB. They stack — L7 for smart routing, L4 for pod-level distribution.

---

## Key Takeaways

1. An HTTP server needs two things: `http.HandleFunc` to register routes and `http.ListenAndServe` to start
2. Handler signature is always `func(w http.ResponseWriter, r *http.Request)` — exact match required
3. `w.WriteHeader(code)` sets status; `fmt.Fprintf(w, body)` writes the body
4. `http.ResponseWriter` satisfies `io.Writer` — same `fmt.Fprintf` pattern as file writing
5. The router (`"/"` catch-all) matches any path not claimed by a more specific route
6. Routing belongs in `main`, not in handler functions — one handler, one responsibility
7. `log.Fatal(http.ListenAndServe(...))` — always handle the error; a port conflict fails silently otherwise
8. The request object `r` only exists inside handlers — it's created per request, not available in `main`
9. L4 LB routes by IP + port only — fast, protocol-agnostic, no SSL termination
10. L7 LB terminates TLS, reads HTTP — path routing, auth, header manipulation
11. NLB = L4 in AWS; ALB = L7 in AWS; Service = L4 in k8s; Ingress = L7 in k8s
12. Every service needs a `/health` endpoint — ALB, k8s probes, and Consul all call it
13. In Kubernetes: Ingress (L7) → Service (L4) → Pod — they stack

---

> **As-Sami' — السَّمِيع — The All-Hearing**
>
> _He hears every request — nothing goes unanswered. Today your server did the same: every curl found its handler, every path found its response. See you on Day 15._

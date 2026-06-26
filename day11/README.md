# Day 11 — Packages & Modules in Go

---

> **بِسْمِ اللهِ الرَّحْمَنِ الرَّحِيم**
>
> **Al-Wahhab — الوَهَّاب — The Bestower**
>
> _He gives without limit and without being asked. Knowledge is His gift — He placed it in the concepts you wrestled with today and the code that finally ran. Begin with His name._

---

## Blog of the Day

[Organizing a Go module — go.dev/doc](https://go.dev/doc/modules/layout)

Read this after the session. It shows the standard layouts for Go modules — single package, multiple packages, and when to split. The patterns used in real SRE tooling (Prometheus, k8s client-go) follow these layouts.

---

## Concept 1: Packages

Every `.go` file begins with `package <name>`. A **package** is all the `.go` files in one directory sharing the same package name. It is Go's unit of code organisation.

```
day11/
├── main.go               ← package main
└── healthcheck/
    └── healthcheck.go    ← package healthcheck
```

You've been writing `package main` every day. That's a special package — it produces an executable binary. All other packages are **library packages** imported and used by other code.

---

## Concept 2: Exported vs Unexported — The Capital Letter Rule

This is the single most important rule in Go packages:

> **Capital first letter = exported** (visible outside the package — public)
> **Lowercase first letter = unexported** (only visible inside the package — private)

No `public` or `private` keywords. Just the case of the first letter. The compiler enforces this at compile time.

```go
// inside package healthcheck

type Server struct { ... }        // ✅ Exported — other packages can use healthcheck.Server{}
type config struct { ... }        // ❌ Unexported — only healthcheck.go can use it

func Check(s Server) string { }   // ✅ Exported — other packages can call healthcheck.Check()
func validate(s Server) bool { }  // ❌ Unexported — only healthcheck.go can call it
```

### The prefix rule — package name as qualifier

When you import a package, you access its exported identifiers using the **package name as a prefix**:

```go
import "day11/healthcheck"

// package name "healthcheck" is the prefix
healthcheck.Server{}     // using the exported struct
healthcheck.Check(s)     // calling the exported function
```

If `Server` were written as `server` (lowercase), `main.go` would get a compile error:
```
cannot refer to unexported name healthcheck.server
```

This forces a clean boundary: the package decides what it exposes. Callers only see the public API.

### SRE analogy

Think of a package like a microservice:
- **Exported** = the API endpoints your service exposes (`/check`, `/status`)
- **Unexported** = the internal database queries, helper functions, config structs that callers never need to know about

---

## Concept 3: Modules

A **module** is one or more packages with a `go.mod` file at the root. It declares:
- The **module path** — the unique name used in import statements
- The **Go version**
- External dependencies (none today)

```
module day11    ← module path

go 1.26.2
```

When `main.go` imports the `healthcheck` package:

```go
import "day11/healthcheck"
```

Go reads: "find the module named `day11`, look inside the `healthcheck/` subdirectory."

The module path is the root. Subdirectories extend it. This is how all Go imports work — including the standard library (`"fmt"` is a package in the Go standard module).

---

## Concept 4: `fmt.Sprintf` vs `fmt.Printf` vs `fmt.Println`

This trips up almost everyone. Three functions, very similar names, completely different behaviour:

### `fmt.Sprintf` — builds and **returns** a string

```go
result := fmt.Sprintf("%s: CRITICAL — CPU %.1f", s.Name, s.CPUUsage)
// result = "web-02: CRITICAL — CPU 95.5"
// nothing is printed — you get a string back
```

**When to use:** inside a function that **returns** a string. You want to build the string and hand it back to the caller.

```go
func Check(s Server) string {
    return fmt.Sprintf("%s: OK", s.Name)   // ✅ right tool — function returns string
}
```

### `fmt.Printf` — builds and **prints** to stdout, returns `(int, error)`

```go
fmt.Printf("%s: CRITICAL — CPU %.1f\n", s.Name, s.CPUUsage)
// prints immediately — returns (bytes written, error)
// you get nothing useful back
```

**When to use:** when you want to print with formatting and control the exact output (including `\n` for newlines yourself).

```go
// ❌ Wrong — trying to return the result of Printf
func Check(s Server) string {
    return fmt.Printf(...)   // compile error — Printf returns (int, error), not string
}
```

### `fmt.Println` — prints with a newline, no format verbs

```go
fmt.Println("web-01: OK")
fmt.Println(healthcheck.Check(s1))   // prints whatever Check returns
```

**When to use:** simple output with no formatting needed. Automatically adds a newline. Takes any value — no format string required.

---

### Quick decision guide

| I want to... | Use |
|-------------|-----|
| Build a string to return from a function | `fmt.Sprintf` |
| Print a formatted string with variables | `fmt.Printf` |
| Print a simple value or string | `fmt.Println` |
| Print and also capture the string | `fmt.Sprintf` then `fmt.Println` |

---

## Format Verbs Reference

| Verb | Type | Example output |
|------|------|---------------|
| `%s` | string | `web-01` |
| `%d` | integer | `5432` |
| `%f` | float (all decimals) | `95.500000` |
| `%.1f` | float, 1 decimal place | `95.5` |
| `%v` | any type, default format | works for everything |
| `%T` | prints the type itself | `healthcheck.Server` |

---

## Project Structure

```
day11/
├── go.mod                        ← module definition (module day11)
├── main.go                       ← package main — entry point, wires everything
└── healthcheck/
    └── healthcheck.go            ← package healthcheck — exported API
```

`main.go` only imports and calls. `healthcheck.go` only defines logic. They never share a directory — that's the separation.

---

## Mistakes Made Today

### Mistake 1 — Missing space in format string

```go
// ❌ Wrong — output: "web-01:OK" (no space)
return fmt.Sprintf("%s:OK", s.Name)
```

```go
// ✅ Correct — output: "web-01: OK"
return fmt.Sprintf("%s: OK", s.Name)
```

Small but matters — inconsistent formatting in alert output makes log parsing and grep harder.

---

### Mistake 2 — Mixed case in severity level

```go
// ❌ Wrong — "Critical" looks unprofessional in SRE tooling
return fmt.Sprintf("%s: Critical - CPU %.1f", s.Name, s.CPUUsage)
```

```go
// ✅ Correct — uppercase severity is the convention (CRITICAL, WARNING, OK)
return fmt.Sprintf("%s: CRITICAL - CPU %.1f", s.Name, s.CPUUsage)
```

Alertmanager, PagerDuty, and Prometheus all use uppercase severity labels. Consistency matters when you're parsing or routing alerts programmatically.

---

## Final Code

**`healthcheck/healthcheck.go`**
```go
package healthcheck

import "fmt"

type Server struct {
	Name     string
	CPUUsage float64
}

func Check(s Server) string {
	if s.CPUUsage > 90 {
		return fmt.Sprintf("%s: CRITICAL - CPU %.1f", s.Name, s.CPUUsage)
	}
	return fmt.Sprintf("%s: OK", s.Name)
}
```

**`main.go`**
```go
package main

import (
	"day11/healthcheck"
	"fmt"
)

func main() {
	s1 := healthcheck.Server{Name: "web-01", CPUUsage: 45.0}
	s2 := healthcheck.Server{Name: "web-02", CPUUsage: 95.5}
	fmt.Println(healthcheck.Check(s1))
	fmt.Println(healthcheck.Check(s2))
}
```

Output:
```
web-01: OK
web-02: CRITICAL - CPU 95.5
```

---

## System Design: API Gateway — Single Entry Point Pattern

### The Problem Without a Gateway

Imagine you have five microservices — alerting, metrics, config, auth, incidents. Every client (web app, mobile, CLI tool) calls each service directly:

```
Client → alerting-service:8001
Client → metrics-service:8002
Client → config-service:8003
Client → auth-service:8004
```

Every service must now implement:
- Authentication — verify the token on every request
- Rate limiting — protect itself from abuse
- TLS — terminate HTTPS itself
- Logging — track who called what
- CORS — handle browser security headers

That's five teams duplicating the same cross-cutting concerns. When the auth logic changes, all five services update. When you want to see all traffic in one place — you can't.

---

### The API Gateway Pattern

A single entry point sits in front of all services:

```
Client → API Gateway → alerting-service
                    → metrics-service
                    → config-service
                    → auth-service
```

The gateway handles all the cross-cutting concerns **once**:

| Concern | Without gateway | With gateway |
|---------|----------------|-------------|
| Auth | Every service validates tokens | Gateway validates, backends trust |
| Rate limiting | Every service implements it | One place, one policy |
| TLS termination | Every service manages certs | Gateway only |
| Logging | Scattered across services | Single access log |
| Routing | Clients know every service URL | Clients know one URL |

---

### What a Gateway Does — The Request Lifecycle

```
1. Client sends: GET https://api.company.com/alerts
2. Gateway receives the request
3. Gateway checks: is this path rate-limited? → No, allow
4. Gateway checks: is there a valid auth token? → Yes, pass
5. Gateway routes: /alerts → alerting-service:8001
6. alerting-service responds
7. Gateway may transform the response (strip internal headers, add correlation ID)
8. Gateway returns response to client
```

All of this happens before the backend ever sees the request.

---

### AWS ALB as an API Gateway

AWS ALB (Application Load Balancer) does the core of this:

```
Listener rule 1: path /api/alerts*  → Target Group: alerting-service
Listener rule 2: path /api/metrics* → Target Group: metrics-service
Listener rule 3: path /api/config*  → Target Group: config-service
Default:         → 404
```

Add a Cognito authorizer or Lambda authorizer → auth at the gateway.
Add WAF → rate limiting and IP filtering at the gateway.
Enable access logs → single place for all traffic.

ALB + WAF + Cognito = a fully capable API Gateway, without running your own software.

---

### Dedicated API Gateways

When ALB isn't enough — more complex transformations, multi-cloud, plugin ecosystem:

| Gateway | Used for |
|---------|---------|
| **Kong** | Plugin-based, on-prem or cloud, Lua/Go plugins |
| **Envoy** | Service mesh sidecar + edge gateway, used by Istio |
| **AWS API Gateway** | Fully managed, serverless-native, Lambda integration |
| **Nginx** | Simple reverse proxy + gateway for smaller setups |
| **Traefik** | Kubernetes-native, auto-discovers services via labels |

In Kubernetes — **Ingress + Ingress Controller** (Nginx, Traefik, Contour) is the API Gateway. The Ingress resource defines routing rules; the controller is the actual gateway.

---

### When to Use an API Gateway

| Situation | Use gateway? |
|-----------|-------------|
| Multiple backend services, one public URL | ✅ Yes |
| Auth must be consistent across all services | ✅ Yes |
| Rate limiting per client API key | ✅ Yes |
| Single service, simple setup | ❌ Overkill |
| Service-to-service internal calls | ❌ Use service mesh instead |

The gateway is for **north-south traffic** (client to cluster). For **east-west traffic** (service to service inside the cluster), use a service mesh (Istio, Linkerd) or direct calls.

---

## Key Takeaways

1. A package is all `.go` files in one directory sharing the same `package` name
2. Capital first letter = exported (public); lowercase = unexported (private) — enforced by the compiler
3. Exported identifiers are accessed with the package name as a prefix: `healthcheck.Server{}`
4. A module is one or more packages + a `go.mod` file that declares the module path
5. Import path = module path + subdirectory: `"day11/healthcheck"`
6. `fmt.Sprintf` — builds and **returns** a string; use inside functions that return string
7. `fmt.Printf` — builds and **prints**; returns `(int, error)`, not a string
8. `fmt.Println` — prints any value with a newline; no format verbs needed
9. Use `%s` for strings, `%d` for ints, `%.1f` for floats to one decimal place
10. An API Gateway is a single entry point — auth, rate limiting, routing, TLS in one place
11. Without a gateway, every service duplicates cross-cutting concerns
12. AWS ALB + WAF + Cognito = a production API Gateway without running your own software
13. Gateway handles north-south traffic (client → cluster); service mesh handles east-west (service → service)

---

> **Al-Mujib — المُجِيب — The Responsive, The Answerer**
>
> _Every call finds its answer with Him — no request goes unheard. Today you built a system where every call finds its right handler: the gateway routes, the package exports, the function responds. See you on Day 12._

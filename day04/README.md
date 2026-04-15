# Day 04 — Structs in Go

---

> **بِسْمِ اللهِ الرَّحْمَنِ الرَّحِيم**
>
> **Al-'Aleem — العَلِيم — The All-Knowing**
>
> _Allah knows everything you are learning, every struggle, every line of code. Nothing is wasted. Begin with His name and trust the process._

---

## Blog of the Day

[JSON and Go — The Go Blog](https://go.dev/blog/json-and-go)

Read this after completing the session. It covers how Go handles JSON encoding/decoding — directly relevant to structs since JSON maps onto struct fields. Preview of Day 15.

---

## What is a Struct?

A struct is a collection of named fields grouped under one type. It's how Go models real-world data.

As an SRE, you deal with this constantly — a server has a hostname, an IP, a region, a status. A struct lets you group all of that into one named type instead of passing loose variables around.

```go
type Server struct {
    Hostname string
    IP       string
    Region   string
    Healthy  bool
}
```

Once defined, you create instances (values) of that struct:

```go
s := Server{Hostname: "web-01", IP: "10.0.0.1", Region: "us-east-1", Healthy: true}
fmt.Println(s.Hostname)  // web-01
```

---

## Methods — Attaching Behaviour to a Struct

A method is a function that belongs to a type. You attach it using a **receiver**.

```go
func (s Server) Status() string {
    // s is the specific Server instance this method was called on
}
```

- `s` is the receiver variable — it gives you access to the struct's fields inside the method
- `Server` is the type the method belongs to
- This is NOT the same as `func Server.Status()` — that syntax doesn't exist in Go

### Why this matters

Define formatting or behaviour once on the struct, reuse it everywhere:

```go
fmt.Println(server.Status())  // same call works for any Server
```

---

## Mistake Made: Wrong Method Syntax

```go
// ❌ Wrong — this is not valid Go
func Server.Status() string {
    if Server.Healthy == true {
```

```go
// ✅ Correct — receiver goes before the method name
func (s Server) Status() string {
    if s.Healthy {
```

The receiver `(s Server)` is what makes this a method. Inside the method, you use `s` (the instance), not `Server` (the type).

---

## Boolean Idiom

```go
// ❌ Redundant
if s.Healthy == true {

// ✅ Idiomatic Go
if s.Healthy {
```

Booleans don't need `== true`. The condition already is true/false.

---

## Slice of Structs

```go
servers := []Server{
    {Hostname: "web-01", IP: "10.0.0.1", Region: "us-east-1", Healthy: true},
    {Hostname: "web-02", IP: "10.0.0.2", Region: "us-east-1", Healthy: false},
    {Hostname: "web-03", IP: "10.0.0.3", Region: "us-east-2", Healthy: false},
}
```

- `[]Server` — a slice where each element is a `Server` struct
- `Server{...}` — creates a single struct value
- `[]Server{...}` — creates a collection of struct values

### Mistake Made: Confusing the two

```go
// ❌ Wrong — Server{} creates one struct, not a slice
Servers := Server{
    {Hostname: "web-01", ...},
}

// ✅ Correct — []Server{} creates a slice of structs
servers := []Server{
    {Hostname: "web-01", ...},
}
```

---

## Package-level vs Function-level Variables

```go
// ❌ Wrong — := cannot be used outside a function
serversHealth := []Server{}

// ✅ Correct — declare inside the function where it's used
func FilterUnhealthy(servers []Server) []Server {
    serversHealth := []Server{}
    ...
}
```

Rule: `:=` only works inside functions. Package-level variables use `var`.
Keep variables as close to where they're used as possible.

---

## append — Must Assign the Result

```go
// ❌ Wrong — append doesn't modify in place
append(serversHealth, server)

// ✅ Correct — capture the returned slice
serversHealth = append(serversHealth, server)
```

`append` returns a new slice. If you don't assign it, nothing happens.

---

## Return Type vs Return Value

```go
// ❌ Wrong — Server is a type, not a value
return Server

// ✅ Correct — serversHealth is the variable holding the result
return serversHealth
```

The function signature says what *type* to return. The `return` statement returns the actual *value*.

---

## Exported vs Unexported Names

In Go, capitalisation controls visibility:

| Name | Meaning |
|------|---------|
| `Server` | Exported — visible outside the package |
| `servers` | Unexported — local to this package |

Local variables should be lowercase. Uppercase is for types, functions, and fields you want to expose publicly.

---

## Final Code

```go
package main

import "fmt"

type Server struct {
	Hostname string
	IP       string
	Region   string
	Healthy  bool
}

func (s Server) Status() string {
	serverHealth := "Unhealthy"
	if s.Healthy {
		serverHealth = "Healthy"
	}
	return fmt.Sprintf("The %s [%s] - %s", s.Hostname, s.Region, serverHealth)
}

func FilterUnhealthy(servers []Server) []Server {
	serversHealth := []Server{}
	for _, server := range servers {
		if !server.Healthy {
			serversHealth = append(serversHealth, server)
		}
	}
	return serversHealth
}

func main() {
	servers := []Server{
		{Hostname: "web-01", IP: "10.0.0.1", Region: "us-east-1", Healthy: true},
		{Hostname: "web-02", IP: "10.0.0.2", Region: "us-east-1", Healthy: false},
		{Hostname: "web-03", IP: "10.0.0.3", Region: "us-east-2", Healthy: false},
	}
	for _, server := range servers {
		fmt.Println(server.Status())
	}
	fmt.Println("Printing only Unhealthy now")
	unhealthyServers := FilterUnhealthy(servers)
	for _, server := range unhealthyServers {
		fmt.Println(server.Status())
	}
}
```

---

---

## System Design: DNS — How the Internet Resolves Names

You type `google.com` in a browser. Your computer doesn't know where that is. DNS translates that name into an IP address like `142.250.80.46`.

Think of it as the phone book of the internet.

---

### The Lookup Chain

```
Your Browser
    → DNS Resolver (your ISP or 8.8.8.8)
        → Root Nameserver (who handles .com?)
            → TLD Nameserver (who handles google.com?)
                → Authoritative Nameserver (what's the IP for google.com?)
                    → returns 142.250.80.46
```

**1. DNS Resolver** — your first stop. Usually your ISP or a public resolver (Cloudflare `1.1.1.1`, Google `8.8.8.8`). Checks its cache first.

**2. Root Nameserver** — knows nothing about `google.com` but knows who handles `.com`. 13 root server clusters globally.

**3. TLD Nameserver** — handles `.com`, `.io`, `.org`. Knows who the authoritative nameserver is for `google.com`.

**4. Authoritative Nameserver** — the source of truth. Owned by the domain owner. Returns the actual IP.

---

### Caching & TTL

Every DNS response includes a **TTL (Time To Live)** — how long to cache the answer in seconds.

- High TTL = fewer lookups, faster, but slow to propagate changes
- Low TTL = frequent lookups, but changes take effect quickly

---

### DNS Record Types

| Record | Purpose | Example |
|--------|---------|---------|
| `A` | Maps hostname → IPv4 | `web-01 → 10.0.0.1` |
| `AAAA` | Maps hostname → IPv6 | |
| `CNAME` | Alias to another hostname | `www → google.com` |
| `MX` | Mail server for a domain | |
| `TXT` | Arbitrary text, used for verification | |
| `NS` | Which nameservers are authoritative | |

---

### TTL & Incident Response

**The problem:** TTL is 24 hours. You need to failover `api.yourcompany.com` to a new IP immediately. Clients who cached it won't see the change for up to 24 hours.

**The fix (proactive):**
1. Keep TTL low during normal operations — 300s (5 min) is common for critical records
2. Before a planned failover, drop TTL to 60s
3. Wait for the old TTL to expire so all clients pick up the new low TTL
4. Then change the IP — propagates in 60s

**The catch:** Lowering the TTL only helps *after* the old TTL expires. If TTL is 24h and you lower it now, clients who cached it an hour ago still have 23 hours left. You have to wait out the old TTL first.

**Rule:** Lower TTL before you need it, not during the crisis.

---

### Route 53 Health Checks — Bypassing TTL Entirely

Route 53 failover routing + health checks solve the TTL problem:
- The resolver keeps checking your endpoint health
- When the health check fails, Route 53 stops returning that IP immediately
- No waiting on TTL — DNS-level failover without the cache problem

---

### SRE Relevance

- **Route 53, Cloud DNS** — managed authoritative DNS with health-check-based routing
- **Split-horizon DNS** — same name resolves to different IPs inside vs outside (e.g., inside k8s vs public)
- **DNS-based load balancing** — return multiple IPs, client picks one
- **Service discovery in k8s** — `my-service.my-namespace.svc.cluster.local` is just DNS

---

## Key Takeaways

1. Structs group related fields under one named type
2. Methods attach behaviour to a struct using a receiver `(s Server)`
3. Use the receiver variable (`s`), not the type name (`Server`), inside methods
4. `if s.Healthy` is idiomatic — no need for `== true`
5. `[]Server{}` is a slice of structs; `Server{}` is a single struct
6. `:=` only works inside functions — package-level needs `var`
7. `append` returns a new slice — always assign it back
8. `return` needs a value, not a type name
9. Lowercase variable names for locals; uppercase for exported identifiers
10. Define methods on structs so formatting/behaviour is defined once and reused everywhere

---

> **Al-Fattah — الفَتَّاح — The Opener**
>
> _He opens doors of understanding for those who seek. You showed up today, you wrote the code, you debugged it yourself. That is the way. See you on Day 05._

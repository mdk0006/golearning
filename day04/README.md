# Day 04 — Structs in Go

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

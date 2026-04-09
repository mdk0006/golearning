# Day 01 — Variables, Types & Zero Values

## Concepts Learned

### 1. Variable Declaration — 3 Ways

```go
// 1. Explicit type
var port int = 8080

// 2. Type inferred by Go
var host = "prometheus.internal"

// 3. Short declaration (most common inside functions)
cpuUsage := 99.5
```

> Use `:=` inside functions. Use `var` at package level (outside functions).

---

### 2. Core Types & Zero Values

| Type | Example | Zero Value |
|------|---------|------------|
| `int` | `8080` | `0` |
| `string` | `"node-01"` | `""` |
| `bool` | `true` | `false` |
| `float64` | `99.5` | `0` |

**Zero value** = what Go assigns automatically if you declare a variable without assigning a value.  
No garbage, no null, always predictable.

```go
var unhealthy bool    // false
var errorCount int    // 0
var errorMsg string   // ""
```

---

### 3. Naming Convention

| Language | Style | Example |
|----------|-------|---------|
| Python | snake_case | `cpu_usage` |
| Go | camelCase | `cpuUsage` |

Always use `camelCase` in Go.

---

### 4. Format Verbs (fmt.Printf)

| Verb | Output |
|------|--------|
| `%v` | Default format of the value |
| `%q` | String wrapped in double quotes — useful to see empty strings |
| `%d` | Integer |
| `%f` | Float |
| `%T` | Type of the variable |

```go
fmt.Printf("%q\n", errorMsg)    // prints: ""
fmt.Printf("%T\n", cpuUsage)   // prints: float64
```

---

## Code Written

[main.go](main.go)

```go
package main

import "fmt"

var port int = 8080
var host = "prometheus.internal"

func main() {
	cpuUsage := 99.5
	healthy := true

	fmt.Printf("The cpu usage is %v\n", cpuUsage)
	fmt.Printf("If cpu is healthy: %v\n", healthy)
	fmt.Printf("The port for prometheus is %v\n", port)
	fmt.Printf("The host name is %v\n", host)

	// Zero values
	var unhealthy bool
	var errorCount int
	var errorMsg string

	fmt.Printf("unhealthy: %v, errorCount: %v, errorMsg: %q\n", unhealthy, errorCount, errorMsg)

	// Type inspection
	fmt.Printf("type of cpuUsage: %T\n", cpuUsage)
}
```

---

## Key Takeaways

- Go initializes every variable — no uninitialized memory surprises
- Use `:=` inside functions, `var` at package level
- `camelCase` not `snake_case`
- `%q` is your friend when debugging empty strings in logs

---

## Next — Day 02
Functions, multiple return values, named returns, and error handling basics.

# Day 02 — Functions in Go

## What is a Function?

A function is a named, reusable block of code that takes inputs and returns outputs.
In Go, functions are the primary unit of logic organization — there are no classes.

---

## Basic Syntax

```go
func functionName(param type) returnType {
    return value
}
```

---

## Multiple Return Values

Go functions can return more than one value. This is used everywhere in Go.

```go
func checkHealth(host string, threshold, cpuUsage float64) (string, bool) {
    ...
    return message, isHealthy
}
```

- Wrap multiple return types in parentheses `(string, bool)`
- Return them comma-separated: `return msg, true`
- Caller captures both: `status, healthy := checkHealth(...)`

---

## Idiomatic Go — Return Early

Instead of declaring variables upfront and returning at the end, return directly from each branch.

**Not idiomatic:**
```go
var msg string
var isHealthy bool
if cpuUsage > threshold {
    msg = "unhealthy"
    isHealthy = false
} else {
    msg = "healthy"
    isHealthy = true
}
return msg, isHealthy
```

**Idiomatic Go:**
```go
if cpuUsage > threshold {
    return fmt.Sprintf("%s is unhealthy, cpu: %.1f", host, cpuUsage), false
}
return fmt.Sprintf("%s is healthy, cpu: %.1f", host, cpuUsage), true
```

No `else` needed — if the `if` fires, it returns and exits. Otherwise execution falls through to the next `return`.

---

## Discarding a Return Value

Use `_` to discard a return value you don't need:

```go
status, _ := checkHealth(host, threshold, cpuUsage)
```

Go will not compile if you declare a variable and never use it. `_` is the idiomatic way to explicitly ignore a value.

---

## Naming Convention

Go uses `camelCase`, not `snake_case`.

| Wrong | Correct |
|-------|---------|
| `is_healthy` | `isHealthy` |
| `cpu_usage` | `cpuUsage` |
| `error_count` | `errorCount` |

---

## fmt.Sprintf

`Sprintf` = **S**tring printf — builds and **returns** a formatted string instead of printing it.

```go
fmt.Printf("cpu: %.1f", 90.1)          // prints directly
msg := fmt.Sprintf("cpu: %.1f", 90.1)  // returns "cpu: 90.1"
```

### Format Verbs

| Verb | Meaning | Output |
|------|---------|--------|
| `%s` | string | `prometheus` |
| `%v` | any value, default format | `true`, `8080` |
| `%d` | integer | `42` |
| `%f` | float | `90.100000` |
| `%.1f` | float, 1 decimal place | `90.1` |

### When to use which

| Function | Use when |
|----------|----------|
| `fmt.Sprintf` | Building a string to store or return |
| `fmt.Printf` | Printing with formatting directly |
| `fmt.Println` | Printing without formatting |

---

## Final Code

```go
package main

import "fmt"

func checkHealth(host string, threshold, cpuUsage float64) (string, bool) {
	if cpuUsage > threshold {
		return fmt.Sprintf("%s is unhealthy, cpu: %.1f", host, cpuUsage), false
	}
	return fmt.Sprintf("%s is healthy, cpu: %.1f", host, cpuUsage), true
}

func main() {
	threshold := 89.0
	cpuUsage := 90.1
	host := "prometheus"
	status, _ := checkHealth(host, threshold, cpuUsage)
	fmt.Println(status)
}
```

---

## Key Takeaways

1. Go functions can return multiple values — use `(type1, type2)` syntax
2. Return early from branches instead of pre-declaring variables
3. Use `_` to discard return values you don't need
4. Go won't compile if a variable is declared but unused
5. Use `camelCase` for all variable and function names
6. `gofmt -w main.go` auto-formats your code — always run it

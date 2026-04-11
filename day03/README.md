# Day 03 — Control Flow in Go

## What is Control Flow?

Control flow is how your program decides what to do and how many times to do it.
Go has three tools: `if/else`, `for`, and `switch`.

---

## 1. if / else

### Basic syntax
```go
if condition {
    // do this
} else {
    // do that
}
```

### if with initialization (Go-specific)
```go
if err := doSomething(); err != nil {
    return err
}
```

Two statements separated by `;`:
1. `err := doSomething()` — runs first, stores the result
2. `err != nil` — the actual condition

The variable `err` only exists inside the `if` block. Once the block ends it's gone.
This keeps scope tight — the Go way.

**Rule of thumb:** Use `if init; condition` when the variable is only needed for that one check.

---

## 2. for — The Only Loop in Go

Go has **no `while`, no `do-while`**. `for` does everything.

### Classic loop
```go
for i := 0; i < 5; i++ {
    fmt.Println(i)
}
```

### Acts like while
```go
for condition {
    // runs until condition is false
}
```

### Infinite loop
```go
for {
    // runs forever, use break to exit
}
```

### Range loop — iterate over a slice
```go
servers := []string{"node-01", "node-02", "node-03"}

for i, server := range servers {
    fmt.Println(i, server)
}
```

- `i` — the index (0, 1, 2...)
- `server` — the value at that index
- Use `_` to discard what you don't need:

```go
for _, server := range servers {  // don't need the index
    fmt.Println(server)
}

for i := range servers {  // don't need the value
    fmt.Println(i)
}
```

**Rule of thumb:** Use `range` for slices and maps. Use classic `for` only when you need precise index control.

---

## 3. switch

Cleaner than a chain of `if/else if`. No `break` needed — Go does **not** fall through by default.

### Basic syntax
```go
switch status {
case "healthy":
    fmt.Println("all good")
case "degraded":
    fmt.Println("watch out")
case "down":
    fmt.Println("alert!")
default:
    fmt.Println("unknown status")
}
```

### Always add a `default` case
If an unexpected value comes in, `default` handles it. Without it, unknown values are silently ignored — hard to debug.

**Rule of thumb:** Treat `default` like the `else` — always include it as a safety net.

### switch vs if/else chain

**Use switch when:** matching one variable against multiple known values
**Use if/else when:** conditions are complex or involve multiple variables

```go
// prefer switch
switch status {
case "healthy": ...
case "down": ...
}

// prefer if/else
if cpuUsage > 90 && memUsage > 80 {
    ...
}
```

---

## 4. Where to Declare Variables — Rule of Thumb

| Situation | Where |
|-----------|-------|
| Used only in one function | Inside that function with `:=` |
| Used by multiple functions | Package-level `var` |
| Fixed/config-like data | Package-level, named clearly |

Keep scope as tight as possible. Declare variables as close to where they're used as you can.

---

## Version 1 — Everything in main

```go
package main

import "fmt"

func main() {
    statuses := []string{"healthy", "degraded", "down"}
    servers := []string{"node-01", "node-02", "node-03"}

    for i, server := range servers {
        switch statuses[i] {
        case "healthy":
            fmt.Printf("The %s is healthy\n", server)
        case "degraded":
            fmt.Printf("The %s is degraded\n", server)
        case "down":
            fmt.Printf("The %s is down\n", server)
        default:
            fmt.Printf("The %s has unknown status\n", server)
        }
    }
}
```

Works, but the loop is doing too much — it's both iterating and handling logic.

---

## Version 2 — Idiomatic Go (logic in its own function)

```go
package main

import "fmt"

func printStatus(server, status string) {
    switch status {
    case "healthy":
        fmt.Printf("The %s is healthy\n", server)
    case "degraded":
        fmt.Printf("The %s is degraded\n", server)
    case "down":
        fmt.Printf("The %s is down\n", server)
    default:
        fmt.Printf("The %s has unknown status\n", server)
    }
}

func main() {
    statuses := []string{"healthy", "degraded", "down"}
    servers := []string{"node-01", "node-02", "node-03"}

    for i, server := range servers {
        printStatus(server, statuses[i])
    }
}
```

**Why this is better:**
- Loop stays thin — just calls a function
- `printStatus` can be tested independently
- Easy to add more servers without touching the logic

**Rule of thumb:** If the body of a loop is more than 2-3 lines, extract it into a function.

---

## Shared Parameter Type Shorthand

When two consecutive parameters have the same type, you can write:

```go
func printStatus(server, status string) {  // both are string
```

Instead of:
```go
func printStatus(server string, status string) {
```

Both are valid. The shorthand is idiomatic when types match.

---

## Key Takeaways

1. `for` is the only loop in Go — it replaces `while` and `do-while`
2. Use `range` to iterate over slices and maps
3. `switch` needs no `break` — Go doesn't fall through by default
4. Always add a `default` case to your switch
5. Use `if init; condition` to keep variable scope tight
6. Keep loops thin — extract logic into functions
7. Declare variables as close to their use as possible
8. Use `_` to discard values you don't need from `range`

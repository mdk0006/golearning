# Day 12 — File I/O in Go

---

> **بِسْمِ اللهِ الرَّحْمَنِ الرَّحِيم**
>
> **Al-Raqib — الرَّقِيب — The Watchful, The Ever-Observant**
>
> _He watches over everything — every file written, every line read, every error ignored. Build systems that are equally watchful — never ignore a failure, always check what comes back. Begin with His name._

---

## Blog of the Day

[Microservices — Martin Fowler](https://martinfowler.com/articles/microservices.html)

Read this after the session. The original article that defined the microservices pattern. Martin Fowler explains the trade-offs clearly — this isn't a hype piece, it's an honest assessment of when microservices help and when they hurt.

---

## Concept 1: The Two Packages for File I/O

Go's standard library splits file work across two packages:

| Package | Purpose |
|---------|---------|
| `os` | Low-level OS interaction — open, create, delete files |
| `bufio` | Buffered reading — read a file efficiently line by line |

You almost always use both together.

---

## Concept 2: Opening vs Creating

```go
file, err := os.Open("servers.txt")    // open for READING only
report, err := os.Create("report.txt") // create for WRITING — wipes existing content
```

- `os.Open` — read only. Fails if file doesn't exist.
- `os.Create` — write only. Creates the file if it doesn't exist. **Wipes it if it does.**

Both return `(file, error)`. Always check the error before using the file.

---

## Concept 3: Always Check File Errors

```go
file, err := os.Open("servers.txt")
if err != nil {
    fmt.Println("error opening file:", err)
    return
}
```

If `os.Open` fails, `file` is `nil`. Passing a `nil` file to `bufio.NewScanner` will panic. The error check is not optional — it's how you find out **why** something failed (file not found, permission denied, disk full).

Same pattern for `os.Create`:

```go
report, err := os.Create("report.txt")
if err != nil {
    fmt.Println("error creating report:", err)
    return
}
```

`os.Create` can fail if the directory doesn't exist, you lack write permission, or the disk is full. Without the check, you'd write to a nil file handle — silent failure, empty output, no clue why.

---

## Concept 4: `defer file.Close()`

```go
file, err := os.Open("servers.txt")
if err != nil { ... }
defer file.Close()   // ← write this immediately after a successful open
```

`defer` schedules a function call to run **when the surrounding function returns** — no matter how it exits (normal return, early return on error, even panic).

Without `defer file.Close()`:
- The OS file descriptor leaks until the program exits
- In long-running services, running out of file descriptors causes cryptic errors

**Rule:** write `defer file.Close()` on the very next line after every successful open. Never wait.

When multiple defers are queued, they run in **reverse order** (LIFO — last in, first out):

```go
defer file.Close()    // queued first — runs second
defer report.Close()  // queued second — runs first
```

---

## Concept 5: Reading Line by Line with `bufio.Scanner`

```go
scanner := bufio.NewScanner(file)
for scanner.Scan() {
    line := scanner.Text()
    // process line
}
```

- `bufio.NewScanner(file)` — wraps the file in a scanner with an internal buffer
- `scanner.Scan()` — advances to the next line, returns `true` if content exists, `false` at EOF
- `scanner.Text()` — returns the current line as a `string`, newline stripped

**Why not `os.ReadFile`?** `os.ReadFile` reads the entire file into memory at once. Fine for small config files. For a file with 100,000 server names, `bufio.Scanner` reads one line at a time — constant memory usage regardless of file size.

---

## Concept 6: Writing to a File with `fmt.Fprintf`

```go
fmt.Fprintf(report, "%s\n", serverHealth(line))
```

`fmt.Fprintf` is `fmt.Printf` but writes to any `io.Writer` instead of stdout. A file satisfies `io.Writer`, so you pass it as the first argument.

| Function | Writes to | Returns |
|----------|-----------|---------|
| `fmt.Println(...)` | stdout | `(int, error)` |
| `fmt.Printf(...)` | stdout | `(int, error)` |
| `fmt.Fprintf(w, ...)` | any `io.Writer` (file, buffer, network) | `(int, error)` |
| `fmt.Sprintf(...)` | returns a string | `string` |

The `\n` adds the newline manually — `scanner.Text()` strips the newline when reading, so you must add it back when writing.

---

## Line by Line Walkthrough

```go
package main
```
Executable program — Go calls `main()` on startup.

---

```go
import ("bufio" "fmt" "os")
```
- `os` — file open/create
- `bufio` — line-by-line scanner
- `fmt` — Fprintf to write to file, Println to stdout

---

```go
func serverHealth(name string) string {
    return name + ": OK"
}
```
Pure function — no file I/O, no side effects. Takes a server name, returns the health string. Kept separate from I/O so it's easy to test and change independently.

---

```go
file, err := os.Open("servers.txt")
if err != nil {
    fmt.Println("error opening file:", err)
    return
}
defer file.Close()
```
Open for reading. Check immediately. Defer the close right after — before touching the file for anything else.

---

```go
report, err := os.Create("report.txt")
if err != nil {
    fmt.Println("error creating report:", err)
    return
}
defer report.Close()
```
Create for writing. Same pattern — check, then defer. Two defers now queued: `report.Close()` runs first, then `file.Close()`.

---

```go
scanner := bufio.NewScanner(file)
for scanner.Scan() {
    line := scanner.Text()
    fmt.Fprintf(report, "%s\n", serverHealth(line))
}
```
Wrap the file in a scanner. Loop until EOF. Each iteration: get the line, build the health string, write it to `report.txt` with a newline.

---

```go
fmt.Println("Report written to report.txt")
```
Confirms completion to the operator. In a real SRE tool this would be a structured log line, not a plain print.

---

## The Full Data Flow

```
servers.txt              main()                       report.txt
───────────              ──────                       ──────────
web-01        →    scanner reads "web-01"
                   serverHealth("web-01") = "web-01: OK"   →   web-01: OK
web-02        →    scanner reads "web-02"
                   serverHealth("web-02") = "web-02: OK"   →   web-02: OK
db-01         →    scanner reads "db-01"
                   serverHealth("db-01")  = "db-01: OK"    →   db-01: OK
db-02         →    scanner reads "db-02"
                   serverHealth("db-02")  = "db-02: OK"    →   db-02: OK
EOF           →    scanner.Scan() = false → loop ends
                   defers run: report.Close(), file.Close()
```

---

## Mistakes Made Today

### Mistake 1 — Ignoring file errors with `_`

```go
// ❌ Wrong — if open fails, file is nil, scanner silently processes nothing
file, _ := os.Open("servers.txt")
report, _ := os.Create("report.txt")
```

```go
// ✅ Correct — check every file operation
file, err := os.Open("servers.txt")
if err != nil {
    fmt.Println("error opening file:", err)
    return
}
```

Silently swallowed errors are the hardest bugs to find in SRE tooling — the tool runs, produces no output, and you have no idea why.

---

### Mistake 2 — Missing `: ` separator in health string

```go
// ❌ Wrong — output: "web-01OK"
return name + "OK"
```

```go
// ✅ Correct — output: "web-01: OK"
return name + ": OK"
```

---

### Mistake 3 — Wrong output filename (`reports.txt` vs `report.txt`)

```go
// ❌ Wrong — creates "reports.txt", not "report.txt"
report, _ := os.Create("reports.txt")
```

A filename typo creates a different file silently. Always verify the output file exists after running — `cat report.txt` or `ls -la`.

---

### Mistake 4 — Missing final confirmation print

The tool ran, wrote the file, but printed nothing. An SRE tool should always signal completion or failure — never finish silently.

```go
fmt.Println("Report written to report.txt")
```

---

## Final Code

**`servers.txt`**
```
web-01
web-02
db-01
db-02
```

**`main.go`**
```go
package main

import (
	"bufio"
	"fmt"
	"os"
)

func serverHealth(name string) string {
	return name + ": OK"
}

func main() {
	file, err := os.Open("servers.txt")
	if err != nil {
		fmt.Println("error opening file:", err)
		return
	}
	defer file.Close()

	report, err := os.Create("report.txt")
	if err != nil {
		fmt.Println("error creating report:", err)
		return
	}
	defer report.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintf(report, "%s\n", serverHealth(line))
	}
	fmt.Println("Report written to report.txt")
}
```

**`report.txt`** (generated):
```
web-01: OK
web-02: OK
db-01: OK
db-02: OK
```

---

## System Design: Monolith vs Microservices

Every SRE will be asked this question — or will inherit a system built on the wrong answer. Neither is universally right.

---

### The Monolith

All features live in one codebase, one deployable unit, one process.

```
┌─────────────────────────────────────┐
│           Alerting Platform         │
│  ┌──────────┐  ┌──────────────────┐ │
│  │  Auth    │  │  Alert Rules     │ │
│  ├──────────┤  ├──────────────────┤ │
│  │  Users   │  │  Notifications   │ │
│  ├──────────┤  ├──────────────────┤ │
│  │  Config  │  │  Incident Mgmt   │ │
│  └──────────┘  └──────────────────┘ │
└─────────────────────────────────────┘
         one binary, one deploy
```

**Strengths:**
- Simple to develop, test, and deploy — one repo, one build, one `go run`
- Easy to trace a request — everything is in the same process, same logs
- No network calls between components — function calls are nanoseconds
- Easy to do cross-cutting changes — rename a field, it's one change everywhere

**Weaknesses:**
- Scaling is all-or-nothing — can't scale just the alerting engine without scaling auth too
- One team's bad deploy takes down the whole system
- Large codebase becomes hard to navigate — "big ball of mud"
- Technology lock-in — every component must use the same language and runtime

---

### Microservices

Each feature is a separate service, deployed independently, communicating over the network.

```
┌──────────┐   ┌──────────┐   ┌──────────────┐
│  Auth    │   │  Users   │   │ Alert Rules  │
│ service  │   │ service  │   │   service    │
└────┬─────┘   └────┬─────┘   └──────┬───────┘
     │              │                │
     └──────────────┴────────────────┘
                API Gateway
```

**Strengths:**
- Scale independently — 10× the alert processing pods without touching auth
- Independent deploys — ship the notification service without touching incident management
- Independent failures — auth service going down doesn't kill alerting (if designed right)
- Technology freedom — auth in Go, ML model in Python, legacy in Java

**Weaknesses:**
- Network calls between services — latency, timeouts, retries, partial failures
- Distributed tracing required — a request touches 5 services, you need to follow it
- Operational complexity — 10 services = 10 deployment pipelines, 10 sets of logs, 10 health checks
- Distributed transactions are hard — updating two services atomically requires saga patterns

---

### The Honest Trade-off Table

| Factor | Monolith | Microservices |
|--------|----------|--------------|
| Development speed (early) | ✅ Fast | ❌ Slow setup |
| Operational complexity | ✅ Low | ❌ High |
| Independent scaling | ❌ No | ✅ Yes |
| Fault isolation | ❌ One bug kills all | ✅ Failure is contained |
| Debugging | ✅ Simple | ❌ Distributed tracing required |
| Team size | ✅ Small teams | ✅ Large independent teams |
| Deploy frequency | ❌ All or nothing | ✅ Independent |

---

### The Real SRE Answer — Start Monolith, Split When It Hurts

The biggest mistake in the industry is building microservices before you understand your domain. Martin Fowler himself says:

> "Don't start with microservices. Start with a monolith, understand the boundaries, then extract services where it actually helps."

**Split a service out when:**
- One component has significantly different scaling needs
- One team is blocked by another team's deployments
- One component needs a different technology (ML model, legacy integration)
- The deployment of one component is risky to others

**Stay monolith when:**
- The team is small (under 10 engineers)
- The domain isn't well understood yet
- Deploy frequency is low
- You can't afford the operational overhead

---

### SRE Relevance

As an SRE you inherit both. Your job changes based on the architecture:

| Architecture | SRE concern |
|-------------|------------|
| Monolith | Single deploy is high risk — one bad change = full outage. Need solid rollback. |
| Microservices | Distributed tracing, service mesh, per-service SLOs, dozens of health checks |

Microservices move complexity from the codebase to the **infrastructure** — which is why SREs are often the ones who feel the pain most.

---

## Key Takeaways

1. `os.Open` opens for reading; `os.Create` opens for writing and wipes existing content
2. Both return `(file, error)` — always check the error before using the file
3. `defer file.Close()` — write it immediately after a successful open, every time
4. Multiple defers run in reverse order (LIFO) — last deferred runs first
5. `bufio.Scanner` reads line by line — constant memory regardless of file size
6. `scanner.Scan()` returns true while there are lines; `scanner.Text()` gives the current line without the newline
7. `fmt.Fprintf(file, ...)` writes formatted output to any file or writer — not to stdout
8. Never ignore file errors with `_` in production — silent failures are the hardest bugs to find
9. Monolith: simple to build, hard to scale independently, one deploy risk
10. Microservices: independently scalable and deployable, but high operational complexity
11. Start with a monolith — split when a specific pain point justifies the operational cost
12. Microservices move complexity from code to infrastructure — SREs feel this most

---

> **Al-Lateef — اللَّطِيف — The Subtle, The Kind**
>
> _He works in ways unseen — the file closes quietly with defer, the buffer drains without noise, the error surfaces before the crash. He placed subtlety in the design of good systems. See you on Day 13._

# Day 20 — CLI Tool in Go

---

> **بِسْمِ اللهِ الرَّحْمَنِ الرَّحِيم**
>
> **Al-Fattah — الفَتَّاح — The Opener, The Granter of Success**
>
> _He opens doors that no one can close. Today you built your first real CLI tool — a door that takes commands and acts on them. Begin with His name._

---

## Blog of the Day

[How I built a CLI in Go — Carolyn Van Slyck](https://carolynvanslyck.com/blog/2020/08/sting-of-the-viper/)

Read this after the session. Covers building real CLI tools in Go, flag parsing, and the popular `cobra` library used by `kubectl`, `helm`, and `docker`.

---

## Concept 1: What Is a CLI Tool?

A CLI tool is a program you run from the terminal with arguments and flags:

```bash
kubectl get pods -n production
aws s3 ls --profile staging
go run main.go --host web-01 --region us-east-1
```

Everything after the binary name is input to your program. Go gives you the `flag` package in the standard library to parse it cleanly.

---

## Concept 2: `os.Args` — Raw Arguments

The simplest way to read input. `os.Args` is a slice of strings — everything on the command line:

```bash
go run main.go check web-01
```

```go
os.Args[0]  // "main" (the program name)
os.Args[1]  // "check"
os.Args[2]  // "web-01"
```

Raw — no parsing, you handle everything manually. Fine for very simple tools. Use `flag` for anything real.

---

## Concept 3: The `flag` Package

Go's standard library `flag` package parses `--name value` style flags:

```bash
go run main.go --host web-01 --timeout 5
```

```go
host    := flag.String("host", "localhost", "hostname to check")
timeout := flag.Int("timeout", 3, "timeout in seconds")
flag.Parse()

fmt.Println(*host)    // "web-01"
fmt.Println(*timeout) // 5
```

Three arguments to every flag:
1. **name** — the flag name (`--host`)
2. **default** — value used if the flag isn't passed
3. **description** — shown in `--help`

`flag.String` and `flag.Int` return **pointers** (`*string`, `*int`) — the value doesn't exist until after `flag.Parse()`, so Go returns a pointer that gets filled in later.

---

## Concept 4: `flag.Parse()` — Must Be Called First

```go
host := flag.String("host", "localhost", "hostname to check")
flag.Parse()
fmt.Println(*host)   // ✅ correct — parse first, read after
```

`flag.Parse()` reads `os.Args` and fills in all the flag pointers. If you read `*host` before calling it, you always get the default value — the actual input hasn't been parsed yet.

**Rule:** define flags → call `flag.Parse()` → read values.

---

## Concept 5: Dereferencing Flag Pointers

```go
host := flag.String("host", "localhost", "hostname to check")
// host is *string, not string

flag.Parse()
fmt.Println(*host)              // dereference to get the string
checkServer(logger, *host, ...)  // pass the value, not the pointer
```

`*host` dereferences the pointer — gives you the actual string. This is the same `*` from Day 5 (Pointers).

---

## Concept 6: `os.Exit(1)` — Signal Failure to the Caller

```go
if err != nil {
    os.Exit(1)
}
```

Unix convention: exit code 0 = success, non-zero = failure. This is what shell scripts, CI pipelines, and monitoring tools check:

```bash
go run main.go --host db-01
echo $?   # prints 1 — failure
```

A health check tool that doesn't exit with a non-zero code on failure is useless to automation.

---

## Concept 7: Handler vs Logger (`slog`)

```go
// Two steps:
handler := slog.NewJSONHandler(os.Stdout, nil)  // WHERE and FORMAT
logger  := slog.New(handler)                     // the caller interface

// Or in one line:
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
```

**Handler** — decides where logs go and what format they use.
**Logger** — what you call (`logger.Info`, `logger.Error`). Passes every call to its handler.

You can create multiple independent loggers:
```go
stdoutLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
fileLogger   := slog.New(slog.NewJSONHandler(logFile, nil))
```

In practice, create one logger in `main` and pass it through the program — which is exactly what this tool does.

---

## Mistakes Made Today

### Mistake 1 — Handler instead of logger

```go
// ❌ Wrong — NewJSONHandler returns a handler, not a logger
logger := slog.NewJSONHandler(os.Stdout, nil)
```

```go
// ✅ Correct — wrap it in slog.New() to get a logger
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
```

`slog.NewJSONHandler` returns a `slog.Handler`. You can't call `.Info()` on a handler — you need a `*slog.Logger`.

---

### Mistake 2 — Error message capitalized

```go
// ❌ Un-idiomatic — error strings should be lowercase
err := fmt.Errorf("Health Check Failed")
```

```go
// ✅ Correct — lowercase, no trailing punctuation
err := fmt.Errorf("health check failed")
```

Same rule from Day 8 — error strings are lowercase in Go so they read naturally when wrapped: `"starting health check: health check failed"`.

---

### Mistake 3 — Missing `*` when passing flags to a function

```go
// ❌ Wrong — host is *string, not string
checkServer(logger, host, region)
```

```go
// ✅ Correct — dereference to pass the actual string value
checkServer(logger, *host, *region)
```

`flag.String` returns a pointer. You must dereference before passing to a function that expects a `string`.

---

## Final Code

```go
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
)

func checkServer(logger *slog.Logger, host, region string) error {
	logger.Info("running health check ", "host", host, "region", region)
	if host == "db-01" {
		err := fmt.Errorf("health check failed")
		logger.Error("check failed", "host", host, "err", err)
		return err
	}
	logger.Info("check passed", "host", host)
	return nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	host := flag.String("host", "localhost", "hostname to check")
	region := flag.String("region", "us-east-1", "region of host")
	flag.Parse()
	err := checkServer(logger, *host, *region)
	if err != nil {
		os.Exit(1)
	}
}
```

**Run:**
```bash
go run main.go --host web-01 --region us-east-1
go run main.go --host db-01 --region us-east-1
```

**Output:**
```json
{"time":"2026-07-18T00:47:28Z","level":"INFO","msg":"running health check ","host":"web-01","region":"us-east-1"}
{"time":"2026-07-18T00:47:28Z","level":"INFO","msg":"check passed","host":"web-01"}

{"time":"2026-07-18T00:47:28Z","level":"INFO","msg":"running health check ","host":"db-01","region":"us-east-1"}
{"time":"2026-07-18T00:47:28Z","level":"ERROR","msg":"check failed","host":"db-01","err":"health check failed"}
exit status 1
```

---

## System Design: URL Shortener — End-to-End

### What Is a URL Shortener?

Takes a long URL and returns a short code:

```
Input:  https://grafana.internal/d/api-latency?orgId=1&from=now-1h&to=now
Output: https://short.ly/x7k2p
```

When someone visits `short.ly/x7k2p`, they get redirected to the full URL. Simple product — rich system design problem.

---

### Core Components

```
Client
  ↓
[API Gateway / Load Balancer]
  ↓              ↓
[Write Service]  [Read Service]
  ↓                    ↓
[Database]         [Cache — Redis]
                        ↓
                   [Database] (cache miss)
```

**Write path** — user submits a long URL, gets a short code back. Low volume, can be slower.

**Read path** — someone visits a short URL, gets redirected. High volume (every click), must be fast.

---

### The Short Code — How to Generate It

You need a unique 6–8 character code for every URL. Three approaches:

**1. Random — `crypto/rand` + base62**
Generate random bytes, encode as base62 (`[a-zA-Z0-9]`). Simple. Risk of collision — check DB before saving.

**2. Hash — MD5/SHA256 of the long URL, take first 6 chars**
Same URL always gets same code. Collision risk is real — two different URLs can hash to the same prefix.

**3. Counter — global auto-increment ID, encode as base62**
ID 1 → `"000001"`, ID 1000000 → `"4c92"`. No collision. Predictable (enumerable). Needs a central counter — single point of failure if not handled carefully.

Production systems (Bitly, TinyURL) typically use counter + base62.

---

### Database Schema

```sql
CREATE TABLE urls (
    code       VARCHAR(8)   PRIMARY KEY,
    long_url   TEXT         NOT NULL,
    created_at TIMESTAMP    DEFAULT NOW(),
    clicks     BIGINT       DEFAULT 0
);
```

- `code` is the primary key — lookups are by code, so this is the hot query
- Index on `long_url` if you want to detect duplicate submissions

---

### Read Path — Why Caching Is Critical

Every click = one redirect = one DB read. A viral link could get millions of clicks.

Without cache:
```
click → DB query → 5ms → redirect
1M clicks/min → DB overwhelmed
```

With Redis cache:
```
click → Redis lookup (cache hit) → 0.1ms → redirect
click → Redis miss → DB query → write to Redis → redirect
```

TTL on cache entries: 24h for popular links, shorter for rarely-clicked ones.

---

### Write Path — Handling Duplicates

If the same long URL is submitted twice:
- **Option A** — return the same short code (check DB before generating)
- **Option B** — generate a new code each time (simpler, but wastes codes)

Most services choose Option A — index on `long_url`, look up before inserting.

---

### Scale: Read-Heavy System

URL shorteners are extremely read-heavy — 100:1 read/write ratio is common. Design decisions that follow from this:

| Decision | Why |
|----------|-----|
| Cache aggressively (Redis) | Reads must be <1ms |
| Read replicas for DB | Spread read load |
| CDN for popular links | Serve redirects at edge |
| Write service separate from read service | Scale independently |
| NoSQL (DynamoDB) for the URL table | Key-value lookup, no joins needed |

---

### What Happens on Click

```
1. Browser hits short.ly/x7k2p
2. Load balancer routes to Read Service
3. Read Service checks Redis: cache hit?
   → YES: return 301/302 redirect immediately
   → NO:  query DB, write to Redis, return redirect
4. Browser follows redirect to long URL
```

**301 vs 302:**
- `301 Permanent` — browser caches the redirect, never hits your service again. Lower load, but you lose click analytics.
- `302 Temporary` — browser always hits your service. More load, but you can count every click.

Analytics-focused services use 302.

---

### Handling Expiry

Links can expire — useful for campaigns, one-time links:

```sql
ALTER TABLE urls ADD COLUMN expires_at TIMESTAMP NULL;
```

Read service checks: if `expires_at < now()` → return 404. Cache TTL should match `expires_at` so expired links don't stay in cache.

---

### Summary — Design Decisions

| Decision | Choice | Reason |
|----------|--------|--------|
| Short code generation | Counter + base62 | No collision, compact |
| Storage | DynamoDB or PostgreSQL | Simple key-value lookup |
| Caching | Redis, TTL 24h | Read-heavy, must be fast |
| Redirect type | 302 | Preserve click analytics |
| Duplicate URLs | Return same code | User experience |
| Scale | Separate read/write services | Read volume >> write volume |

---

## Key Takeaways

1. `flag.String(name, default, description)` defines a CLI flag — returns `*string`
2. `flag.Parse()` must be called before reading any flag values
3. Dereference flag pointers with `*` before using: `*host`, `*region`
4. `--help` is free — `flag` generates it from your flag descriptions
5. `os.Exit(1)` signals failure to the shell — required for automation and CI
6. `slog.NewJSONHandler` returns a handler (format + destination); `slog.New(handler)` returns the logger (caller interface)
7. Create one logger in `main`, pass it through — don't create loggers inside functions
8. Multiple loggers are possible — each is independent with its own handler
9. URL shortener = write path (generate code) + read path (redirect) — design them separately
10. Read path must be fast — cache with Redis, 100:1 read/write ratio is common
11. Counter + base62 is the standard short code strategy — no collision risk
12. 302 redirect preserves analytics; 301 lets the browser cache the redirect
13. Cache TTL must match link expiry — expired links must not be served from cache

---

> **Al-Hadi — الهَادِي — The Guide**
>
> _He guides to the straight path — no wandering when He leads. A CLI tool does the same: takes input, follows a clear path, exits with a signal anyone can read. See you on Day 21._

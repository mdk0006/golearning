# Day 18 — Testing in Go

---

> **بِسْمِ اللهِ الرَّحْمَنِ الرَّحِيم**
>
> **Al-Muhaymin — المُهَيْمِن — The Guardian, The Overseer**
>
> _He watches over all things — nothing escapes His sight. Tests watch over your code the same way — nothing should slip through unverified. Begin with His name._

---

## Blog of the Day

[Table-Driven Tests — The Go Blog](https://go.dev/blog/subtests)

Read this after the session. The official Go blog post on subtests and table-driven testing. Covers `t.Run`, parallel subtests, and how the pattern scales to hundreds of cases cleanly.

---

## Concept 1: Go's Built-in Testing

Go has testing built into the language — no third-party framework needed. Every test file ends in `_test.go` and uses the `testing` package.

```
day18/
├── healthcheck.go       ← the code
├── healthcheck_test.go  ← the tests
└── go.mod
```

Run tests:
```bash
go test ./...          # run all tests
go test -v ./...       # verbose — shows each subtest result
go test -run TestCheckCPU ./...   # run only one test function
```

---

## Concept 2: Test Function Signature

Every test function must follow this exact signature:

```go
func TestSomething(t *testing.T) {
```

- Name starts with `Test` (capital T) — the runner finds it automatically
- Takes `*testing.T` as the only parameter
- No return value

`t` gives you methods to report failures:

| Method | What it does |
|--------|-------------|
| `t.Errorf("got %s, want %s", got, want)` | marks failed, continues running |
| `t.Fatalf("msg")` | marks failed, stops the test immediately |
| `t.Logf("msg")` | prints only when test fails — for debugging |

---

## Concept 3: Table-Driven Tests — The Go Idiom

Instead of one function per test case, Go uses a slice of test cases — a table:

```go
func TestCheckCPU(t *testing.T) {
    tests := []struct {
        name     string
        usage    float64
        expected string
    }{
        {"healthy",  45.0, "OK"},
        {"boundary", 90.0, "OK"},
        {"critical", 95.0, "CRITICAL"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := CheckCPU(tt.usage)
            if got != tt.expected {
                t.Errorf("got %s, want %s", got, tt.expected)
            }
        })
    }
}
```

**Why this pattern:**
- Adding a new test case = one new line in the slice
- Each case runs independently — one failure doesn't stop others
- Test names appear in output — easy to see which case failed
- No code duplication — the loop handles all cases

---

## Concept 4: Anonymous Struct Slice

```go
tests := []struct {
    name     string
    usage    float64
    expected string
}{
    {"healthy", 45.0, "OK"},
}
```

`[]struct{ ... }` is a slice of an anonymous struct — a struct defined inline with no type name. You only need a named type if you're going to use it in multiple places. For a test table used once, anonymous is cleaner.

Each `{"healthy", 45.0, "OK"}` is a struct literal — values assigned in order to match the field declarations.

---

## Concept 5: `t.Run` — Subtests

```go
t.Run(tt.name, func(t *testing.T) {
    // test body
})
```

`t.Run` creates a named subtest. Each subtest:
- Runs independently — if `"boundary"` fails, `"critical"` still runs
- Has its own `t` — failures are scoped to that subtest
- Appears in verbose output with its full name: `TestCheckCPU/boundary`

The inner `t` shadows the outer `t` — inside the subtest, `t` is the subtest's reporter, not the parent test's.

---

## Concept 6: The Error Convention

```go
t.Errorf("got %s, want %s", got, tt.expected)
```

Go convention: always state **what you got first, then what you wanted**. Makes reading failures natural:

```
--- FAIL: TestCheckCPU/critical (0.00s)
    healthcheck_test.go:22: got OK, want CRITICAL
```

"got OK, want CRITICAL" — immediately clear what went wrong and what was expected.

---

## Package Structure for Tests

```go
// healthcheck.go
package healthcheck

// healthcheck_test.go
package healthcheck   // same package — can access all exported functions directly
```

Test files in the same package don't need to import their own code. `CheckCPU` is directly available because both files share `package healthcheck`.

If you used `package healthcheck_test` (external test package), you'd need to import it — tests would only see exported identifiers. Useful for testing the public API from a user's perspective.

---

## Line by Line — `healthcheck_test.go`

```go
package healthcheck
```
Same package as the code — direct access to `CheckCPU` and `CheckDisk`.

---

```go
import "testing"
```
Go's built-in test package — provides `*testing.T`.

---

```go
func TestCheckCPU(t *testing.T) {
```
Test function. Starts with `Test` — found automatically by `go test`. `t` is your reporter.

---

```go
tests := []struct {
    name     string
    usage    float64
    expected string
}{
```
Anonymous struct slice — the test table. Three fields per case: test name, input, expected output. Defined and initialized in one block.

---

```go
{"healthy",  45.0, "OK"},
{"boundary", 90.0, "OK"},
{"critical", 95.0, "CRITICAL"},
```
Three test cases. `boundary` at exactly 90.0 — checks the threshold boundary is not over-inclusive. `critical` at 95.0 — confirms the CRITICAL branch triggers above 90.

---

```go
for _, tt := range tests {
```
Loops the table. `tt` is each test case struct — `tt.name`, `tt.usage`, `tt.expected`.

---

```go
t.Run(tt.name, func(t *testing.T) {
```
Creates a named subtest. Runs independently. `t` inside is this subtest's reporter.

---

```go
got := CheckCPU(tt.usage)
if got != tt.expected {
    t.Errorf("got %s, want %s", got, tt.expected)
}
```
Call the function, compare result. If wrong — report the failure with got/want message. `t.Errorf` marks failed but keeps running — the next subtest still runs.

---

## Mistakes Made Today

### Mistake 1 — `expected` field typed as `float64` instead of `string`

```go
// ❌ Wrong — expected is a string ("OK", "WARNING"), not a float
tests := []struct {
    name     string
    freeGB   float64
    expected float64   // ← wrong type
}
```

```go
// ✅ Correct
tests := []struct {
    name     string
    freeGB   float64
    expected string
}
```

Go catches this at compile time — you can't put `"OK"` into a `float64` field.

---

### Mistake 2 — Wrong format verb and swapped arguments

```go
// ❌ Wrong — %.1d is for integers, and got/want are swapped
t.Errorf("got %.1d wants %s", got, tt.expected)
```

```go
// ✅ Correct — %s for strings, got first then want
t.Errorf("got %s, want %s", got, tt.expected)
```

---

### Mistake 3 — Typo in expected value

```go
// ❌ Wrong — "CRTICAL" vs actual return value "CRITICAL"
{"critical", 95.0, "CRTICAL"},
```

```go
// ✅ Correct
{"critical", 95.0, "CRITICAL"},
```

A typo in the expected value makes the test fail — not because the code is wrong, but because the test itself has a bug. This is why test names and values deserve the same care as production code.

---

## Final Code

**`healthcheck.go`**
```go
package healthcheck

func CheckCPU(usage float64) string {
	if usage > 90 {
		return "CRITICAL"
	}
	return "OK"
}

func CheckDisk(freeGB float64) string {
	if freeGB < 10 {
		return "WARNING"
	}
	return "OK"
}
```

**`healthcheck_test.go`**
```go
package healthcheck

import "testing"

func TestCheckCPU(t *testing.T) {
	tests := []struct {
		name     string
		usage    float64
		expected string
	}{
		{"healthy", 45.0, "OK"},
		{"boundary", 90.0, "OK"},
		{"critical", 95.0, "CRITICAL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckCPU(tt.usage)
			if got != tt.expected {
				t.Errorf("got %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestCheckDisk(t *testing.T) {
	tests := []struct {
		name     string
		freeGB   float64
		expected string
	}{
		{"healthy", 20.0, "OK"},
		{"boundary", 10.0, "OK"},
		{"warning", 5.0, "WARNING"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckDisk(tt.freeGB)
			if got != tt.expected {
				t.Errorf("got %s wants %s", got, tt.expected)
			}
		})
	}
}
```

Output (`go test -v ./...`):
```
=== RUN   TestCheckCPU
=== RUN   TestCheckCPU/healthy
=== RUN   TestCheckCPU/boundary
=== RUN   TestCheckCPU/critical
--- PASS: TestCheckCPU (0.00s)
=== RUN   TestCheckDisk
=== RUN   TestCheckDisk/healthy
=== RUN   TestCheckDisk/boundary
=== RUN   TestCheckDisk/warning
--- PASS: TestCheckDisk (0.00s)
PASS
ok  	day18
```

---

## System Design: Circuit Breaker — Fail Fast, Prevent Cascade

A circuit breaker prevents a failing downstream service from taking down your entire system. It's one of the most important resilience patterns in distributed SRE work.

---

### The Problem — Cascade Failures

Without a circuit breaker:

```
alerting-service calls config-service
config-service is slow/down
alerting-service waits (timeout: 30s)
alerting-service has 1000 requests waiting
alerting-service runs out of goroutines/threads
alerting-service goes down too
→ cascade failure — one sick service takes down healthy services
```

Each slow call ties up a thread waiting. Enough slow calls and the caller runs out of capacity.

---

### The Circuit Breaker Pattern

Named after the electrical circuit breaker — trips when overloaded to protect the rest of the circuit.

Three states:

```
CLOSED → OPEN → HALF-OPEN → CLOSED
```

**CLOSED (normal operation)**
Requests pass through. Failures are counted. If failures exceed a threshold within a time window, the circuit trips to OPEN.

```
Request → [CLOSED] → config-service
              ↓ 5 failures in 10s
           TRIPS TO OPEN
```

**OPEN (failing fast)**
All requests are immediately rejected — no network call made. Returns an error instantly. Downstream service gets a chance to recover.

```
Request → [OPEN] → immediately returns error
                   (no call to config-service)
```

After a timeout (e.g. 30 seconds), circuit moves to HALF-OPEN.

**HALF-OPEN (testing recovery)**
Allows a small number of requests through. If they succeed, circuit closes. If they fail, circuit opens again.

```
Request → [HALF-OPEN] → config-service
              ↓ succeeds
           CLOSES (normal)
              ↓ fails
           OPENS again (wait longer)
```

---

### State Diagram

```
         failures > threshold
CLOSED ──────────────────────► OPEN
  ▲                               │
  │                    timeout    │
  │         HALF-OPEN ◄───────────┘
  │              │
  └──────────────┘
     probe succeeds
```

---

### What the Caller Sees

```go
result, err := circuitBreaker.Call(func() (interface{}, error) {
    return configService.Get("/config")
})

if err == ErrCircuitOpen {
    // use cached config or default — don't crash
    return cachedConfig, nil
}
```

The circuit breaker wraps calls. When open, the caller gets an error immediately and can use a fallback — cached data, default config, degraded mode.

---

### SRE Relevance

| Without circuit breaker | With circuit breaker |
|------------------------|---------------------|
| Slow service cascades failures | Failures contained, caller gets fast error |
| All threads blocked waiting | Threads freed immediately |
| Recovery takes minutes | Recovery detected in seconds (HALF-OPEN) |
| No visibility into failure rate | Clear state: OPEN/CLOSED/HALF-OPEN |

**Go libraries:** `sony/gobreaker`, `afex/hystrix-go`
**Service mesh:** Istio and Envoy have built-in circuit breaking at the proxy level — no application code changes needed.

---

### Circuit Breaker vs Retry

They complement each other:

- **Retry** — try again on transient failure. Good for flaky networks, brief blips.
- **Circuit breaker** — stop trying after repeated failure. Good for sustained outages.

Without a circuit breaker, retries make cascades **worse** — each retry hammers a service that's already struggling. Circuit breaker + retry: retry a few times, then open the circuit and stop.

---

## Key Takeaways

1. Test files end in `_test.go` and use `package testing` — no external framework needed
2. Test functions start with `Test`, take `*testing.T`, return nothing
3. `t.Errorf` — fail and continue; `t.Fatalf` — fail and stop immediately
4. Table-driven tests: anonymous struct slice, one row per case, loop with `t.Run`
5. Adding a test case = one line in the table — no new function needed
6. `t.Run(name, func)` creates a named subtest — runs independently
7. Error message convention: `"got X, want Y"` — actual first, expected second
8. Test files in the same package access all exported functions directly
9. Circuit breaker has three states: CLOSED (normal), OPEN (failing fast), HALF-OPEN (probing)
10. CLOSED → OPEN when failure threshold exceeded; OPEN → HALF-OPEN after timeout; HALF-OPEN → CLOSED on success
11. Circuit breaker fails fast — returns error immediately when OPEN, no network call made
12. Prevents cascade failures — a slow downstream service can't exhaust upstream goroutines
13. Retry + circuit breaker work together — retry for transient failures, circuit breaker for sustained outages
14. Istio/Envoy provide circuit breaking at the proxy level without application code changes

---

> **Al-Wakeel — الوَكِيل — The Trustee, The Disposer of Affairs**
>
> _Entrust your affairs to Him — He protects what matters. A circuit breaker does the same for your system: when something can't be trusted, it cuts the connection and protects the rest. See you on Day 19._

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Purpose

This is Danish's **3-month Go learning journal** (started April 2026): Golang beginner → advanced, plus System Design and AIOps. Danish is an SRE (CKA, AWS/GCP Pro) using Go to build SRE tooling. It is **not** a production codebase — every `dayXX/` is a tiny standalone exercise paired with learning notes. Optimize for teaching and reinforcement, not for shipping software.

## Structure

```
golearning/
├── README.md          # master 30-day plan + progress checklist
└── dayXX/
    ├── main.go        # the day's Go exercise (package main, single file)
    ├── go.mod         # each day is its own module (module dayXX)
    └── README.md      # learning notes for that day
```

Each day is a **self-contained Go module**. There is no shared code across days — a new `go.mod` is created per day and mistakes/iterations stay scoped to that day.

## Common Commands

```bash
cd dayXX
go run main.go       # run the day's exercise
gofmt -w main.go     # format before committing (required)
go vet ./...         # optional sanity check
```

Initialize a new day:

```bash
mkdir dayNN && cd dayNN
go mod init dayNN
# create main.go and README.md
```

There is no test suite, no linter config, no CI. `gofmt` is the only formatting gate.

## Day README Format (REQUIRED)

Every `dayXX/README.md` must open and close with an **Allah's name + meaning + motivation block** (see the memory file `feedback_session_format.md` and existing day READMEs). This is non-negotiable — the user has explicitly asked for it.

Structure of a day README:
1. Opening: `بِسْمِ اللهِ الرَّحْمَنِ الرَّحِيم` + an Allah's name, meaning, and a short motivation
2. Blog of the Day (link + one-line why it matters)
3. Concept explanation (concept first, always)
4. Mistakes made + corrections (teaches through Danish's own errors)
5. Final code
6. System design topic for the day
7. Key takeaways
8. Closing: another Allah's name + meaning + motivation

The `README.md` is a learning artifact — it should capture **what Danish got wrong and why the correction is idiomatic**, not just the final answer.

## Teaching Rules (from global CLAUDE.md — reinforced here)

These override the default "just solve it" behavior:

1. **Concept first** — explain before any code, every time.
2. **Danish writes the code** — never give the full solution upfront. Hints, not answers.
3. **SRE/infra examples only** — servers, health checks, alerts, k8s, pipelines, DNS, observability. No `Animal`/`Shape`/`Dog` toy examples.
4. **Review after he writes** — once Danish submits code, point out Go idioms, best practices, and mistakes.
5. **System design** — he designs first, then compare trade-offs.
6. **No spoon-feeding** — challenge at the level of his current day.

When Danish asks "how do I do X?", the default answer is a hint toward the concept, not the code.

## Progress Tracking

When a day is completed:
- Flip the `Status` column in root `README.md` to ✅
- Tick the matching checkbox in the Month N checklist
- Link the new `dayXX/README.md` from the root table

## Conventions Already Established

- Go version: `1.26.2` (in each `go.mod`).
- Examples use SRE vocabulary: `Server`, `Hostname`, `Region`, `Healthy`, `cpuUsage`, `prometheus`, `web-01`.
- Variables are lowercase; exported types/functions are uppercase. Local slices use `[]T{}` literal initialization.
- `gofmt` tabs, no extra tooling.
- Commit messages follow the existing style: `Day NN: <short change>` (see `git log`).

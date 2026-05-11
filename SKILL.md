---
name: crap4go
description: Use when the user asks for a CRAP report, cyclomatic complexity analysis, or code quality metrics on a Go project
---

# crap4go - CRAP Metric for Go

Computes the **CRAP** (Change Risk Anti-Pattern) score for every Go function and method. CRAP combines cyclomatic complexity with test coverage to identify functions that are both complex and under-tested.

## Setup

Run from the root of a Go module:

```bash
go run github.com/unclebob/crap4go/cmd/crap4go@latest
```

Or install it:

```bash
go install github.com/unclebob/crap4go/cmd/crap4go@latest
crap4go
```

## Usage

```bash
# Analyze all non-test Go source files under the module
crap4go

# Filter to specific path fragments
crap4go internal/combat movement
```

crap4go automatically deletes old coverage reports, runs `go test ./... -coverprofile=target/coverage/coverage.out`, and then analyzes the results.

### Output

A table sorted by CRAP score, worst first:

```
CRAP Report
===========
Function                       Package                               CC   Cov%     CRAP
-------------------------------------------------------------------------------------
Widget.Run                     widget                                12   45.0%    130.2
simple                         widget                                 1  100.0%      1.0
```

## Interpreting Scores

| CRAP Score | Meaning |
|-----------|---------|
| 1-5       | Clean - low complexity, well tested |
| 5-30      | Moderate - consider refactoring or adding tests |
| 30+       | Crappy - high complexity with poor coverage |

## How It Works

1. Deletes old coverage reports and runs Go coverage
2. Finds non-test `.go` files, excluding `.git`, `target`, and `vendor`
3. Extracts functions and methods with line ranges
4. Computes cyclomatic complexity (`if`, `for`, `range`, switch cases, select clauses, `&&`, `||`)
5. Reads Go coverage profile segments for per-function statement coverage
6. Applies CRAP formula: `CC² x (1 - cov)³ + CC`
7. Sorts by CRAP score descending and prints report

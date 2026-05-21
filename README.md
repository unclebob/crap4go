# crap4go

**CRAP** (Change Risk Anti-Pattern) metric for Go projects.

Combines cyclomatic complexity with test coverage to identify functions and methods that are both complex and under-tested - the riskiest code to change.

## Quick Start

Install or run with `go run`:

```bash
go run github.com/unclebob/crap4go/cmd/crap4go@latest
```

From a Go module:

```bash
crap4go    # deletes old coverage, runs go test ./..., analyzes
```

crap4go automatically deletes stale coverage reports, runs:

```bash
go test ./... -coverprofile=target/coverage/coverage.out
```

and then analyzes the results.

Override the coverage test command when a project needs a narrower or tagged
test invocation:

```bash
crap4go --test-command "go test -tags appunit ./internal/app ./internal/sim"
```

crap4go appends `-coverprofile=target/coverage/coverage.out` to the command.
If you need to control where the coverage flag appears, include
`{coverprofile}` in the command:

```bash
crap4go --test-command "go test ./internal/app {coverprofile} -run TestEditor"
```

Source-file analysis runs in parallel and defaults to half the available
logical CPUs. Override it when needed:

```bash
crap4go --max-workers 4
```

## Output

```
CRAP Report
===========
Function                       Package                               CC   Cov%     CRAP
-------------------------------------------------------------------------------------
Widget.Run                     widget                                12   45.0%    130.2
simple                         widget                                 1  100.0%      1.0
```

## Filtering

Pass module name fragments as arguments to filter:

```bash
crap4go internal/combat movement    # only files matching those path fragments
```

## Coverage Mapping Notes

crap4go reads Go coverage profiles from `target/coverage/coverage.out`.
Coverage is matched to functions by source file and line range. If `go test`
emits module-qualified source paths, crap4go also matches by path suffix so
normal local paths still resolve.

Files skipped during analysis:

- `_test.go` files
- `target/`
- `vendor/`
- `.git/`

## CRAP Formula

```
CRAP(fn) = CC² x (1 - coverage)³ + CC
```

- **CC** = cyclomatic complexity (decision points + 1)
- **coverage** = fraction of covered Go coverage statements in the function range

| Score | Risk |
|-------|------|
| 1-5   | Low - clean code |
| 5-30  | Moderate - refactor or add tests |
| 30+   | High - complex and under-tested |

## What It Counts

Decision points that increase cyclomatic complexity:

- `if`
- `for`
- `range`
- `switch` and `type switch` case clauses
- `select` communication clauses
- `&&`
- `||`

## Claude Code Skill

crap4go includes a `SKILL.md` for use as a Claude Code skill. Add it to your
project's `.claude/settings.json`:

```json
{
  "skills": [
    "https://github.com/unclebob/crap4go/blob/master/SKILL.md"
  ]
}
```

Then ask Claude Code for a "CRAP report" and it will know how to set up and run the tool.

## Development

```bash
go test ./...                          # run tests
go run ./cmd/crap4go                   # run on own source
go run ./cmd/crap4go internal/coverage # filter to one module
```

## License

Copyright (c) Robert C. Martin. All rights reserved.

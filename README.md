# timesheet-txt

`timesheet-txt` is a small command-line tool for logging work in a plain text file (todo.txt format) and generating time reports.

## What this project does

- Stores work entries in a text file (default: `~/.timesheet.txt`)
- Supports adding, listing, and deleting entries
- Generates report output in two formats:
  - `simple` (daily lines with totals)
  - `summary` (grouped totals per task)

## Project structure (quick analysis)

- `timesheet-cli.go` – CLI entrypoint and command routing
- `file/` – file handling (`ReadFile`, `WriteFile`, default path)
- `model/` – parsing for date/duration/work arguments
- `report/` – report generation and aggregation logic
- `util/` – generic helpers (errors, string/array helpers)
- `Makefile` – test, lint, and cross-platform release targets

## Prerequisites

- Go (toolchain)
- Git

## Build

### 1) Clone

```bash
git clone https://github.com/perhenrik/timesheet-txt
cd timesheet-txt
```

### 2) Install dependencies

```bash
go mod tidy
```

### 3) Build binary

```bash
go build -o timesheet .
```

### 4) Run tests

```bash
go test ./...
```

### 5) Makefile targets

```bash
make test      # run tests + coverage summary
make lint      # run go fmt + go vet
make build     # build bin/timesheet
make release   # build release binaries (darwin/linux/windows, amd64+arm64)
make dev-tui   # run TUI with auto-rebuild/restart via air
```

`make release` writes binaries to `release/` with names like:

`timesheet-0.2.0-darwin-arm64`

## CI

GitHub Actions runs `make lint`, `make test`, and `make build` on pushes and pull requests.

## Usage

```text
timesheet [-f filename] action [parameters]
```

- `-f` lets you override the timesheet file path
- default file: `~/.timesheet.txt`

### Commands

#### Add

```bash
timesheet add [date] +<project> [task:<taskname>] hours:<number>
```

Example:

```bash
timesheet add 2026-05-12 +client-a task:api hours:6.5
```

#### Start stopwatch

```bash
timesheet start +<project> [task:<taskname>]
```

Examples:

```bash
timesheet start +client-a task:api
timesheet start +internal task:planning
```

If a stopwatch is already running, it is stopped automatically before the new one starts.

#### Stop stopwatch

```bash
timesheet stop
```

Stops the running stopwatch and writes the calculated `hours` into the timesheet entry.

#### TUI

```bash
timesheet tui
```

Starts a report-first terminal UI where you can:

- view report output for date/period/type
- start and stop stopwatch sessions
- choose project/task from existing values or type your own

Useful keys in TUI:

- `tab` / `shift+tab`: move focus
- `up` / `down`: move focus, or cycle project/task options when those fields are focused
- `enter` or `r`: refresh report
- `t`: toggle report type (`simple`/`summary`)
- `s`: start stopwatch from selected project/task
- `x`: stop running stopwatch
- `n` / `p`: cycle known project/task options
- `q`: quit

Date input accepts digits and `-`, auto-formats as `YYYY-MM-DD`, and supports flexible date entry like `2026-5-3`.
Validation hints for date and period are shown directly under the filter row in the header.
TUI colors adapt to your terminal theme (light/dark) and respect terminal default background.

For TUI development with auto-rebuild/restart on file changes:

```bash
make dev-tui
```

`make dev-tui` runs a project-local watcher (`cmd/devtuiwatch`) that rebuilds and restarts `./tmp/timesheet tui` on Go file changes while preserving normal TUI keyboard input.

#### List

```bash
timesheet list
```

Shows all entries with IDs you can use for delete.

#### Delete

```bash
timesheet delete <id>
```

Example:

```bash
timesheet delete 3
```

#### Report

```bash
timesheet report [date] [period] [type]
```

- `date`: report end date (defaults to now)
- `period`: duration counting backward from `date` (defaults to `5d`)
- `type`: `simple` or `summary` (defaults to simple-style output)

Examples:

```bash
timesheet report
timesheet report 2026-05-12 5d
timesheet report 2026-05-12 2w summary
```

Supported period units:

- `m` = minutes
- `h` = hours
- `d` = days
- `w` = weeks

## License

This project is licensed under GNU GPL v3. See [LICENSE](./LICENSE).

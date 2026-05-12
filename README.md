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
```

`make release` writes binaries to `release/` with names like:

`timesheet-0.1.0-darwin-arm64`

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

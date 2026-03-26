# TaskHash Development Context

TaskHash is a portable, zero-dependency Go tool designed for code signature hashing and task optimization. It prevents redundant execution of development tasks (linting, testing, building) by tracking file hashes and skipping execution when no changes are detected.

## Project Overview

- **Language:** Go (1.26+)
- **Architecture:** Primarily a single-file CLI application (`taskhash.go`).
- **Configuration:** Project settings are stored in `taskhash.json`.
- **State:** Task signatures are tracked in `.code_signatures.json` (by default).
- **Hooks:** Installs both `pre-commit` (enforcement) and `post-commit` (detection of bypass).

## Building and Running

The project uses a `Makefile` for standard operations:

| Task | Command | Description |
|------|---------|-------------|
| **Build** | `make build` | Compiles the `taskhash` binary. |
| **Test** | `make test` | Runs all tests with the race detector enabled. |
| **Coverage**| `make test-cover` | Runs tests and generates an HTML coverage report. |
| **Format** | `make fmt` | Formats Go code using `go fmt`. |
| **Lint** | `make lint` | Performs static analysis using `go vet`. |
| **Profile** | `make prof` | Runs benchmarks and generates CPU/Memory profiles. |
| **Install** | `make install` | Builds and moves the binary to `/usr/local/bin` (requires sudo). |

Manual build: `go build -o taskhash taskhash.go`

### Useful Binary Commands

| Command | Description |
|---------|-------------|
| `taskhash install <path>` | Copy binary to a path (e.g. `tools/taskhash/taskhash`). |
| `taskhash status` | Check if all tasks are up to date. |
| `taskhash up` | Robust self-upgrade to latest GitHub release. |
| `taskhash install-hook` | Installs bypass-resistant hooks. |

## Development Conventions

### Coding Style
- **Single File:** The core logic resides in `taskhash.go`. Logic is organized using clear header comments (e.g., `// Output Formatting`).
- **Formatting:** Standard `go fmt` is mandatory.
- **Terminal Output:** Uses ANSI escape codes for rich terminal feedback. Helper functions like `success()`, `failure()`, and `cyanBold()` are used to maintain consistency.

### Testing Practices
- **Table-Driven Tests:** Prefer table-driven patterns for logic testing (see `taskhash_test.go`).
- **Temporary Directories:** Use `t.TempDir()` for all filesystem-related tests to ensure isolation.
- **Benchmarks:** Performance-critical paths (like hash calculation) should be benchmarked using `go test -bench`.

### Framework Support
When adding support for new frameworks, update the `detectFramework()` function in `taskhash.go` to include appropriate `includes` and `excludes` patterns.

## Task Queuing System (Internal)
The project directory contains a `docs/tasks/` folder which appears to be a local mission tracking system (using `.usv` and `.md` files).
- **Active Tasks:** `docs/tasks/active/`
- **Pending Tasks:** `docs/tasks/pending/`
- **Completed Tasks:** `docs/tasks/completed/`

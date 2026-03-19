# TaskHash

A portable, zero-dependency tool for code signature hashing and task optimization. It prevents redundant execution of linting, testing, and other tasks by tracking code hashes.

## Features

- **Portable**: Compiled Go binary works on Linux, macOS, and Windows
- **Auto-detection**: Automatically detects Python, JavaScript/TypeScript, Go, Rust, and Ruby projects
- **Task-aware**: Tracks separate hashes for different tasks (lint, test, build)
- **Per-task includes**: Each task can hash different files for optimal skipping
- **Pre-commit hook**: Optional automatic installation of git hooks

## Quick Start

```bash
# Download and install taskhash, then initialize in your repo
curl -fsSL https://raw.githubusercontent.com/bizkite-co/taskhash/main/install.sh | bash

# Or if you have the source:
./install.sh

# Or if taskhash is already in your PATH:
taskhash init
```

This will:
1. Download/install the taskhash binary
2. Detect your framework (Python, JS/TS, Go, Rust, Ruby)
3. Create `taskhash.json` with sensible defaults
4. Install the pre-commit hook

## Commands

| Command | Description |
|---------|-------------|
| `taskhash init` | Detect framework, create config, install hook |
| `taskhash init --no-hook` | Detect framework, create config only |
| `taskhash enforce` | Run all tasks sequentially (lint → test → build) |
| `taskhash check <task>` | Check if task hash matches |
| `taskhash update <task>` | Update task hash manually |
| `taskhash install-hook` | Install pre-commit hook |
| `taskhash remove-hook` | Remove pre-commit hook |
| `taskhash reinstall-hook` | Reinstall pre-commit hook |

## Configuration (`taskhash.json`)

```json
{
  "includes": ["src", "lib"],
  "excludes": [".git", "node_modules", "__pycache__"],
  "store": ".code_signatures.json",
  "runner": "make",
  "tasks": {
    "lint": {},
    "test": {"includes": ["src", "tests"]},
    "build": {}
  }
}
```

### Options

| Field | Description | Default |
|-------|-------------|---------|
| `includes` | Files/dirs to hash (global) | `["src"]` |
| `excludes` | Files/dirs to skip | `.git`, `node_modules`, etc. |
| `store` | Where to save hashes | `.code_signatures.json` |
| `runner` | Command prefix | `make` |
| `tasks` | Task definitions | `lint`, `test`, `build` |

### Per-task Options

```json
{
  "tasks": {
    "lint": {},
    "test": {"includes": ["src", "tests"]},
    "build": {}
  }
}
```

Tasks inherit global `includes` if not overridden. Commands are `{runner} {task}` (e.g., `make lint`).

## Output Examples

First run (no hashes):
```
lint: no hash found, running make lint...
test: no hash found, running make test...
build: no hash found, running make build...
```

Subsequent runs (hashes match):
```
lint: hash match (0f6048a4), skipping.
test: hash match (0f6048a4), skipping.
build: hash match (0f6048a4), skipping.
```

Code changed:
```
lint: hash mismatch (was abc123, now def456), running make lint...
lint: hash updated (def456)
test: hash mismatch (was abc123, now def456), running make test...
test: hash updated (def456)
```

## Framework Detection

`taskhash init` auto-detects:

| Framework | Includes |
|-----------|----------|
| Python (pyproject.toml) | `src`, `tests`, `pyproject.toml` |
| JavaScript/TypeScript (package.json) | `src`, `package.json` |
| Go (go.mod) | `.` (everything) |
| Rust (Cargo.toml) | `src`, `Cargo.toml` |
| Ruby (Gemfile) | `lib`, `Gemfile` |

## Makefile Integration

Update existing targets to use taskhash:

```makefile
lint:
	@if taskhash check lint; then \
		echo "Lint hash matches. Skipping."; \
	else \
		ruff check . --fix && mypy . && \
		taskhash update lint; \
	fi

test: lint
	@if taskhash check test; then \
		echo "Test hash matches. Skipping."; \
	else \
		pytest tests/ && \
		taskhash update test; \
	fi
```

## Compilation

```bash
go build -o taskhash taskhash.go
```

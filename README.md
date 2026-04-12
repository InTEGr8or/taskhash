# TaskHash

A portable, zero-dependency tool for code signature hashing and task optimization. It prevents redundant execution of linting, testing, and other tasks by tracking code hashes.

## Features

- **Portable**: Compiled Go binary works on Linux, macOS, and Windows
- **Auto-detection**: Automatically detects Python, JavaScript/TypeScript, Go, Rust, and Ruby projects
- **Task-aware**: Tracks separate hashes for different tasks (lint, test, build)
- **Per-task includes**: Each task can hash different files for optimal skipping
- **Pre-commit hook**: Optional automatic installation of git hooks

## Quick Start

### Linux/macOS (Bash)
```bash
# Download and install taskhash, then initialize in your repo
curl -H "Cache-Control: no-cache" -fsSL https://raw.githubusercontent.com/InTEGr8or/taskhash/main/install.sh | bash
```

### Windows (PowerShell)
```powershell
# Download and install taskhash, then initialize in your repo
iwr -useb https://raw.githubusercontent.com/InTEGr8or/taskhash/main/install.ps1 | iex
```

# Or if you have the source:
```bash
./install.sh # Linux/macOS
# or
.\install.ps1 # Windows
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
| `taskhash enforce` | Run all tasks sequentially (lint → test → build) |
| `taskhash status` | Show hash match status for all tasks |
| `taskhash check <task>` | Check if task hash matches |
| `taskhash update <task>` | Update task hash manually |
| `taskhash install <path>` | Copy taskhash to a specific location (e.g. tools/taskhash/taskhash) |
| `taskhash install-hook` | Install pre-commit and post-commit hooks |
| `taskhash remove-hook` | Remove git hooks |
| `taskhash up` | Upgrade taskhash to latest version |

## Installation & Recommendations

For a consistent experience across team members, we recommend "vendoring" the `taskhash` binary into your repository:

1. **Tools Directory (Recommended):** Place it in `tools/taskhash/taskhash`. This keeps your project root clean and ensures any script or Makefile can find it reliably.
   ```bash
   ./taskhash install tools/taskhash/taskhash
   ```
2. **Global (Path):** Install to `/usr/local/bin` if you want it available everywhere on your machine.
   ```bash
   sudo make install
   ```
3. **Project Root:** Keep it in the root as `./taskhash` for simple, single-repo use.

### Bypassing Hooks

The `taskhash install-hook` command installs both a `pre-commit` and a `post-commit` hook. While Git allows bypassing `pre-commit` with the `--no-verify` flag, the `post-commit` hook will still run and issue a warning if the code signatures are stale, ensuring that bypassed checks are always detected.

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

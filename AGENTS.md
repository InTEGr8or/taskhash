# OpenCode Agent Instructions

This repository is a Go-based tool for task optimization.

## Development Workflow

- **Build**: `make build`
- **Lint**: `make lint` (runs `go vet`)
- **Test**: `make test` (runs with `-race`)
- **Format**: `make fmt`

## Entrypoint

- **Main**: `taskhash.go` is the primary entrypoint.
- **Install Scripts**: `install.sh` (Linux/macOS), `install.ps1` (Windows).

## Performance Analysis

To analyze performance, use the following:

- **Profiling**: `make prof`
- **Analysis**: `go tool pprof cpu.prof` or `go tool pprof mem.prof`

## Notes

- **Usage Docs**: See `README.md` for CLI command documentation and configuration (`taskhash.json`) specifications.
- **Self-Hosting**: This tool uses standard Go paradigms. Avoid adding unnecessary dependencies to keep it portable and zero-dependency.
- **Windows Support**: Windows binaries are generated via GitHub Actions (`.github/workflows/release.yml`).

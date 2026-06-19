# Development

## Makefile

A Makefile is provided for common tasks:

| Command | What it does |
|---|---|
| `make build` | Build both `cli` and `tui` binaries into `./bin/` |
| `make test` | Run all tests with race detector |
| `make vet` | Run `go vet` static analysis |
| `make lint` | Run golangci-lint (if installed) |
| `make tidy` | Tidy module dependencies |
| `make clean` | Remove build artifacts |
| `make run-cli` | Quick CLI run via `go run` |
| `make run-tui` | Quick TUI run via `go run` |
| `make install` | Install binaries to `$GOPATH/bin` |
| `make all` | Build, vet, test (CI-style) |

Set `RACE=0` to disable the race detector for faster test runs:

```bash
make test RACE=0
```

## Pre-push hook

A pre-push hook is configured at `.githooks/pre-push` that runs `make all` (build + vet + test) before every `git push`. The hook directory is set via `git config core.hooksPath .githooks` — already configured in this repo. New clones should run:

```bash
git config core.hooksPath .githooks
```

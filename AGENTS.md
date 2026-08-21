# AGENTS.md - Guide for AI Agents and Contributors

This file guides automated agents (and humans) working on the `gandalf` project.

## Project Overview

`gandalf` is a secure password generator CLI written in Go, themed around *The Lord of the Rings*.

- **Module:** `github.com/wellingtonlope/gandalf` (`go.mod:1`)
- **Binary:** `gandalf` via `cmd/gandalf/main.go:1`
- **Core logic:** `internal/generator/generator.go:1` - cryptographically secure generation using `crypto/rand`
- **Easter egg:** `internal/mellon/mellon.go:1` - `mellon` / `friend` reveals the Doors of Durin inscription
- **Language:** All code, comments, and documentation MUST be in English

## Project Structure

```
.
├── cmd/gandalf/main.go          # CLI entrypoint, flag parsing, mellon dispatch
├── internal/generator/          # Password generation logic + table tests
│   ├── generator.go
│   └── generator_test.go
├── internal/mellon/             # Easter egg logic + table tests
│   ├── mellon.go
│   └── mellon_test.go
├── Makefile                     # Build, test, lint, fmt tasks
├── go.mod
├── README.md
└── AGENTS.md                    # This file
```

## How to Build and Run

```bash
make build              # bin/gandalf
./bin/gandalf -l 16
./bin/gandalf -l 32 --no-symbols
./bin/gandalf -c 5 -l 12
./bin/gandalf mellon    # easter egg
./bin/gandalf friend    # also works
./bin/gandalf --help
```

## Mandatory Pre-Commit Checks

**Before EVERY commit, run:**

```bash
make fmt    # format code (go fmt + gofmt)
make lint   # golangci-lint or fallback go vet
make test   # go test ./... -race -cover
```

Or the single combined target:

```bash
make check  # fmt-check + lint + test
```

Do NOT commit if any of these fail. Fix issues first.

- `make fmt` - formats with `go fmt ./...` and `gofmt -s -w`
- `make fmt-check` - verifies formatting without modifying files (CI)
- `make lint` - runs `golangci-lint run ./...` if installed, otherwise `go vet ./...`
  - Install linter: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- `make vet` - `go vet ./...` only
- `make test` - `go test ./... -count=1 -race -cover`
- `make verify` - `make check` + `make build` (full verification)

## Coding Conventions

1.  **Language:** English only for code, comments, commit messages, docs.
2.  **Formatting:** Always run `make fmt` before commit. No manual formatting debates.
3.  **Linting:** Keep `go vet` clean. If `golangci-lint` is available, keep it clean.
4.  **Error Handling:** Return errors with clear English messages (`internal/generator/generator.go:52`). Wrap with context if needed.
5.  **Security:** Use `crypto/rand` for passwords, never `math/rand`. See `internal/generator/generator.go:65`.
6.  **Tests:** Table-driven tests are REQUIRED for all new logic. See `internal/generator/generator_test.go:1` and `internal/mellon/mellon_test.go:1` as reference.
    - Use `tests := []struct{ name, input, want, wantErr }` pattern
    - Use `t.Run` for each case
    - Cover valid, invalid, edge (0, 1, 256, 257), empty charset, etc.
7.  **Flags:** CLI uses `flag` stdlib. Keep short (`-l`, `-c`, `-h`) and long (`--length`, `--count`, `--help`) variants in sync (`cmd/gandalf/main.go:35`).
8.  **Easter Egg:** `mellon` logic must stay case-insensitive in `main.go` but `IsMellon` checks exact lower-case. Do not break `gandalf mellon` / `gandalf friend`.
9.  **No Comments Overkill:** Keep comments concise, English, godoc-style.
10. **Changelog:** Every commit that changes behavior, features, fixes, or tests MUST update `CHANGELOG.md` under `[Unreleased]` following [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) (`Added`/`Changed`/`Fixed`/`Tests`). Reference `file:line` (e.g., `internal/generator/generator.go:13`). Do not leave `Unreleased` empty if the commit is user-visible.

## Testing Guidelines

- Tests must be table-driven.
- Run `make test` or `go test ./... -race -cover`.
- Cover:
  - `Options.Validate()` - length bounds, empty charset
  - `Options.Charset()` - combinations
  - `Generate()` - length, charset membership, error cases
  - `GenerateMultiple()` - count validation
  - `IsMellon()` - true/false cases, case sensitivity
  - `Reveal()` - output contains expected substrings (capture stdout)

Example table test structure:

```go
tests := []struct{
    name string
    opts Options
    wantErr bool
}{
    {name: "valid default", opts: DefaultOptions(), wantErr: false},
    {name: "invalid zero length", opts: Options{Length: 0, UseLower: true}, wantErr: true},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) { ... })
}
```

## Git Workflow

1.  Make changes
2.  Update `CHANGELOG.md` under `[Unreleased]` if the change is user-visible (features, fixes, tests, behavior)
3.  Run `make check` (or `make fmt && make lint && make test`)
4.  Fix failures
5.  Commit (commit message in English)
6.  Push

## Common Tasks

| Task | Command |
|------|---------|
| Build | `make build` |
| Test | `make test` |
| Lint | `make lint` |
| Format | `make fmt` |
| Verify all | `make verify` |
| Easter egg | `make mellon` |
| Tidy deps | `make tidy` |

## Notes for Agents

- Read relevant files entirely before editing.
- Prefer editing existing files; do not create new files unless required.
- Verify changes by running the mandatory checks; do not assume.
- Keep the LOTR theme subtle and fun, but code professional.
- If you change the module path or add dependencies, run `go mod tidy`.


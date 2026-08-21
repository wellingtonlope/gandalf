# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased](https://github.com/wellingtonlope/gandalf/compare/v0.2.2...HEAD)

## [0.2.2](https://github.com/wellingtonlope/gandalf/releases/tag/v0.2.2) - 2026-08-21
### Fixed
- `clipboard`: when using `-x` / `--copy` with `-c` (multiple passwords), now copies **all** passwords joined by newline instead of only the first (`cmd/gandalf/main.go:98`).

## [0.2.1](https://github.com/wellingtonlope/gandalf/releases/tag/v0.2.1) - 2026-08-21
### Added
- `generator`: guaranteed charset coverage — every generated password now contains at least one character from each enabled set (`lower`, `upper`, `digits`, `symbols`) and is shuffled with a crypto-secure Fisher-Yates shuffle (`internal/generator/generator.go:85`).
- `generator`: validation that `length` must be at least the number of enabled character sets (e.g., `length 2` with 4 sets returns `length must be at least 4...`) (`internal/generator/generator.go:68`).

### Tests
- Added `TestGenerate_GuaranteesCharset` (5 scenarios, 20-50 iterations) proving the previous algorithm missed digits/symbols and now always passes.
- Added `TestGenerate_ValidateRequiresLengthForCharsets` (5 cases) for the new validation.

## [0.2.0](https://github.com/wellingtonlope/gandalf/releases/tag/v0.2.0) - 2026-08-21
### Added
- `clipboard`: new flag `-x, --copy` to copy passwords to the system clipboard (one per line), using `github.com/atotto/clipboard` (`cmd/gandalf/main.go:61`, `internal/clipboard/clipboard.go:1`).
- `clipboard`: `pbcopy` (macOS), `xclip`/`xsel`/`wl-copy` (Linux), `clip` (Windows) support via `atotto/clipboard`.
- `main`: refactored `run()` to be testable with injected `copyFn` and `io.Writer` (`cmd/gandalf/main.go:27`).
- `README`: documented `-x, --copy` in options table.

### Tests
- Added `internal/clipboard` package with 100% coverage (`internal/clipboard/clipboard_test.go:1`).
- Added `cmd/gandalf` tests for `--copy` / `-x` (success, error, nil, multiple passwords) with `map[string]struct{}` and `testify/assert` (`cmd/gandalf/main_test.go:132`).

## [0.1.0](https://github.com/wellingtonlope/gandalf/releases/tag/v0.1.0) - 2026-08-20
### Added
- Initial release of `gandalf` — secure password generator CLI (`cmd/gandalf/main.go:1`).
- Cryptographically secure generation with `crypto/rand` (`internal/generator/generator.go:72`).
- CLI flags: `-l, --length`, `-c, --count`, `--no-upper`, `--no-lower`, `--no-digits`, `--no-symbols`, `-h, --help`, `-v, --version` (`cmd/gandalf/main.go:49`).
- Easter egg `mellon` / `friend` revealing the Doors of Durin inscription (`internal/mellon/mellon.go:1`).
- Table-driven tests with `testify/assert` and `map[string]struct{}` (`internal/generator/generator_test.go:1`, `internal/mellon/mellon_test.go:1`).
- Cross-platform builds for `linux`/`darwin`/`windows` `amd64`/`arm64` (`Makefile:76`, `dist/`).
- Release automation via GitHub Actions (`release.yml:1`) publishing `*.tar.gz`, `*.zip`, `checksums.txt`.
- Homebrew tap at `wellingtonlope/homebrew-tap` with auto-update (`Formula/gandalf.rb:1`, `.github/scripts/update-tap.sh:1`, `TAP_TOKEN`).
- Installers `install.sh` (Linux/macOS) and `install.ps1` (Windows).
- Pre-commit checks `make fmt` / `make lint` / `make test` / `make check` (`AGENTS.md:1`, `Makefile:58`).

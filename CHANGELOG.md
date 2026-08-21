# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased](https://github.com/wellingtonlope/gandalf/compare/v0.3.1...HEAD)

## [0.3.1](https://github.com/wellingtonlope/gandalf/compare/v0.3.0...v0.3.1) - 2026-08-21
### Added
- `completion`: new subcommand `gandalf completion zsh` generating zsh completion for subcommands only (`check`, `completion`, `mellon`, `friend`, `help`, `version`) (`cmd/gandalf/main.go:43`, `internal/completion/completion.go:1`).
- `docs`: document zsh completion install (`README.md:66`).

### Tests
- `completion`: add table-driven tests for zsh script (subcommands, no flags) and supported shells (`internal/completion/completion_test.go:1`).
- `cmd/gandalf`: add completion tests — zsh success, case-insensitive, missing/unsupported shell, help, unknown flag (`cmd/gandalf/main_test.go:500`).

## [0.3.0](https://github.com/wellingtonlope/gandalf/compare/v0.2.2...v0.3.0) - 2026-08-21
### Added
- `lore`: new package `internal/lore/lore.go:1` with Middle-earth themed strength ranks `Hobbit (0-27)` < `Orc (28-35)` < `Ranger (36-59)` < `Elf (60-79)` < `Dwarf (80-99)` < `Wizard (100-127)` < `Sauron (128+)`, entropy `len*log2(charset)`, thematic penalties (common passwords, Middle-earth words, sequential, repeating) and crack time at 10B guesses/sec (`internal/lore/lore.go:120`, `internal/lore/lore.go:220`).
- `check`: new subcommand `gandalf check <password>` and `echo <password> | gandalf check` with stdin support, `gandalf check --help` and Middle-earth themed output (emoji, quote, advice) (`cmd/gandalf/main.go:40`, `cmd/gandalf/main.go:122`, `cmd/gandalf/main.go:239`).
- `docs`: document `gandalf check` with ranks and examples (`README.md:51`).

### Changed
- `generator`: expose `randReader` and `randInt` for testability to simulate `crypto/rand` failures (`internal/generator/generator.go:13`, `internal/generator/generator.go:17`).
- `docs`: require `CHANGELOG.md` update on every commit under `[Unreleased]` via `AGENTS.md:85` (`AGENTS.md:10`, `AGENTS.md:114`).

### Tests
- `generator`: achieve 100% coverage (`93.1%` → `100%`, total `95.6%` → `98.7%`) by testing `crypto/rand` error propagation for guaranteed chars, fill and shuffle phases (`internal/generator/generator_test.go:290`, `internal/generator/generator_test.go:347`).
- `generator`: fix flaky `TestGenerate_Randomness` (2 passwords `NotEqual`) to statistical check of 20 passwords (`>=18` unique) (`internal/generator/generator_test.go:188`) and increase `TestGenerate_CharsetFiltering` iterations `10` → `50` (`internal/generator/generator_test.go:163`).
- `generator`: add `TestRequiredCount` and convert all table tests from `map` to `[]struct` slice for deterministic order (`internal/generator/generator_test.go:23`, `internal/generator/generator_test.go:73`).
- `mellon`: fix `os.Stdout` leak with `t.Cleanup` and helper `captureReveal`, convert tables to slices, consolidate `Portal` assertions (`internal/mellon/mellon_test.go:99`, `internal/mellon/mellon_test.go:137`).
- `clipboard`: fix `WriteAll` restoration with `t.Cleanup`, dedup `TestCopy_EmptyText` and `TestWriteAll_Mockable` into single table-driven `TestCopy` (`internal/clipboard/clipboard_test.go:11`).
- `cmd/gandalf`: add missing CLI edge cases — flag parse errors (`-l abc`, `--unknown`), `count` validation (`-c 0`), `length < requiredCount` (`-l 2` with 4 sets), hint verification (`Hint: try 'gandalf --help'`), and `--no-*` filtering verification with 20 iterations (`cmd/gandalf/main_test.go:63`, `cmd/gandalf/main_test.go:124`).
- `cmd/gandalf`: expand `mellon` tests for `TrimSpace`/`ToLower` (`"  mellon  "`, `" FrIeNd "`) and negative case `"amigo"` not triggering, plus `t.Cleanup` capture helper (`cmd/gandalf/main_test.go:186`, `cmd/gandalf/main_test.go:287`).
- `cmd/gandalf`: dedup duplicate `copy with error` case and convert all `map` tables to slices (`cmd/gandalf/main_test.go:15`, `cmd/gandalf/main_test.go:132`).
- `lore`: add comprehensive table-driven tests for ranks, entropy, warnings and penalties (`95.8%` coverage) (`internal/lore/lore_test.go:1`).
- `cmd/gandalf`: add `check` subcommand tests — human output for all ranks, stdin handling, errors, help, multiple args and emoji (`cmd/gandalf/main_test.go:271`).

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

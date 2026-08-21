# gandalf 🧙‍♂️

> "Speak, friend, and enter."

*Not all who wander are careless with passwords.* A cryptographically secure generator forged in Go — speak, friend, and conjure passwords as strong as the Doors of Durin.

## Installation

### Homebrew (macOS / Linux) — no Go required

```bash
brew tap wellingtonlope/tap
brew install gandalf

# verify
gandalf --version
gandalf --help
```

Upgrade:

```bash
brew update && brew upgrade gandalf
```

### With Go

```bash
go install github.com/wellingtonlope/gandalf/cmd/gandalf@latest
```

Or build locally:

```bash
make build
./bin/gandalf -l 16
```

> Need Windows or a direct download? Grab a pre-built binary from [GitHub Releases](https://github.com/wellingtonlope/gandalf/releases) (Linux, macOS, Windows).

## Usage

```bash
gandalf                  # 16-character password
gandalf -l 24            # 24-character password
gandalf -l 32 --no-symbols
gandalf -c 5 -l 12       # 5 passwords
gandalf --help
```

### Check Password Strength (Middle-earth Ranks)

Evaluate any password with thematic lore ranks from Hobbit (weak) to Sauron (legendary):

```bash
gandalf check "password123"          # 🔴 Hobbit - weak
gandalf check "xK9$m2PqL8@vB4!"      # 🔵 Dwarf - strong
echo "mySecret" | gandalf check      # via stdin (avoids shell history)
gandalf check --help                 # show all ranks
```

Ranks: `Hobbit (0-27)` < `Orc (28-35)` < `Ranger (36-59)` < `Elf (60-79)` < `Dwarf (80-99)` < `Wizard (100-127)` < `Sauron (128+)`

Each check shows entropy, charset, crack time, warnings (common words, Middle-earth terms, sequences, repeats) and advice.

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `-l, --length` | password length | `16` |
| `-c, --count` | number of passwords | `1` |
| `--no-upper` | exclude uppercase | `false` |
| `--no-lower` | exclude lowercase | `false` |
| `--no-digits` | exclude digits | `false` |
| `--no-symbols` | exclude symbols | `false` |
| `-x, --copy` | copy first password to clipboard | `false` |
| `-h, --help` | show help | |
| `-v, --version` | show version | |

## Development

```bash
make test   # run tests
make lint   # run linter (golangci-lint or go vet)
make fmt    # format code
make check  # fmt + lint + test
make dist   # cross-compile for Linux/macOS/Windows (amd64/arm64) to dist/
```

See `AGENTS.md` for agent guidelines and pre-commit checks.

# gandalf 🧙‍♂️

> "Speak, friend, and enter."

*Not all who wander are careless with passwords.* A cryptographically secure generator forged in Go — speak, friend, and conjure passwords as strong as the Doors of Durin.

## Installation

```bash
go install github.com/wellingtonlope/gandalf/cmd/gandalf@latest
```

Or build locally:

```bash
make build
./bin/gandalf -l 16
```

## Usage

```bash
gandalf                  # 16-character password
gandalf -l 24            # 24-character password
gandalf -l 32 --no-symbols
gandalf -c 5 -l 12       # 5 passwords
gandalf --help
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `-l, --length` | password length | `16` |
| `-c, --count` | number of passwords | `1` |
| `--no-upper` | exclude uppercase | `false` |
| `--no-lower` | exclude lowercase | `false` |
| `--no-digits` | exclude digits | `false` |
| `--no-symbols` | exclude symbols | `false` |
| `-h, --help` | show help | |
| `-v, --version` | show version | |

## The Secret

```bash
gandalf mellon
# or
gandalf friend
```

Displays the inscription on the Doors of Durin.

## Development

```bash
make test   # run tests
make lint   # run linter (golangci-lint or go vet)
make fmt    # format code
make check  # fmt + lint + test
```

See `AGENTS.md` for agent guidelines and pre-commit checks.

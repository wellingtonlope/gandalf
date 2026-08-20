# gandalf 🧙‍♂️

> "Speak, friend, and enter."

*Not all who wander are careless with passwords.* A cryptographically secure generator forged in Go — speak, friend, and conjure passwords as strong as the Doors of Durin.

## Installation

### With Go

```bash
go install github.com/wellingtonlope/gandalf/cmd/gandalf@latest
```

Or build locally:

```bash
make build
./bin/gandalf -l 16
```

### Without Go (pre-built binaries)

Download from [GitHub Releases](https://github.com/wellingtonlope/gandalf/releases) — available for **Linux**, **macOS** (Intel & Apple Silicon), and **Windows**.

#### Linux / macOS (quick install)

```bash
curl -sSL https://raw.githubusercontent.com/wellingtonlope/gandalf/main/install.sh | bash
# custom version or directory:
curl -sSL https://raw.githubusercontent.com/wellingtonlope/gandalf/main/install.sh | bash -s -- --version v0.1.0 --dir /usr/local/bin

gandalf --help
gandalf -l 24
```

The script auto-detects `linux`/`darwin` and `amd64`/`arm64`, downloads the matching `gandalf-*.tar.gz` from the latest Release, and installs to `/usr/local/bin` (with `sudo` if needed).

#### Linux / macOS (manual)

```bash
# Example: Linux amd64, replace version and platform as needed
curl -LO https://github.com/wellingtonlope/gandalf/releases/download/v0.1.0/gandalf-linux-amd64.tar.gz
tar -xzf gandalf-linux-amd64.tar.gz
sudo mv gandalf-linux-amd64 /usr/local/bin/gandalf
chmod +x /usr/local/bin/gandalf

# macOS Apple Silicon
curl -LO https://github.com/wellingtonlope/gandalf/releases/download/v0.1.0/gandalf-darwin-arm64.tar.gz
tar -xzf gandalf-darwin-arm64.tar.gz
sudo mv gandalf-darwin-arm64 /usr/local/bin/gandalf
```

Available assets: `gandalf-linux-amd64.tar.gz`, `gandalf-linux-arm64.tar.gz`, `gandalf-darwin-amd64.tar.gz`, `gandalf-darwin-arm64.tar.gz`, `checksums.txt`.

#### Windows

1. **Manual:** Go to [Releases](https://github.com/wellingtonlope/gandalf/releases), download `gandalf-windows-amd64.zip` (or `-arm64` for ARM), extract `gandalf.exe`, move it to a folder in your `PATH` (e.g. `%USERPROFILE%\bin`) and add that folder to `PATH` via System Properties → Environment Variables.
2. **PowerShell script (one-liner):**
   ```powershell
   iwr -useb https://raw.githubusercontent.com/wellingtonlope/gandalf/main/install.ps1 | iex
   # custom:
   .\install.ps1 -Version v0.1.0 -Dir "$env:USERPROFILE\bin"
   ```
   The script detects `amd64`/`arm64`, downloads the correct `.zip`, extracts `gandalf.exe` to `~/bin`, and adds it to the user `PATH` (restart terminal after).

```powershell
gandalf --help
gandalf -l 24 --no-symbols
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

## Development

```bash
make test   # run tests
make lint   # run linter (golangci-lint or go vet)
make fmt    # format code
make check  # fmt + lint + test
make dist   # cross-compile for Linux/macOS/Windows (amd64/arm64) to dist/
```

See `AGENTS.md` for agent guidelines and pre-commit checks.

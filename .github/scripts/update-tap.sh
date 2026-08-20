#!/usr/bin/env bash
set -euo pipefail

# update-tap.sh - Update Homebrew tap formula with new version and sha256
# Usage:
#   ./.github/scripts/update-tap.sh v0.2.0
#   ./.github/scripts/update-tap.sh v0.2.0 /path/to/Formula/gandalf.rb /path/to/checksums.txt
#   VERSION=v0.2.0 ./update-tap.sh
#
# In CI, called after release publishes checksums.txt.
#   - If checksums file not found, it downloads from GitHub Releases.

REPO="wellingtonlope/gandalf"
TAP_REPO="wellingtonlope/homebrew-tap"

VERSION="${1:-${VERSION:-}}"
FORMULA_PATH="${2:-}"
CHECKSUMS_FILE="${3:-}"

if [[ -z "$VERSION" ]]; then
  echo "Usage: $0 <version> [formula_path] [checksums_file]" >&2
  echo "  version e.g. v0.2.0" >&2
  exit 1
fi

# Normalize version: allow v0.2.0 or 0.2.0
TAG="$VERSION"
if [[ "$TAG" != v* ]]; then
  TAG="v$TAG"
fi
VER="${TAG#v}"

# Detect formula path
if [[ -z "$FORMULA_PATH" ]]; then
  if [[ -f "tap/Formula/gandalf.rb" ]]; then
    FORMULA_PATH="tap/Formula/gandalf.rb"
  elif [[ -f "../homebrew-tap/Formula/gandalf.rb" ]]; then
    FORMULA_PATH="../homebrew-tap/Formula/gandalf.rb"
  elif [[ -f "$HOME/git/homebrew-tap/Formula/gandalf.rb" ]]; then
    FORMULA_PATH="$HOME/git/homebrew-tap/Formula/gandalf.rb"
  elif [[ -f "Formula/gandalf.rb" ]]; then
    FORMULA_PATH="Formula/gandalf.rb"
  else
    FORMULA_PATH="tap/Formula/gandalf.rb"
  fi
fi

# Detect checksums file
if [[ -z "$CHECKSUMS_FILE" ]]; then
  if [[ -f "dist/checksums.txt" ]]; then
    CHECKSUMS_FILE="dist/checksums.txt"
  elif [[ -f "checksums.txt" ]]; then
    CHECKSUMS_FILE="checksums.txt"
  elif [[ -f "/tmp/checksums.txt" ]]; then
    CHECKSUMS_FILE="/tmp/checksums.txt"
  fi
fi

# Download if not found
if [[ -z "$CHECKSUMS_FILE" || ! -f "$CHECKSUMS_FILE" ]]; then
  echo "Checksums not found locally, downloading from release $TAG..."
  TMP=$(mktemp)
  URL="https://github.com/${REPO}/releases/download/${TAG}/checksums.txt"
  echo "  URL: $URL"
  if ! curl -fsSL "$URL" -o "$TMP"; then
    echo "Failed to download checksums.txt from $URL" >&2
    echo "Ensure release $TAG exists and has checksums.txt asset." >&2
    exit 1
  fi
  CHECKSUMS_FILE="$TMP"
fi

echo "Version: $TAG ($VER)"
echo "Formula: $FORMULA_PATH"
echo "Checksums: $CHECKSUMS_FILE"
cat "$CHECKSUMS_FILE"

if [[ ! -f "$FORMULA_PATH" ]]; then
  echo "Formula not found: $FORMULA_PATH" >&2
  exit 1
fi

# Extract sha256 for each asset
get_sha() {
  local asset="$1"
  grep -E " $asset\$" "$CHECKSUMS_FILE" | awk '{print $1}' | head -1
}

SHA_DARWIN_ARM64=$(get_sha "gandalf-darwin-arm64.tar.gz")
SHA_DARWIN_AMD64=$(get_sha "gandalf-darwin-amd64.tar.gz")
SHA_LINUX_ARM64=$(get_sha "gandalf-linux-arm64.tar.gz")
SHA_LINUX_AMD64=$(get_sha "gandalf-linux-amd64.tar.gz")

if [[ -z "$SHA_DARWIN_ARM64" || -z "$SHA_DARWIN_AMD64" || -z "$SHA_LINUX_ARM64" || -z "$SHA_LINUX_AMD64" ]]; then
  echo "Could not extract all shas from checksums.txt" >&2
  echo "Found:" >&2
  echo "  darwin-arm64: $SHA_DARWIN_ARM64" >&2
  echo "  darwin-amd64: $SHA_DARWIN_AMD64" >&2
  echo "  linux-arm64: $SHA_LINUX_ARM64" >&2
  echo "  linux-amd64: $SHA_LINUX_AMD64" >&2
  exit 1
fi

echo "Shas:"
echo "  darwin-arm64: $SHA_DARWIN_ARM64"
echo "  darwin-amd64: $SHA_DARWIN_AMD64"
echo "  linux-arm64: $SHA_LINUX_ARM64"
echo "  linux-amd64: $SHA_LINUX_AMD64"

# Update formula: version, urls, shas
# Use temporary file and sed
cp "$FORMULA_PATH" "$FORMULA_PATH.bak"

# Update version line
sed -i -E "s/^[[:space:]]*version \".*\"/  version \"$VER\"/" "$FORMULA_PATH"

# Update URLs: replace /v<old>/ with /${TAG}/
sed -i -E "s|releases/download/v[^/]+/|releases/download/${TAG}/|g" "$FORMULA_PATH"

# Update sha256 per platform: we need to replace the sha after each url's block.
# Simplest: replace all 4 shas sequentially by matching the asset name context.
# Approach: replace the sha lines following each platform's url.
# For robustness, replace the 4 placeholder shas directly in order:
# darwin-arm64 first occurrence, darwin-amd64 second, etc. But sed alone is fragile.
# Instead, do per-file replacement using awk or python.

# Use python for precise replacement if available, else fallback to sed per sha.

if command -v python3 >/dev/null 2>&1; then
  python3 - "$FORMULA_PATH" "$SHA_DARWIN_ARM64" "$SHA_DARWIN_AMD64" "$SHA_LINUX_ARM64" "$SHA_LINUX_AMD64" <<'PY'
import re, sys
path, sha_darwin_arm64, sha_darwin_amd64, sha_linux_arm64, sha_linux_amd64 = sys.argv[1:6]
with open(path) as f:
    content = f.read()

# Replace shas in order of appearance: darwin arm64, darwin amd64, linux arm64, linux amd64
# The formula has 4 sha256 lines in that order; replace sequentially.
shas = [sha_darwin_arm64, sha_darwin_amd64, sha_linux_arm64, sha_linux_amd64]
idx = 0
def repl(m):
    global idx
    if idx < len(shas):
        result = f'      sha256 "{shas[idx]}"'
        idx += 1
        return result
    return m.group(0)

new_content, n = re.subn(r'[ ]*sha256 "[a-f0-9]{64}"', repl, content)
with open(path, 'w') as f:
    f.write(new_content)
print(f"Updated {n} sha256 lines via python")
PY
else
  # Fallback: sequential sed (assumes order darwin-arm64, darwin-amd64, linux-arm64, linux-amd64)
  # We need to replace nth occurrence - use ed or manual
  echo "python3 not found, using sed fallback (may be less precise)"
  # Replace all shas with first sha then fix? Instead, use a temp file and iterate.
  # For fallback, just replace each sha individually by searching surrounding url pattern would be complex.
  # As fallback, we will replace the 4 shas in order using a loop with awk.
  awk -v a="$SHA_DARWIN_ARM64" -v b="$SHA_DARWIN_AMD64" -v c="$SHA_LINUX_ARM64" -v d="$SHA_LINUX_AMD64" '
  BEGIN { n=0 }
  /sha256 "/ {
    n++
    if (n==1) sub(/sha256 "[^"]*"/, "sha256 \"" a "\"")
    else if (n==2) sub(/sha256 "[^"]*"/, "sha256 \"" b "\"")
    else if (n==3) sub(/sha256 "[^"]*"/, "sha256 \"" c "\"")
    else if (n==4) sub(/sha256 "[^"]*"/, "sha256 \"" d "\"")
  }
  { print }
  ' "$FORMULA_PATH.bak" > "$FORMULA_PATH"
fi

echo "Updated formula:"
cat "$FORMULA_PATH"
echo ""
echo "Diff:"
diff -u "$FORMULA_PATH.bak" "$FORMULA_PATH" || true
rm -f "$FORMULA_PATH.bak"

echo "Done. Verify with: cat $FORMULA_PATH"

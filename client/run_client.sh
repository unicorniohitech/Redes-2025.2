#!/usr/bin/env bash
set -euo pipefail

# Usage: ./check-and-run-client.sh [HOST] [PORT]
# Defaults: HOST=localhost PORT=8000
HOST="${1:-localhost}"
PORT="${2:-8000}"

echo "Using HOST=${HOST}, PORT=${PORT}"

# Check Go toolchain: verify `go version` returns output
GOVER=$(go version 2>/dev/null)
if [ -z "${GOVER}" ]; then
  echo "Error: 'go version' returned no output. Install Go and ensure 'go' is in PATH."
  exit 1
fi

echo "Go found: ${GOVER}"

echo "Running client locally with HOST=${HOST}, PORT=${PORT}"

# Move to project root (assumes script is in client/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}/.." || exit 1

# Prefer running a built binary at tcp/bin/tcp if present; otherwise build it and run
BIN_DIR="./tcp/bin"
BIN_PATH="${BIN_DIR}/tcp"
mkdir -p "${BIN_DIR}"
if [ -x "${BIN_PATH}" ]; then
  echo "Found ${BIN_PATH} — running it"
  "${BIN_PATH}" -mode=client -address="${HOST}" -port=${PORT}
  exit $?
fi

echo "Building client binary to ${BIN_PATH}"
if ! go build -o "${BIN_PATH}" ./tcp; then
  echo "Error: Failed to build tcp client binary."
  exit 1
fi

echo "Running built client binary"
"${BIN_PATH}" -mode=client -address="${HOST}" -port=${PORT}
exit $?
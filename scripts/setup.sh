#!/usr/bin/env bash
# Setup toolchain for hermetic/reproducible builds.
# Run from repo root: ./scripts/setup.sh
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

REQUIRED_NODE="${REQUIRED_NODE:-24.14.0}"

run() {
  echo "==> $*"
  "$@"
}

cmp_version_ge() {
  local current="$1" required="$2"
  local first
  first="$(printf "%s\n%s\n" "$required" "$current" | sort -V | head -n1)"
  [[ "$first" == "$required" ]]
}

# Node
setup_node() {
  if [[ -f .nvmrc ]]; then
    local want
    want="$(cat .nvmrc | tr -d '[:space:]')"
    if [[ -s "${NVM_DIR:-$HOME/.nvm}/nvm.sh" ]]; then
      # npm-run shells often export npm_config_prefix, which breaks nvm activation.
      unset npm_config_prefix NPM_CONFIG_PREFIX
      echo "==> nvm install $want && nvm use"
      source "${NVM_DIR:-$HOME/.nvm}/nvm.sh"
      nvm install "$want" 2>/dev/null || true
      nvm use
    elif command -v node >/dev/null 2>&1; then
      local current
      current="$(node -v | sed 's/^v//')"
      if ! cmp_version_ge "$current" "$want"; then
        echo "node v$current < required v$want (use nvm or install from nodejs.org)"
        exit 1
      fi
      echo "node ok: v$current"
    else
      echo "missing node; install nvm or node >= $want"
      exit 1
    fi
  else
    echo "no .nvmrc; skipping node version check"
  fi
}

# npm deps (includes buf via @bufbuild/buf)
setup_npm() {
  if [[ -d node_modules ]]; then
    run npm install
  else
    run npm ci
  fi
}

# buf (from node_modules or system)
verify_buf() {
  if [[ -x node_modules/.bin/buf ]]; then
    run node_modules/.bin/buf --version
    return
  fi
  if command -v buf >/dev/null 2>&1; then
    run buf --version
    return
  fi
  echo "buf not found; run npm ci first"
  exit 1
}

# Go (optional)
verify_go() {
  if [[ -d platform-cli ]] || [[ -d services/omnichannel/backend ]]; then
    if ! command -v go >/dev/null 2>&1; then
      echo "go not found; install from go.dev"
      exit 1
    fi
    run go version
  fi
}

# Rust (optional)
verify_rust() {
  if [[ -d cli ]] || [[ -d services/x_stream_bot ]]; then
    if ! command -v cargo >/dev/null 2>&1; then
      echo "cargo not found; install from rustup.rs"
      exit 1
    fi
    run cargo --version
  fi
}

# Python3 (optional)
verify_python() {
  if [[ -f scripts/ci/materialize_platform_configs.py ]]; then
    if ! command -v python3 >/dev/null 2>&1; then
      echo "python3 not found"
      exit 1
    fi
    run python3 --version
  fi
}

setup_node
setup_npm
verify_buf
verify_go
verify_rust
verify_python

echo ""
echo "setup: toolchain ready"

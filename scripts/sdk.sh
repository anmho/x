#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
NX_BIN="${NX_BIN:-}"

usage() {
  cat <<USAGE
Usage: scripts/sdk.sh <command>

Nx flow wrapper for SDK commands. Delegates to "nx run sdk:<target>".

Commands:
  lint
  generate-server | generate-go
  generate-es
  push
  publish-es
  publish-go
  publish | publish-all
  list

Examples:
  ./scripts/sdk.sh lint
  ./scripts/sdk.sh generate-es
  SDK_VERSION=v1.2.0 ./scripts/sdk.sh publish-all
USAGE
}

resolve_nx() {
  if [[ -n "$NX_BIN" ]]; then
    return
  fi

  if command -v nx >/dev/null 2>&1; then
    NX_BIN="$(command -v nx)"
    return
  fi

  if [[ -x "$ROOT_DIR/node_modules/.bin/nx" ]]; then
    NX_BIN="$ROOT_DIR/node_modules/.bin/nx"
    return
  fi

  echo "missing required command: nx" >&2
  echo "hint: run npm ci at repo root" >&2
  exit 1
}

main() {
  if [[ $# -lt 1 ]]; then
    usage
    exit 1
  fi

  local cmd="$1"
  shift || true
  local target=""

  case "$cmd" in
    lint)
      target="lint"
      ;;
    generate-server|generate-go)
      target="generate-go"
      ;;
    generate-es)
      target="generate-es"
      ;;
    push)
      target="push"
      ;;
    publish-es)
      target="publish-es"
      ;;
    publish-go)
      target="publish-go"
      ;;
    publish|publish-all)
      target="publish"
      ;;
    list)
      target="list"
      ;;
    *)
      usage
      exit 1
      ;;
  esac

  resolve_nx
  cd "$ROOT_DIR"
  "$NX_BIN" run "sdk:$target" "$@"
}

main "$@"

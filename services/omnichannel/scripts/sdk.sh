#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROTO_DIR="$ROOT_DIR/backend/proto"

MODULE_FULL_NAME="${BSR_MODULE:-buf.build/anmhela/omnichannel}"
SDK_VERSION="${SDK_VERSION:-}"
ES_PLUGIN="${ES_PLUGIN:-buf.build/connectrpc/es}"
GO_PLUGIN="${GO_PLUGIN:-buf.build/connectrpc/go}"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

usage() {
  cat <<USAGE
Usage: scripts/sdk.sh <command>

Commands:
  lint            Run buf lint (+breaking check when main branch is available)
  generate-server Generate local Go server stubs (for backend runtime)
  generate-es     Generate local TypeScript/ES client (for frontend; use nx run sdk:generate-es)
  push            Push proto module to Buf Schema Registry (BSR)
  publish-es      Publish ConnectRPC TypeScript SDK version in BSR
  publish-go      Publish ConnectRPC Go SDK version in BSR
  publish-all     Push module and publish both ES + Go SDK versions
  list            List SDK versions for configured plugins

Required env for publish-*:
  SDK_VERSION=vX.Y.Z

Optional env:
  BSR_MODULE=buf.build/anmhela/omnichannel
  ES_PLUGIN=buf.build/connectrpc/es
  GO_PLUGIN=buf.build/connectrpc/go
  SKIP_BREAKING=1
USAGE
}

run_buf() {
  (cd "$PROTO_DIR" && buf "$@")
}

publish_sdk() {
  local plugin="$1"

  if [[ -z "$SDK_VERSION" ]]; then
    echo "SDK_VERSION is required for publishing (example: SDK_VERSION=v1.2.0)" >&2
    exit 1
  fi

  # Publishes a generated SDK package version in BSR for the module+plugin pair.
  run_buf registry sdk version create "$MODULE_FULL_NAME" \
    --plugin "$plugin" \
    --version "$SDK_VERSION"
}

main() {
  if [[ $# -lt 1 ]]; then
    usage
    exit 1
  fi

  require_cmd buf

  local cmd="$1"
  shift || true

  case "$cmd" in
    lint)
      run_buf lint
      if [[ "${SKIP_BREAKING:-0}" == "1" ]]; then
        echo "Skipping breaking-change check (SKIP_BREAKING=1)"
      elif git -C "$ROOT_DIR" rev-parse --verify main >/dev/null 2>&1; then
        run_buf breaking --against '.git#branch=main'
      else
        echo "Skipping breaking-change check: 'main' branch not found locally"
      fi
      ;;
    generate-server)
      run_buf generate --template buf.gen.server.yaml
      ;;
    generate-es)
      run_buf generate --template buf.gen.client.yaml
      ;;
    push)
      run_buf push
      ;;
    publish-es)
      publish_sdk "$ES_PLUGIN"
      ;;
    publish-go)
      publish_sdk "$GO_PLUGIN"
      ;;
    publish-all)
      run_buf push
      publish_sdk "$ES_PLUGIN"
      publish_sdk "$GO_PLUGIN"
      ;;
    list)
      run_buf registry sdk version list "$MODULE_FULL_NAME" --plugin "$ES_PLUGIN"
      run_buf registry sdk version list "$MODULE_FULL_NAME" --plugin "$GO_PLUGIN"
      ;;
    *)
      usage
      exit 1
      ;;
  esac
}

main "$@"

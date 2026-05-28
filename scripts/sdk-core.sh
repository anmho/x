#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROTO_DIR="$ROOT_DIR/services/omnichannel/backend/proto"

MODULE_FULL_NAME="${BSR_MODULE:-buf.build/anmhela/omnichannel}"
SDK_VERSION="${SDK_VERSION:-}"
ES_PLUGIN="${ES_PLUGIN:-buf.build/connectrpc/es}"
GO_PLUGIN="${GO_PLUGIN:-buf.build/connectrpc/go}"
BUF_BIN=""

ensure_path() {
  export PATH="$ROOT_DIR/node_modules/.bin:$PATH"
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

resolve_buf() {
  if [[ -x "$ROOT_DIR/node_modules/.bin/buf" ]]; then
    BUF_BIN="$ROOT_DIR/node_modules/.bin/buf"
    return
  fi

  if command -v buf >/dev/null 2>&1; then
    BUF_BIN="$(command -v buf)"
    return
  fi

  echo "missing required command: buf" >&2
  echo "hint: install dependencies at repo root (npm ci) or install buf globally" >&2
  exit 1
}

usage() {
  cat <<USAGE
Usage: scripts/sdk-core.sh <command>

Commands:
  lint            Run buf lint (+breaking check when main branch is available)
  generate-server Generate local Go server stubs (for backend runtime)
  generate-es     Generate local TypeScript/ES client (for frontend)
  push            Push proto module to Buf Schema Registry (BSR)
  publish-es      Publish ConnectRPC TypeScript SDK version in BSR
  publish-go      Publish ConnectRPC Go SDK version in BSR
  publish-all     Push module and publish both ES + Go SDK versions
  list            Resolve SDK versions for configured plugins

Optional env:
  SDK_VERSION=vX.Y.Z
  BSR_MODULE=buf.build/anmhela/omnichannel
  ES_PLUGIN=buf.build/connectrpc/es
  GO_PLUGIN=buf.build/connectrpc/go
  SKIP_BREAKING=1
USAGE
}

run_buf() {
  (cd "$PROTO_DIR" && "$BUF_BIN" "$@")
}

has_proto_files() {
  local count
  count=$(find "$PROTO_DIR" -name "*.proto" -maxdepth 10 2>/dev/null | wc -l | tr -d ' ')
  [[ "$count" -gt 0 ]]
}

has_sdk_version_create() {
  run_buf registry sdk version --help | rg -q 'create'
}

has_sdk_version_list() {
  run_buf registry sdk version --help | rg -q 'list'
}

resolve_sdk_version() {
  local plugin="$1"
  run_buf registry sdk version --module "$MODULE_FULL_NAME" --plugin "$plugin"
}

publish_sdk() {
  local plugin="$1"

  if has_sdk_version_create; then
    if [[ -n "$SDK_VERSION" ]]; then
      run_buf registry sdk version create \
        --module "$MODULE_FULL_NAME" \
        --plugin "$plugin" \
        --version "$SDK_VERSION"
      return
    fi

    resolve_sdk_version "$plugin"
    return
  fi

  if [[ -n "$SDK_VERSION" ]]; then
    echo "SDK_VERSION provided, but this buf CLI does not support explicit sdk version create; resolving generated version instead."
  fi

  resolve_sdk_version "$plugin"
}

main() {
  if [[ $# -lt 1 ]]; then
    usage
    exit 1
  fi

  ensure_path
  resolve_buf

  local cmd="$1"
  shift || true

  case "$cmd" in
    lint)
      run_buf lint
      if [[ "${SKIP_BREAKING:-0}" == "1" ]]; then
        echo "Skipping breaking-change check (SKIP_BREAKING=1)"
      elif git -C "$ROOT_DIR" rev-parse --verify main >/dev/null 2>&1; then
        if git -C "$ROOT_DIR" ls-tree -r --name-only main -- services/omnichannel/backend/proto | rg -q '\.proto$'; then
          run_buf breaking --against "$ROOT_DIR/.git#branch=main,subdir=services/omnichannel/backend/proto"
        else
          echo "Skipping breaking-change check: no proto files found in main baseline"
        fi
      else
        echo "Skipping breaking-change check: 'main' branch not found locally"
      fi
      ;;
    generate-server)
      if ! has_proto_files; then
        echo "sdk-core: no .proto files found in $PROTO_DIR — skipping generate-server"
        exit 0
      fi
      require_cmd protoc-gen-go
      require_cmd protoc-gen-connect-go
      run_buf generate --template buf.gen.server.yaml
      ;;
    generate-es)
      if ! has_proto_files; then
        echo "sdk-core: no .proto files found in $PROTO_DIR — skipping generate-es"
        exit 0
      fi
      require_cmd protoc-gen-es
      require_cmd protoc-gen-connect-es
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
      if has_sdk_version_list; then
        run_buf registry sdk version list --module "$MODULE_FULL_NAME" --plugin "$ES_PLUGIN"
        run_buf registry sdk version list --module "$MODULE_FULL_NAME" --plugin "$GO_PLUGIN"
      else
        resolve_sdk_version "$ES_PLUGIN"
        resolve_sdk_version "$GO_PLUGIN"
      fi
      ;;
    *)
      usage
      exit 1
      ;;
  esac
}

main "$@"

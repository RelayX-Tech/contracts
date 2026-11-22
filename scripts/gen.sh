#!/usr/bin/env bash
# 固定版本生成 Go stubs：protoc 3.19.4 / protoc-gen-go v1.36.6 / protoc-gen-go-grpc 1.5.1。
# 生成物入库；CI 重新生成后 `git diff --exit-code` 校验一致。
set -euo pipefail
cd "$(dirname "$0")/.."

PROTO_INCLUDE="${PROTO_INCLUDE:-$(go env GOPATH)/include}"

protoc -I . -I "$PROTO_INCLUDE" \
  --go_out=. --go_opt=module=github.com/relayx-tech/contracts \
  --go-grpc_out=. --go-grpc_opt=module=github.com/relayx-tech/contracts \
  core/v1/*.proto agent/v1/*.proto

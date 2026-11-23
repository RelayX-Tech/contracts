#!/usr/bin/env bash
# 依赖边界扫描：contracts 是开放合同仓库，不得依赖 core / panel / agent 实现 module。
set -euo pipefail
cd "$(dirname "$0")/.."

forbidden='github.com/relayx-tech/(core|panel|agent|clients)'
if go list -m all | grep -Eq "$forbidden"; then
  echo "boundary: contracts 依赖了禁止的实现 module" >&2
  go list -m all | grep -E "$forbidden" >&2
  exit 1
fi

# 生成物必须来自 proto，禁止手写 gRPC 类型：gen/ 下只允许 protoc 产物。
if find gen -type f ! -name '*.pb.go' | grep -q .; then
  echo "boundary: gen/ 下存在非 protoc 生成文件" >&2
  find gen -type f ! -name '*.pb.go' >&2
  exit 1
fi

echo "boundary: ok"

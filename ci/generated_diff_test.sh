#!/usr/bin/env bash
# REG-02：合同生成物一致性——重新生成 stubs 后与入库内容逐字节比对（无 diff）。
set -euo pipefail
cd "$(dirname "$0")/.."

if ! command -v protoc >/dev/null 2>&1; then
  echo "generated_diff_test: skip（protoc 未安装）"
  exit 0
fi
if ! command -v protoc-gen-go >/dev/null 2>&1 || ! command -v protoc-gen-go-grpc >/dev/null 2>&1; then
  echo "generated_diff_test: skip（protoc-gen-go / protoc-gen-go-grpc 未安装）"
  exit 0
fi

backup=$(mktemp -d)
trap 'rm -rf "$backup"' EXIT
cp -a gen "$backup/gen"

./scripts/gen.sh

if ! diff -r "$backup/gen" gen >/tmp/generated_diff_test.diff 2>&1; then
  # 工具链版本头随环境变化（CI 固定 protoc 3.19.4 / 插件版本），不作为内容差异。
  remaining=$(grep -E '^[<>]' /tmp/generated_diff_test.diff \
    | grep -vE '^[<>] *// - (protoc|protoc-gen-go|protoc-gen-go-grpc) ' || true)
  if [ -n "$remaining" ]; then
    echo "generated_diff_test: 生成物与 proto 不一致（请重新生成并入库）" >&2
    cat /tmp/generated_diff_test.diff >&2
    exit 1
  fi
  echo "generated_diff_test: 仅工具链版本头差异（内容一致）"
fi

echo "generated_diff_test: ok"

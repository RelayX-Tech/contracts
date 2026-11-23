#!/usr/bin/env bash
# AC43：开源 / 闭源产物边界——core-control / core-node 为闭源二进制；Agent 开源且无闭源 SDK。
set -euo pipefail
cd "$(dirname "$0")/.."

fail=0
note() { echo "artifact_boundary_test: $*" >&2; fail=1; }

# 1. 不存在独立 SDK 目录（无 xCore SDK）。
for root in . ../core ../agent ../panel; do
  [ -d "$root" ] || continue
  if find "$root" -maxdepth 3 -type d -iname 'sdk' | grep -q .; then
    note "存在 SDK 目录（不允许）：$(find "$root" -maxdepth 3 -type d -iname 'sdk')"
  fi
done

# 2. core 提供 core-control / core-node 两个闭源命令。
[ -d ../core/cmd/core-control ] || note "缺少 core/cmd/core-control"
[ -d ../core/cmd/core-node ] || note "缺少 core/cmd/core-node"

# 3. Agent / Panel 不内嵌闭源 core，也不依赖其 module。
for repo in ../agent ../panel; do
  [ -d "$repo" ] || continue
  if grep -rInE --include='*.go' 'github.com/relayx-tech/core' "$repo" 2>/dev/null | grep -v '_test\.go:' | grep -q .; then
    note "$(basename "$repo") 依赖了闭源 module github.com/relayx-tech/core"
  fi
done

# 4. Agent 不内嵌闭源二进制 / 源码目录。
for dir in ../agent/internal/core ../agent/core ../agent/vendor/github.com/relayx-tech/core; do
  [ -e "$dir" ] && note "Agent 内嵌闭源 core：$dir"
done

if [ "$fail" -ne 0 ]; then
  exit 1
fi
echo "artifact_boundary_test: ok"

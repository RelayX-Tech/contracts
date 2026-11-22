#!/usr/bin/env bash
# 生成物一致性门禁：重新生成 stubs 后 gen/ 必须无 diff（合同与生成物不漂移）。
set -euo pipefail
cd "$(dirname "$0")/.."

if ! git diff --quiet -- gen/; then
  echo "generated artifacts are stale; 请运行 scripts/gen.sh 并提交生成物" >&2
  git diff --stat -- gen/ >&2
  exit 1
fi

if [ -n "$(git status --porcelain -- gen/)" ]; then
  echo "gen/ 存在未提交的生成物变更" >&2
  git status --porcelain -- gen/ >&2
  exit 1
fi

echo "breaking_change_gate: ok"

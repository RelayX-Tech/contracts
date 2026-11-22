#!/usr/bin/env bash
# 非目标扫描：合同中不得出现 TUN / TUIC / xdrive / 闭源 SDK / HTTPS 控制契约。
set -euo pipefail
cd "$(dirname "$0")/.."

fail=0
scan() {
  local pattern="$1" desc="$2"
  if grep -rInE "$pattern" core agent gen 2>/dev/null | grep -vE '_test\.go:' | grep -q .; then
    echo "non_goals_scan: 发现禁止项：$desc" >&2
    grep -rInE "$pattern" core agent gen 2>/dev/null | grep -vE '_test\.go:' >&2
    fail=1
  fi
}

scan 'PROTOCOL_TUIC|PROTOCOL_TUN[^N]' 'TUIC / TUN 入站'
scan 'xdrive' 'xdrive 协议'
scan 'sdk/' '闭源 SDK 目录引用'
scan 'https://.*control|/api/v1/control' 'HTTPS 控制契约'

if [ "$fail" -ne 0 ]; then
  exit 1
fi
echo "non_goals_scan: ok"

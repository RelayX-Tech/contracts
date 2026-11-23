#!/usr/bin/env bash
# AC47：非目标全仓扫描——无闭源 SDK、无 Agent HTTPS/JSON 控制、无 Panel 直连 core、
# 无 TUN/TUIC/xdrive、无 License / 授权码。默认扫描本仓 + 同级 core / agent / panel。
set -euo pipefail
cd "$(dirname "$0")/.."

ROOTS=("$@")
if [ ${#ROOTS[@]} -eq 0 ]; then
  ROOTS=(".")
  for sib in ../core ../agent ../panel; do
    [ -d "$sib" ] && ROOTS+=("$sib")
  done
fi

fail=0

scan_code() {
  local pattern="$1" desc="$2" hits
  hits=$(grep -rInE --include='*.go' --include='*.proto' "$pattern" "${ROOTS[@]}" 2>/dev/null | grep -v '_test\.go:' || true)
  if [ -n "$hits" ]; then
    echo "non_goals_scan_test: 代码出现禁止项：$desc" >&2
    echo "$hits" >&2
    fail=1
  fi
}

scan_docs() {
  local pattern="$1" desc="$2" hits
  hits=$(grep -rInE --include='*.md' "$pattern" "${ROOTS[@]}" 2>/dev/null \
    | grep -vE '非目标|已移除|不引入|不实现|不新增|不提供|不包含|禁止|移除' || true)
  if [ -n "$hits" ]; then
    echo "non_goals_scan_test: 有效文档出现禁止项：$desc" >&2
    echo "$hits" >&2
    fail=1
  fi
}

scan_code 'PROTOCOL_TUIC|PROTOCOL_TUN([^A-Z]|$)' 'TUIC / TUN 入站'
scan_code 'xdrive' 'xdrive 协议'
scan_code 'LicenseService|LicenseCode|授权码' 'License / 授权码'
scan_code '/api/v1/control|https://[^ ]*control' 'HTTPS 控制契约'
scan_docs 'TUIC|xdrive' 'TUIC / xdrive（文档须标注为非目标/已移除）'

# 开源组件不得依赖闭源实现 module。
hits=$(grep -rInE --include='*.go' 'relayx-tech/core' ../agent ../panel 2>/dev/null | grep -v '_test\.go:' || true)
if [ -n "$hits" ]; then
  echo "non_goals_scan_test: Agent / Panel 依赖了闭源 core" >&2
  echo "$hits" >&2
  fail=1
fi

if [ "$fail" -ne 0 ]; then
  exit 1
fi
echo "non_goals_scan_test: ok"

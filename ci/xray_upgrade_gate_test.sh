#!/usr/bin/env bash
# AC26：Xray 升级门禁——固定版本一致性（compat / xrayconf / go.mod）+ 真实二进制回归（可选）。
# 升级内置 Xray 时须同步更新本脚本的 EXPECTED_COMMIT 与 compat.yaml、go.mod、xrayconf.XrayVersion。
set -euo pipefail
cd "$(dirname "$0")/.."

EXPECTED_VERSION="26.11.9"
EXPECTED_COMMIT="676eaf66f8f0"

fail=0
note() { echo "xray_upgrade_gate_test: $*" >&2; fail=1; }

compat=$(awk -F'"' '/^xray:/{print $2}' compat/v1/compat.yaml)
[ "$compat" = "$EXPECTED_VERSION" ] || note "compat.yaml xray=$compat，期望 $EXPECTED_VERSION"

const=$(grep -rhoE 'XrayVersion[[:space:]]*=[[:space:]]*"[^"]+"' ../core/internal/xrayconf/*.go 2>/dev/null | head -1 | sed -E 's/.*"([^"]+)"/\1/')
[ "$const" = "$EXPECTED_VERSION" ] || note "xrayconf.XrayVersion=$const，期望 $EXPECTED_VERSION"

grep -q "xray-core v.*$EXPECTED_COMMIT" ../core/go.mod || note "core/go.mod 未固定到 commit $EXPECTED_COMMIT"

# 真实二进制回归（设置 XRAY_BIN / RELAXY_XRAY_BIN 时执行；否则跳过）。
XRAY="${XRAY_BIN:-${RELAYX_XRAY_BIN:-}}"
if [ -n "$XRAY" ] && [ -x "$XRAY" ]; then
  ( cd ../core && XRAY_BIN="$XRAY" RELAXY_XRAY_BIN="$XRAY" ./ci/xray-test.sh )
  ( cd ../core && XRAY_BIN="$XRAY" RELAXY_XRAY_BIN="$XRAY" ./ci/integration.sh )
else
  echo "xray_upgrade_gate_test: 未设置 XRAY_BIN，跳过真实 golden / 集成回归"
fi

if [ "$fail" -ne 0 ]; then
  exit 1
fi
echo "xray_upgrade_gate_test: ok"

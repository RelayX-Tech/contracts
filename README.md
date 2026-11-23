# RelayX Contracts

RelayX 开放合同仓库（**开源**）。

## 内容

- `core/v1`：`ConfigService`、`TunnelService`、`ToolsService`、`VersionService`、`HealthService`、`NodeRuntimeService`。
- `agent/v1`：`AgentControlService`。
- 生成的 Go stubs、错误码与版本矩阵。

## 目录

```text
contracts/
├── core/v1/            # core-control / core-node 合同（config / service / tunnel / tools / version / health / node）
├── agent/v1/           # Panel ↔ Agent 合同（独立定义，不跨包 import core.v1）
├── errors/v1/          # 跨服务错误码唯一来源（error-codes.yaml）
├── compat/v1/          # 合同 / core / schema / agent 兼容范围（xray 只读）
├── gen/go/{core,agent}/v1/  # protoc 生成物（入库，CI diff 校验）
└── scripts/{gen.sh,gen.ps1} # 固定版本生成脚本
```

## 生成

固定版本：protoc 3.19.4、protoc-gen-go v1.36.6、protoc-gen-go-grpc 1.5.1、goctl 1.10.2。

```bash
# Linux / macOS / CI
PROTO_INCLUDE="$(go env GOPATH)/include" ./scripts/gen.sh
# Windows
pwsh ./scripts/gen.ps1
```

生成物入库；CI 重新生成后 `git diff --exit-code`，合同与生成物不漂移。

## 边界

- 只包含合同（proto / 生成物），不包含实现。
- 合同独立 semver；Panel / Agent / core 声明兼容范围。
- Xray 版本不参与合同版本协商。
- 不手写 gRPC 类型；禁止复制 proto 到服务仓。

## 依赖约束

Panel / Agent 只依赖本仓库的合同，不 import 闭源 `core` module。

```bash
go test ./...              # 合同一致性 / 错误码 / 兼容矩阵
./ci/boundary.sh           # 依赖边界
./ci/non_goals_scan.sh     # 非目标扫描
./ci/breaking_change_gate.sh # 生成物一致性
```

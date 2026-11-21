# RelayX Contracts

RelayX 开放合同仓库（**开源**）。

## 内容

- `core/v1`：`ConfigService`、`TunnelService`、`ToolsService`、`VersionService`、`HealthService`、`NodeRuntimeService`。
- `agent/v1`：`AgentControlService`。
- 生成的 Go stubs、错误码与版本矩阵。

## 边界

- 只包含合同（proto / 生成物），不包含实现。
- 合同独立 semver；Panel / Agent / core 声明兼容范围。
- Xray 版本不参与合同版本协商。

## 依赖约束

Panel / Agent 只依赖本仓库的合同，不 import 闭源 `core` module。

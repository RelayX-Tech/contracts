# 固定版本生成 Go stubs（Windows 开发环境）。
# 生成物入库；CI 重新生成后 `git diff --exit-code` 校验一致。
$ErrorActionPreference = "Stop"
Set-Location (Join-Path $PSScriptRoot "..")

$protoInclude = if ($env:PROTO_INCLUDE) { $env:PROTO_INCLUDE } else { Join-Path (go env GOPATH) "include" }

$protos = @(Get-ChildItem core/v1/*.proto, agent/v1/*.proto | ForEach-Object { $_.FullName })

protoc -I . -I $protoInclude `
  --go_out=. --go_opt=module=github.com/relayx-tech/contracts `
  --go-grpc_out=. --go-grpc_opt=module=github.com/relayx-tech/contracts `
  $protos

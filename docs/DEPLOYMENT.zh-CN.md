# 部署指南

[English](DEPLOYMENT.md)

本文说明 Open Wallet Auth 的生产部署方式。认证服务建议作为独立系统部署，业务应用通过 API 调用认证能力，并通过 JWKS 本地校验 access token。

## 运行组件

- Open Wallet Auth 服务：Go HTTP API 和内置管理控制台。
- PostgreSQL：持久化身份、接入应用、会话、绑定关系和审计数据。
- Redis：在开启相关配置时，用于验证码存储、OAuth state 存储和限流。
- 管理控制台：由 `admin-web` 构建，并由 Go 服务在 `/` 路径提供访问。

## 构建

Docker 镜像会同时构建 Go API 和管理控制台：

```bash
docker build -t open-wallet-auth:local .
```

镜像内包含：

- `/app/open-wallet-auth`
- `/app/admin-web/dist`
- `/app/configs/config.example.yaml`
- `/app/migrations`

## 配置

从示例配置创建运行配置，并确保不要提交到 Git：

```bash
cp configs/config.example.yaml configs/config.yaml
```

生产环境需要重点配置：

- 高强度 `management.admin_token` 和管理后台管理员密码。
- 严格的 `cors.allowed_origins`；生产环境不要使用 `*`。
- 持久化 JWT 私钥/公钥路径或密钥内容。
- 开启短信/邮件验证时，配置真实短信和邮件服务商。
- OAuth 回调地址必须匹配认证服务公网域名，例如 `https://auth.example.com/api/v1/oauth/github/callback`。

## 数据库迁移

启动或升级服务前，先执行版本化 SQL 迁移：

```bash
go run ./cmd/migrate -direction up
```

容器化部署时，建议在一次性迁移任务或发布脚本中先执行迁移，再替换正在运行的服务实例。

## 启动

本地二进制启动：

```bash
go run ./cmd/server
```

容器启动：

```bash
docker run --rm \
  -p 8080:8080 \
  -v "$PWD/configs/config.yaml:/app/configs/config.yaml:ro" \
  open-wallet-auth:local
```

## 健康检查

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/.well-known/jwks.json
```

管理控制台访问地址：

```text
http://localhost:8080/
```

## 发布检查

- `go mod tidy` 无变更。
- `CGO_ENABLED=0 go test ./...` 通过。
- `go vet ./...` 通过。
- `staticcheck ./...` 通过。
- `admin-web` 目录下 `npm run build` 通过。
- `docker build -t open-wallet-auth:local .` 通过。
- 未提交运行配置、token、OAuth secret、数据库密码或私钥。

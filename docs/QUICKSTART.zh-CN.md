# 快速开始

这份文档只解决一件事：在本地把认证服务跑起来，并打开管理后台和统一登录页。

## 前置条件

- Go
- Docker 或可用的 PostgreSQL / Redis
- Node.js，仅在需要构建管理后台时使用

## 1. 准备配置

```bash
cp configs/config.example.yaml configs/config.yaml
```

本地开发可以先使用示例配置。生产环境不要直接使用示例密钥和默认管理凭证。

## 2. 启动依赖

```bash
docker compose up -d postgres redis
```

如果你不用 Docker，也可以自己准备 PostgreSQL 和 Redis，然后修改 `configs/config.yaml`。

## 3. 执行数据库迁移

```bash
go run ./cmd/migrate -direction up
```

迁移文件在 `migrations/` 目录。生产环境建议显式执行迁移，不建议启动时自动迁移。

## 4. 启动认证服务

```bash
OWA_HTTP_PORT=8081 go run ./cmd/server
```

健康检查：

```bash
curl http://localhost:8081/healthz
```

JWKS：

```bash
curl http://localhost:8081/.well-known/jwks.json
```

## 5. 打开管理后台

```text
http://localhost:8081/console/login
```

管理后台用于创建接入应用、查看身份用户、查看登录会话和调整运行期配置。

## 6. 打开统一登录页

```text
http://localhost:8081/login?client_id=default&return_uri=http://localhost:5173/callback
```

统一登录页是给业务系统用户使用的入口。业务系统应该跳转到这个地址，而不是让用户直接理解认证中台。

## 下一步

- 接入业务系统：[INTEGRATION.zh-CN.md](INTEGRATION.zh-CN.md)
- 配置短信、邮件、OAuth：[PROVIDERS.zh-CN.md](PROVIDERS.zh-CN.md)
- 看代码结构：[CODEBASE_GUIDE.zh-CN.md](CODEBASE_GUIDE.zh-CN.md)

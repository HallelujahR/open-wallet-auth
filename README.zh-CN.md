# Open Wallet Auth

[English](README.md)

Open Wallet Auth 是一个自托管统一认证服务，用来把多个业务系统的登录入口、第三方登录、钱包登录和 Token 签发集中管理起来。

它只负责“认证身份”和“签发凭证”，不接管业务系统自己的用户资料、权限、订单、内容或业务数据。

## 适合什么场景

- 已经有多个系统，希望统一登录和统一身份。
- 希望业务系统接入邮箱、手机号、Google、GitHub、钱包等登录方式。
- 希望业务后端通过 JWT/JWKS 本地校验用户身份。
- 希望认证能力独立部署，业务系统保持自己的数据模型。

## 你需要理解的最小模型

```text
业务系统前端
  -> 跳转到 Open Wallet Auth 统一登录页
  -> 用户完成登录
  -> 返回业务系统 callback
  -> 业务系统用 access token 换取或创建自己的本地登录态

业务系统后端
  -> 使用 JWKS 校验 access token
  -> 使用 JWT sub 关联自己的本地用户表
```

## 核心能力

- 邮箱密码注册和登录
- 手机验证码登录
- Google / GitHub OAuth 登录与绑定
- 浏览器钱包登录与绑定
- JWT access token、refresh token、JWKS
- 多业务系统 client 接入
- 登录日志、登录会话、安全操作审计
- 管理后台：应用、身份用户、会话、审计、运行配置
- 可视化配置登录页品牌、OAuth、短信、邮件、CORS 等运行期配置

## 正在使用

Open Wallet Auth 已被 [Lianxi Labs](https://lianxilabs.com/) 用于生产环境，为 [BlockX](https://blockx.lianxilabs.com) 和 [Label Service](https://label.lianxilabs.com) 等业务系统提供统一认证能力。

## 快速开始

详细步骤见：[快速开始](docs/QUICKSTART.zh-CN.md)

```bash
cp configs/config.example.yaml configs/config.yaml
docker compose up -d postgres redis
go run ./cmd/migrate -direction up
OWA_HTTP_PORT=8081 go run ./cmd/server
```

打开管理后台：

```text
http://localhost:8081/console/login
```

打开业务统一登录页：

```text
http://localhost:8081/login?client_id=default&return_uri=http://localhost:5173/callback
```

## 怎么接入一个业务系统

1. 在管理后台创建应用 client。
2. 配置允许的回调地址。
3. 业务前端把用户跳转到 `/login?client_id=...&return_uri=...`。
4. 用户登录成功后，认证服务把 access token 返回给业务回调页。
5. 业务后端校验 token，并用 `sub` 关联自己的用户。

完整说明见：[业务系统接入指南](docs/INTEGRATION.zh-CN.md)

如果希望少写协议代码，可以使用内置 SDK：

- [Web SDK](sdk/web/README.md)：统一登录跳转、回调解析、钱包和 OAuth 辅助方法。
- [Node SDK](sdk/node/README.md)：Node 后端调用登录、注册、profile 校验。
- [Go SDK](sdk/go/README.md)：Go 后端调用登录、注册、profile 校验。

## 文档导航

- [快速开始](docs/QUICKSTART.zh-CN.md)：本地启动和第一个登录流程。
- [接入指南](docs/INTEGRATION.zh-CN.md)：业务系统如何接入统一登录。
- [SDK 说明](sdk/README.md)：前端和后端 SDK 的职责边界。
- [配置说明](docs/PROVIDERS.zh-CN.md)：短信、邮件、OAuth、运行期配置。
- [部署指南](docs/DEPLOYMENT.zh-CN.md)：生产部署、迁移、健康检查。
- [代码导览](docs/CODEBASE_GUIDE.zh-CN.md)：目录结构和从哪里开始看代码。
- [架构说明](docs/ARCHITECTURE.zh-CN.md)：分层边界和依赖方向。
- [文档地图](docs/README.zh-CN.md)：每份文档负责什么。

## 技术栈

- Go
- PostgreSQL
- Redis
- JWT / JWKS
- React + Vite 管理后台
- Docker Compose 本地依赖

## 开发命令

```bash
go test ./...
staticcheck ./...
npm run build --prefix admin-web
```

## License

Apache-2.0

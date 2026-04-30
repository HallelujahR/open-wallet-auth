# Open Wallet Auth

[English](README.md)

Open Wallet Auth 是一个自托管的 Web2 + Web3 统一认证服务，适合需要账号密码登录、钱包签名登录、JWT/JWKS、多系统共享身份的应用。

认证服务只负责认证、签发 Token、暴露 JWKS。你的业务系统仍然保留自己的用户资料、权限、订单、内容和业务数据。

## 功能

- 邮箱密码注册和登录
- 邮箱验证码发送和校验
- 手机号验证码登录
- Redis 验证码存储和验证码发送/校验限流
- 密码登录和钱包 nonce 创建限流
- EVM 钱包签名登录，返回 SIWE-compatible 签名消息
- 已登录用户绑定钱包接口
- Google 和 GitHub OAuth start/callback 流程
- 已登录用户绑定 OAuth 账号授权流程
- RS256 签名的 JWT access token
- JWKS 公钥端点，业务系统可以本地校验 token
- Refresh token 持久化和轮换
- Refresh token 会话管理和吊销接口
- 已登录用户修改密码接口
- 邮箱验证码重置密码接口
- 密码重置后吊销已有 refresh token 会话
- 已登录用户绑定邮箱和手机号接口
- 用户侧解绑邮箱、手机号、钱包和 OAuth，并保护至少保留一种登录方式
- 当前用户资料查询和展示资料更新
- 通过 `client_id` 和 JWT audience 支持多系统接入
- 登录日志和用户-系统关系记录
- 密码登录失败审计记录
- 修改密码、绑定、解绑等安全操作审计记录
- 内部身份管理 API，支持用户、绑定关系和登录日志查询
- 安全操作审计事件管理 API
- 钱包和 OAuth 账号绑定的管理端解绑接口
- 生产环境配置安全检查
- 浏览器 CORS 配置
- 浏览器钱包登录示例
- Gin API JWT 校验示例

## 当前状态

项目已经不只是最小 MVP，适合做本地接入测试和架构验证。生产部署必须使用显式 SQL 迁移、强管理 token、真实短信/邮件服务商、严格 CORS 来源和持久化 JWT 密钥文件。

## 架构

项目采用 Clean Architecture 思路，明确区分：

- HTTP delivery
- usecase
- domain model
- repository interface
- infrastructure adapter

架构说明见：[docs/ARCHITECTURE.zh-CN.md](docs/ARCHITECTURE.zh-CN.md)

## 接入

- [接入指南](docs/INTEGRATION.zh-CN.md)
- [通用认证前端 Demo](examples/universal-auth-demo)
- [身份管理控制台 Demo](examples/admin-console)
- [短信和邮件服务商接入](docs/PROVIDERS.zh-CN.md)
- [浏览器钱包登录示例](examples/browser-wallet-login)
- [Gin API JWT 校验示例](examples/gin-api)

典型接入流程：

1. 为业务系统创建一个 client。
2. 前端请求钱包 nonce。
3. 调用钱包对返回的 message 签名。
4. 用签名换取 access token 和 refresh token。
5. 业务后端通过 JWKS 本地校验 access token。
6. 业务库使用 JWT 的 `sub` 作为 `auth_user_id` 关联本地用户资料。

## 快速启动

```bash
cp configs/config.example.yaml configs/config.yaml
docker compose up -d postgres redis
go run ./cmd/migrate -direction up
go run ./cmd/server
```

健康检查：

```bash
curl http://localhost:8080/healthz
```

JWKS 公钥：

```bash
curl http://localhost:8080/.well-known/jwks.json
```

## API 示例

账号注册：

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"client_id":"default","username":"alice","email":"alice@example.com","password":"password123"}'
```

账号登录：

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"client_id":"default","email":"alice@example.com","password":"password123"}'
```

当前用户：

```bash
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <access_token>"
```

当前用户资料：

```bash
curl http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer <access_token>"
```

更新展示资料：

```bash
curl -X PATCH http://localhost:8080/api/v1/profile \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer <access_token>" \
  -d '{"username":"alice_new","avatar":"https://example.com/avatar.png"}'
```

修改当前用户密码：

```bash
curl -X PATCH http://localhost:8080/api/v1/auth/password \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer <access_token>" \
  -d '{"current_password":"password123","new_password":"new-password123"}'
```

使用邮箱验证码重置密码：

```bash
curl -X POST http://localhost:8080/api/v1/auth/password/reset \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","code":"123456","new_password":"new-password123"}'
```

给当前用户绑定邮箱：

```bash
curl -X POST http://localhost:8080/api/v1/auth/bind/email \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer <access_token>" \
  -d '{"email":"alice@example.com","code":"123456"}'
```

给当前用户绑定手机号：

```bash
curl -X POST http://localhost:8080/api/v1/auth/bind/phone \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer <access_token>" \
  -d '{"phone":"+8613800000000","code":"123456"}'
```

给当前用户绑定 Google 或 GitHub 账号：

```bash
curl "http://localhost:8080/api/v1/oauth/github/bind/start?client_id=default&redirect_uri=http://localhost:8081/oauth/callback" \
  -H "Authorization: Bearer <access_token>"
```

解绑当前用户的登录方式：

```bash
curl -X DELETE http://localhost:8080/api/v1/auth/bind/email \
  -H "Authorization: Bearer <access_token>"

curl -X DELETE http://localhost:8080/api/v1/auth/wallets/<wallet_id> \
  -H "Authorization: Bearer <access_token>"
```

刷新 Token：

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H 'Content-Type: application/json' \
  -d '{"refresh_token":"<refresh_token>"}'
```

退出登录：

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H 'Content-Type: application/json' \
  -d '{"refresh_token":"<refresh_token>"}'
```

创建接入应用：

```bash
curl -X POST http://localhost:8080/api/v1/clients \
  -H 'Content-Type: application/json' \
  -H 'X-Admin-Token: dev-admin-token' \
  -d '{"client_id":"example-app","name":"Example App"}'
```

查询内部身份用户：

```bash
curl http://localhost:8080/api/v1/admin/users \
  -H 'X-Admin-Token: dev-admin-token'
```

创建钱包 nonce：

```bash
curl -X POST http://localhost:8080/api/v1/wallet/nonce \
  -H 'Content-Type: application/json' \
  -d '{"address":"0x0000000000000000000000000000000000000001","domain":"example.com","chain_id":1}'
```

校验钱包签名：

```bash
curl -X POST http://localhost:8080/api/v1/wallet/verify \
  -H 'Content-Type: application/json' \
  -d '{"client_id":"default","address":"<wallet_address>","nonce":"<nonce>","signature":"<signature>"}'
```

给当前用户绑定钱包：

```bash
curl -X POST http://localhost:8080/api/v1/wallet/bind \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer <access_token>" \
  -d '{"address":"<wallet_address>","nonce":"<nonce>","signature":"<signature>"}'
```

## 配置

示例配置在 [configs/config.example.yaml](configs/config.example.yaml)。

关键配置：

- `database.dsn`：PostgreSQL 连接地址
- `jwt.issuer`：JWT issuer，业务系统校验 token 时需要匹配
- `jwt.private_key_path`：RSA 私钥路径
- `jwt.public_key_path`：RSA 公钥路径
- `wallet.nonce_ttl`：钱包签名挑战有效期
- `wallet.rate_limit_*`：钱包 nonce 创建限流配置
- `auth.rate_limit_*`：密码登录限流配置
- `phone.code_ttl`：手机号验证码有效期
- `phone.code_store`：验证码存储方式，支持 `memory` 或 `redis`
- `phone.dev_code`：本地开发验证码
- `phone.rate_limit_*`：手机号验证码发送和校验限流配置
- `phone.enabled`：是否开启手机号验证码登录
- `phone.provider.*`：短信服务商配置，支持 `noop`、`webhook`、`aliyun_sms`
- `email.verification_enabled`：是否开启邮箱验证接口
- `email.code_store`：验证码存储方式，支持 `memory` 或 `redis`
- `email.rate_limit_*`：邮箱验证码发送和校验限流配置
- `email.provider.*`：邮件服务商配置，支持 `noop`、`webhook`、`smtp`
- `redis.enabled`：是否启用 Redis 适配器，用于验证码存储和限流
- `oauth.google.*`：Google OAuth 凭据和端点
- `oauth.github.*`：GitHub OAuth 凭据和端点
- `management.admin_token`：管理接口 token，生产环境必须使用强随机非默认值
- `http.cors_allowed_origins`：允许调用认证服务的浏览器来源

当 `app.env=production` 时，服务启动会拒绝危险配置，例如 `database.auto_migrate=true`、暴露开发验证码、启用登录/验证但仍使用 `noop` 短信或邮件服务商、通配/`null` CORS 来源、弱管理 token 或缺失 JWT 密钥文件。

## 测试

```bash
CGO_ENABLED=0 go test ./...
CGO_ENABLED=0 go vet ./...
CGO_ENABLED=0 go build ./cmd/server
CGO_ENABLED=0 go build ./cmd/migrate
```

## 后续计划

- 更正式的管理后台权限模型
- 更多框架接入示例

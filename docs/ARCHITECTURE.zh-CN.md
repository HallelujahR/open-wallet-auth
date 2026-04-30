# 架构说明

[English](ARCHITECTURE.md)

Open Wallet Auth 采用 Clean Architecture，核心目标是让认证业务、HTTP 交付、数据库、JWT、Redis、OAuth 服务商等外部依赖保持清晰边界。

## 依赖方向

```text
delivery/http -> usecase -> domain
infrastructure -> repository interfaces
```

`domain` 层不能导入 Gin、GORM、Redis、JWT 库或数据库包。`usecase` 层只依赖领域对象和仓储接口，不直接依赖具体基础设施实现。

## 目录结构

```text
cmd/server
  服务进程入口。负责加载配置、初始化日志并启动应用。

cmd/migrate
  生产数据库迁移命令。执行 migrations 目录中的版本化 SQL 文件。

internal/app
  依赖装配和应用生命周期管理。

internal/domain
  核心实体和值对象，不依赖框架。

internal/usecase
  业务用例编排。依赖 domain 和 repository 接口。

internal/repository
  持久化端口/接口定义。

internal/delivery/http
  HTTP handler、DTO、中间件、路由和统一响应。

internal/infrastructure
  PostgreSQL、JWT、哈希、日志、配置、消息服务商等外部适配器。

api
  OpenAPI 规范。

migrations
  SQL 数据库迁移脚本。

examples
  通用业务系统接入示例。
```

## 当前边界

- 密码登录、已登录修改密码、邮箱验证码重置密码位于 `internal/usecase/auth`。
- 钱包登录和已登录绑定钱包位于 `internal/usecase/wallet`；EVM 地址和签名细节隔离在 `internal/infrastructure/wallet`。
- 手机号登录位于 `internal/usecase/phone`；验证码存储通过 `repository.PhoneCodeRepository` 抽象。
- 邮箱验证位于 `internal/usecase/email`；短信/邮件发送通过 usecase provider port 抽象，并由 `internal/infrastructure/message` 实现。
- 验证码、密码登录、钱包 nonce 的限流通过 `repository.RateLimiter` 抽象；Redis 和 no-op 实现在 infrastructure 层。
- OAuth 登录位于 `internal/usecase/oauth`；服务商 HTTP 交换和 state 存储隔离在 `internal/infrastructure/oauth`。
- Client 管理和 JWT audience 动态解析位于 `internal/usecase/client`。
- 内部身份管理位于 `internal/usecase/admin`；它可以查看身份用户、登录活动、钱包绑定和 OAuth 绑定，但不拥有业务系统自己的资料或权限。
- Refresh token 持久化和会话吊销通过 `repository.RefreshTokenRepository` 抽象。
- 钱包绑定和一次性挑战值通过 `repository.WalletRepository` 抽象。
- 登录活动、用户-业务系统关系、安全操作审计通过 `repository.ActivityRepository` 抽象。
- JWT 签发、校验和 JWKS 生成位于 `internal/infrastructure/jwt`。
- HTTP handler 不直接访问数据库。
- CORS 由 HTTP 中间件按运行配置处理；业务 client 归属仍由 client usecase 管理。

## 已知边界与后续方向

- Refresh token rotation 已暴露为仓储端口，便于存储适配器实现原子化 revoke-and-create。
- 密码重置会通过 refresh-token 仓储端口吊销用户已有 refresh token 会话。
- 邮箱和手机号绑定遵循显式归属规则：未绑定值可绑定到当前用户，当前用户已绑定值幂等成功，其他用户已占用则拒绝。
- OAuth 登录不会再按 provider email 自动合并账号；已有身份需要在登录后显式绑定 OAuth 账号。
- 用户侧解绑有“至少保留一种登录方式”保护，避免账号失去所有登录入口。
- 资料更新只允许展示字段 `username`、`avatar`；邮箱和手机号仍走验证码绑定流程。
- 登录审计记录密码、refresh、钱包、手机号、OAuth 的成功登录；密码登录失败以尽力而为方式记录。
- 安全操作审计记录修改密码、重置密码、绑定邮箱/手机、解绑邮箱/手机/钱包/OAuth。
- 生产启动检查会在打开数据库连接前拒绝危险运行配置。
- Client 和身份管理接口当前使用 `X-Admin-Token` 保护；完整 admin/RBAC 模型暂未落地。

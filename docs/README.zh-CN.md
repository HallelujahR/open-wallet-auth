# 文档地图

[English](README.md)

本目录把产品说明、架构、业务接入、服务商配置、部署、发布检查分开维护。代码变更后，应更新对应主题文档，避免同一段说明散落在多处后互相冲突。

## 文档分工

| 文档 | 负责内容 | 不负责内容 |
| --- | --- | --- |
| [ARCHITECTURE.zh-CN.md](ARCHITECTURE.zh-CN.md) | 分层边界、依赖方向、目录结构、模块职责 | API 示例、部署命令、服务商密钥配置 |
| [INTEGRATION.zh-CN.md](INTEGRATION.zh-CN.md) | 业务系统如何调用认证服务、如何校验 JWT/JWKS | 内部实现细节、生产运维步骤 |
| [PROVIDERS.zh-CN.md](PROVIDERS.zh-CN.md) | 短信、邮件、OAuth 服务商配置模型 | 业务系统接入流程、数据库迁移 |
| [DEPLOYMENT.zh-CN.md](DEPLOYMENT.zh-CN.md) | 生产构建、运行配置、数据库迁移、启动、健康检查 | 架构设计理由、API 使用教程 |
| [OPEN_SOURCE_READINESS.zh-CN.md](OPEN_SOURCE_READINESS.zh-CN.md) | 发布前检查、CI、安全扫描、打包检查 | 运行时用户指南或功能教程 |

## 准确性规则

- 功能清单必须和 `internal/delivery/http/router` 中注册的 HTTP 路由一致。
- 架构说明必须和 `internal/app`、`internal/usecase`、`internal/domain`、`internal/repository`、`internal/infrastructure` 的依赖关系一致。
- 部署说明必须和 `Dockerfile`、`docker-compose.yml`、`cmd/server`、`cmd/migrate` 一致。
- 服务商文档必须和 `internal/infrastructure/message`、`internal/infrastructure/oauth`、`configs/config.example.yaml` 一致。
- 示例只能使用非真实密钥的占位值。

## 当前功能边界

- 邮箱密码注册和登录由 `/api/v1/auth/register`、`/api/v1/auth/login` 提供。
- 邮箱验证码由 `/api/v1/email/code`、`/api/v1/email/verify` 提供；邮箱验证码用于重置密码和绑定邮箱，不是独立的邮箱验证码登录方式。
- 手机号验证码登录由 `/api/v1/phone/code`、`/api/v1/phone/login` 提供。
- 钱包登录由 `/api/v1/wallet/nonce`、`/api/v1/wallet/verify` 提供；已登录绑定钱包使用 `/api/v1/wallet/bind`。
- OAuth 登录和绑定使用 `/api/v1/oauth/:provider/start`、`/api/v1/oauth/:provider/bind/start`、`/api/v1/oauth/:provider/callback`。
- 当前后端已实现用户资料、联系方式绑定、解绑、refresh token 会话、管理端身份管理、client 管理和审计日志。
- 正式 admin RBAC 模型尚未实现；管理接口当前依赖管理后台登录和 `X-Admin-Token`。

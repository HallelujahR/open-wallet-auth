# 文档地图

[English](README.md)

这个目录按“读者要完成的任务”组织，而不是按技术模块堆叠。第一次看项目时，建议按下面顺序阅读。

## 推荐阅读路径

1. [快速开始](QUICKSTART.zh-CN.md)  
   在本地启动服务，确认管理后台和统一登录页可用。

2. [业务系统接入指南](INTEGRATION.zh-CN.md)  
   理解 client、回调地址、access token、JWKS、业务本地用户之间的关系。

3. [SDK 说明](../sdk/README.md)  
   看 Web、Node、Go SDK 分别封装了什么，什么时候该用、什么时候不该用。

4. [配置说明](PROVIDERS.zh-CN.md)  
   配置 Google、GitHub、短信、邮件、CORS、登录页品牌等。

5. [部署指南](DEPLOYMENT.zh-CN.md)  
   生产环境如何迁移数据库、准备密钥、启动服务和做健康检查。

6. [代码导览](CODEBASE_GUIDE.zh-CN.md)  
   看代码时从哪里进、每个目录负责什么、一个请求如何流转。

7. [架构说明](ARCHITECTURE.zh-CN.md)  
   深入理解分层边界、依赖方向和扩展点。

## 文档分工

| 文档 | 解决的问题 |
| --- | --- |
| [QUICKSTART.zh-CN.md](QUICKSTART.zh-CN.md) | 本地启动和基础验证 |
| [INTEGRATION.zh-CN.md](INTEGRATION.zh-CN.md) | 业务系统怎么接入 |
| [SDK 说明](../sdk/README.md) | SDK 职责和选择 |
| [PROVIDERS.zh-CN.md](PROVIDERS.zh-CN.md) | 登录方式和服务商怎么配置 |
| [DEPLOYMENT.zh-CN.md](DEPLOYMENT.zh-CN.md) | 怎么上线和运维 |
| [CODEBASE_GUIDE.zh-CN.md](CODEBASE_GUIDE.zh-CN.md) | 怎么读代码 |
| [ARCHITECTURE.zh-CN.md](ARCHITECTURE.zh-CN.md) | 为什么这样分层 |
| [OPEN_SOURCE_READINESS.zh-CN.md](OPEN_SOURCE_READINESS.zh-CN.md) | 发布前检查项 |

## 文档维护原则

- README 只放项目定位、最短启动路径和文档导航。
- 接入流程只写在 Integration；SDK 只写封装职责和最小示例，避免重复解释业务流程。
- 代码结构只写在 Codebase Guide 和 Architecture。
- 服务商、密钥、开关只写在 Providers / Configuration 类文档。
- 英文文档优先保证 README、Quick Start、Integration 可读；深入实现细节以中文文档为准。

## 当前功能边界

- 邮箱密码注册和登录：`/api/v1/auth/register`、`/api/v1/auth/login`
- 手机验证码登录：`/api/v1/phone/code`、`/api/v1/phone/login`
- 钱包登录：`/api/v1/wallet/nonce`、`/api/v1/wallet/verify`
- OAuth 登录和绑定：`/api/v1/oauth/:provider/start`、`/api/v1/oauth/:provider/bind/start`、`/api/v1/oauth/:provider/callback`
- 用户资料、联系方式绑定、解绑、refresh token 会话、client 管理和审计日志已实现。
- 正式 admin RBAC 模型暂未实现；管理接口当前依赖管理后台登录和 `X-Admin-Token`。

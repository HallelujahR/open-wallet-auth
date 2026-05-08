# 代码导览

这份文档帮助你第一次读代码时知道从哪里开始。它不解释所有实现细节，只解释目录职责和主要调用链。

## 最短阅读路径

1. `cmd/server`：服务如何启动。
2. `internal/app`：依赖如何装配。
3. `internal/delivery/http/router`：HTTP 路由有哪些。
4. `internal/usecase`：核心业务流程在哪里。
5. `internal/infrastructure/postgres/repository`：数据如何落库。

## 目录职责

```text
cmd/
  server/      认证服务入口
  migrate/     数据库迁移入口

configs/
  配置模板、JWT 密钥路径等本地运行配置

migrations/
  PostgreSQL SQL 迁移文件

internal/app/
  组装配置、数据库、Redis、JWT、Usecase、HTTP Server

internal/domain/
  核心领域对象，比如用户、钱包、OAuth、token、审计事件

internal/usecase/
  业务流程编排，比如登录、注册、绑定、解绑、刷新 token、管理查询

internal/repository/
  仓储接口，也就是 usecase 依赖的数据访问端口

internal/infrastructure/
  外部实现，比如 PostgreSQL、Redis、JWT、短信、邮件、OAuth HTTP 客户端

internal/delivery/http/
  HTTP handler、DTO、中间件、路由和统一响应

admin-web/
  React + Vite 管理后台，同时包含业务统一登录页

examples/
  接入示例
```

## 一个登录请求怎么走

```text
POST /api/v1/auth/login
  -> internal/delivery/http/handler
  -> internal/usecase/auth
  -> internal/repository.UserRepository
  -> internal/infrastructure/postgres/repository
  -> internal/infrastructure/jwt
  -> HTTP response
```

HTTP 层只做参数解析和响应包装，真正的业务判断在 usecase。

## 一个业务系统怎么接入

```text
业务前端 /login
  -> 跳转到认证服务 /login?client_id=...&return_uri=...
  -> 用户完成登录
  -> 返回业务前端 /auth/oauth/callback#access_token=...
  -> 业务前端把 access token 交给业务后端
  -> 业务后端校验 token 并创建自己的登录态
```

认证中台不应该直接管理业务系统自己的角色、订单、内容、业务权限。它只提供身份凭证。

## 代码边界规则

- `domain` 不导入 Gin、GORM、Redis、JWT 等基础设施包。
- `usecase` 依赖 repository 接口，不直接操作数据库。
- `handler` 不直接查数据库。
- `infrastructure` 负责把接口落到具体技术实现。
- 管理后台配置项要通过 API 保存到运行期配置，不要硬编码公司品牌或服务商密钥。

## 常见修改从哪里开始

| 需求 | 推荐入口 |
| --- | --- |
| 新增一个 HTTP API | `internal/delivery/http/router` 和对应 handler |
| 新增业务流程 | `internal/usecase/<模块>` |
| 新增数据表 | `migrations/`、`internal/infrastructure/postgres/model`、repository 实现 |
| 新增短信/邮件服务商 | `internal/infrastructure/message` |
| 新增 OAuth 服务商 | `internal/infrastructure/oauth` 和 settings 配置 |
| 调整管理后台页面 | `admin-web/src/pages` |
| 调整统一登录页 | `admin-web/src/pages/UnifiedLoginPage.tsx` |

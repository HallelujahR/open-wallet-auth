# SDK

Open Wallet Auth provides small SDKs for common integration points. They do not replace the auth service; they wrap the stable HTTP contract so business systems do not need to hand-write login URLs, token callback parsing, profile validation, or identity migration calls.

Open Wallet Auth 提供轻量 SDK，用来封装业务系统接入认证中台时最常见的重复代码。SDK 不替代认证服务本身，而是把统一登录 URL、回调 token 解析、profile 校验、老用户迁移等固定协议收进清晰的方法里。

## Packages

| Package | Runtime | Use case |
| --- | --- | --- |
| `@open-wallet-auth/web` | Browser | Hosted login redirect, callback parsing, wallet/OAuth helper calls |
| `@open-wallet-auth/node` | Node.js service | Email/password identity login, registration, profile validation |
| `github.com/HallelujahR/open-wallet-auth/sdk/go` | Go service | Email/password identity login, registration, profile validation |

## Recommended Integration

1. Frontend uses `@open-wallet-auth/web` to redirect users to `/login` and consume the returned token.
2. Business backend exchanges the identity token for its own local business token.
3. Business backend uses the Node or Go SDK only for trusted service-side operations such as legacy user migration and profile validation.

## 推荐接入方式

1. 前端使用 `@open-wallet-auth/web` 跳转统一登录页，并解析回跳 token。
2. 业务后端用中台 token 换取自己的业务 token，继续保持自身业务权限和数据模型。
3. 业务后端只在可信服务端使用 Node 或 Go SDK，例如老用户迁移、账号注册、profile 校验。

## Repository Layout

```text
sdk/
  web/    Browser SDK, no framework dependency
  node/   Node.js service SDK, CommonJS output
  go/     Go service SDK, independent Go module
```

The SDKs are intentionally small and dependency-light. If a project needs a richer framework integration, build it as a thin wrapper on top of these packages instead of adding framework-specific logic here.

SDK 会保持小而稳定。如果某个业务系统需要 React/Vue/Nest/Gin 等框架级封装，建议在业务侧基于这些 SDK 再包一层薄适配，而不是把框架逻辑塞进通用 SDK。

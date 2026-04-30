# Gin API JWT Verification Example / Gin API JWT 校验示例

This example shows how a business API verifies access tokens issued by Open Wallet Auth.

The API does not call the auth service for every request. It fetches JWKS, caches public keys, and verifies JWT signature, issuer, audience, and expiry locally.

这个示例展示业务 API 如何校验 Open Wallet Auth 签发的 access token。

业务 API 不需要每次请求都回调认证服务。它会拉取 JWKS、缓存公钥，并在本地校验 JWT 签名、issuer、audience 和过期时间。

## Run / 运行

Start Open Wallet Auth first:

先启动 Open Wallet Auth：

```bash
go run ./cmd/server
```

Start the example API:

启动示例业务 API：

```bash
OWA_JWKS_URL=http://localhost:8080/.well-known/jwks.json \
OWA_ISSUER=open-wallet-auth \
OWA_AUDIENCE=default \
go run ./examples/gin-api
```

Call a public endpoint:

调用公开接口：

```bash
curl http://localhost:8090/public
```

Call a protected endpoint:

调用受保护接口：

```bash
curl http://localhost:8090/profile \
  -H "Authorization: Bearer <access_token>"
```

## What To Copy Into Your App / 可以复制到业务系统的代码

- `JWTMiddleware`
- `JWKSVerifier`
- `AuthClaims`

The protected handler can read the authenticated user from Gin context:

受保护的 handler 可以从 Gin context 中读取认证用户：

```go
claims := MustAuthClaims(c)
authUserID := claims.UserID
wallets := claims.Wallets
```

Use `authUserID` to create or find your local business profile record.

业务系统可以用 `authUserID` 创建或查找自己的本地业务用户资料。

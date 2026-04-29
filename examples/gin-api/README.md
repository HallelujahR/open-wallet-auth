# Gin API JWT Verification Example

This example shows how a business API verifies access tokens issued by Open Wallet Auth.

The API does not call the auth service for every request. It fetches JWKS, caches public keys, and verifies JWT signature, issuer, audience, and expiry locally.

## Run

Start Open Wallet Auth first:

```bash
go run ./cmd/server
```

Start the example API:

```bash
OWA_JWKS_URL=http://localhost:8080/.well-known/jwks.json \
OWA_ISSUER=open-wallet-auth \
OWA_AUDIENCE=default \
go run ./examples/gin-api
```

Call a public endpoint:

```bash
curl http://localhost:8090/public
```

Call a protected endpoint:

```bash
curl http://localhost:8090/profile \
  -H "Authorization: Bearer <access_token>"
```

## What To Copy Into Your App

- `JWTMiddleware`
- `JWKSVerifier`
- `AuthClaims`

The protected handler can read the authenticated user from Gin context:

```go
claims := MustAuthClaims(c)
authUserID := claims.UserID
wallets := claims.Wallets
```

Use `authUserID` to create or find your local business profile record.

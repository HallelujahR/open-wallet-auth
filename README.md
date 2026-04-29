# Open Wallet Auth

[简体中文](README.zh-CN.md)

Open Wallet Auth is a self-hosted Web2 + Web3 authentication service for applications that want password login, wallet signature login, JWT/JWKS, and shared identity across multiple systems.

The service owns authentication. Your business applications still own their own profiles, permissions, orders, content, and domain data.

## Features

- Email/password registration and login
- Phone verification-code login
- EVM wallet signature login with SIWE-compatible messages
- Google and GitHub OAuth start/callback flow
- JWT access tokens signed with RS256
- JWKS endpoint for local token verification in business APIs
- Refresh token persistence and rotation
- Multi-client login with `client_id` and JWT audience
- Login activity and user-client tracking
- Browser CORS configuration
- Browser wallet login example
- Gin API JWT verification example

## Status

This project is in early development. It is suitable for local integration testing and architecture validation. Production use still requires additional hardening such as rate limiting, failed-login auditing, operational migrations, and stronger management APIs.

## Architecture

The service follows Clean Architecture with explicit boundaries between:

- HTTP delivery
- usecases
- domain models
- repository interfaces
- infrastructure adapters

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for project layout and dependency rules.

## Integration

- [Chinese integration guide](docs/INTEGRATION.zh-CN.md)
- [Universal auth frontend demo](examples/universal-auth-demo)
- [Browser wallet login example](examples/browser-wallet-login)
- [Gin API JWT verification example](examples/gin-api)

Typical integration flow:

1. Create a client for your business application.
2. Use the browser wallet example or your own UI to request a nonce.
3. Ask the wallet to sign the returned message.
4. Exchange the signature for an access token and refresh token.
5. Verify access tokens locally in your business API through JWKS.
6. Use the JWT `sub` claim as `auth_user_id` in your own business database.

## Quick Start

```bash
cp configs/config.example.yaml configs/config.yaml
docker compose up -d postgres redis
go run ./cmd/server
```

Health check:

```bash
curl http://localhost:8080/healthz
```

JWKS:

```bash
curl http://localhost:8080/.well-known/jwks.json
```

## API Examples

Register:

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"client_id":"default","username":"alice","email":"alice@example.com","password":"password123"}'
```

Login:

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"client_id":"default","email":"alice@example.com","password":"password123"}'
```

Current user:

```bash
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <access_token>"
```

Refresh token:

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H 'Content-Type: application/json' \
  -d '{"refresh_token":"<refresh_token>"}'
```

Logout:

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H 'Content-Type: application/json' \
  -d '{"refresh_token":"<refresh_token>"}'
```

Create a client:

```bash
curl -X POST http://localhost:8080/api/v1/clients \
  -H 'Content-Type: application/json' \
  -H 'X-Admin-Token: dev-admin-token' \
  -d '{"client_id":"example-app","name":"Example App"}'
```

Create a wallet nonce:

```bash
curl -X POST http://localhost:8080/api/v1/wallet/nonce \
  -H 'Content-Type: application/json' \
  -d '{"address":"0x0000000000000000000000000000000000000001","domain":"example.com","chain_id":1}'
```

Verify a wallet signature:

```bash
curl -X POST http://localhost:8080/api/v1/wallet/verify \
  -H 'Content-Type: application/json' \
  -d '{"client_id":"default","address":"<wallet_address>","nonce":"<nonce>","signature":"<signature>"}'
```

## Configuration

Example configuration lives in [configs/config.example.yaml](configs/config.example.yaml).

Important settings:

- `database.dsn`: PostgreSQL DSN
- `jwt.issuer`: JWT issuer expected by business APIs
- `jwt.private_key_path`: RSA private key path
- `jwt.public_key_path`: RSA public key path
- `wallet.nonce_ttl`: wallet challenge lifetime
- `phone.code_ttl`: phone verification-code lifetime
- `phone.dev_code`: local development phone code
- `oauth.google.*`: Google OAuth credentials and endpoints
- `oauth.github.*`: GitHub OAuth credentials and endpoints
- `management.admin_token`: token for management APIs in development
- `http.cors_allowed_origins`: browser origins allowed to call the auth service

## Testing

```bash
CGO_ENABLED=0 go test ./...
CGO_ENABLED=0 go vet ./...
CGO_ENABLED=0 go build ./cmd/server
```

## Roadmap

- Rate limiting for login and nonce endpoints
- Failed-login auditing
- Production migration command
- Wallet binding and unbinding APIs
- Account linking between password users and wallet users
- User and wallet management APIs
- Stronger admin/RBAC model for service management
- More framework integration examples

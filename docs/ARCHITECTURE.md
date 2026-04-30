# Architecture

[简体中文](ARCHITECTURE.zh-CN.md)

Open Wallet Auth follows Clean Architecture with a small, explicit set of layers.

## Dependency Direction

```text
delivery/http -> usecase -> domain
infrastructure -> repository interfaces
```

The domain layer must not import Gin, GORM, Redis, JWT libraries, or database packages.

## Directories

```text
cmd/server
  Process entrypoint. It loads config, initializes logging, and starts the app.

cmd/migrate
  Production database migration command. It runs versioned SQL files from migrations.

internal/app
  Dependency wiring and application lifecycle.

internal/domain
  Core entities and value objects. No framework dependencies.

internal/usecase
  Business orchestration. It depends on domain types and repository interfaces.

internal/repository
  Ports/interfaces for persistence.

internal/delivery/http
  HTTP handlers, DTOs, middleware, routing, and response envelopes.

internal/infrastructure
  Adapters for PostgreSQL, JWT, hashing, logging, config, and other external dependencies.

api
  OpenAPI specification.

migrations
  SQL schema migrations.

examples
  Integration examples for generic applications.
```

## Current Boundaries

- Password auth, authenticated password change, and email-code password reset live in `internal/usecase/auth`.
- Wallet auth and authenticated wallet binding live in `internal/usecase/wallet`; EVM address and signature details are isolated in `internal/infrastructure/wallet`.
- Phone auth lives in `internal/usecase/phone`; verification-code storage is behind `repository.PhoneCodeRepository`.
- Email verification lives in `internal/usecase/email`; message delivery is behind usecase provider ports implemented by `internal/infrastructure/message`.
- Rate limiting for verification codes, password login, and wallet nonce creation is behind `repository.RateLimiter`; Redis and no-op implementations live in infrastructure.
- OAuth auth lives in `internal/usecase/oauth`; provider HTTP exchange and state storage are isolated in `internal/infrastructure/oauth`.
- Client management and dynamic audience resolution live in `internal/usecase/client`.
- Internal identity management lives in `internal/usecase/admin`; it can inspect identity users, login activity, wallet bindings, and OAuth bindings, but it does not own business-system profiles or permissions.
- Refresh token persistence and session revocation are behind `repository.RefreshTokenRepository`.
- Wallet bindings and one-time challenges are behind `repository.WalletRepository`.
- Login activity and user-client tracking are behind `repository.ActivityRepository`.
- Security operation audit is also behind `repository.ActivityRepository`; usecases record sensitive account changes without importing database adapters.
- JWT signing, verification, and JWKS generation live in `internal/infrastructure/jwt`.
- HTTP handlers do not access the database directly.
- Browser CORS is handled as HTTP middleware from runtime config; business client ownership still belongs to the client usecase.

## Known Gaps

- Refresh token rotation is exposed as a repository port so storage adapters can make revoke-and-create atomic.
- Password reset revokes existing refresh-token sessions through the refresh-token repository port.
- Email and phone binding follow explicit ownership rules: unbound values attach to the current user, current-user values are idempotent, and values owned by another user are rejected.
- OAuth login no longer auto-merges by provider email; users must explicitly bind OAuth accounts while authenticated when an existing identity should own the provider account.
- User-side unbinding is protected by a last-login-method check so an account cannot remove every way to sign in.
- Profile updates are limited to display fields (`username`, `avatar`); email and phone remain verification-code binding flows.
- Login auditing records successful password, refresh, wallet, phone, and OAuth logins; password-login failures are recorded as best-effort audit events.
- Security operation auditing records password changes, password resets, email/phone binding, and user-side unbinding of email, phone, wallet, and OAuth accounts.
- Production startup validation rejects unsafe runtime settings before opening database connections.
- Client and identity management are protected by `X-Admin-Token`; a first-class admin/RBAC model is still pending.

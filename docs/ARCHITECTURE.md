# Architecture

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

- Password auth lives in `internal/usecase/auth`.
- Wallet auth lives in `internal/usecase/wallet`; EVM address and signature details are isolated in `internal/infrastructure/wallet`.
- Phone auth lives in `internal/usecase/phone`; verification-code storage is behind `repository.PhoneCodeRepository`.
- OAuth auth lives in `internal/usecase/oauth`; provider HTTP exchange and state storage are isolated in `internal/infrastructure/oauth`.
- Client management and dynamic audience resolution live in `internal/usecase/client`.
- Refresh token persistence is behind `repository.RefreshTokenRepository`.
- Wallet bindings and one-time challenges are behind `repository.WalletRepository`.
- Login activity and user-client tracking are behind `repository.ActivityRepository`.
- JWT signing, verification, and JWKS generation live in `internal/infrastructure/jwt`.
- HTTP handlers do not access the database directly.
- Browser CORS is handled as HTTP middleware from runtime config; business client ownership still belongs to the client usecase.

## Known Gaps

- Refresh token rotation should be made transactional as the repository layer matures.
- Failed login auditing is still pending; current activity recording covers successful registration, login, refresh, and wallet login flows.
- Client management is protected by `X-Admin-Token`; a first-class admin/RBAC model is still pending.

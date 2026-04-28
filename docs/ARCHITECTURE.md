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
- Refresh token persistence is behind `repository.RefreshTokenRepository`.
- JWT signing, verification, and JWKS generation live in `internal/infrastructure/jwt`.
- HTTP handlers do not access the database directly.

## Known Gaps

- Refresh token rotation should be made transactional as the repository layer matures.
- Login logs and user-client activity tracking are still pending.
- Client management APIs are still pending; only the `default` client is bootstrapped.

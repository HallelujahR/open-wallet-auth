# Documentation Map

[简体中文](README.zh-CN.md)

This directory keeps product, architecture, integration, provider, deployment, and release-readiness documents separate. When code changes, update the document that owns the affected topic instead of duplicating the same explanation in multiple places.

## Ownership

| Document | Owns | Does not own |
| --- | --- | --- |
| [ARCHITECTURE.md](ARCHITECTURE.md) | Layer boundaries, dependency direction, project layout, module responsibilities | API examples, deployment commands, provider credentials |
| [INTEGRATION.md](INTEGRATION.md) | How a business application calls the auth service and verifies JWT/JWKS | Internal implementation details, production operations |
| [PROVIDERS.md](PROVIDERS.md) | SMS, email, and OAuth provider configuration model | Business-system integration flow, database migrations |
| [DEPLOYMENT.md](DEPLOYMENT.md) | Production build, runtime config, migration, start, health checks | Architecture rationale, API usage examples |
| [OPEN_SOURCE_READINESS.md](OPEN_SOURCE_READINESS.md) | Release checklist, CI, security scan, packaging checks | Runtime user guide or feature tutorial |

## Accuracy Rules

- Feature lists must match registered HTTP routes in `internal/delivery/http/router`.
- Architecture descriptions must match dependencies in `internal/app`, `internal/usecase`, `internal/domain`, `internal/repository`, and `internal/infrastructure`.
- Deployment instructions must match `Dockerfile`, `docker-compose.yml`, `cmd/server`, and `cmd/migrate`.
- Provider documentation must match `internal/infrastructure/message`, `internal/infrastructure/oauth`, and `configs/config.example.yaml`.
- Examples must use non-secret placeholder values only.

## Current Functional Boundaries

- Email/password registration and login are handled by `/api/v1/auth/register` and `/api/v1/auth/login`.
- Email verification codes are available through `/api/v1/email/code` and `/api/v1/email/verify`; email codes are used by password reset and email binding flows, not as a standalone email-code login flow.
- Phone code login is handled by `/api/v1/phone/code` and `/api/v1/phone/login`.
- Wallet login is handled by `/api/v1/wallet/nonce` and `/api/v1/wallet/verify`; authenticated wallet binding uses `/api/v1/wallet/bind`.
- OAuth login and binding use `/api/v1/oauth/:provider/start`, `/api/v1/oauth/:provider/bind/start`, and `/api/v1/oauth/:provider/callback`.
- User profile, contact binding, unbinding, refresh token sessions, admin identity management, client management, and audit logs are implemented in the current backend.
- A first-class admin RBAC model is not implemented yet; management APIs currently rely on management login and `X-Admin-Token`.

# Quick Start

This guide gets the auth service running locally and opens both the management console and hosted login page.

## Prerequisites

- Go
- Docker, or your own PostgreSQL / Redis
- Node.js, only if you need to build the management console

## 1. Prepare Configuration

```bash
cp configs/config.example.yaml configs/config.yaml
```

The example config is suitable for local development. Do not use default secrets or default management credentials in production.

## 2. Start Dependencies

```bash
docker compose up -d postgres redis
```

If you do not use Docker, prepare PostgreSQL and Redis yourself and update `configs/config.yaml`.

## 3. Run Migrations

```bash
go run ./cmd/migrate -direction up
```

Migration files live in `migrations/`. Production deployments should run migrations explicitly.

## 4. Start The Server

```bash
OWA_HTTP_PORT=8081 go run ./cmd/server
```

Health check:

```bash
curl http://localhost:8081/healthz
```

JWKS:

```bash
curl http://localhost:8081/.well-known/jwks.json
```

## 5. Open The Management Console

```text
http://localhost:8081/console/login
```

Use the console to manage clients, identities, sessions, audit logs, and runtime settings.

## 6. Open The Hosted Login Page

```text
http://localhost:8081/login?client_id=default&return_uri=http://localhost:5173/callback
```

This page is for end users of business applications. Business applications redirect users here.

## Next Steps

- [Integration](INTEGRATION.md)
- [Providers](PROVIDERS.md)
- [Architecture](ARCHITECTURE.md)

# Deployment Guide

[简体中文](DEPLOYMENT.zh-CN.md)

This guide describes a production-oriented deployment path for Open Wallet Auth. The service should be deployed as an independent authentication system; business applications call its APIs and verify access tokens through JWKS.

## Runtime Components

- Open Wallet Auth server: Go HTTP API and embedded admin console.
- PostgreSQL: persistent identity, client, session, binding, and audit data.
- `system_settings`: PostgreSQL table for admin-console editable OAuth/SMS/email provider settings.
- Redis: verification-code storage, OAuth state storage, and rate limiting when those options are enabled.
- Admin console: built from `admin-web` and served by the Go server at `/`.

## Build

The Docker image builds both the Go API and the admin console:

```bash
docker build -t open-wallet-auth:local .
```

The image includes:

- `/app/open-wallet-auth`
- `/app/admin-web/dist`
- `/app/configs/config.example.yaml`
- `/app/migrations`

## Configuration

Create a runtime config from the example and keep it outside Git:

```bash
cp configs/config.example.yaml configs/config.yaml
```

Production deployments should set:

- Strong `management.admin_token` and management admin password.
- Strict `cors.allowed_origins`; do not use `*` in production.
- Persistent JWT private/public key paths or key material.
- Real SMS and email providers when verification is enabled.
- OAuth callback URLs that match the public auth domain, for example `https://auth.example.com/api/v1/oauth/github/callback`.

The admin console can edit Google/GitHub OAuth credentials, SMS provider settings, email provider settings, and phone/email feature switches. Secret values are not returned by read APIs; submitting an empty secret keeps the existing value.

Startup-level settings such as database DSN, Redis address, JWT key paths, HTTP port, and production safety policy should remain in environment variables or `config.yaml`.

## Database Migration

Run versioned SQL migrations before starting or upgrading the service:

```bash
go run ./cmd/migrate -direction up
```

For containerized deployment, run the same command from a one-off migration job or a release script before replacing the running server.

## Start

Local binary:

```bash
go run ./cmd/server
```

Container:

```bash
docker run --rm \
  -p 8080:8080 \
  -v "$PWD/configs/config.yaml:/app/configs/config.yaml:ro" \
  open-wallet-auth:local
```

## Health Checks

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/.well-known/jwks.json
```

The admin console is available at:

```text
http://localhost:8080/
```

## Release Checklist

- `go mod tidy` has no diff.
- `CGO_ENABLED=0 go test ./...` passes.
- `go vet ./...` passes.
- `staticcheck ./...` passes.
- `npm run build` passes in `admin-web`.
- `docker build -t open-wallet-auth:local .` passes.
- No runtime config, token, OAuth secret, database password, or private key is committed.

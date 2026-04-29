# Open Wallet Auth

Open Wallet Auth is a self-hosted authentication service for Web2 and Web3 apps.

It is designed to support password login, wallet signature login, JWT/JWKS, and multi-app SSO.

## Status

This project is in early development.

## Architecture

The service follows Clean Architecture with explicit boundaries between:

- HTTP delivery
- usecases
- domain models
- repositories
- infrastructure adapters

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for the current project layout and dependency rules.

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

Wallet nonce:

```bash
curl -X POST http://localhost:8080/api/v1/wallet/nonce \
  -H 'Content-Type: application/json' \
  -d '{"address":"0x0000000000000000000000000000000000000001","domain":"example.com","chain_id":1}'
```

Wallet login:

```bash
curl -X POST http://localhost:8080/api/v1/wallet/verify \
  -H 'Content-Type: application/json' \
  -d '{"client_id":"default","address":"<wallet_address>","nonce":"<nonce>","signature":"<signature>"}'
```

Refresh token:

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
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

## Roadmap

- Password login
- JWT RS256 and JWKS
- Refresh token rotation
- EVM wallet SIWE-compatible login
- Multi-client audience support
- Go Gin integration example
- NestJS integration example

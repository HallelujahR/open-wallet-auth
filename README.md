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

## Roadmap

- Password login
- JWT RS256 and JWKS
- Refresh token rotation
- EVM wallet SIWE login
- Multi-client audience support
- Go Gin integration example
- NestJS integration example

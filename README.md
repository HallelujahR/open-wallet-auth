# Open Wallet Auth

[中文](README.zh-CN.md)

Open Wallet Auth is a self-hosted identity service for applications that need centralized login, OAuth, wallet authentication, JWT issuing, and JWKS verification.

It owns authentication and tokens. Your business applications still own their profiles, permissions, orders, content, and product data.

## When To Use It

- You have multiple applications and want one shared login service.
- You want email, phone, Google, GitHub, and wallet login behind one hosted login page.
- Your backend services should verify JWTs locally through JWKS.
- You want authentication deployed separately from business systems.

## Minimal Model

```text
Business frontend
  -> redirects to Open Wallet Auth hosted login
  -> user signs in
  -> auth service redirects back to the business callback
  -> business app creates its own local session

Business backend
  -> verifies access token through JWKS
  -> maps JWT sub to its local user record
```

## Features

- Email/password registration and login
- Phone verification-code login
- Google and GitHub OAuth login and binding
- Browser wallet login and binding
- JWT access tokens, refresh tokens, and JWKS
- Multi-application client registry
- Login logs, sessions, and security audit events
- Management console for applications, identities, sessions, audit logs, and runtime settings
- Runtime configuration for hosted-login branding, OAuth, SMS, email, and CORS

## Production Usage

Open Wallet Auth is used in production by [Lianxi Labs](https://lianxilabs.com/) to provide centralized authentication for business systems including [BlockX](https://blockx.lianxilabs.com) and [Label Service](https://label.lianxilabs.com).

## Quick Start

See the full guide: [Quick Start](docs/QUICKSTART.md)

```bash
cp configs/config.example.yaml configs/config.yaml
docker compose up -d postgres redis
go run ./cmd/migrate -direction up
OWA_HTTP_PORT=8081 go run ./cmd/server
```

Management console:

```text
http://localhost:8081/console/login
```

Hosted login page:

```text
http://localhost:8081/login?client_id=default&return_uri=http://localhost:5173/callback
```

## Integration Flow

1. Create a client application in the management console.
2. Configure allowed redirect URIs.
3. Redirect users to `/login?client_id=...&return_uri=...`.
4. Receive the access token on the business callback page.
5. Verify the token and map `sub` to your local user.

Full guide: [Integration](docs/INTEGRATION.md)

## Documentation

- [Quick Start](docs/QUICKSTART.md)
- [Integration](docs/INTEGRATION.md)
- [Providers and Runtime Configuration](docs/PROVIDERS.md)
- [Deployment](docs/DEPLOYMENT.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Documentation Map](docs/README.md)

The Chinese documentation is currently more detailed for implementation notes:

- [中文文档地图](docs/README.zh-CN.md)
- [代码导览](docs/CODEBASE_GUIDE.zh-CN.md)

## Tech Stack

- Go
- PostgreSQL
- Redis
- JWT / JWKS
- React + Vite management console
- Docker Compose for local dependencies

## Development

```bash
go test ./...
staticcheck ./...
npm run build --prefix admin-web
```

## License

Apache-2.0

# Release Readiness Checklist

[简体中文](OPEN_SOURCE_READINESS.zh-CN.md)

This checklist is for release preparation. It does not define runtime architecture or integration behavior.

## Repository

- `README.md` and `README.zh-CN.md` are both updated.
- `LICENSE`, `CONTRIBUTING.md`, `SECURITY.md`, issue templates, and PR template exist.
- Documentation ownership is clear in [docs/README.md](README.md).

## Engineering

```bash
go mod tidy
git diff --exit-code -- go.mod go.sum
test -z "$(gofmt -l .)"
CGO_ENABLED=0 go test ./...
CGO_ENABLED=0 go vet ./...
staticcheck ./...
npm run build
```

## Security

```bash
govulncheck ./...
gitleaks detect --source .
```

The latest local `govulncheck` run reported no reachable vulnerabilities after dependency upgrades. `gitleaks` still needs to be run in an environment where the tool is installed.

## Packaging

- Dockerfile uses a multi-stage build.
- The image contains the Go server, `admin-web/dist`, example config, and migrations.
- Runtime secrets and generated local config files are ignored by Git.

## Documentation

- Feature lists match routes in `internal/delivery/http/router`.
- Architecture descriptions match the code boundaries.
- Deployment commands match `Dockerfile`, `cmd/server`, and `cmd/migrate`.
- Provider examples use placeholders, not real secrets.

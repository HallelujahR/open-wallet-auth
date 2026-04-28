# Contributing

Thanks for helping improve Open Wallet Auth.

## Development

```bash
cp configs/config.example.yaml configs/config.yaml
go test ./...
go run ./cmd/server
```

## Commit Messages

Use Conventional Commits:

```text
feat(auth): add password login
fix(wallet): reject reused nonce
docs(readme): add docker quick start
```

## Pull Requests

Every PR should include:

- Purpose
- Main changes
- Test result
- Database migration impact
- API impact
- Security considerations

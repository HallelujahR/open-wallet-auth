# Contributing / 贡献指南

Thanks for helping improve Open Wallet Auth.

感谢你帮助改进 Open Wallet Auth。

## Development / 本地开发

```bash
cp configs/config.example.yaml configs/config.yaml
go test ./...
go run ./cmd/server
```

## Commit Messages / 提交信息

Use Conventional Commits:

请使用 Conventional Commits：

```text
feat(auth): add password login
fix(wallet): reject reused nonce
docs(readme): add docker quick start
```

## Pull Requests / Pull Request 要求

Every PR should include:

每个 PR 应包含：

- Purpose
- Main changes
- Test result
- Database migration impact
- API impact
- Security considerations
- 目的
- 主要改动
- 测试结果
- 数据库迁移影响
- API 影响
- 安全注意事项

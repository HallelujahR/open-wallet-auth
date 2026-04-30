# 开源发布收敛检查

本文记录开源发布前应持续执行的检查项，适合作为 `v0.x` 发布前的人工 checklist。

## 已覆盖

- 根目录存在 `LICENSE`，当前使用 Apache 2.0。
- 存在 `CONTRIBUTING.md`、`SECURITY.md`、`CODE_OF_CONDUCT.md`、`CHANGELOG.md`。
- `.github` 下存在 Issue / PR 模板。
- README 已提供中英双版本、CI/Go Report/License 徽章、快速启动和 Mermaid 架构图。
- 已提供 `Dockerfile` 和 `docker-compose.yml`。
- GitHub Actions 已执行 `go mod tidy` 检查、`gofmt`、`go vet`、`staticcheck`、`go test`、race test 和 Docker build。
- `go mod tidy` 当前无变更。
- 已移除重量级 `github.com/ethereum/go-ethereum` 依赖，钱包签名校验改为轻量 `secp256k1` + Keccak 实现。
- 架构依赖扫描未发现 `domain/usecase/repository` 反向依赖 `infrastructure`、Gin、GORM、Redis 等实现细节。

## 本地已执行

```bash
go mod tidy
staticcheck ./...
CGO_ENABLED=0 GOCACHE=/tmp/open-wallet-auth-gocache GOTMPDIR=/tmp go test ./...
git diff --check
```

## 仍需正式工具复核

当前本机未安装以下工具，因此本轮只做了内置 grep 和已有工具检查：

- `govulncheck`
- `gitleaks`
- `trufflehog`
- `golangci-lint`

正式发布前建议执行：

```bash
govulncheck ./...
gitleaks detect --source .
trufflehog git file://. --only-verified
golangci-lint run
```

## 已知风险

- 当前依赖图仍受 Gin/GORM/Viper 等框架影响，但已避免为钱包验签引入完整以太坊客户端依赖。
- Git 历史中曾出现旧私有测试 audience 名称。它不是密钥，但如果目标是完全干净的公开历史，应使用历史重写或 orphan public branch 单独处理。
- `examples/universal-auth-demo/index.html` 是无构建工具的单文件 Demo，便于复制和打开，但长期维护时可以拆分为 `style.css` 和 `app.js`。

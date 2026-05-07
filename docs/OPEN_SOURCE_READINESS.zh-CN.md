# 发布准备检查

本文记录版本发布前应持续执行的检查项，适合作为 `v0.x` 发布前的人工 checklist。本文只负责发布质量检查，不定义运行架构或业务接入行为。

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
git diff --exit-code -- go.mod go.sum
test -z "$(gofmt -l .)"
CGO_ENABLED=0 go test ./...
CGO_ENABLED=0 go vet ./...
staticcheck ./...
govulncheck ./...
npm run build
git diff --check
```

## 仍需正式工具复核

当前本机未安装以下工具，因此本轮没有完成 Git 历史密钥扫描和 golangci-lint 聚合检查：

- `gitleaks`
- `trufflehog`
- `golangci-lint`

正式发布前建议补充执行：

```bash
gitleaks detect --source .
trufflehog git file://. --only-verified
golangci-lint run
```

最近一次本地 `govulncheck ./...` 已通过，可达漏洞为 0。

## 已知风险

- 当前依赖图仍受 Gin/GORM/Viper 等框架影响，但核心业务层没有反向依赖 Gin、GORM、Redis 等基础设施实现。
- Git 历史扫描仍需使用 `gitleaks` 或 `trufflehog` 在正式发布前复核。
- `examples/universal-auth-demo/index.html` 是无构建工具的单文件 Demo，便于复制和打开，但长期维护时可以拆分为 `style.css` 和 `app.js`。

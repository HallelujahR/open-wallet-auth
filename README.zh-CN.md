# Open Wallet Auth

Open Wallet Auth 是一个独立、开源、自托管的 Web2 + Web3 统一认证服务。

目标能力：

- 账号密码登录
- 钱包签名登录
- JWT / JWKS
- 多应用共享登录身份
- 用户与钱包地址绑定

架构说明见：[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)

## 当前状态

项目处于早期开发阶段。

## 快速启动

```bash
cp configs/config.example.yaml configs/config.yaml
docker compose up -d postgres redis
go run ./cmd/server
```

健康检查：

```bash
curl http://localhost:8080/healthz
```

JWKS 公钥：

```bash
curl http://localhost:8080/.well-known/jwks.json
```

账号注册：

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"client_id":"default","username":"alice","email":"alice@example.com","password":"password123"}'
```

账号登录：

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"client_id":"default","email":"alice@example.com","password":"password123"}'
```

当前用户：

```bash
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <access_token>"
```

刷新 Token：

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H 'Content-Type: application/json' \
  -d '{"refresh_token":"<refresh_token>"}'
```

退出登录：

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H 'Content-Type: application/json' \
  -d '{"refresh_token":"<refresh_token>"}'
```

创建接入应用：

```bash
curl -X POST http://localhost:8080/api/v1/clients \
  -H 'Content-Type: application/json' \
  -H 'X-Admin-Token: dev-admin-token' \
  -d '{"client_id":"example-app","name":"Example App"}'
```

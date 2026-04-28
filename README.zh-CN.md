# Open Wallet Auth

Open Wallet Auth 是一个独立、开源、自托管的 Web2 + Web3 统一认证服务。

目标能力：

- 账号密码登录
- 钱包签名登录
- JWT / JWKS
- 多应用共享登录身份
- 用户与钱包地址绑定

## 当前状态

项目处于早期开发阶段。

## 快速启动

```bash
cp configs/config.example.yaml configs/config.yaml
go run ./cmd/server
```

健康检查：

```bash
curl http://localhost:8080/healthz
```

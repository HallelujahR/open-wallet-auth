# 接入指南

这份文档描述业务系统如何接入 Open Wallet Auth。认证服务只负责登录、签发 Token、暴露 JWKS；业务系统仍然保留自己的业务用户资料、订单、内容、权限等数据。

## 接入模型

```text
Browser UI
  -> Open Wallet Auth: 创建 nonce、校验签名、获取 JWT
  -> Business API: 携带 Authorization: Bearer <access_token>

Business API
  -> Open Wallet Auth JWKS: 获取公钥
  -> Local JWT Middleware: 校验 issuer、audience、signature、expiry
  -> Business Logic: 使用 user_id/client_id/wallets 做业务处理
```

## 前端钱包登录流程

1. 检查浏览器是否存在 `window.ethereum`。
2. 调用 `eth_requestAccounts` 获取钱包地址。
3. 调用认证服务 `POST /api/v1/wallet/nonce`，传入钱包地址、当前域名、链 ID。
4. 使用钱包对返回的 `message` 调用 `personal_sign`。
5. 调用认证服务 `POST /api/v1/wallet/verify`，传入 `client_id`、地址、nonce、签名。
6. 保存返回的 `access_token` 和 `refresh_token`。
7. 调用业务接口时，在请求头中携带 `Authorization: Bearer <access_token>`。

## 业务系统需要做什么

业务系统不需要保存用户密码，也不需要自己验证钱包签名。

业务系统需要做三件事：

- 创建自己的 `client_id`，例如 `example-app`。
- 在 API 网关或后端中校验 JWT。
- 用 JWT 中的 `sub` 作为认证服务用户 ID，用 `client_id` 区分登录来源，用 `wallets` 读取已验证的钱包地址。

## 用户表怎么处理

认证服务的 `users` 表只表示统一身份，不等于业务系统的完整用户资料表。

推荐业务系统保留自己的轻量用户资料表：

```sql
CREATE TABLE app_users (
  id VARCHAR(64) PRIMARY KEY,
  auth_user_id VARCHAR(64) NOT NULL UNIQUE,
  display_name VARCHAR(128),
  avatar VARCHAR(512),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

首次收到一个合法 JWT 时，如果本业务库没有对应的 `auth_user_id`，可以自动创建一条业务资料记录。

## 多系统登录识别

Open Wallet Auth 通过 `client_id` 识别用户登录的是哪个系统。

- `client_id=default`：默认应用
- `client_id=example-app`：示例业务系统
- `audience`：写入 JWT 的受众，业务系统校验 JWT 时必须匹配
- `user_clients`：认证服务侧记录用户和业务系统的登录关系

这样两个系统可以共用一套登录服务，但每个系统仍然能知道“这个用户是登录到我这里”。

## 推荐落地顺序

1. 部署认证服务和 PostgreSQL。
2. 创建业务系统对应的 client。
3. 在业务系统前端接入钱包登录按钮。
4. 在业务系统后端加入 JWT middleware。
5. 根据 JWT 的 `sub` 自动创建或关联本业务用户资料。
6. 后续再补用户中心、钱包绑定管理、后台管理权限。

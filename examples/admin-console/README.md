# Admin Console Demo / 身份管理控制台 Demo

Lightweight internal console for identity management APIs.

这是一个轻量级内部身份管理控制台 Demo，用于调用统一身份管理接口。

## 运行

Open `index.html` in a browser, then set:

在浏览器中打开 `index.html`，然后设置：

- API base URL, for example `http://localhost:8080`
- Admin token, matching `OWA_MANAGEMENT_ADMIN_TOKEN`
- API 基础地址，例如 `http://localhost:8080`
- 管理 Token，对应 `OWA_MANAGEMENT_ADMIN_TOKEN`

The demo calls:

该 Demo 会调用：

- `GET /api/v1/admin/users`
- `GET /api/v1/admin/users/{user_id}`
- `PATCH /api/v1/admin/users/{user_id}/status`
- `GET /api/v1/admin/login-logs`

This page is only a demo for internal operations. Production deployments should protect it behind company SSO, VPN, or an internal gateway.

该页面只用于内部运营/管理演示。生产环境应放在公司 SSO、VPN 或内部网关之后，不应直接暴露到公网。

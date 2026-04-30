# Admin Console Demo

Lightweight internal console for identity management APIs.

## Run

Open `index.html` in a browser, then set:

- API base URL, for example `http://localhost:8080`
- Admin token, matching `OWA_MANAGEMENT_ADMIN_TOKEN`

The demo calls:

- `GET /api/v1/admin/users`
- `GET /api/v1/admin/users/{user_id}`
- `PATCH /api/v1/admin/users/{user_id}/status`
- `GET /api/v1/admin/login-logs`

This page is only a demo for internal operations. Production deployments should protect it behind company SSO, VPN, or an internal gateway.

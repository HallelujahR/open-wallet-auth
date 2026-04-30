# Universal Auth Demo / 通用认证 Demo

This is a browser-only developer demo for Open Wallet Auth.

It is designed to show how a product frontend can wire multiple login methods, profile APIs, binding flows, token operations, and security-audit queries into one page.

这是一个纯浏览器开发者 Demo，用来展示业务前端如何在一个页面中接入多种登录方式、用户资料接口、账号绑定/解绑、token 操作和安全审计查询。

## Supported Now / 当前支持

- Email registration
- Email login
- Email verification code
- Phone code login with the local development code
- EVM wallet login
- Wallet provider selection through EIP-6963 when the browser exposes multiple wallets
- Google OAuth start/callback when provider credentials are configured
- GitHub OAuth start/callback when provider credentials are configured
- Current-user profile loading and update
- Email, phone, wallet, and OAuth binding/unbinding
- Refresh/logout token operations
- Security event list through the management API

- 邮箱注册
- 邮箱登录
- 邮箱验证码发送和校验
- 使用本地开发验证码的手机号登录
- EVM 钱包登录
- 浏览器暴露多个钱包时，通过 EIP-6963 选择钱包
- 配置服务商凭据后的 Google OAuth start/callback
- 配置服务商凭据后的 GitHub OAuth start/callback
- 当前用户资料加载和更新
- 邮箱、手机号、钱包、OAuth 绑定和解绑
- refresh/logout token 操作
- 通过管理接口查询安全操作审计事件

## Run / 运行

Start Open Wallet Auth:

启动 Open Wallet Auth：

```bash
OWA_HTTP_PORT=8081 go run ./cmd/server
```

Serve the examples directory:

启动 examples 静态服务：

```bash
python3 -m http.server 5173
```

Open in Chrome or another browser with a wallet extension:

在安装了钱包扩展的 Chrome 或其他浏览器中打开：

```text
http://localhost:5173/examples/universal-auth-demo/
```

Use this Auth Base URL:

页面中使用这个认证服务地址：

```text
http://localhost:8081
```

The Codex in-app browser cannot load wallet extensions, so wallet login must be tested in a normal browser with MetaMask or another EIP-1193 wallet installed.

For local phone and email verification, the default development code is `123456`.

Codex 内置浏览器无法加载钱包扩展，所以钱包登录需要在安装 MetaMask 或其他 EIP-1193 钱包的普通浏览器中测试。

本地手机号和邮箱验证码默认开发验证码是 `123456`。

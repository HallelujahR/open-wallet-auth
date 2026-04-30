# Browser Wallet Login Example / 浏览器钱包登录示例

This is a minimal browser-only example for EVM wallet login.

It does not use Next.js, Vite, or any build tool. Open `index.html` in a browser with MetaMask or another EIP-1193 wallet installed.

这是一个最小化的纯浏览器 EVM 钱包登录示例。

它不依赖 Next.js、Vite 或任何构建工具。请在安装了 MetaMask 或其他 EIP-1193 钱包的浏览器中打开 `index.html`。

## Run / 运行

Start Open Wallet Auth first:

先启动 Open Wallet Auth：

```bash
go run ./cmd/server
```

Then open:

然后打开：

```text
examples/browser-wallet-login/index.html
```

Default settings:

默认配置：

- Auth base URL: `http://localhost:8080`
- Client ID: `default`
- Chain ID: `1`
- 认证服务地址：`http://localhost:8080`
- 业务系统 Client ID：`default`
- 链 ID：`1`

## Flow / 流程

1. Request wallet accounts with `eth_requestAccounts`.
2. Create a nonce with `POST /api/v1/wallet/nonce`.
3. Ask the wallet to sign the returned message with `personal_sign`.
4. Verify the signature with `POST /api/v1/wallet/verify`.
5. Store the returned token pair in `localStorage`.

1. 使用 `eth_requestAccounts` 请求钱包账号。
2. 使用 `POST /api/v1/wallet/nonce` 创建 nonce。
3. 调用钱包用 `personal_sign` 签名后端返回的 message。
4. 使用 `POST /api/v1/wallet/verify` 校验签名。
5. 将返回的 token pair 保存到 `localStorage`。

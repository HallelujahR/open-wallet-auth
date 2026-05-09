# Open Wallet Auth Web SDK

Browser SDK for hosted-login integrations. It keeps frontend projects from duplicating login URL construction, redirect-state storage, callback hash parsing, wallet verification calls, and OAuth start calls.

浏览器端 SDK，用于业务前端接入统一认证登录页。它负责统一登录地址生成、登录前页面保存、回调 token 解析、钱包签名请求和 OAuth 启动请求，业务页面不需要再手写这些协议细节。

## Install

```bash
npm install @open-wallet-auth/web
```

For local monorepo development before publishing:

```json
{
  "dependencies": {
    "@open-wallet-auth/web": "file:../open-wallet-auth/sdk/web"
  }
}
```

## Basic Usage

```ts
import { createAuthClient } from "@open-wallet-auth/web";

export const auth = createAuthClient({
  authBaseURL: "https://auth.example.com",
  clientID: "my-app",
  returnURI: `${window.location.origin}/auth/oauth/callback`,
});

auth.login();
```

## Callback Page

```ts
const returned = auth.parseCallback(window.location.hash);

if (returned) {
  const businessToken = await exchangeIdentityToken(returned.accessToken);
  const redirect = auth.consumeRedirect("/");
  window.location.replace(redirect);
}
```

## Notes

- `authBaseURL` points to the deployed auth service.
- `clientID` must match the application configured in the auth console.
- `returnURI` must be allowed by the auth service application settings.
- The SDK stores only the post-login redirect path by default; business tokens should still be managed by the business system.

## 说明

- `authBaseURL` 指向认证中台服务地址。
- `clientID` 必须和中台管理后台配置的接入应用一致。
- `returnURI` 必须在接入应用允许的回调地址中。
- SDK 默认只保存登录完成后的回跳路径；业务 token 仍应由业务系统自己管理。

# Open Wallet Auth Node SDK

Service-side SDK for exchanging and validating Open Wallet Auth identities.

服务端 SDK，用于业务后端调用认证中台登录、注册和 profile 校验。

## Install

```bash
npm install @open-wallet-auth/node
```

For local monorepo development before publishing:

```json
{
  "dependencies": {
    "@open-wallet-auth/node": "file:../open-wallet-auth/sdk/node"
  }
}
```

## Usage

```ts
import { createIdentityClient } from "@open-wallet-auth/node";

const identity = createIdentityClient({
  authBaseURL: process.env.OPEN_WALLET_AUTH_URL!,
  clientID: process.env.OPEN_WALLET_AUTH_CLIENT_ID!,
});

const result = await identity.login({
  email: "user@example.com",
  password: "password",
});

const profile = await identity.profile(result.token.access_token);
```

## Notes

- Use this SDK only in trusted server-side code.
- Do not expose service-side configuration or profile validation logic to the browser.
- Business systems should still issue their own local token after validating the identity token.

## 说明

- 这个 SDK 只应该在可信后端使用。
- 不要把服务端配置、profile 校验逻辑暴露给浏览器。
- 业务系统校验中台身份后，仍建议签发自己的本地业务 token。

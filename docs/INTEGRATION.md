# Integration Guide

[简体中文](INTEGRATION.zh-CN.md)

This document explains how a business application integrates with Open Wallet Auth. The auth service owns authentication, token issuance, and JWKS. Business applications still own business profiles, permissions, orders, content, and domain data.

## Integration Model

```text
Browser UI
  -> Open Wallet Auth: register, login, create wallet nonce, verify signature, start OAuth, receive JWT
  -> Business API: send Authorization: Bearer <access_token>

Business API
  -> Open Wallet Auth JWKS: fetch public keys
  -> Local JWT Middleware: verify issuer, audience, signature, expiry
  -> Business Logic: use sub/client_id as identity references
```

## What the Business System Must Do

- Create a `client_id` for the application, for example `example-app`.
- Add JWT verification middleware in the API gateway or backend service.
- Use the JWT `sub` claim as `auth_user_id` in the business database.
- Keep business-specific profile, role, permission, order, and content tables in the business system.

The business system does not store auth passwords and does not verify wallet signatures itself.

## Login Methods

- Email/password: `/api/v1/auth/register`, `/api/v1/auth/login`
- Phone code login: `/api/v1/phone/code`, `/api/v1/phone/login`
- Wallet login: `/api/v1/wallet/nonce`, `/api/v1/wallet/verify`
- OAuth login: `/api/v1/oauth/:provider/start`, `/api/v1/oauth/:provider/callback`
- Email verification code: `/api/v1/email/code`, `/api/v1/email/verify`

Email verification codes are used for email verification, password reset, and email binding. They are not currently a standalone email-code login method.

## Unified Login Page

For a more consistent user experience, a business application can skip its own login form and redirect users to the Open Wallet Auth login page:

```text
https://auth.example.com/login?client_id=example-app&return_uri=https%3A%2F%2Fapp.example.com%2Fauth%2Fcallback
```

Parameters:

- `client_id`: the application ID registered in the auth service.
- `return_uri`: the business callback URL that receives the auth result.

The application display name is read from the auth service client registry, not from URL parameters, so users cannot spoof the source application name by editing the URL.

After email, phone, Google, GitHub, or wallet login succeeds, the page redirects to:

```text
https://app.example.com/auth/callback#access_token=...&token_type=Bearer&expires_at=...
```

The business callback page reads the `access_token` from the URL fragment and exchanges it with its own backend for a local business token or session. This keeps password, OAuth, and wallet UI inside the auth service.

The Web SDK is recommended for hosted-login redirects and callback parsing:

```ts
import { createAuthClient } from "@open-wallet-auth/web";

const auth = createAuthClient({
  authBaseURL: "https://auth.example.com",
  clientID: "example-app",
  returnURI: `${window.location.origin}/auth/callback`,
});

auth.login({ redirect: window.location.pathname });
```

Use the Node or Go SDK only for trusted backend work such as legacy user migration, email/password identity login, and profile validation. SDKs wrap the identity protocol; business systems still own their local user tables and business tokens.

Hosted login page branding and login methods are managed in the admin console:

- Path: `/console/settings`
- Tab: `登录页 Login`
- Editable fields: brand name, brand mark, subtitle, registration, phone login, GitHub, Google, and wallet login.

## Legacy Password Migration

Existing applications often store passwords with old algorithms such as `hmac_sha1`, `sha1`, or `md5`. Open Wallet Auth can import those legacy hashes into `legacy_credentials` and transparently upgrade them after the first successful login.

Recommended flow:

1. Import or match the old application user by email or phone.
2. Store the old hash in `legacy_credentials` with `source`, `hash_type`, `password_hash`, and optional `salt`.
3. Add the user to the target application's allow-list when `clients.whitelist_enabled` is enabled.
4. On password login, Open Wallet Auth checks bcrypt first. If bcrypt fails, it checks active legacy credentials for the user.
5. After a legacy credential succeeds, Open Wallet Auth immediately stores a bcrypt hash in `users.password_hash` and marks the legacy credential as `migrated`.

This feature is intended for generic application migrations. Business systems should not add product-specific password logic to the auth service.

## Wallet Login Flow

1. Detect EIP-1193 wallet providers in the browser.
2. Ask the user to choose a wallet when multiple injected wallets exist.
3. Request accounts through the selected provider.
4. Call `POST /api/v1/wallet/nonce` with address, domain, and chain ID.
5. Ask the wallet to sign the returned message.
6. Call `POST /api/v1/wallet/verify` with `client_id`, address, nonce, and signature.
7. Store the returned access token and refresh token.
8. Call business APIs with `Authorization: Bearer <access_token>`.

## Business User Table

The auth service `users` table is the shared identity table. It is not the full business user profile.

Recommended business-side profile table:

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

When a valid JWT is received for the first time, the business system can create a local profile row for that `auth_user_id`.

## Multi-Application Login

Open Wallet Auth uses `client_id` to identify the business application.

- Empty `client_id` is normalized to `default`.
- Each business application should create its own explicit `client_id`.
- The JWT `audience` must match what the business API expects.
- `user_clients` records which applications a user has logged into.

This lets multiple systems share authentication while still keeping system-specific business data separate.

## References

- [Universal auth demo](../examples/universal-auth-demo)
- [Browser wallet login example](../examples/browser-wallet-login)
- [Gin API JWT verification example](../examples/gin-api)
- [Provider configuration](PROVIDERS.md)

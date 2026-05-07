# Provider Configuration

[简体中文](PROVIDERS.zh-CN.md)

This document describes SMS, email, and OAuth provider configuration. Business-system JWT integration is covered in [INTEGRATION.md](INTEGRATION.md), and production runtime operations are covered in [DEPLOYMENT.md](DEPLOYMENT.md).

These provider settings can be edited in the admin console under System Settings. Secret fields are redacted when read back; leave a secret input empty to keep the existing value.

## SMS and Email Providers

Open Wallet Auth supports four message-provider modes:

- `noop`: does not send real messages; useful for local development.
- `webhook`: forwards verification-code delivery requests to your own HTTP service.
- `smtp`: sends email verification codes through SMTP.
- `aliyun_sms`: sends phone verification codes through Aliyun SendSms API.

## Feature Switches

```yaml
phone:
  enabled: true

email:
  verification_enabled: true
```

When disabled, related APIs return explicit disabled errors:

- `PHONE_LOGIN_DISABLED`
- `EMAIL_VERIFICATION_DISABLED`

## Local Development Codes

```yaml
phone:
  dev_code: "123456"
  expose_dev_code: true

email:
  dev_code: "123456"
  expose_dev_code: true
```

`expose_dev_code=true` returns `dev_code` in API responses for local demos only. Production should set both `expose_dev_code` values to `false`.

## Webhook Provider

```yaml
phone:
  provider:
    type: webhook
    webhook:
      url: https://your-message-service.example.com/messages
      bearer_token: your-secret-token
    headers:
      X-Provider: open-wallet-auth

email:
  provider:
    type: webhook
    webhook:
      url: https://your-message-service.example.com/messages
      bearer_token: your-secret-token
    headers:
      X-Provider: open-wallet-auth
```

SMS payload:

```json
{
  "type": "sms",
  "phone": "+8613800000000",
  "code": "123456"
}
```

Email payload:

```json
{
  "type": "email",
  "email": "alice@example.com",
  "subject": "Your Open Wallet Auth verification code",
  "code": "123456"
}
```

Any `2xx` response is treated as success.

## SMTP Email

```yaml
email:
  provider:
    type: smtp
    smtp:
      host: smtp.example.com
      port: 587
      username: noreply@example.com
      password: your-smtp-password
      from: noreply@example.com
```

If `from` is empty, the provider uses `username`.

## Aliyun SMS

```yaml
phone:
  provider:
    type: aliyun_sms
    aliyun_sms:
      access_key_id: your-access-key-id
      access_key_secret: your-access-key-secret
      sign_name: your-approved-sign-name
      template_code: SMS_000000000
      region_id: cn-hangzhou
      endpoint: https://dysmsapi.aliyuncs.com
```

The implementation sends the verification code as template variable `code`, so the Aliyun template must contain `${code}`.

## OAuth Providers

Google and GitHub OAuth providers are configured under `oauth.google` and `oauth.github`. The service supports default credentials and host-specific tenant credentials.

Default credentials:

```yaml
oauth:
  github:
    client_id: your-github-client-id
    client_secret: your-github-client-secret
```

Host-specific credentials:

```yaml
oauth:
  github:
    tenant_credentials:
      - host: blockx.example.com
        client_id: your-blockx-github-client-id
        client_secret: your-blockx-github-client-secret
```

Use tenant credentials when different business domains need different OAuth apps or callback settings.

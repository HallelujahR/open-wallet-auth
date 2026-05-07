# 短信和邮件服务商接入

Open Wallet Auth 提供 `noop`、`webhook`、`smtp`、`aliyun_sms` 四种消息服务商模式。

这些服务商配置可以在管理后台“系统配置”页面可视化编辑。密钥字段读取时会脱敏；保存时密钥输入框留空表示保留已有值。

- `noop`：不真正发送消息，适合本地开发。
- `webhook`：认证服务把验证码发送请求转发到你自己的 HTTP 服务。
- `smtp`：直接通过 SMTP 发送邮箱验证码。
- `aliyun_sms`：直接通过阿里云短信 SendSms API 发送手机号验证码。

## 开关

```yaml
phone:
  enabled: true

email:
  verification_enabled: true
```

关闭后，对应接口会返回明确的禁用错误：

- `PHONE_LOGIN_DISABLED`
- `EMAIL_VERIFICATION_DISABLED`

## 本地开发验证码

```yaml
phone:
  dev_code: "123456"
  expose_dev_code: true

email:
  dev_code: "123456"
  expose_dev_code: true
```

`expose_dev_code=true` 时，接口响应会返回 `dev_code`，方便本地 Demo 调试。

生产环境建议：

```yaml
phone:
  expose_dev_code: false

email:
  expose_dev_code: false
```

## Webhook 配置

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

## 短信 Webhook Payload

```json
{
  "type": "sms",
  "phone": "+8613800000000",
  "code": "123456"
}
```

## 邮件 Webhook Payload

```json
{
  "type": "email",
  "email": "alice@example.com",
  "subject": "Your Open Wallet Auth verification code",
  "code": "123456"
}
```

你的 Webhook 服务只需要返回任意 `2xx` 状态码，认证服务就会认为发送成功。非 `2xx` 会被视为发送失败。

## SMTP 邮件配置

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

`from` 为空时默认使用 `username`。真实密码建议通过环境变量或服务器密钥管理注入，不要提交到仓库。

## 阿里云短信配置

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

当前实现会把验证码作为模板变量 `code` 发送，因此阿里云短信模板需要包含 `${code}`。签名和模板必须先在阿里云控制台审核通过。

## 自定义 Go 适配器

如果你想直接在项目内接入某个云厂商 SDK，可以实现：

```go
type SMSProvider interface {
    SendSMS(ctx context.Context, msg message.SMSMessage) error
}

type EmailProvider interface {
    SendEmail(ctx context.Context, msg message.EmailMessage) error
}
```

然后在 `internal/app/wire_services.go` 中替换默认 provider wiring。

## OAuth Provider 配置

Google 和 GitHub OAuth 配置位于 `oauth.google`、`oauth.github`。当前实现支持默认凭据，也支持按业务域名区分的 tenant credentials。

默认凭据：

```yaml
oauth:
  github:
    client_id: your-github-client-id
    client_secret: your-github-client-secret
```

按域名配置独立凭据：

```yaml
oauth:
  github:
    tenant_credentials:
      - host: blockx.example.com
        client_id: your-blockx-github-client-id
        client_secret: your-blockx-github-client-secret
```

当不同业务域名需要配置不同 OAuth App 或不同回调地址时，使用 tenant credentials。

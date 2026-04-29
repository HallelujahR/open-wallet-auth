# 短信和邮件服务商接入

Open Wallet Auth 默认提供 `noop` 和 `webhook` 两种消息服务商模式。

- `noop`：不真正发送消息，适合本地开发。
- `webhook`：认证服务把验证码发送请求转发到你自己的 HTTP 服务，由你的服务再对接阿里云短信、腾讯云短信、SendGrid、Resend、企业邮件网关等。

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
    webhook_url: https://your-message-service.example.com/messages
    bearer_token: your-secret-token
    headers:
      X-Provider: open-wallet-auth

email:
  provider:
    type: webhook
    webhook_url: https://your-message-service.example.com/messages
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

然后在 `internal/app/app.go` 中替换默认 provider wiring。

# Open Wallet Auth Go SDK

Go service-side SDK for business backends.

Go 服务端 SDK，用于业务后端调用认证中台登录、注册和 profile 校验。

## Install

```bash
go get github.com/HallelujahR/open-wallet-auth/sdk/go
```

For local monorepo development before tagging a release:

```go
replace github.com/HallelujahR/open-wallet-auth/sdk/go => ../open-wallet-auth/sdk/go
```

## Usage

```go
client := owa.NewClient(owa.Config{
    BaseURL:  "https://auth.example.com",
    ClientID: "my-app",
})

user, err := client.Profile(ctx, accessToken)
if err != nil {
    return err
}
_ = user.ID
```

## Notes

- Use this SDK in trusted Go services, not in browser code.
- `ErrInvalidCredentials` and `ErrEmailExists` are sentinel errors for stable business mapping.
- Business services should keep their own local user model and token after identity validation.

## 说明

- 这个 SDK 用于可信 Go 服务端，不用于浏览器。
- `ErrInvalidCredentials` 和 `ErrEmailExists` 是稳定哨兵错误，方便业务系统映射自己的错误返回。
- 业务系统校验中台身份后，仍应保留自己的业务用户模型和业务 token。

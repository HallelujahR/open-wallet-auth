# 认证中台管理控制台

这是 Open Wallet Auth 的管理后台前端，用于公司内部查看和管理统一身份、接入应用、登录会话、登录审计和安全事件。

## 技术栈

- React
- TypeScript
- Vite
- Ant Design
- ECharts

## 本地运行

```bash
cd admin-web
npm install
npm run dev
```

默认访问：

```text
http://localhost:5174
```

后端地址通过环境变量配置，未配置时默认连接 `http://localhost:8081`：

```bash
VITE_AUTH_API_BASE_URL=http://localhost:8081 npm run dev
```

登录页只需要填写：

- 管理员账号，对应后端配置 `OWA_MANAGEMENT_ADMIN_USERNAME`
- 管理员密码，对应后端配置 `OWA_MANAGEMENT_ADMIN_PASSWORD`

## 当前页面

- 登录页：账号密码登录认证管理后台
- 管理概览：聚合应用、身份、会话、登录审计和安全事件
- 接入应用：查看和创建业务系统 client
- 身份用户：查看身份用户、绑定关系、会话，支持禁用/启用用户
- 登录会话：查看和吊销刷新令牌会话
- 登录审计：查看登录审计日志
- 安全操作：查看敏感操作审计事件
- 系统设置：查看管理边界和基础运行配置

## 管理边界

该后台只管理认证服务自身的数据：

- 统一身份
- 登录方式
- 接入应用
- 会话
- 登录审计
- 安全事件

它不管理业务系统数据，例如 blockx 积分、labelService API Key、业务权限、订单、标签、报告等。

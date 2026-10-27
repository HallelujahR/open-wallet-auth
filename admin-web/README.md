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

## 界面与登录行为

- 统一登录页 `/login` 支持同一浏览器的多账号选择、单账号退出和全部退出。
- `src/theme/index.ts` 集中配置 Ant Design 的颜色、圆角、控件和表格主题；`src/styles/globals.css` 管理布局与响应式规则。
- 登录页与后台登录页共用 `AuthShell`；控制台页面共用 `PageHeader`，手机端使用抽屉导航。
- 页面按路由懒加载，ECharts 只在管理概览加载。
- 管理概览通过 `/api/v1/admin/dashboard` 读取完整统计，按浏览器时区展示最近七天，不受列表分页影响。
- 开发环境默认请求 `http://localhost:8081`，生产构建默认同源；可用 `VITE_AUTH_API_BASE_URL` 覆盖。

新接入应用应配置完整回调地址白名单（带 state 的流程必须配置），详见 [接入指南](../docs/INTEGRATION.zh-CN.md)。

## 管理边界

该后台只管理认证服务自身的数据：

- 统一身份
- 登录方式
- 接入应用
- 会话
- 登录审计
- 安全事件

它不管理业务系统数据，例如 blockx 积分、labelService API Key、业务权限、订单、标签、报告等。

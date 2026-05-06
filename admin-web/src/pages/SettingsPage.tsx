import { Card, Descriptions, Typography } from "antd";
import { getAdminSession } from "../store/authStore";

export function SettingsPage() {
  const session = getAdminSession();
  return (
    <div>
      <Typography.Title level={3}>系统设置</Typography.Title>
      <Card title="运行配置概览">
        <Descriptions bordered column={1}>
          <Descriptions.Item label="认证服务地址">{session?.baseUrl || "-"}</Descriptions.Item>
          <Descriptions.Item label="管理鉴权方式">X-Admin-Token</Descriptions.Item>
          <Descriptions.Item label="JWKS 地址">{session?.baseUrl ? `${session.baseUrl}/.well-known/jwks.json` : "-"}</Descriptions.Item>
          <Descriptions.Item label="管理边界">仅管理身份、应用、会话和安全审计，不管理业务系统权限与数据</Descriptions.Item>
        </Descriptions>
      </Card>
    </div>
  );
}

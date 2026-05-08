import {
  ApiOutlined,
  AppstoreOutlined,
  AuditOutlined,
  DashboardOutlined,
  LogoutOutlined,
  SafetyOutlined,
  SettingOutlined,
  TeamOutlined,
} from "@ant-design/icons";
import { Avatar, Button, Layout, Menu, Space, Typography } from "antd";
import { Outlet, useLocation, useNavigate } from "react-router-dom";
import { clearAdminSession, getAdminSession } from "../store/authStore";

const { Header, Sider, Content } = Layout;

const navItems = [
  { key: "/console", icon: <DashboardOutlined />, label: "管理概览" },
  { key: "/console/applications", icon: <AppstoreOutlined />, label: "接入应用" },
  { key: "/console/identities", icon: <TeamOutlined />, label: "身份用户" },
  { key: "/console/sessions", icon: <SafetyOutlined />, label: "登录会话" },
  { key: "/console/audit-logs", icon: <AuditOutlined />, label: "登录审计" },
  { key: "/console/security-events", icon: <ApiOutlined />, label: "安全操作" },
  { key: "/console/settings", icon: <SettingOutlined />, label: "系统设置" },
];

export function AdminLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const session = getAdminSession();

  const logout = () => {
    clearAdminSession();
    navigate("/console/login", { replace: true });
  };

  return (
    <Layout className="admin-shell">
      <Sider width={232} className="admin-sider">
        <div className="sidebar-brand">
          <div className="brand-symbol">∞</div>
          <div>
            <strong>CORE AUTH</strong>
            <span>统一认证控制台</span>
          </div>
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={navItems}
          onClick={(item) => navigate(item.key)}
          className="admin-menu"
        />
      </Sider>
      <Layout>
        <Header className="admin-header">
          <Typography.Text strong>统一认证管理后台</Typography.Text>
          <Space>
            <span className="node-status">API: {session?.baseUrl || "未配置"}</span>
            <Avatar size={32}>A</Avatar>
            <Button icon={<LogoutOutlined />} onClick={logout}>
              退出
            </Button>
          </Space>
        </Header>
        <Content className="admin-content">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}

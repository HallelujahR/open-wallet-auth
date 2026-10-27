import {
  ApiOutlined,
  AppstoreOutlined,
  AuditOutlined,
  DashboardOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  SafetyOutlined,
  SettingOutlined,
  TeamOutlined,
} from "@ant-design/icons";
import {
  Avatar,
  Breadcrumb,
  Button,
  Drawer,
  Grid,
  Layout,
  Menu,
  Space,
  Tooltip,
} from "antd";
import { useState } from "react";
import { Outlet, useLocation, useNavigate } from "react-router-dom";
import { clearAdminSession } from "../store/authStore";
const { Header, Sider, Content } = Layout;
const navItems = [
  { key: "/console", icon: <DashboardOutlined />, label: "管理概览" },
  {
    key: "/console/applications",
    icon: <AppstoreOutlined />,
    label: "接入应用",
  },
  { key: "/console/identities", icon: <TeamOutlined />, label: "身份用户" },
  { key: "/console/sessions", icon: <SafetyOutlined />, label: "登录会话" },
  { key: "/console/audit-logs", icon: <AuditOutlined />, label: "登录审计" },
  { key: "/console/security-events", icon: <ApiOutlined />, label: "安全操作" },
  { key: "/console/settings", icon: <SettingOutlined />, label: "系统设置" },
];
export function AdminLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const screens = Grid.useBreakpoint();
  const [collapsed, setCollapsed] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const mobile = !screens.lg;
  const title =
    navItems.find((item) => item.key === location.pathname)?.label || "控制台";
  const menu = (
    <Menu
      mode="inline"
      selectedKeys={[location.pathname]}
      items={navItems}
      onClick={(item) => {
        navigate(item.key);
        setMobileOpen(false);
      }}
      className="admin-menu"
    />
  );
  const brand = (
    <div className="sidebar-brand">
      <span className="brand-mark">O</span>
      {(!collapsed || mobile) && (
        <div>
          <strong>Open Wallet Auth</strong>
          <span>身份管理控制台</span>
        </div>
      )}
    </div>
  );
  return (
    <Layout className="admin-shell">
      {!mobile && (
        <Sider
          width={240}
          collapsedWidth={80}
          collapsed={collapsed}
          className="admin-sider"
        >
          {brand}
          {!collapsed && <div className="nav-caption">工作空间</div>}
          {menu}
          <div className="sidebar-footer">
            <SafetyOutlined />
            {!collapsed && <span>统一身份 · 安全连接</span>}
          </div>
        </Sider>
      )}
      <Drawer
        placement="left"
        width={260}
        open={mobileOpen}
        onClose={() => setMobileOpen(false)}
        styles={{ body: { padding: 0 } }}
        title="导航"
      >
        {brand}
        {menu}
      </Drawer>
      <Layout className="admin-main">
        <Header className="admin-header">
          <Space size={16}>
            <Button
              type="text"
              aria-label={
                mobile ? "打开导航" : collapsed ? "展开导航" : "收起导航"
              }
              icon={
                collapsed || mobile ? (
                  <MenuUnfoldOutlined />
                ) : (
                  <MenuFoldOutlined />
                )
              }
              onClick={() =>
                mobile ? setMobileOpen(true) : setCollapsed(!collapsed)
              }
            />
            <Breadcrumb items={[{ title: "控制台" }, { title }]} />
          </Space>
          <Space size={12}>
            <span className="admin-role">管理员</span>
            <Avatar size={30} className="admin-avatar">
              A
            </Avatar>
            <Tooltip title="退出登录">
              <Button
                type="text"
                aria-label="退出登录"
                icon={<LogoutOutlined />}
                onClick={() => {
                  clearAdminSession();
                  navigate("/console/login", { replace: true });
                }}
              />
            </Tooltip>
          </Space>
        </Header>
        <Content className="admin-content">
          <div className="content-container">
            <Outlet />
          </div>
        </Content>
      </Layout>
    </Layout>
  );
}

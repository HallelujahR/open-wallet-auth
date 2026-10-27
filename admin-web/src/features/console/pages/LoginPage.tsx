import {
  ArrowRightOutlined,
  LockOutlined,
  UserOutlined,
} from "@ant-design/icons";
import { App, Button, Form, Input, Typography } from "antd";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { adminApi } from "../../../api/admin";
import { authApiBaseUrl } from "../../../config";
import { saveAdminSession } from "../../../store/authStore";
import { AuthShell } from "../../../components/AuthShell";

type FormValues = { username: string; password: string };
export function AdminLoginPage() {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const submit = async (values: FormValues) => {
    setLoading(true);
    try {
      const result = await adminApi.login(values);
      saveAdminSession({
        baseUrl: authApiBaseUrl,
        adminToken: result.admin_token,
      });
      navigate("/console", { replace: true });
    } catch (err) {
      message.error(err instanceof Error ? err.message : "登录失败");
    } finally {
      setLoading(false);
    }
  };
  return (
    <AuthShell admin context="管理控制台">
      <div className="auth-card-heading">
        <span className="eyebrow">ADMIN ACCESS</span>
        <Typography.Title level={2}>登录控制台</Typography.Title>
        <Typography.Paragraph type="secondary">
          使用管理员账号，管理你的认证服务。
        </Typography.Paragraph>
      </div>
      <Form<FormValues>
        layout="vertical"
        onFinish={submit}
        requiredMark={false}
        size="large"
        disabled={loading}
      >
        <Form.Item
          label="管理员账号"
          name="username"
          rules={[{ required: true, message: "请输入管理员账号" }]}
        >
          <Input
            prefix={<UserOutlined />}
            placeholder="输入管理员账号"
            autoComplete="username"
          />
        </Form.Item>
        <Form.Item
          label="密码"
          name="password"
          rules={[{ required: true, message: "请输入密码" }]}
        >
          <Input.Password
            prefix={<LockOutlined />}
            placeholder="输入管理员密码"
            autoComplete="current-password"
          />
        </Form.Item>
        <Button
          htmlType="submit"
          type="primary"
          block
          size="large"
          loading={loading}
          icon={<ArrowRightOutlined />}
          iconPosition="end"
        >
          登录控制台
        </Button>
      </Form>
      <p className="auth-card-note">仅供管理员使用</p>
    </AuthShell>
  );
}

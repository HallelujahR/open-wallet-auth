import { LockOutlined, UserOutlined } from "@ant-design/icons";
import { Button, Card, Form, Input, Typography, message } from "antd";
import { useNavigate } from "react-router-dom";
import { adminApi } from "../../../api/admin";
import { authApiBaseUrl } from "../../../config";
import { saveAdminSession } from "../../../store/authStore";

type FormValues = {
  username: string;
  password: string;
};

export function AdminLoginPage() {
  const navigate = useNavigate();

  const submit = async (values: FormValues) => {
    try {
      const result = await adminApi.login({
        username: values.username,
        password: values.password,
      });
      saveAdminSession({
        baseUrl: authApiBaseUrl,
        adminToken: result.admin_token,
      });
      message.success("登录成功");
      navigate("/console", { replace: true });
    } catch (err: any) {
      message.error(err.message || "登录失败");
    }
  };

  return (
    <main className="login-page">
      <section className="login-hero">
        <div className="hero-network">
          <span className="node n1" />
          <span className="node n2" />
          <span className="node n3" />
          <span className="node n4" />
        </div>
        <div className="hero-copy">
          <div className="hero-mark">∞</div>
          <h1>认证服务<br />管理后台</h1>
          <p>统一管理接入应用、身份用户、登录会话与安全审计</p>
        </div>
      </section>
      <section className="login-panel">
        <Card className="login-card" bordered={false}>
          <Typography.Title level={2}>登录认证管理后台</Typography.Title>
          <Form<FormValues>
            layout="vertical"
            initialValues={{
              username: "admin",
            }}
            onFinish={submit}
          >
            <Form.Item label="管理员账号" name="username" rules={[{ required: true, message: "请输入管理员账号" }]}>
              <Input prefix={<UserOutlined />} placeholder="admin" autoComplete="username" />
            </Form.Item>
            <Form.Item label="管理员密码" name="password" rules={[{ required: true, message: "请输入管理员密码" }]}>
              <Input.Password prefix={<LockOutlined />} placeholder="请输入管理后台密码" autoComplete="current-password" />
            </Form.Item>
            <Button htmlType="submit" type="primary" block size="large" className="login-submit">
              登录
            </Button>
          </Form>
          <div className="login-footnote">该后台仅用于认证中台运维管理，不是业务系统用户登录页。</div>
          <div className="node-bar">生产环境请使用部署配置中的管理后台账号</div>
        </Card>
      </section>
    </main>
  );
}

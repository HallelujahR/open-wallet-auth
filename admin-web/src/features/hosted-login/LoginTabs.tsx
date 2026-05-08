import {
  LockOutlined,
  MailOutlined,
  MobileOutlined,
  PhoneOutlined,
  UserAddOutlined,
} from "@ant-design/icons";
import { Button, Form, Input, type TabsProps } from "antd";
import type {
  PasswordValues,
  PhoneValues,
  RegisterValues,
  UnifiedLoginConfig,
} from "./types";

type BuildLoginTabsInput = {
  loading: boolean;
  phoneCodeLoading: boolean;
  pageConfig: UnifiedLoginConfig;
  submitPasswordLogin: (values: PasswordValues) => void;
  sendPhoneCode: (phone: string) => void;
  submitPhoneLogin: (values: PhoneValues) => void;
  submitRegister: (values: RegisterValues) => void;
};

// buildLoginTabs keeps optional login methods type-safe for Ant Design Tabs.
// buildLoginTabs 以类型安全的方式为 Ant Design Tabs 组装可选登录方式。
export function buildLoginTabs(input: BuildLoginTabsInput): TabsProps["items"] {
  const items: TabsProps["items"] = [
    {
      key: "password",
      label: "邮箱登录",
      children: (
        <Form<PasswordValues>
          layout="vertical"
          onFinish={input.submitPasswordLogin}
        >
          <Form.Item
            label="邮箱 Email"
            name="email"
            rules={[
              { required: true, type: "email", message: "请输入有效邮箱" },
            ]}
          >
            <Input
              prefix={<MailOutlined />}
              placeholder="name@example.com"
              autoComplete="email"
            />
          </Form.Item>
          <Form.Item
            label="密码 Password"
            name="password"
            rules={[{ required: true, message: "请输入密码" }]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="请输入密码"
              autoComplete="current-password"
            />
          </Form.Item>
          <Button
            type="primary"
            htmlType="submit"
            loading={input.loading}
            block
            size="large"
          >
            登录
          </Button>
        </Form>
      ),
    },
  ];

  if (input.pageConfig.enablePhone) {
    items.push({
      key: "phone",
      label: "手机登录",
      children: (
        <PhoneLoginForm
          loading={input.loading}
          codeLoading={input.phoneCodeLoading}
          onSendCode={input.sendPhoneCode}
          onFinish={input.submitPhoneLogin}
        />
      ),
    });
  }

  if (input.pageConfig.enableRegister) {
    items.push({
      key: "register",
      label: "注册账号",
      children: (
        <Form<RegisterValues> layout="vertical" onFinish={input.submitRegister}>
          <Form.Item label="昵称 Username" name="username">
            <Input
              prefix={<UserAddOutlined />}
              placeholder="可选，默认使用邮箱前缀"
              autoComplete="username"
            />
          </Form.Item>
          <Form.Item
            label="邮箱 Email"
            name="email"
            rules={[
              { required: true, type: "email", message: "请输入有效邮箱" },
            ]}
          >
            <Input
              prefix={<MailOutlined />}
              placeholder="name@example.com"
              autoComplete="email"
            />
          </Form.Item>
          <Form.Item
            label="密码 Password"
            name="password"
            rules={[{ required: true, min: 8, message: "密码至少 8 位" }]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="至少 8 位密码"
              autoComplete="new-password"
            />
          </Form.Item>
          <Button
            type="primary"
            htmlType="submit"
            loading={input.loading}
            block
            size="large"
          >
            创建账号并登录
          </Button>
        </Form>
      ),
    });
  }

  return items;
}

function PhoneLoginForm({
  loading,
  codeLoading,
  onSendCode,
  onFinish,
}: {
  loading: boolean;
  codeLoading: boolean;
  onSendCode: (phone: string) => void;
  onFinish: (values: PhoneValues) => void;
}) {
  const [form] = Form.useForm<PhoneValues>();

  return (
    <Form<PhoneValues> form={form} layout="vertical" onFinish={onFinish}>
      <Form.Item
        label="手机号 Phone"
        name="phone"
        rules={[{ required: true, message: "请输入手机号" }]}
      >
        <Input
          prefix={<PhoneOutlined />}
          placeholder="+8613800000000 或 13800000000"
          autoComplete="tel"
        />
      </Form.Item>
      <Form.Item
        label="验证码 Code"
        name="code"
        rules={[{ required: true, message: "请输入验证码" }]}
      >
        <Input
          prefix={<MobileOutlined />}
          placeholder="请输入短信验证码"
          addonAfter={
            <Button
              type="link"
              size="small"
              loading={codeLoading}
              onClick={() => onSendCode(form.getFieldValue("phone") || "")}
            >
              发送
            </Button>
          }
        />
      </Form.Item>
      <Button
        type="primary"
        htmlType="submit"
        loading={loading}
        block
        size="large"
      >
        手机号登录
      </Button>
    </Form>
  );
}

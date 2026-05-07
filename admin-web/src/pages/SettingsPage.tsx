import { SaveOutlined, SyncOutlined } from "@ant-design/icons";
import { Button, Card, Col, Form, Input, InputNumber, Row, Select, Space, Switch, Tabs, Typography, message } from "antd";
import { useEffect, useState } from "react";
import { adminApi } from "../api/admin";
import type { RuntimeSettingsResult, SecretStatus } from "../types/api";

const providerOptions = [
  { label: "不发送 Noop", value: "noop" },
  { label: "Webhook 回调", value: "webhook" },
  { label: "SMTP 邮件", value: "smtp" },
  { label: "阿里云短信 Aliyun SMS", value: "aliyun_sms" },
];

// SecretHint shows whether a secret exists without exposing the secret value.
// SecretHint 只展示密钥是否已配置，不展示密钥明文。
function SecretHint({ status }: { status?: SecretStatus }) {
  if (!status?.configured) return <Typography.Text type="secondary">未配置</Typography.Text>;
  return <Typography.Text type="secondary">已配置：{status.masked}</Typography.Text>;
}

// SettingsPage lets operators edit provider settings from the admin console.
// SettingsPage 允许运维人员在管理后台可视化编辑服务商配置。
export function SettingsPage() {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [secrets, setSecrets] = useState<RuntimeSettingsResult["secrets"]>({});

  const loadSettings = async () => {
    setLoading(true);
    try {
      const result = await adminApi.getSettings();
      form.setFieldsValue(result.settings);
      setSecrets(result.secrets || {});
    } catch (error: any) {
      message.error(error.message || "加载系统配置失败");
    } finally {
      setLoading(false);
    }
  };

  const saveSettings = async () => {
    const values = await form.validateFields();
    setSaving(true);
    try {
      const result = await adminApi.updateSettings(values);
      form.setFieldsValue(result.settings);
      setSecrets(result.secrets || {});
      message.success("系统配置已保存");
    } catch (error: any) {
      message.error(error.message || "保存系统配置失败");
    } finally {
      setSaving(false);
    }
  };

  useEffect(() => {
    loadSettings();
  }, []);

  return (
    <div>
      <Space className="page-title-row" align="center">
        <div>
          <Typography.Title level={3}>系统配置</Typography.Title>
          <Typography.Text type="secondary">
            可视化配置第三方登录 OAuth、短信 SMS、邮件 Email 服务商。密钥 Secret 字段不会回显，保存时留空表示保留当前值。
          </Typography.Text>
        </div>
        <Space>
          <Button icon={<SyncOutlined />} onClick={loadSettings} loading={loading}>
            刷新
          </Button>
          <Button type="primary" icon={<SaveOutlined />} onClick={saveSettings} loading={saving}>
            保存配置
          </Button>
        </Space>
      </Space>

      <Form form={form} layout="vertical" disabled={loading}>
        <Tabs
          items={[
            {
              key: "oauth",
              label: "第三方登录 OAuth",
              children: (
                <Row gutter={[16, 16]}>
                  <Col span={12}>
                    <OAuthProviderCard provider="google" title="Google OAuth" secrets={secrets} />
                  </Col>
                  <Col span={12}>
                    <OAuthProviderCard provider="github" title="GitHub OAuth" secrets={secrets} />
                  </Col>
                </Row>
              ),
            },
            {
              key: "message",
              label: "短信 SMS 与邮件 Email",
              children: (
                <Row gutter={[16, 16]}>
                  <Col span={12}>
                    <MessageProviderCard root="phone" title="手机号验证码" secretPrefix="phone.provider" secrets={secrets} />
                  </Col>
                  <Col span={12}>
                    <MessageProviderCard root="email" title="邮箱验证码" secretPrefix="email.provider" secrets={secrets} />
                  </Col>
                </Row>
              ),
            },
          ]}
        />
      </Form>
    </div>
  );
}

function OAuthProviderCard({ provider, title, secrets }: { provider: "google" | "github"; title: string; secrets: RuntimeSettingsResult["secrets"] }) {
  const base = ["oauth", provider];
  return (
    <Card title={title}>
      <Typography.Paragraph type="secondary">
        配置第三方 OAuth 登录参数。默认配置适用于所有业务域名；如果 GitHub 这类平台要求一个应用只能绑定一个回调地址，可以在下方按业务域名配置独立凭据。
      </Typography.Paragraph>
      <Row gutter={12}>
        <Col span={12}>
          <Form.Item label="应用 ID Client ID" name={[...base, "client_id"]} extra="第三方平台创建 OAuth 应用后获得的公开应用标识。">
            <Input placeholder="请输入 OAuth Client ID" />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item label="应用密钥 Client Secret" name={[...base, "client_secret"]} extra={<SecretHint status={secrets[`oauth.${provider}.client_secret`]} />}>
            <Input.Password placeholder="留空则保留当前密钥" autoComplete="new-password" />
          </Form.Item>
        </Col>
      </Row>
      <Form.Item label="授权地址 Auth URL" name={[...base, "auth_url"]} extra="用户点击第三方登录后，浏览器会跳转到这个授权地址。">
        <Input />
      </Form.Item>
      <Form.Item label="令牌交换地址 Token URL" name={[...base, "token_url"]} extra="认证服务用授权码 code 换取第三方 access token 的地址。">
        <Input />
      </Form.Item>
      <Form.Item label="用户信息地址 UserInfo URL" name={[...base, "user_info_url"]} extra="认证服务用第三方 access token 拉取用户资料的地址。">
        <Input />
      </Form.Item>
      <Form.Item label="授权范围 Scopes" name={[...base, "scopes"]} extra="需要向第三方申请的权限范围，例如 email、profile、read:user。">
        <Select mode="tags" tokenSeparators={[",", " "]} />
      </Form.Item>
      <Form.List name={[...base, "tenant_credentials"]}>
        {(fields, { add, remove }) => (
          <div>
            <Space style={{ marginBottom: 8 }}>
              <Typography.Text strong>按业务域名配置 Tenant Credentials</Typography.Text>
              <Button size="small" onClick={() => add({ host: "", client_id: "", client_secret: "" })}>
                添加域名
              </Button>
            </Space>
            {fields.map((field) => (
              <Card key={field.key} size="small" style={{ marginBottom: 8 }}>
                <Row gutter={12}>
                  <Col span={7}>
                    <Form.Item label="业务域名 Host" name={[field.name, "host"]} extra="例如 blockx.example.com，用于匹配回调或返回地址。">
                      <Input placeholder="blockx.example.com" />
                    </Form.Item>
                  </Col>
                  <Col span={7}>
                    <Form.Item label="应用 ID Client ID" name={[field.name, "client_id"]}>
                      <Input />
                    </Form.Item>
                  </Col>
                  <Col span={7}>
                    <Form.Item label="应用密钥 Client Secret" name={[field.name, "client_secret"]} extra="留空则保留该域名当前密钥。">
                      <Input.Password placeholder="留空则保留当前密钥" autoComplete="new-password" />
                    </Form.Item>
                  </Col>
                  <Col span={3}>
                    <Form.Item label="操作">
                      <Button danger onClick={() => remove(field.name)}>
                        删除
                      </Button>
                    </Form.Item>
                  </Col>
                </Row>
              </Card>
            ))}
          </div>
        )}
      </Form.List>
    </Card>
  );
}

function MessageProviderCard({
  root,
  title,
  secretPrefix,
  secrets,
}: {
  root: "phone" | "email";
  title: string;
  secretPrefix: string;
  secrets: RuntimeSettingsResult["secrets"];
}) {
  return (
    <Card title={title}>
      <Typography.Paragraph type="secondary">
        配置验证码发送服务商。Noop 适合本地调试；Webhook 适合接入自有消息网关；SMTP 用于邮件；Aliyun SMS 用于阿里云短信。
      </Typography.Paragraph>
      <Form.Item
        label={root === "phone" ? "启用手机号登录 Phone Login" : "启用邮箱验证 Email Verification"}
        name={[root, root === "phone" ? "enabled" : "verification_enabled"]}
        valuePropName="checked"
        extra={root === "phone" ? "关闭后手机号验证码登录接口会返回禁用错误。" : "关闭后邮箱验证码发送和校验接口会返回禁用错误。"}
      >
        <Switch />
      </Form.Item>
      <Form.Item label="服务商类型 Provider Type" name={[root, "provider", "type"]} extra="选择当前验证码实际发送方式。">
        <Select options={providerOptions} />
      </Form.Item>
      <Typography.Title level={5}>Webhook 回调</Typography.Title>
      <Form.Item label="回调地址 Webhook URL" name={[root, "provider", "webhook", "url"]} extra="认证服务会把验证码发送请求 POST 到这个地址。">
        <Input />
      </Form.Item>
      <Form.Item label="访问令牌 Bearer Token" name={[root, "provider", "webhook", "bearer_token"]} extra={<SecretHint status={secrets[`${secretPrefix}.webhook.bearer_token`]} />}>
        <Input.Password placeholder="留空则保留当前密钥" autoComplete="new-password" />
      </Form.Item>

      <Typography.Title level={5}>SMTP 邮件</Typography.Title>
      <Row gutter={12}>
        <Col span={16}>
          <Form.Item label="服务器地址 SMTP Host" name={[root, "provider", "smtp", "host"]}>
            <Input />
          </Form.Item>
        </Col>
        <Col span={8}>
          <Form.Item label="端口 Port" name={[root, "provider", "smtp", "port"]}>
            <InputNumber min={1} max={65535} style={{ width: "100%" }} />
          </Form.Item>
        </Col>
      </Row>
      <Form.Item label="账号 Username" name={[root, "provider", "smtp", "username"]}>
        <Input />
      </Form.Item>
      <Form.Item label="密码 Password" name={[root, "provider", "smtp", "password"]} extra={<SecretHint status={secrets[`${secretPrefix}.smtp.password`]} />}>
        <Input.Password placeholder="留空则保留当前密钥" autoComplete="new-password" />
      </Form.Item>
      <Form.Item label="发件人 From" name={[root, "provider", "smtp", "from"]} extra="为空时默认使用 SMTP Username。">
        <Input />
      </Form.Item>

      <Typography.Title level={5}>阿里云短信 Aliyun SMS</Typography.Title>
      <Form.Item label="访问密钥 ID Access Key ID" name={[root, "provider", "aliyun_sms", "access_key_id"]}>
        <Input />
      </Form.Item>
      <Form.Item label="访问密钥 Secret Access Key Secret" name={[root, "provider", "aliyun_sms", "access_key_secret"]} extra={<SecretHint status={secrets[`${secretPrefix}.aliyun_sms.access_key_secret`]} />}>
        <Input.Password placeholder="留空则保留当前密钥" autoComplete="new-password" />
      </Form.Item>
      <Row gutter={12}>
        <Col span={12}>
          <Form.Item label="短信签名 Sign Name" name={[root, "provider", "aliyun_sms", "sign_name"]}>
            <Input />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item label="模板编号 Template Code" name={[root, "provider", "aliyun_sms", "template_code"]} extra="短信模板需要包含验证码变量 ${code}。">
            <Input />
          </Form.Item>
        </Col>
      </Row>
      <Row gutter={12}>
        <Col span={12}>
          <Form.Item label="地域 Region" name={[root, "provider", "aliyun_sms", "region_id"]}>
            <Input />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item label="接口地址 Endpoint" name={[root, "provider", "aliyun_sms", "endpoint"]}>
            <Input />
          </Form.Item>
        </Col>
      </Row>
    </Card>
  );
}

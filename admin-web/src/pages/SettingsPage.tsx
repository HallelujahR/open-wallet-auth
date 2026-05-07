import { SaveOutlined, SyncOutlined } from "@ant-design/icons";
import { Button, Card, Col, Form, Input, InputNumber, Row, Select, Space, Switch, Tabs, Typography, message } from "antd";
import { useEffect, useState } from "react";
import { adminApi } from "../api/admin";
import type { RuntimeSettingsResult, SecretStatus } from "../types/api";

const providerOptions = [
  { label: "不发送 noop", value: "noop" },
  { label: "Webhook", value: "webhook" },
  { label: "SMTP 邮件", value: "smtp" },
  { label: "阿里云短信", value: "aliyun_sms" },
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
          <Typography.Text type="secondary">配置 GitHub、Google、短信和邮件服务商。密钥字段留空表示保留当前值。</Typography.Text>
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
              label: "第三方登录",
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
              label: "短信与邮件",
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
      <Row gutter={12}>
        <Col span={12}>
          <Form.Item label="Client ID" name={[...base, "client_id"]}>
            <Input placeholder="OAuth Client ID" />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item label="Client Secret" name={[...base, "client_secret"]} extra={<SecretHint status={secrets[`oauth.${provider}.client_secret`]} />}>
            <Input.Password placeholder="留空则保留当前密钥" autoComplete="new-password" />
          </Form.Item>
        </Col>
      </Row>
      <Form.Item label="Auth URL" name={[...base, "auth_url"]}>
        <Input />
      </Form.Item>
      <Form.Item label="Token URL" name={[...base, "token_url"]}>
        <Input />
      </Form.Item>
      <Form.Item label="UserInfo URL" name={[...base, "user_info_url"]}>
        <Input />
      </Form.Item>
      <Form.Item label="Scopes" name={[...base, "scopes"]}>
        <Select mode="tags" tokenSeparators={[",", " "]} />
      </Form.Item>
      <Form.List name={[...base, "tenant_credentials"]}>
        {(fields, { add, remove }) => (
          <div>
            <Space style={{ marginBottom: 8 }}>
              <Typography.Text strong>按业务域名配置</Typography.Text>
              <Button size="small" onClick={() => add({ host: "", client_id: "", client_secret: "" })}>
                添加域名
              </Button>
            </Space>
            {fields.map((field) => (
              <Card key={field.key} size="small" style={{ marginBottom: 8 }}>
                <Row gutter={12}>
                  <Col span={7}>
                    <Form.Item label="域名" name={[field.name, "host"]}>
                      <Input placeholder="blockx.example.com" />
                    </Form.Item>
                  </Col>
                  <Col span={7}>
                    <Form.Item label="Client ID" name={[field.name, "client_id"]}>
                      <Input />
                    </Form.Item>
                  </Col>
                  <Col span={7}>
                    <Form.Item label="Client Secret" name={[field.name, "client_secret"]} extra="留空则保留该域名当前密钥">
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
      <Form.Item label={root === "phone" ? "启用手机号登录" : "启用邮箱验证"} name={[root, root === "phone" ? "enabled" : "verification_enabled"]} valuePropName="checked">
        <Switch />
      </Form.Item>
      <Form.Item label="服务商类型" name={[root, "provider", "type"]}>
        <Select options={providerOptions} />
      </Form.Item>
      <Typography.Title level={5}>Webhook</Typography.Title>
      <Form.Item label="URL" name={[root, "provider", "webhook", "url"]}>
        <Input />
      </Form.Item>
      <Form.Item label="Bearer Token" name={[root, "provider", "webhook", "bearer_token"]} extra={<SecretHint status={secrets[`${secretPrefix}.webhook.bearer_token`]} />}>
        <Input.Password placeholder="留空则保留当前密钥" autoComplete="new-password" />
      </Form.Item>

      <Typography.Title level={5}>SMTP</Typography.Title>
      <Row gutter={12}>
        <Col span={16}>
          <Form.Item label="Host" name={[root, "provider", "smtp", "host"]}>
            <Input />
          </Form.Item>
        </Col>
        <Col span={8}>
          <Form.Item label="Port" name={[root, "provider", "smtp", "port"]}>
            <InputNumber min={1} max={65535} style={{ width: "100%" }} />
          </Form.Item>
        </Col>
      </Row>
      <Form.Item label="Username" name={[root, "provider", "smtp", "username"]}>
        <Input />
      </Form.Item>
      <Form.Item label="Password" name={[root, "provider", "smtp", "password"]} extra={<SecretHint status={secrets[`${secretPrefix}.smtp.password`]} />}>
        <Input.Password placeholder="留空则保留当前密钥" autoComplete="new-password" />
      </Form.Item>
      <Form.Item label="From" name={[root, "provider", "smtp", "from"]}>
        <Input />
      </Form.Item>

      <Typography.Title level={5}>阿里云短信</Typography.Title>
      <Form.Item label="Access Key ID" name={[root, "provider", "aliyun_sms", "access_key_id"]}>
        <Input />
      </Form.Item>
      <Form.Item label="Access Key Secret" name={[root, "provider", "aliyun_sms", "access_key_secret"]} extra={<SecretHint status={secrets[`${secretPrefix}.aliyun_sms.access_key_secret`]} />}>
        <Input.Password placeholder="留空则保留当前密钥" autoComplete="new-password" />
      </Form.Item>
      <Row gutter={12}>
        <Col span={12}>
          <Form.Item label="签名名称" name={[root, "provider", "aliyun_sms", "sign_name"]}>
            <Input />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item label="模板 Code" name={[root, "provider", "aliyun_sms", "template_code"]}>
            <Input />
          </Form.Item>
        </Col>
      </Row>
      <Row gutter={12}>
        <Col span={12}>
          <Form.Item label="Region" name={[root, "provider", "aliyun_sms", "region_id"]}>
            <Input />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item label="Endpoint" name={[root, "provider", "aliyun_sms", "endpoint"]}>
            <Input />
          </Form.Item>
        </Col>
      </Row>
    </Card>
  );
}

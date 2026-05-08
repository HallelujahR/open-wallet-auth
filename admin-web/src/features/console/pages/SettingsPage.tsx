import { CopyOutlined, SaveOutlined, SyncOutlined } from "@ant-design/icons";
import { Alert, Button, Card, Col, Descriptions, Form, Input, InputNumber, Modal, Row, Select, Space, Switch, Tabs, Typography, message } from "antd";
import { useEffect, useState } from "react";
import { adminApi } from "../../../api/admin";
import { authApiBaseUrl } from "../../../config";
import type { ReadonlySettings, RuntimeSettingsResult, SecretStatus } from "../../../types/api";

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
  const [readonly, setReadonly] = useState<ReadonlySettings | null>(null);

  const loadSettings = async () => {
    setLoading(true);
    try {
      const result = await adminApi.getSettings();
      form.setFieldsValue(result.settings);
      setSecrets(result.secrets || {});
      setReadonly(result.readonly);
    } catch (error: any) {
      message.error(error.message || "加载系统配置失败");
    } finally {
      setLoading(false);
    }
  };

  const saveSettings = async () => {
    const values = await form.validateFields();
    Modal.confirm({
      title: "确认保存系统配置？",
      content: "本次修改会立即影响 OAuth 登录、验证码发送或浏览器跨域访问。密钥输入框留空时会保留原密钥。",
      okText: "确认保存",
      cancelText: "取消",
      onOk: async () => {
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
      },
    });
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
              key: "runtime",
              label: "只读配置 Runtime",
              children: <ReadonlySettingsPanel settings={readonly} />,
            },
            {
              key: "access",
              label: "访问控制 CORS",
              children: <AccessSettingsCard />,
            },
            {
              key: "login",
              label: "登录页 Login",
              children: <LoginPageSettingsCard />,
            },
            {
              key: "oauth",
              label: "第三方登录 OAuth",
              children: (
                <Row gutter={[16, 16]}>
                  <Col span={24}>
                    <OAuthCallbackCard />
                  </Col>
                  <Col xs={24} xl={12}>
                    <OAuthProviderCard provider="google" title="Google OAuth" secrets={secrets} />
                  </Col>
                  <Col xs={24} xl={12}>
                    <OAuthProviderCard provider="github" title="GitHub OAuth" secrets={secrets} />
                  </Col>
                </Row>
              ),
            },
            {
              key: "message",
              label: "短信 SMS 与邮件 Email",
              children: <MessageSettingsTabs secrets={secrets} />,
            },
          ]}
        />
      </Form>
    </div>
  );
}

function ReadonlySettingsPanel({ settings }: { settings: ReadonlySettings | null }) {
  if (!settings) {
    return <Card loading />;
  }
  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Alert
        type="info"
        showIcon
        message="这些是启动级配置，只展示不允许在页面修改。"
        description="数据库 DSN、Redis 地址、HTTP 端口、JWT 密钥路径等配置通常来自 config.yaml 或环境变量，修改后需要重启服务。为了避免误操作导致服务不可用，管理后台只做只读展示。"
      />
      <Card title="服务信息 App / HTTP">
        <Descriptions bordered column={1}>
          <Descriptions.Item label="服务名称">{settings.app.name}</Descriptions.Item>
          <Descriptions.Item label="运行环境">{settings.app.env}</Descriptions.Item>
          <Descriptions.Item label="监听地址">{settings.http.host}</Descriptions.Item>
          <Descriptions.Item label="监听端口">{settings.http.port}</Descriptions.Item>
        </Descriptions>
      </Card>
      <Card title="数据库 Database">
        <Descriptions bordered column={1}>
          <Descriptions.Item label="驱动 Driver">{settings.database.driver}</Descriptions.Item>
          <Descriptions.Item label="连接地址 DSN">{settings.database.dsn || "-"}</Descriptions.Item>
          <Descriptions.Item label="自动迁移 Auto Migrate">{settings.database.auto_migrate ? "开启" : "关闭"}</Descriptions.Item>
        </Descriptions>
      </Card>
      <Card title="Redis">
        <Descriptions bordered column={1}>
          <Descriptions.Item label="是否启用">{settings.redis.enabled ? "启用" : "未启用"}</Descriptions.Item>
          <Descriptions.Item label="地址 Addr">{settings.redis.addr || "-"}</Descriptions.Item>
          <Descriptions.Item label="密码 Password">{settings.redis.password || "未配置"}</Descriptions.Item>
          <Descriptions.Item label="DB">{settings.redis.db}</Descriptions.Item>
        </Descriptions>
      </Card>
      <Card title="JWT / JWKS">
        <Descriptions bordered column={1}>
          <Descriptions.Item label="签发方 Issuer">{settings.jwt.issuer}</Descriptions.Item>
          <Descriptions.Item label="Access Token TTL">{settings.jwt.access_token_ttl}</Descriptions.Item>
          <Descriptions.Item label="Refresh Token TTL">{settings.jwt.refresh_token_ttl}</Descriptions.Item>
          <Descriptions.Item label="私钥路径 Private Key Path">{settings.jwt.private_key_path}</Descriptions.Item>
          <Descriptions.Item label="公钥路径 Public Key Path">{settings.jwt.public_key_path}</Descriptions.Item>
          <Descriptions.Item label="当前 Key ID">{settings.jwt.active_key_id}</Descriptions.Item>
        </Descriptions>
      </Card>
    </Space>
  );
}

function LoginPageSettingsCard() {
  return (
    <Card title="统一登录页 Open Wallet Auth">
      <Typography.Paragraph type="secondary">
        这里配置业务用户看到的统一登录页。保存后立即生效，业务系统只需要传 client_id 和 return_uri，访问来源名称会从接入应用配置中读取。
      </Typography.Paragraph>
      <Row gutter={12}>
        <Col xs={24} lg={10}>
          <Form.Item label="品牌名称 Brand Name" name={["login", "brand_name"]} rules={[{ required: true, message: "请输入品牌名称" }]}>
            <Input placeholder="Open Wallet Auth" />
          </Form.Item>
        </Col>
        <Col xs={24} lg={4}>
          <Form.Item label="品牌标识 Mark" name={["login", "brand_mark"]} rules={[{ required: true, message: "请输入品牌标识" }]}>
            <Input placeholder="L" maxLength={4} />
          </Form.Item>
        </Col>
        <Col xs={24} lg={10}>
          <Form.Item label="副标题 Subtitle" name={["login", "subtitle"]} extra="为空时不展示副标题，适合保持登录页简洁。">
            <Input placeholder="可选，例如 Unified Identity Gateway" />
          </Form.Item>
        </Col>
      </Row>
      <Row gutter={[12, 12]}>
        <Col xs={24} md={8} lg={4}>
          <Form.Item label="注册" name={["login", "enable_register"]} valuePropName="checked">
            <Switch />
          </Form.Item>
        </Col>
        <Col xs={24} md={8} lg={4}>
          <Form.Item label="手机号" name={["login", "enable_phone"]} valuePropName="checked">
            <Switch />
          </Form.Item>
        </Col>
        <Col xs={24} md={8} lg={4}>
          <Form.Item label="GitHub" name={["login", "enable_github"]} valuePropName="checked">
            <Switch />
          </Form.Item>
        </Col>
        <Col xs={24} md={8} lg={4}>
          <Form.Item label="Google" name={["login", "enable_google"]} valuePropName="checked">
            <Switch />
          </Form.Item>
        </Col>
        <Col xs={24} md={8} lg={4}>
          <Form.Item label="钱包" name={["login", "enable_wallet"]} valuePropName="checked">
            <Switch />
          </Form.Item>
        </Col>
      </Row>
    </Card>
  );
}

function AccessSettingsCard() {
  return (
    <Card title="浏览器跨域来源 CORS Allowed Origins">
      <Typography.Paragraph type="secondary">
        这里配置允许哪些业务前端从浏览器直接调用认证服务，例如 BlockX 或 LabelService 的前端域名。保存后立即生效，不需要重启。
      </Typography.Paragraph>
      <Form.Item
        label="Allowed Origins"
        name={["http", "cors_allowed_origins"]}
        extra="填写完整 origin，例如 http://localhost:5173 或 https://blockx.example.com。生产环境不建议使用 * 或 null。"
      >
        <Select mode="tags" tokenSeparators={[",", " "]} placeholder="https://blockx.example.com" />
      </Form.Item>
    </Card>
  );
}

function OAuthCallbackCard() {
  const callbackURL = `${authApiBaseUrl}/api/v1/oauth/{provider}/callback`;
  const copy = async (value: string) => {
    await navigator.clipboard.writeText(value);
    message.success("已复制回调地址");
  };

  return (
    <Card title="OAuth 回调地址 Callback URL">
      <Typography.Paragraph type="secondary">
        Google 和 GitHub 控制台里的回调地址需要指向认证服务后端。下面的地址会根据当前认证服务地址自动生成，实际配置时把 {"{provider}"} 替换为 google 或 github。
      </Typography.Paragraph>
      <Input.Group compact>
        <Input value={callbackURL} readOnly style={{ width: "calc(100% - 96px)" }} />
        <Button icon={<CopyOutlined />} onClick={() => copy(callbackURL)}>
          复制
        </Button>
      </Input.Group>
    </Card>
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
          <Form.Item label="Client ID" name={[...base, "client_id"]} extra="第三方平台创建 OAuth 应用后获得的公开应用标识。">
            <Input placeholder="请输入 OAuth Client ID" />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item label="Client Secret" name={[...base, "client_secret"]} extra={<SecretHint status={secrets[`oauth.${provider}.client_secret`]} />}>
            <Input.Password placeholder="留空则保留当前密钥" autoComplete="new-password" />
          </Form.Item>
        </Col>
      </Row>
      <Form.Item label="Auth URL" name={[...base, "auth_url"]} extra="用户点击第三方登录后，浏览器会跳转到这个授权地址。">
        <Input />
      </Form.Item>
      <Form.Item label="Token URL" name={[...base, "token_url"]} extra="认证服务用授权码 code 换取第三方 access token 的地址。">
        <Input />
      </Form.Item>
      <Form.Item label="UserInfo URL" name={[...base, "user_info_url"]} extra="认证服务用第三方 access token 拉取用户资料的地址。">
        <Input />
      </Form.Item>
      <Form.Item label="Scopes" name={[...base, "scopes"]} extra="需要向第三方申请的权限范围，例如 email、profile、read:user。">
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
                    <Form.Item label="Host" name={[field.name, "host"]} extra="例如 blockx.example.com，用于匹配回调或返回地址。">
                      <Input placeholder="blockx.example.com" />
                    </Form.Item>
                  </Col>
                  <Col span={7}>
                    <Form.Item label="Client ID" name={[field.name, "client_id"]}>
                      <Input />
                    </Form.Item>
                  </Col>
                  <Col span={7}>
                    <Form.Item label="Client Secret" name={[field.name, "client_secret"]} extra="留空则保留该域名当前密钥。">
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

function MessageSettingsTabs({ secrets }: { secrets: RuntimeSettingsResult["secrets"] }) {
  return (
    <Tabs
      type="card"
      items={[
        {
          key: "phone",
          label: "手机号 Phone",
          children: <MessageProviderCard root="phone" title="手机号验证码 Phone Code" secretPrefix="phone.provider" secrets={secrets} />,
        },
        {
          key: "email",
          label: "邮箱 Email",
          children: <MessageProviderCard root="email" title="邮箱验证码 Email Code" secretPrefix="email.provider" secrets={secrets} />,
        },
      ]}
    />
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
  const providerType = Form.useWatch([root, "provider", "type"]) || "noop";
  return (
    <Card title={title}>
      <Typography.Paragraph type="secondary">
        配置验证码发送服务商。Noop 适合本地调试；Webhook 适合接入自有消息网关；SMTP 用于邮件；Aliyun SMS 用于阿里云短信。
      </Typography.Paragraph>
      <Form.Item
        label={root === "phone" ? "Phone Login" : "Email Verification"}
        name={[root, root === "phone" ? "enabled" : "verification_enabled"]}
        valuePropName="checked"
        extra={root === "phone" ? "关闭后手机号验证码登录接口会返回禁用错误。" : "关闭后邮箱验证码发送和校验接口会返回禁用错误。"}
      >
        <Switch />
      </Form.Item>
      <Form.Item label="Provider Type" name={[root, "provider", "type"]} extra="选择当前验证码实际发送方式。">
        <Select options={providerOptions} />
      </Form.Item>

      {providerType === "noop" && (
        <Card size="small">
          <Typography.Text type="secondary">Noop 不会发送真实短信或邮件，适合本地开发和演示环境。</Typography.Text>
        </Card>
      )}

      {providerType === "webhook" && (
        <>
          <Typography.Title level={5}>Webhook 回调</Typography.Title>
          <Form.Item label="Webhook URL" name={[root, "provider", "webhook", "url"]} extra="认证服务会把验证码发送请求 POST 到这个地址。">
            <Input />
          </Form.Item>
          <Form.Item label="Bearer Token" name={[root, "provider", "webhook", "bearer_token"]} extra={<SecretHint status={secrets[`${secretPrefix}.webhook.bearer_token`]} />}>
            <Input.Password placeholder="留空则保留当前密钥" autoComplete="new-password" />
          </Form.Item>
        </>
      )}

      {root === "email" && providerType === "smtp" && (
        <>
          <Typography.Title level={5}>SMTP 邮件</Typography.Title>
          <Row gutter={12}>
            <Col span={16}>
              <Form.Item label="SMTP Host" name={[root, "provider", "smtp", "host"]}>
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
          <Form.Item label="From" name={[root, "provider", "smtp", "from"]} extra="为空时默认使用 SMTP Username。">
            <Input />
          </Form.Item>
        </>
      )}

      {root === "phone" && providerType === "aliyun_sms" && (
        <>
          <Typography.Title level={5}>阿里云短信 Aliyun SMS</Typography.Title>
          <Form.Item label="Access Key ID" name={[root, "provider", "aliyun_sms", "access_key_id"]}>
            <Input />
          </Form.Item>
          <Form.Item label="Access Key Secret" name={[root, "provider", "aliyun_sms", "access_key_secret"]} extra={<SecretHint status={secrets[`${secretPrefix}.aliyun_sms.access_key_secret`]} />}>
            <Input.Password placeholder="留空则保留当前密钥" autoComplete="new-password" />
          </Form.Item>
          <Row gutter={12}>
            <Col span={12}>
              <Form.Item label="Sign Name" name={[root, "provider", "aliyun_sms", "sign_name"]}>
                <Input />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="Template Code" name={[root, "provider", "aliyun_sms", "template_code"]} extra="短信模板需要包含验证码变量 ${code}。">
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
        </>
      )}

      {root === "phone" && providerType === "smtp" && (
        <Card size="small">
          <Typography.Text type="secondary">SMTP 只适用于邮箱验证码。手机号验证码请选择 Webhook 或 Aliyun SMS。</Typography.Text>
        </Card>
      )}
      {root === "email" && providerType === "aliyun_sms" && (
        <Card size="small">
          <Typography.Text type="secondary">Aliyun SMS 只适用于手机号验证码。邮箱验证码请选择 Webhook 或 SMTP。</Typography.Text>
        </Card>
      )}
    </Card>
  );
}

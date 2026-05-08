import {
  GithubOutlined,
  GoogleOutlined,
  LockOutlined,
  MailOutlined,
  MobileOutlined,
  PhoneOutlined,
  UserAddOutlined,
  WalletOutlined,
} from "@ant-design/icons";
import { Alert, Button, Form, Input, Tabs, Typography, message, type TabsProps } from "antd";
import { useEffect, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { publicAuthApi, type OAuthProvider } from "../api/auth";
import { authApiBaseUrl } from "../config";
import type { AuthResult } from "../types/api";

type PasswordValues = {
  email: string;
  password: string;
};

type RegisterValues = PasswordValues & {
  username?: string;
};

type PhoneValues = {
  phone: string;
  code: string;
};

type WalletProvider = {
  request: (input: { method: string; params?: unknown[] }) => Promise<unknown>;
};

type WalletOption = {
  id: string;
  name: string;
  icon?: string;
  provider: WalletProvider;
};

type EIP6963ProviderDetail = {
  info: {
    uuid: string;
    name?: string;
    icon?: string;
    rdns?: string;
  };
  provider: WalletProvider;
};

// UnifiedLoginPage is the branded login surface shared by all business applications.
// UnifiedLoginPage 是所有业务系统共享的品牌化登录入口。
export function UnifiedLoginPage() {
  const [searchParams] = useSearchParams();
  const [loading, setLoading] = useState(false);
  const [phoneCodeLoading, setPhoneCodeLoading] = useState(false);
  const [wallets, setWallets] = useState<WalletOption[]>([]);
  const [successToken, setSuccessToken] = useState("");
  const [pageConfig, setPageConfig] = useState<UnifiedLoginConfig>(() => defaultUnifiedLoginConfig());
  const [appName, setAppName] = useState("");

  const clientID = searchParams.get("client_id") || "default";
  const returnURI = searchParams.get("return_uri") || "";
  const redirectURI = `${authApiBaseUrl}/api/v1/oauth/{provider}/callback`;

  const fallbackWallet = useMemo<WalletOption | null>(() => {
    const ethereum = (window as unknown as { ethereum?: WalletProvider }).ethereum;
    if (!ethereum?.request) return null;
    return { id: "browser-wallet", name: "浏览器钱包", provider: ethereum };
  }, []);

  const availableWallets = wallets.length > 0 ? wallets : fallbackWallet ? [fallbackWallet] : [];

  useEffect(() => {
    const onProvider = (event: Event) => {
      const detail = (event as CustomEvent<EIP6963ProviderDetail>).detail;
      if (!detail?.provider?.request || !detail.info?.uuid) return;
      setWallets((items) => {
        if (items.some((item) => item.id === detail.info.uuid)) return items;
        return [
          ...items,
          {
            id: detail.info.uuid,
            name: detail.info.name || detail.info.rdns || "浏览器钱包",
            icon: detail.info.icon,
            provider: detail.provider,
          },
        ];
      });
    };

    window.addEventListener("eip6963:announceProvider", onProvider as EventListener);
    window.dispatchEvent(new Event("eip6963:requestProvider"));
    return () => window.removeEventListener("eip6963:announceProvider", onProvider as EventListener);
  }, []);

  useEffect(() => {
    let cancelled = false;
    const loadLoginConfig = async () => {
      try {
        const result = await publicAuthApi.loginConfig(clientID);
        if (cancelled) return;
        setAppName(result.client.name || result.client.client_id);
        setPageConfig({
          brandName: result.login.brand_name || "Open Wallet Auth",
          brandMark: result.login.brand_mark || "L",
          subtitle: result.login.subtitle || "",
          enableRegister: result.login.enable_register,
          enablePhone: result.login.enable_phone,
          enableGitHub: result.login.enable_github,
          enableGoogle: result.login.enable_google,
          enableWallet: result.login.enable_wallet,
        });
      } catch (err: any) {
        if (!cancelled) {
          setAppName(clientID);
          message.error(err.message || "加载登录配置失败");
        }
      }
    };

    loadLoginConfig();
    return () => {
      cancelled = true;
    };
  }, [clientID]);

  // finishLogin returns the identity token to the business callback through a URL fragment.
  // finishLogin 将认证结果通过 URL fragment 返回业务系统，避免 token 进入服务端访问日志。
  const finishLogin = (result: AuthResult) => {
    if (!returnURI) {
      setSuccessToken(result.token.access_token);
      message.success("登录成功");
      return;
    }
    const values = new URLSearchParams({
      access_token: result.token.access_token,
      token_type: result.token.token_type || "Bearer",
      expires_at: result.token.expires_at,
    });
    window.location.href = `${returnURI}#${values.toString()}`;
  };

  const submitPasswordLogin = async (values: PasswordValues) => {
    try {
      setLoading(true);
      const result = await publicAuthApi.login({
        client_id: clientID,
        email: values.email.trim(),
        password: values.password,
      });
      finishLogin(result);
    } catch (err: any) {
      message.error(err.message || "登录失败");
    } finally {
      setLoading(false);
    }
  };

  const submitRegister = async (values: RegisterValues) => {
    try {
      setLoading(true);
      const email = values.email.trim();
      const result = await publicAuthApi.register({
        client_id: clientID,
        username: values.username?.trim() || email.split("@")[0],
        email,
        password: values.password,
      });
      finishLogin(result);
    } catch (err: any) {
      message.error(err.message || "注册失败");
    } finally {
      setLoading(false);
    }
  };

  const sendPhoneCode = async (phone: string) => {
    if (!phone.trim()) {
      message.warning("请先输入手机号");
      return;
    }
    try {
      setPhoneCodeLoading(true);
      const result = await publicAuthApi.sendPhoneCode({ client_id: clientID, phone: phone.trim() });
      message.success(result.dev_code ? `验证码已生成：${result.dev_code}` : "验证码已发送");
    } catch (err: any) {
      message.error(err.message || "发送验证码失败");
    } finally {
      setPhoneCodeLoading(false);
    }
  };

  const submitPhoneLogin = async (values: PhoneValues) => {
    try {
      setLoading(true);
      const result = await publicAuthApi.phoneLogin({
        client_id: clientID,
        phone: values.phone.trim(),
        code: values.code.trim(),
      });
      finishLogin(result);
    } catch (err: any) {
      message.error(err.message || "手机号登录失败");
    } finally {
      setLoading(false);
    }
  };

  const loginWithWallet = async (provider: WalletProvider) => {
    try {
      setLoading(true);
      const accounts = (await provider.request({ method: "eth_requestAccounts" })) as string[];
      const address = accounts?.[0];
      if (!address) {
        message.error("未获取到钱包地址");
        return;
      }
      const chainHex = (await provider.request({ method: "eth_chainId" })) as string;
      const chainID = Number.parseInt(chainHex, 16) || 1;
      const nonce = await publicAuthApi.walletNonce({
        address,
        chain_id: chainID,
        domain: window.location.host,
      });
      const signature = (await provider.request({
        method: "personal_sign",
        params: [nonce.message, address],
      })) as string;
      const result = await publicAuthApi.walletVerify({
        client_id: clientID,
        address,
        nonce: nonce.nonce,
        signature,
      });
      finishLogin(result);
    } catch (err: any) {
      message.error(err.message || "钱包登录失败");
    } finally {
      setLoading(false);
    }
  };

  const startOAuth = async (provider: OAuthProvider) => {
    try {
      setLoading(true);
      const result = await publicAuthApi.startOAuth(provider, {
        client_id: clientID,
        redirect_uri: redirectURI.replace("{provider}", provider),
        return_uri: returnURI || undefined,
      });
      window.location.href = result.auth_url;
    } catch (err: any) {
      message.error(err.message || "第三方登录暂不可用");
    } finally {
      setLoading(false);
    }
  };

  const loginTabs = buildLoginTabs({
    loading,
    phoneCodeLoading,
    pageConfig,
    submitPasswordLogin,
    sendPhoneCode,
    submitPhoneLogin,
    submitRegister,
  });

  return (
    <main className="unified-login-page">
      <section className="unified-brand">
        <div className="unified-brand-inner">
          <div className="hosted-brand-mark">{pageConfig.brandMark}</div>
          <Typography.Title level={1}>{pageConfig.brandName}</Typography.Title>
          {pageConfig.subtitle && <p className="unified-subtitle">{pageConfig.subtitle}</p>}
          <div className="unified-app-banner">
            <span>登录后将继续访问 {appName || clientID}</span>
          </div>
        </div>
      </section>

      <section className="unified-panel">
        <div className="unified-card">
          <div className="unified-card-head">
            <Typography.Title level={2}>登录账号</Typography.Title>
          </div>

          {successToken && (
            <Alert
              type="success"
              showIcon
              className="unified-alert"
              message="登录成功"
              description="当前没有配置 return_uri，已在本页保留认证结果，方便本地调试。"
            />
          )}

          <Tabs
            defaultActiveKey="password"
            items={loginTabs}
          />

          {(pageConfig.enableGitHub || pageConfig.enableGoogle || pageConfig.enableWallet) && (
            <>
              <div className="unified-divider">
                <span>其他登录方式</span>
              </div>

              <div className="unified-provider-grid">
                {pageConfig.enableGitHub && (
                  <Button icon={<GithubOutlined />} onClick={() => startOAuth("github")} disabled={loading}>
                    GitHub
                  </Button>
                )}
                {pageConfig.enableGoogle && (
                  <Button icon={<GoogleOutlined />} onClick={() => startOAuth("google")} disabled={loading}>
                    Google
                  </Button>
                )}
                {pageConfig.enableWallet && (
                  availableWallets.length === 0 ? (
                    <Button icon={<WalletOutlined />} disabled>
                      钱包
                    </Button>
                  ) : (
                    availableWallets.map((wallet) => (
                      <Button key={wallet.id} icon={<WalletOutlined />} onClick={() => loginWithWallet(wallet.provider)} disabled={loading}>
                        {wallet.name}
                      </Button>
                    ))
                  )
                )}
              </div>
            </>
          )}
        </div>
      </section>
    </main>
  );
}

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
function buildLoginTabs(input: BuildLoginTabsInput): TabsProps["items"] {
  const items: TabsProps["items"] = [
    {
      key: "password",
      label: "邮箱登录",
      children: (
        <Form<PasswordValues> layout="vertical" onFinish={input.submitPasswordLogin}>
          <Form.Item label="邮箱 Email" name="email" rules={[{ required: true, type: "email", message: "请输入有效邮箱" }]}>
            <Input prefix={<MailOutlined />} placeholder="name@example.com" autoComplete="email" />
          </Form.Item>
          <Form.Item label="密码 Password" name="password" rules={[{ required: true, message: "请输入密码" }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="请输入密码" autoComplete="current-password" />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={input.loading} block size="large">
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
            <Input prefix={<UserAddOutlined />} placeholder="可选，默认使用邮箱前缀" autoComplete="username" />
          </Form.Item>
          <Form.Item label="邮箱 Email" name="email" rules={[{ required: true, type: "email", message: "请输入有效邮箱" }]}>
            <Input prefix={<MailOutlined />} placeholder="name@example.com" autoComplete="email" />
          </Form.Item>
          <Form.Item label="密码 Password" name="password" rules={[{ required: true, min: 8, message: "密码至少 8 位" }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="至少 8 位密码" autoComplete="new-password" />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={input.loading} block size="large">
            创建账号并登录
          </Button>
        </Form>
      ),
    });
  }

  return items;
}

type UnifiedLoginConfig = {
  brandName: string;
  brandMark: string;
  subtitle: string;
  enableRegister: boolean;
  enablePhone: boolean;
  enableGitHub: boolean;
  enableGoogle: boolean;
  enableWallet: boolean;
};

// defaultUnifiedLoginConfig keeps the page usable while remote config is loading.
// defaultUnifiedLoginConfig 在远程配置加载前提供可用的页面默认值。
function defaultUnifiedLoginConfig(): UnifiedLoginConfig {
  return {
    brandName: "Open Wallet Auth",
    brandMark: "L",
    subtitle: "",
    enableRegister: true,
    enablePhone: true,
    enableGitHub: true,
    enableGoogle: true,
    enableWallet: true,
  };
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
      <Form.Item label="手机号 Phone" name="phone" rules={[{ required: true, message: "请输入手机号" }]}>
        <Input prefix={<PhoneOutlined />} placeholder="+8613800000000 或 13800000000" autoComplete="tel" />
      </Form.Item>
      <Form.Item label="验证码 Code" name="code" rules={[{ required: true, message: "请输入验证码" }]}>
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
      <Button type="primary" htmlType="submit" loading={loading} block size="large">
        手机号登录
      </Button>
    </Form>
  );
}

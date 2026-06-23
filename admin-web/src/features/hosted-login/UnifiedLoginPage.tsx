import {
  GithubOutlined,
  GoogleOutlined,
  WalletOutlined,
} from "@ant-design/icons";
import { Alert, Button, Tabs, Typography, message } from "antd";
import { useEffect, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { publicAuthApi, type OAuthProvider } from "../../api/auth";
import { ApiError } from "../../api/client";
import { authApiBaseUrl } from "../../config";
import type { AuthResult, AuthUser } from "../../types/api";
import { buildLoginTabs } from "./LoginTabs";
import type {
  EIP6963ProviderDetail,
  PasswordValues,
  PhoneValues,
  RegisterValues,
  UnifiedLoginConfig,
  WalletOption,
  WalletProvider,
} from "./types";

// UnifiedLoginPage is the branded login surface shared by all business applications.
// UnifiedLoginPage 是所有业务系统共享的品牌化登录入口。
export function UnifiedLoginPage() {
  const [searchParams] = useSearchParams();
  const [loading, setLoading] = useState(false);
  const [phoneCodeLoading, setPhoneCodeLoading] = useState(false);
  const [wallets, setWallets] = useState<WalletOption[]>([]);
  const [successToken, setSuccessToken] = useState("");
  const [pageConfig, setPageConfig] = useState<UnifiedLoginConfig>(() =>
    defaultUnifiedLoginConfig(),
  );
  const [appName, setAppName] = useState("");
  const [sessionUser, setSessionUser] = useState<AuthUser | null>(null);

  const clientID = searchParams.get("client_id") || "default";
  const returnURI = searchParams.get("return_uri") || "";
  const redirectURI = `${authApiBaseUrl}/api/v1/oauth/{provider}/callback`;

  const fallbackWallet = useMemo<WalletOption | null>(() => {
    const ethereum = (window as unknown as { ethereum?: WalletProvider })
      .ethereum;
    if (!ethereum?.request) return null;
    return { id: "browser-wallet", name: "浏览器钱包", provider: ethereum };
  }, []);

  const availableWallets =
    wallets.length > 0 ? wallets : fallbackWallet ? [fallbackWallet] : [];

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

    window.addEventListener(
      "eip6963:announceProvider",
      onProvider as EventListener,
    );
    window.dispatchEvent(new Event("eip6963:requestProvider"));
    return () =>
      window.removeEventListener(
        "eip6963:announceProvider",
        onProvider as EventListener,
      );
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
          brandMark: result.login.brand_mark || "O",
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
    const loadSession = async () => {
      setSessionUser(null);
      try {
        const user = await publicAuthApi.session(clientID);
        if (!cancelled) setSessionUser(user);
      } catch {
        if (!cancelled) setSessionUser(null);
      }
    };

    loadLoginConfig();
    loadSession();
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
    window.location.href = appendTokenToReturnURI(returnURI, values);
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

  const loginWithSession = async () => {
    try {
      setLoading(true);
      const result = await publicAuthApi.sessionLogin({ client_id: clientID });
      finishLogin(result);
    } catch (err: any) {
      setSessionUser(null);
      message.error(err.message || "当前登录状态已失效，请重新登录");
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
      const result = await publicAuthApi.sendPhoneCode({
        client_id: clientID,
        phone: phone.trim(),
      });
      message.success(
        result.dev_code ? `验证码已生成：${result.dev_code}` : "验证码已发送",
      );
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
      const accounts = (await provider.request({
        method: "eth_requestAccounts",
      })) as string[];
      const address = accounts?.[0];
      if (!address) {
        message.error("未获取到钱包地址");
        return;
      }
      const chainHex = (await provider.request({
        method: "eth_chainId",
      })) as string;
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
      message.error(oauthLoginErrorMessage(provider, err));
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
          {pageConfig.subtitle && (
            <p className="unified-subtitle">{pageConfig.subtitle}</p>
          )}
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

          {sessionUser && (
            <Alert
              type="info"
              showIcon
              className="unified-alert"
              message="检测到已登录账号"
              description={
                <div className="unified-session-row">
                  <span>
                    {sessionUser.username || sessionUser.email || sessionUser.id}
                  </span>
                  <Button type="primary" size="small" loading={loading} onClick={loginWithSession}>
                    一键登录
                  </Button>
                </div>
              }
            />
          )}

          <Tabs defaultActiveKey="password" items={loginTabs} />

          {(pageConfig.enableGitHub ||
            pageConfig.enableGoogle ||
            pageConfig.enableWallet) && (
            <>
              <div className="unified-divider">
                <span>其他登录方式</span>
              </div>

              <div className="unified-provider-grid">
                {pageConfig.enableGitHub && (
                  <Button
                    icon={<GithubOutlined />}
                    onClick={() => startOAuth("github")}
                    disabled={loading}
                  >
                    GitHub
                  </Button>
                )}
                {pageConfig.enableGoogle && (
                  <Button
                    icon={<GoogleOutlined />}
                    onClick={() => startOAuth("google")}
                    disabled={loading}
                  >
                    Google
                  </Button>
                )}
                {pageConfig.enableWallet &&
                  (availableWallets.length === 0 ? (
                    <Button icon={<WalletOutlined />} disabled>
                      钱包
                    </Button>
                  ) : (
                    availableWallets.map((wallet) => (
                      <Button
                        key={wallet.id}
                        icon={<WalletOutlined />}
                        onClick={() => loginWithWallet(wallet.provider)}
                        disabled={loading}
                      >
                        {wallet.name}
                      </Button>
                    ))
                  ))}
              </div>
            </>
          )}
        </div>
      </section>
    </main>
  );
}

// appendTokenToReturnURI keeps token outside the server-side request path.
// appendTokenToReturnURI 兼容普通 history 路由和 Vue/React hash 路由业务系统。
function appendTokenToReturnURI(returnURI: string, values: URLSearchParams) {
  const tokenQuery = values.toString();
  const hashIndex = returnURI.indexOf("#");
  if (hashIndex < 0) {
    return `${returnURI}#${tokenQuery}`;
  }

  const base = returnURI.slice(0, hashIndex);
  const hashRoute = returnURI.slice(hashIndex + 1);
  const separator = hashRoute.includes("?") ? "&" : "?";
  return `${base}#${hashRoute}${separator}${tokenQuery}`;
}

// oauthLoginErrorMessage keeps user-facing login prompts clear while preserving
// precise configuration hints for operators testing the hosted login page.
function oauthLoginErrorMessage(provider: OAuthProvider, error: unknown) {
  const name = provider === "github" ? "GitHub" : "Google";
  if (error instanceof ApiError && error.code === "OAUTH_PROVIDER_FAILED") {
    const detail = error.message.toLowerCase();
    if (detail.includes("redirect_uri")) {
      return `${name} 登录未配置当前业务域名，请在系统配置里添加对应 Tenant Credentials。`;
    }
    return `${name} 登录尚未配置，请先在系统配置里填写 OAuth Client ID 和 Client Secret。`;
  }
  if (error instanceof Error && error.message) {
    return error.message;
  }
  return "第三方登录暂不可用";
}

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

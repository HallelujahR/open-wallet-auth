import {
  ArrowLeftOutlined,
  ArrowRightOutlined,
  GithubOutlined,
  GoogleOutlined,
  LogoutOutlined,
  PlusOutlined,
  UserOutlined,
  WalletOutlined,
} from "@ant-design/icons";
import {
  Alert,
  App,
  Avatar,
  Button,
  Divider,
  Modal,
  Popconfirm,
  Result,
  Skeleton,
  Tabs,
  Tooltip,
  Typography,
} from "antd";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { publicAuthApi, type OAuthProvider } from "../../api/auth";
import { ApiError } from "../../api/client";
import { authApiBaseUrl } from "../../config";
import { AuthShell } from "../../components/AuthShell";
import type {
  AuthResult,
  BrowserAccount,
  PublicLoginConfig,
} from "../../types/api";
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

export function UnifiedLoginPage() {
  const { message } = App.useApp();
  const [searchParams] = useSearchParams();
  const clientID = searchParams.get("client_id") || "default";
  const returnURI = searchParams.get("return_uri") || "";
  const clientState = searchParams.get("state") || "";
  const contextKey = JSON.stringify([clientID, returnURI, clientState]);
  const currentContext = useRef(contextKey);
  currentContext.current = contextKey;
  const [config, setConfig] = useState<{
    key: string;
    value: PublicLoginConfig;
  } | null>(null);
  const [configError, setConfigError] = useState("");
  const [retry, setRetry] = useState(0);
  const [accounts, setAccounts] = useState<BrowserAccount[]>([]);
  const [accountLoading, setAccountLoading] = useState(true);
  const [accountError, setAccountError] = useState("");
  const [addingAccount, setAddingAccount] = useState(false);
  const [loading, setLoading] = useState(false);
  const [phoneCodeLoading, setPhoneCodeLoading] = useState(false);
  const [wallets, setWallets] = useState<WalletOption[]>([]);
  const [walletChoice, setWalletChoice] = useState<{
    provider: WalletProvider;
    addresses: string[];
  } | null>(null);
  const [successName, setSuccessName] = useState<string | null>(null);
  const busy = useRef(false);
  const sessionRequest = useRef(0);
  const pageConfig = config?.key === contextKey ? config.value : null;
  const loginConfig: UnifiedLoginConfig = {
    brandName: pageConfig?.login.brand_name || "Open Wallet Auth",
    brandMark: pageConfig?.login.brand_mark || "O",
    subtitle: pageConfig?.login.subtitle || "",
    enableRegister: !!pageConfig?.login.enable_register,
    enablePhone: !!pageConfig?.login.enable_phone,
    enableGitHub: !!pageConfig?.login.enable_github,
    enableGoogle: !!pageConfig?.login.enable_google,
    enableWallet: !!pageConfig?.login.enable_wallet,
  };
  const reloadAccounts = useCallback(async () => {
    const request = ++sessionRequest.current;
    setAccountLoading(true);
    try {
      const result = await publicAuthApi.sessions(clientID);
      if (request === sessionRequest.current) {
        setAccounts(result.items);
        setAccountError("");
      }
    } catch (err) {
      if (request === sessionRequest.current) setAccountError(errorText(err));
    } finally {
      if (request === sessionRequest.current) setAccountLoading(false);
    }
  }, [clientID]);

  useEffect(() => {
    let cancelled = false;
    setConfig(null);
    setConfigError("");
    setAddingAccount(false);
    setSuccessName(null);
    setAccounts([]);
    publicAuthApi
      .loginConfig(clientID, returnURI, clientState)
      .then((value) => {
        if (!cancelled) setConfig({ key: contextKey, value });
      })
      .catch((err) => {
        if (!cancelled) setConfigError(errorText(err));
      });
    void reloadAccounts();
    const onFocus = () => {
      if (!busy.current) void reloadAccounts();
    };
    window.addEventListener("focus", onFocus);
    return () => {
      cancelled = true;
      sessionRequest.current++;
      window.removeEventListener("focus", onFocus);
    };
  }, [clientID, returnURI, clientState, contextKey, reloadAccounts, retry]);

  // Refresh display settings without resetting forms or the selected account.
  const configReady = !!pageConfig;
  useEffect(() => {
    if (!configReady) return;
    let cancelled = false;
    let pending = false;
    const refreshConfig = async () => {
      if (pending || document.visibilityState === "hidden") return;
      pending = true;
      try {
        const value = await publicAuthApi.loginConfig(
          clientID,
          returnURI,
          clientState,
        );
        if (!cancelled) setConfig({ key: contextKey, value });
      } catch {
        // A transient refresh failure should not interrupt an active login.
      } finally {
        pending = false;
      }
    };
    const timer = window.setInterval(refreshConfig, 5000);
    window.addEventListener("focus", refreshConfig);
    document.addEventListener("visibilitychange", refreshConfig);
    return () => {
      cancelled = true;
      window.clearInterval(timer);
      window.removeEventListener("focus", refreshConfig);
      document.removeEventListener("visibilitychange", refreshConfig);
    };
  }, [clientID, returnURI, clientState, contextKey, configReady]);

  const fallbackWallet = useMemo(
    () => (window as unknown as { ethereum?: WalletProvider }).ethereum,
    [],
  );
  const availableWallets = wallets.length
    ? wallets
    : fallbackWallet?.request
      ? [{ id: "browser", name: "浏览器钱包", provider: fallbackWallet }]
      : [];
  useEffect(() => {
    const onProvider = (event: Event) => {
      const detail = (event as CustomEvent<EIP6963ProviderDetail>).detail;
      if (!detail?.provider?.request || !detail.info?.uuid) return;
      setWallets((items) =>
        items.some((item) => item.id === detail.info.uuid)
          ? items
          : [
              ...items,
              {
                id: detail.info.uuid,
                name: detail.info.name || "浏览器钱包",
                provider: detail.provider,
              },
            ],
      );
    };
    window.addEventListener("eip6963:announceProvider", onProvider);
    window.dispatchEvent(new Event("eip6963:requestProvider"));
    return () =>
      window.removeEventListener("eip6963:announceProvider", onProvider);
  }, []);

  const run = async (action: () => Promise<void>) => {
    if (busy.current || !pageConfig) return;
    busy.current = true;
    setLoading(true);
    try {
      await action();
    } catch (err) {
      message.error(errorText(err));
    } finally {
      busy.current = false;
      setLoading(false);
    }
  };
  const finishLogin = (result: AuthResult) => {
    // A response started for another application must never redirect the new page.
    if (currentContext.current !== contextKey || !pageConfig) return;
    if (!pageConfig.return_uri) {
      setSuccessName(
        result.user.username ||
          result.user.email ||
          result.user.phone ||
          "当前账号",
      );
      setAddingAccount(false);
      void reloadAccounts();
      return;
    }
    const values = new URLSearchParams({
      access_token: result.token.access_token,
      token_type: result.token.token_type || "Bearer",
      expires_at: result.token.expires_at,
    });
    window.location.assign(
      appendTokenToReturnURI(pageConfig.return_uri, values, clientState),
    );
  };
  const submitPasswordLogin = (values: PasswordValues) =>
    run(async () => {
      finishLogin(
        await publicAuthApi.login({
          client_id: clientID,
          return_uri: pageConfig!.return_uri,
          email: values.email.trim(),
          password: values.password,
        }),
      );
    });
  const submitRegister = (values: RegisterValues) =>
    run(async () => {
      const email = values.email.trim();
      finishLogin(
        await publicAuthApi.register({
          client_id: clientID,
          return_uri: pageConfig!.return_uri,
          username: values.username?.trim() || email.split("@")[0],
          email,
          password: values.password,
        }),
      );
    });
  const submitPhoneLogin = (values: PhoneValues) =>
    run(async () => {
      finishLogin(
        await publicAuthApi.phoneLogin({
          client_id: clientID,
          return_uri: pageConfig!.return_uri,
          phone: values.phone.trim(),
          code: values.code.trim(),
        }),
      );
    });
  const sendPhoneCode = async (phone: string) => {
    if (!phone.trim()) {
      message.warning("请先输入手机号");
      return false;
    }
    setPhoneCodeLoading(true);
    try {
      const result = await publicAuthApi.sendPhoneCode({
        client_id: clientID,
        phone: phone.trim(),
      });
      message.success(
        result.dev_code ? `验证码：${result.dev_code}` : "验证码已发送，请查收",
      );
      return true;
    } catch (err) {
      message.error(errorText(err));
      return false;
    } finally {
      setPhoneCodeLoading(false);
    }
  };
  const loginWithSession = (account: BrowserAccount) =>
    run(async () => {
      try {
        const result = await publicAuthApi.sessionLogin({
          client_id: clientID,
          session_id: account.session_id,
          return_uri: pageConfig!.return_uri,
        });
        if (result.user.id !== account.id)
          throw new Error("账号状态已变化，请重新选择");
        finishLogin(result);
      } catch (err) {
        await reloadAccounts();
        throw err;
      }
    });
  const logout = (sessionID?: string) =>
    run(async () => {
      await publicAuthApi.logoutSession(
        sessionID ? { session_id: sessionID } : { all: true },
      );
      await reloadAccounts();
      message.success(sessionID ? "已退出此账号" : "已退出本浏览器全部账号");
    });
  const signWallet = async (provider: WalletProvider, address: string) => {
    const chainHex = (await provider.request({
      method: "eth_chainId",
    })) as string;
    const nonce = await publicAuthApi.walletNonce({
      address,
      chain_id: Number.parseInt(chainHex, 16) || 1,
      domain: window.location.host,
    });
    const signature = (await provider.request({
      method: "personal_sign",
      params: [nonce.message, address],
    })) as string;
    finishLogin(
      await publicAuthApi.walletVerify({
        client_id: clientID,
        return_uri: pageConfig!.return_uri,
        address,
        nonce: nonce.nonce,
        signature,
      }),
    );
  };
  const loginWithWallet = (provider: WalletProvider) =>
    run(async () => {
      const addresses = (await provider.request({
        method: "eth_requestAccounts",
      })) as string[];
      if (!addresses?.length) throw new Error("未获取到钱包地址");
      if (addresses.length > 1) {
        setWalletChoice({ provider, addresses });
        return;
      }
      await signWallet(provider, addresses[0]);
    });
  const startOAuth = (provider: OAuthProvider) =>
    run(async () => {
      const result = await publicAuthApi.startOAuth(provider, {
        client_id: clientID,
        redirect_uri: `${authApiBaseUrl}/api/v1/oauth/${provider}/callback`,
        return_uri: pageConfig!.return_uri || undefined,
        client_state: clientState || undefined,
      });
      if (currentContext.current === contextKey)
        window.location.assign(result.auth_url);
    });
  const chooseAccount = accounts.length > 0 && !addingAccount;
  return (
    <AuthShell
      heroTitle={pageConfig?.login.hero_title || undefined}
      brand={loginConfig.brandName}
      mark={loginConfig.brandMark}
      context={pageConfig?.client.name || "统一身份认证"}
    >
      {configError ? (
        <Result
          status="warning"
          title="暂时无法登录"
          subTitle={configError}
          extra={
            <Button onClick={() => setRetry((value) => value + 1)}>重试</Button>
          }
        />
      ) : !pageConfig || (accountLoading && !accounts.length) ? (
        <div className="auth-skeleton">
          <Skeleton active paragraph={{ rows: 6 }} />
        </div>
      ) : successName !== null ? (
        <Result
          status="success"
          title="登录成功"
          subTitle={`已登录为 ${successName}`}
          extra={
            <Button type="primary" onClick={() => setSuccessName(null)}>
              管理登录账号
            </Button>
          }
        />
      ) : (
        <>
          {addingAccount && accounts.length > 0 && (
            <Button
              type="text"
              className="auth-back"
              icon={<ArrowLeftOutlined />}
              onClick={() => setAddingAccount(false)}
            >
              返回账号列表
            </Button>
          )}
          <div className="auth-card-heading">
            <span className="eyebrow">
              {chooseAccount ? "YOUR ACCOUNTS" : "WELCOME BACK"}
            </span>
            <Typography.Title level={2}>
              {chooseAccount
                ? "选择一个账号"
                : addingAccount
                  ? "登录其他账号"
                  : "欢迎回来"}
            </Typography.Title>
            <Typography.Paragraph type="secondary">
              继续访问 <strong>{pageConfig.client.name || clientID}</strong>
            </Typography.Paragraph>
          </div>
          {accountError && (
            <Alert
              type="warning"
              showIcon
              message="无法读取已登录账号"
              description={accountError}
              action={
                <Button size="small" onClick={() => void reloadAccounts()}>
                  重试
                </Button>
              }
              className="page-alert"
            />
          )}
          {chooseAccount ? (
            <div className="account-picker">
              <div className="account-list">
                {accounts.map((account) => (
                  <div className="account-row" key={account.session_id}>
                    <button
                      className="account-choice"
                      disabled={loading || accountLoading}
                      onClick={() => void loginWithSession(account)}
                    >
                      <Avatar
                        size={44}
                        icon={<UserOutlined />}
                        src={account.avatar}
                        className="account-avatar"
                      />
                      <span className="account-info">
                        <strong>
                          {account.username ||
                            account.email ||
                            account.phone ||
                            "已登录账号"}
                        </strong>
                        <span>
                          {account.email ||
                            account.phone ||
                            `ID · ${account.id.slice(-12)}`}
                        </span>
                      </span>
                      <ArrowRightOutlined className="account-arrow" />
                    </button>
                    <Popconfirm
                      title="退出这个账号？"
                      description="将移除此浏览器中的登录状态。"
                      onConfirm={() => logout(account.session_id)}
                    >
                      <Tooltip title="退出此账号">
                        <Button
                          type="text"
                          aria-label={`退出 ${account.username || account.email || account.id}`}
                          disabled={loading}
                          icon={<LogoutOutlined />}
                        />
                      </Tooltip>
                    </Popconfirm>
                  </div>
                ))}
              </div>
              <Button
                block
                size="large"
                icon={<PlusOutlined />}
                className="other-account"
                disabled={loading}
                onClick={() => setAddingAccount(true)}
              >
                使用其他账号
              </Button>
              <Popconfirm
                title="退出本浏览器的全部账号？"
                onConfirm={() => logout()}
              >
                <Button
                  type="text"
                  block
                  disabled={loading}
                  className="signout-all"
                >
                  退出全部账号
                </Button>
              </Popconfirm>
            </div>
          ) : (
            <>
              <Tabs
                key={contextKey}
                defaultActiveKey="password"
                items={buildLoginTabs({
                  loading: loading || accountLoading,
                  phoneCodeLoading,
                  pageConfig: loginConfig,
                  submitPasswordLogin,
                  submitRegister,
                  submitPhoneLogin,
                  sendPhoneCode,
                })}
              />
              {(loginConfig.enableGoogle ||
                loginConfig.enableGitHub ||
                loginConfig.enableWallet) && (
                <>
                  <Divider plain className="auth-divider">
                    或通过以下方式继续
                  </Divider>
                  <div className="provider-buttons">
                    {loginConfig.enableGoogle && (
                      <Button
                        icon={<GoogleOutlined />}
                        disabled={loading}
                        onClick={() => void startOAuth("google")}
                      >
                        Google
                      </Button>
                    )}
                    {loginConfig.enableGitHub && (
                      <Button
                        icon={<GithubOutlined />}
                        disabled={loading}
                        onClick={() => void startOAuth("github")}
                      >
                        GitHub
                      </Button>
                    )}
                    {loginConfig.enableWallet &&
                      (availableWallets.length ? (
                        availableWallets.map((wallet) => (
                          <Button
                            key={wallet.id}
                            icon={<WalletOutlined />}
                            disabled={loading}
                            onClick={() =>
                              void loginWithWallet(wallet.provider)
                            }
                          >
                            {wallet.name}
                          </Button>
                        ))
                      ) : (
                        <Tooltip title="请先安装或打开浏览器钱包">
                          <Button icon={<WalletOutlined />} disabled>
                            钱包
                          </Button>
                        </Tooltip>
                      ))}
                  </div>
                </>
              )}
            </>
          )}
          {loginConfig.subtitle && (
            <p className="auth-card-note">{loginConfig.subtitle}</p>
          )}
        </>
      )}
      <Modal
        title="选择钱包地址"
        open={!!walletChoice}
        footer={null}
        onCancel={() => setWalletChoice(null)}
      >
        {walletChoice?.addresses.map((address) => (
          <Button
            block
            key={address}
            className="wallet-choice"
            onClick={() => {
              const choice = walletChoice;
              setWalletChoice(null);
              void run(() => signWallet(choice.provider, address));
            }}
          >
            {address}
          </Button>
        ))}
      </Modal>
    </AuthShell>
  );
}

function errorText(error: unknown) {
  if (error instanceof ApiError && error.code === "OAUTH_PROVIDER_FAILED")
    return "该第三方登录暂不可用，请使用其他方式或联系管理员";
  if (error instanceof ApiError && error.code === "AUTH_INVALID_REFRESH_TOKEN")
    return "登录状态已过期，请重新登录";
  return error instanceof Error ? error.message : "操作失败，请稍后重试";
}
function appendTokenToReturnURI(
  returnURI: string,
  values: URLSearchParams,
  state: string,
) {
  if (state) values.set("state", state);
  const hashIndex = returnURI.indexOf("#");
  if (hashIndex < 0) return `${returnURI}#${values}`;
  const base = returnURI.slice(0, hashIndex);
  const route = returnURI.slice(hashIndex + 1);
  return `${base}#${route}${route.includes("?") ? "&" : "?"}${values}`;
}

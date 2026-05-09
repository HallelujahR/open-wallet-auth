const DEFAULT_TOKEN_KEY = "owa_access_token";
const DEFAULT_REDIRECT_KEY = "owa_post_login_redirect";

function trimSlash(value) {
  return String(value || "").replace(/\/+$/, "");
}

function requireConfig(config) {
  if (!config?.authBaseURL) throw new Error("authBaseURL is required");
  if (!config?.clientID) throw new Error("clientID is required");
  return {
    tokenKey: DEFAULT_TOKEN_KEY,
    redirectKey: DEFAULT_REDIRECT_KEY,
    ...config,
    authBaseURL: trimSlash(config.authBaseURL),
  };
}

async function readJSON(response) {
  const body = await response.json().catch(() => null);
  if (!response.ok || (body?.code && body.code !== "OK")) {
    throw new Error(body?.message || `Open Wallet Auth request failed: ${response.status}`);
  }
  return body;
}

// createAuthClient builds the browser SDK facade used by business frontends.
// createAuthClient 创建浏览器端 SDK 门面，业务前端只需要关心登录入口和回调处理。
export function createAuthClient(config) {
  const cfg = requireConfig(config);

  return {
    config: cfg,

    // buildLoginURL returns the hosted-login URL without changing browser state.
    // buildLoginURL 只生成统一登录地址，不修改浏览器状态。
    buildLoginURL(options = {}) {
      return buildLoginURL(cfg, options);
    },

    // login stores the current page and redirects to the hosted login page.
    // login 保存当前页面后跳转统一登录页。
    login(options = {}) {
      const redirect = options.redirect || window.location.pathname + window.location.search;
      window.localStorage.setItem(cfg.redirectKey, redirect);
      window.location.assign(buildLoginURL(cfg, options));
    },

    // parseCallback reads the callback hash from the hosted login redirect.
    // parseCallback 解析统一登录页回调携带的 hash。
    parseCallback(hash = window.location.hash) {
      return parseCallbackHash(hash);
    },

    // consumeRedirect reads and clears the post-login redirect path.
    // consumeRedirect 读取并清理登录完成后的业务回跳路径。
    consumeRedirect(fallback = "/") {
      const redirect = window.localStorage.getItem(cfg.redirectKey) || fallback;
      window.localStorage.removeItem(cfg.redirectKey);
      return redirect;
    },

    // setAccessToken stores an identity access token when the frontend chooses to keep it.
    // setAccessToken 在前端需要保留中台 token 时写入本地存储。
    setAccessToken(token) {
      window.localStorage.setItem(cfg.tokenKey, token);
    },

    // getAccessToken reads the stored identity access token.
    // getAccessToken 读取本地保存的中台 token。
    getAccessToken() {
      return window.localStorage.getItem(cfg.tokenKey) || "";
    },

    // clearAccessToken removes the stored identity access token.
    // clearAccessToken 清理本地保存的中台 token。
    clearAccessToken() {
      window.localStorage.removeItem(cfg.tokenKey);
    },

    // authHeader builds a Bearer header for APIs that accept identity tokens directly.
    // authHeader 为直接接受中台 token 的接口生成 Bearer 请求头。
    authHeader(token = this.getAccessToken()) {
      return token ? { Authorization: `Bearer ${token}` } : {};
    },

    // loginWithPassword signs in through the identity center.
    // loginWithPassword 使用邮箱密码登录认证中台。
    async loginWithPassword(input) {
      return postAuth(cfg, "/api/v1/auth/login", input);
    },

    // register creates an identity-center account.
    // register 在认证中台创建统一身份账号。
    async register(input) {
      return postAuth(cfg, "/api/v1/auth/register", input);
    },

    // requestWalletNonce creates a sign-in challenge for browser wallet login.
    // requestWalletNonce 为浏览器钱包登录创建签名挑战。
    async requestWalletNonce(input) {
      return post(cfg, "/api/v1/wallet/nonce", input);
    },

    // verifyWalletSignature verifies wallet ownership and returns an identity token.
    // verifyWalletSignature 校验钱包签名并返回中台 token。
    async verifyWalletSignature(input) {
      return post(cfg, "/api/v1/wallet/verify", { client_id: cfg.clientID, ...input });
    },

    // startOAuthLogin asks the auth service for a provider authorization URL.
    // startOAuthLogin 获取第三方 OAuth 授权地址。
    async startOAuthLogin(provider, input = {}) {
      const params = new URLSearchParams({
        client_id: cfg.clientID,
        redirect_uri: input.redirectURI || `${cfg.authBaseURL}/api/v1/oauth/${provider}/callback`,
      });
      if (input.returnURI) params.set("return_uri", input.returnURI);
      const response = await fetch(`${cfg.authBaseURL}/api/v1/oauth/${provider}/start?${params}`);
      return readJSON(response);
    },
  };
}

export function buildLoginURL(config, options = {}) {
  const cfg = requireConfig(config);
  const params = new URLSearchParams({
    client_id: cfg.clientID,
    return_uri: options.returnURI || cfg.returnURI || `${window.location.origin}/auth/oauth/callback`,
  });
  return `${cfg.authBaseURL}/login?${params.toString()}`;
}

// parseCallbackHash reads the token returned by the hosted login page.
// parseCallbackHash 解析统一登录页回跳时携带的 access token。
export function parseCallbackHash(hash) {
  const params = new URLSearchParams(String(hash || "").replace(/^#/, ""));
  const accessToken = params.get("access_token");
  if (!accessToken) return null;
  return {
    accessToken,
    tokenType: params.get("token_type") || "Bearer",
    expiresAt: params.get("expires_at") || "",
  };
}

async function postAuth(config, path, input) {
  return post(config, path, { client_id: config.clientID, ...input });
}

async function post(config, path, input) {
  const response = await fetch(`${config.authBaseURL}${path}`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  return readJSON(response);
}

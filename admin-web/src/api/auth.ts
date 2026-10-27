import { request } from "./client";
import type {
  AuthResult,
  AuthUser,
  BrowserAccount,
  OAuthStartResult,
  PublicLoginConfig,
  WalletNonceResult,
} from "../types/api";

export type OAuthProvider = "github" | "google";

// publicAuthApi wraps user-facing identity endpoints without admin credentials.
// publicAuthApi 封装面向业务用户的公开认证接口，不携带管理后台凭据。
export const publicAuthApi = {
  loginConfig(clientID: string, returnURI = "", state = "") {
    return request<PublicLoginConfig>({
      url: "/api/v1/public/login-config",
      method: "GET",
      params: { client_id: clientID, return_uri: returnURI, state },
      skipAdminAuth: true,
    });
  },
  register(data: {
    client_id: string;
    return_uri?: string;
    username: string;
    email: string;
    password: string;
  }) {
    return request<AuthResult>({
      url: "/api/v1/auth/register",
      method: "POST",
      data,
      skipAdminAuth: true,
    });
  },
  login(data: {
    client_id: string;
    return_uri?: string;
    email: string;
    password: string;
  }) {
    return request<AuthResult>({
      url: "/api/v1/auth/login",
      method: "POST",
      data,
      skipAdminAuth: true,
    });
  },
  sessions(clientID: string) {
    return request<{ items: BrowserAccount[] }>({
      url: "/api/v1/auth/sessions",
      method: "GET",
      params: { client_id: clientID },
      skipAdminAuth: true,
    });
  },
  logoutSession(data: { session_id?: string; all?: boolean }) {
    return request<{ logged_out: boolean }>({
      url: "/api/v1/auth/sessions/logout",
      method: "POST",
      data,
      skipAdminAuth: true,
    });
  },
  session(clientID: string) {
    return request<AuthUser>({
      url: "/api/v1/auth/session",
      method: "GET",
      params: { client_id: clientID },
      skipAdminAuth: true,
    });
  },
  sessionLogin(data: {
    client_id: string;
    session_id: string;
    return_uri?: string;
  }) {
    return request<AuthResult>({
      url: "/api/v1/auth/session/login",
      method: "POST",
      data,
      skipAdminAuth: true,
    });
  },
  sendPhoneCode(data: { client_id: string; phone: string }) {
    return request<{ phone: string; expires_at: string; dev_code?: string }>({
      url: "/api/v1/phone/code",
      method: "POST",
      data,
      skipAdminAuth: true,
    });
  },
  phoneLogin(data: {
    client_id: string;
    return_uri?: string;
    phone: string;
    code: string;
  }) {
    return request<AuthResult>({
      url: "/api/v1/phone/login",
      method: "POST",
      data,
      skipAdminAuth: true,
    });
  },
  walletNonce(data: { address: string; domain: string; chain_id: number }) {
    return request<WalletNonceResult>({
      url: "/api/v1/wallet/nonce",
      method: "POST",
      data,
      skipAdminAuth: true,
    });
  },
  walletVerify(data: {
    client_id: string;
    return_uri?: string;
    address: string;
    nonce: string;
    signature: string;
  }) {
    return request<AuthResult>({
      url: "/api/v1/wallet/verify",
      method: "POST",
      data,
      skipAdminAuth: true,
    });
  },
  startOAuth(
    provider: OAuthProvider,
    data: {
      client_id: string;
      redirect_uri: string;
      return_uri?: string;
      client_state?: string;
    },
  ) {
    return request<OAuthStartResult>({
      url: `/api/v1/oauth/${provider}/start`,
      method: "GET",
      params: data,
      skipAdminAuth: true,
    });
  },
};

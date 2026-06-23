import { request } from "./client";
import { authApiBaseUrl } from "../config";
import type {
  AdminLoginResult,
  Client,
  ClientCreateInput,
  ClientMember,
  ClientMemberInput,
  HealthStatus,
  IdentityDetail,
  IdentityUser,
  LoginLog,
  PageResult,
  SecurityEvent,
  Session,
  RuntimeSettings,
  RuntimeSettingsResult,
} from "../types/api";

export const adminApi = {
  login(data: { username: string; password: string }) {
    return request<AdminLoginResult>({
      baseURL: authApiBaseUrl,
      url: "/api/v1/admin/login",
      method: "POST",
      data,
      skipAdminAuth: true,
    });
  },
  health() {
    return request<HealthStatus>({
      url: "/healthz",
      method: "GET",
      skipAdminAuth: true,
    });
  },
  listUsers(params: { keyword?: string; status?: string; page?: number; page_size?: number }) {
    return request<PageResult<IdentityUser>>({ url: "/api/v1/admin/users", method: "GET", params });
  },
  getUser(userId: string) {
    return request<IdentityDetail>({ url: `/api/v1/admin/users/${userId}`, method: "GET" });
  },
  updateUserStatus(userId: string, status: string) {
    return request<{ updated: boolean }>({
      url: `/api/v1/admin/users/${userId}/status`,
      method: "PATCH",
      data: { status },
    });
  },
  setUserPassword(userId: string, password: string) {
    return request<{ password_updated: boolean }>({
      url: `/api/v1/admin/users/${userId}/password`,
      method: "PATCH",
      data: { password },
    });
  },
  revokeUserSessions(userId: string, clientId?: string) {
    return request<{ revoked: number }>({
      url: `/api/v1/admin/users/${userId}/sessions`,
      method: "DELETE",
      params: clientId ? { client_id: clientId } : undefined,
    });
  },
  unbindWallet(userId: string, walletId: string) {
    return request<{ unbound: boolean }>({
      url: `/api/v1/admin/users/${userId}/wallets/${walletId}`,
      method: "DELETE",
    });
  },
  unbindOAuthAccount(userId: string, accountId: string) {
    return request<{ unbound: boolean }>({
      url: `/api/v1/admin/users/${userId}/oauth-accounts/${accountId}`,
      method: "DELETE",
    });
  },
  listSessions(params: { user_id?: string; client_id?: string; active_only?: boolean }) {
    return request<{ items: Session[] }>({ url: "/api/v1/admin/sessions", method: "GET", params });
  },
  revokeSession(sessionId: string) {
    return request<{ revoked: boolean }>({ url: `/api/v1/admin/sessions/${sessionId}`, method: "DELETE" });
  },
  listLoginLogs(params: { user_id?: string; client_id?: string; page?: number; page_size?: number }) {
    return request<PageResult<LoginLog>>({ url: "/api/v1/admin/login-logs", method: "GET", params });
  },
  listSecurityEvents(params: { user_id?: string; event_type?: string; page?: number; page_size?: number }) {
    return request<PageResult<SecurityEvent>>({ url: "/api/v1/admin/security-events", method: "GET", params });
  },
  listClients() {
    return request<Client[]>({ url: "/api/v1/admin/clients", method: "GET" });
  },
  createClient(input: ClientCreateInput) {
    return request<Client>({ url: "/api/v1/admin/clients", method: "POST", data: input });
  },
  updateClientAccessPolicy(clientId: string, whitelistEnabled: boolean) {
    return request<Client>({
      url: `/api/v1/admin/clients/${clientId}`,
      method: "PATCH",
      data: { whitelist_enabled: whitelistEnabled },
    });
  },
  listClientMembers(clientId: string) {
    return request<ClientMember[]>({ url: `/api/v1/admin/clients/${clientId}/members`, method: "GET" });
  },
  addClientMember(clientId: string, input: ClientMemberInput) {
    return request<ClientMember>({ url: `/api/v1/admin/clients/${clientId}/members`, method: "POST", data: input });
  },
  updateClientMember(clientId: string, memberId: string, input: ClientMemberInput) {
    return request<{ updated: boolean }>({
      url: `/api/v1/admin/clients/${clientId}/members/${memberId}`,
      method: "PATCH",
      data: input,
    });
  },
  deleteClientMember(clientId: string, memberId: string) {
    return request<{ deleted: boolean }>({
      url: `/api/v1/admin/clients/${clientId}/members/${memberId}`,
      method: "DELETE",
    });
  },
  getSettings() {
    return request<RuntimeSettingsResult>({ url: "/api/v1/admin/settings", method: "GET" });
  },
  updateSettings(input: RuntimeSettings) {
    return request<RuntimeSettingsResult>({ url: "/api/v1/admin/settings", method: "PUT", data: input });
  },
};

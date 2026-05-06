const BASE_URL_KEY = "owa_admin_base_url";
const ADMIN_TOKEN_KEY = "owa_admin_token";

export type AdminSession = {
  baseUrl: string;
  adminToken: string;
};

export function getAdminSession(): AdminSession | null {
  const baseUrl = localStorage.getItem(BASE_URL_KEY);
  const adminToken = localStorage.getItem(ADMIN_TOKEN_KEY);
  if (!baseUrl || !adminToken) return null;
  return { baseUrl, adminToken };
}

export function saveAdminSession(session: AdminSession) {
  localStorage.setItem(BASE_URL_KEY, session.baseUrl.replace(/\/$/, ""));
  localStorage.setItem(ADMIN_TOKEN_KEY, session.adminToken);
}

export function clearAdminSession() {
  localStorage.removeItem(BASE_URL_KEY);
  localStorage.removeItem(ADMIN_TOKEN_KEY);
}

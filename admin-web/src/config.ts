const DEFAULT_AUTH_API_BASE_URL = "http://localhost:8081";

// authApiBaseUrl is the single backend entry configured at build time.
// authApiBaseUrl 是管理后台唯一的后端入口，由构建环境变量统一配置。
export const authApiBaseUrl = (import.meta.env.VITE_AUTH_API_BASE_URL || DEFAULT_AUTH_API_BASE_URL).replace(/\/$/, "");

const DEFAULT_AUTH_API_BASE_URL = "http://localhost:8081";

// runtimeAuthApiBaseUrl keeps the console same-origin when it is served by the auth service.
// runtimeAuthApiBaseUrl 让管理台跟随当前访问域名请求 API，避免 HTTP/HTTPS 混用触发跨域预检。
function runtimeAuthApiBaseUrl() {
  if (typeof window !== "undefined" && window.location.origin) {
    return window.location.origin;
  }
  return DEFAULT_AUTH_API_BASE_URL;
}

// authApiBaseUrl is the single backend entry configured at build time or resolved at runtime.
// authApiBaseUrl 是管理后台唯一的后端入口，优先使用构建配置，否则使用当前页面同源地址。
export const authApiBaseUrl = (import.meta.env.VITE_AUTH_API_BASE_URL || runtimeAuthApiBaseUrl()).replace(/\/$/, "");

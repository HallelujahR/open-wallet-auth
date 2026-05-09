import axios from "axios";
import { authApiBaseUrl } from "../config";
import { getAdminSession } from "../store/authStore";
import type { ApiEnvelope } from "../types/api";

declare module "axios" {
  export interface InternalAxiosRequestConfig {
    skipAdminAuth?: boolean;
  }
  export interface AxiosRequestConfig {
    skipAdminAuth?: boolean;
  }
}

export class ApiError extends Error {
  code: string;

  constructor(code: string, message: string) {
    super(message);
    this.name = "ApiError";
    this.code = code;
  }
}

const http = axios.create({
  timeout: 15000,
  withCredentials: true,
  headers: {
    "Content-Type": "application/json",
  },
});

http.interceptors.request.use((config) => {
  const session = getAdminSession();
  config.baseURL = config.baseURL || session?.baseUrl || authApiBaseUrl;
  if (session && !config.skipAdminAuth) {
    config.headers["X-Admin-Token"] = session.adminToken;
  }
  return config;
});

export async function request<T>(config: Parameters<typeof http.request>[0]) {
  try {
    const response = await http.request<ApiEnvelope<T>>(config);
    if (response.data.code !== "OK") {
      throw new ApiError(response.data.code, response.data.message);
    }
    return response.data.data;
  } catch (error: any) {
    if (error instanceof ApiError) throw error;
    const body = error?.response?.data;
    if (body?.code && body?.message) {
      throw new ApiError(body.code, body.message);
    }
    throw new ApiError("NETWORK_ERROR", error?.message || "网络请求失败");
  }
}

"use strict";

function trimSlash(value) {
  return String(value || "").replace(/\/+$/, "");
}

// createIdentityClient builds a trusted server-side SDK client.
// createIdentityClient 创建可信服务端 SDK 客户端，业务后端用它调用认证中台。
function createIdentityClient(config) {
  if (!config || !config.authBaseURL) throw new Error("authBaseURL is required");
  if (!config.clientID) throw new Error("clientID is required");
  const baseURL = trimSlash(config.authBaseURL);
  const clientID = config.clientID;
  const fetchImpl = config.fetch || globalThis.fetch;
  if (!fetchImpl) throw new Error("fetch is not available in this Node.js runtime");

  // request posts one identity command and returns the normalized data payload.
  // request 发送一个身份类命令，并只返回标准 data 数据。
  async function request(path, body) {
    const response = await fetchImpl(`${baseURL}${path}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ client_id: clientID, ...body }),
    });
    const payload = await response.json().catch(() => null);
    if (!response.ok || payload?.code !== "OK") {
      const error = new Error(payload?.message || `Open Wallet Auth request failed: ${response.status}`);
      error.status = response.status;
      error.code = payload?.code || "";
      throw error;
    }
    return payload.data;
  }

  return {
    // login signs in with email/password through the identity center.
    // login 使用邮箱密码登录认证中台。
    login(input) {
      return request("/api/v1/auth/login", input);
    },
    // register creates an identity-center user.
    // register 在认证中台创建统一身份用户。
    register(input) {
      return request("/api/v1/auth/register", input);
    },
    // profile validates an access token and returns the normalized identity user.
    // profile 校验 access token，并返回标准身份用户。
    async profile(accessToken) {
      const response = await fetchImpl(`${baseURL}/api/v1/profile`, {
        method: "GET",
        headers: {
          Authorization: `Bearer ${String(accessToken || "").trim()}`,
          "X-Client-ID": clientID,
        },
      });
      const payload = await response.json().catch(() => null);
      if (!response.ok || payload?.code !== "OK" || !payload?.data?.id) {
        const error = new Error(payload?.message || "Open Wallet Auth profile is invalid");
        error.status = response.status;
        error.code = payload?.code || "";
        throw error;
      }
      return payload.data;
    },
  };
}

module.exports = { createIdentityClient };

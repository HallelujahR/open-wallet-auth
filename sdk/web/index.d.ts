export type AuthClientConfig = {
  authBaseURL: string;
  clientID: string;
  returnURI?: string;
  tokenKey?: string;
  redirectKey?: string;
};

export type LoginURLInput = {
  returnURI?: string;
  redirect?: string;
};

export type CallbackToken = {
  accessToken: string;
  tokenType: string;
  expiresAt: string;
};

export type AuthClient = {
  config: AuthClientConfig & { tokenKey: string; redirectKey: string };
  buildLoginURL(options?: LoginURLInput): string;
  login(options?: LoginURLInput): void;
  parseCallback(hash?: string): CallbackToken | null;
  consumeRedirect(fallback?: string): string;
  setAccessToken(token: string): void;
  getAccessToken(): string;
  clearAccessToken(): void;
  authHeader(token?: string): Record<string, string>;
  loginWithPassword(input: { email: string; password: string }): Promise<any>;
  register(input: { username?: string; email: string; password: string }): Promise<any>;
  requestWalletNonce(input: { address: string; chain_id: number; domain: string }): Promise<any>;
  verifyWalletSignature(input: { address: string; nonce: string; signature: string }): Promise<any>;
  startOAuthLogin(provider: "github" | "google" | string, input?: { redirectURI?: string; returnURI?: string }): Promise<any>;
};

export function createAuthClient(config: AuthClientConfig): AuthClient;
export function buildLoginURL(config: AuthClientConfig, options?: LoginURLInput): string;
export function parseCallbackHash(hash: string): CallbackToken | null;

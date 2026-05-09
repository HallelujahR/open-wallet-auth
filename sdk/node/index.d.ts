export type IdentityClientConfig = {
  authBaseURL: string;
  clientID: string;
  fetch?: typeof fetch;
};

export type IdentityUser = {
  id: string;
  username?: string;
  email?: string;
  avatar?: string;
  status?: string;
};

export type IdentityTokenResult = {
  user?: IdentityUser;
  token?: {
    access_token: string;
    refresh_token?: string;
    expires_at?: string;
    token_type?: string;
  };
};

export type IdentityClient = {
  login(input: { email: string; password: string }): Promise<IdentityTokenResult>;
  register(input: { username?: string; email: string; password: string }): Promise<IdentityTokenResult>;
  profile(accessToken: string): Promise<IdentityUser>;
};

export function createIdentityClient(config: IdentityClientConfig): IdentityClient;

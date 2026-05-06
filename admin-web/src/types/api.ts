export type ApiEnvelope<T> = {
  code: string;
  message: string;
  request_id: string;
  data: T;
};

export type PageResult<T> = {
  items: T[];
  total: number;
  page: number;
  page_size: number;
};

export type AdminLoginResult = {
  token_type: string;
  admin_token: string;
};

export type HealthStatus = {
  service: string;
  env: string;
  status: string;
  started_at: string;
};

export type IdentityStatus = "active" | "suspended" | "deleted";

export type IdentityUser = {
  id: string;
  username: string;
  email?: string;
  phone?: string;
  avatar?: string;
  status: IdentityStatus | string;
  last_login_at?: string;
  created_at: string;
  updated_at: string;
};

export type UserClient = {
  client_id: string;
  first_login_at: string;
  last_login_at: string;
  login_count: number;
  status: string;
};

export type WalletBinding = {
  id: string;
  chain_type: string;
  address: string;
  is_primary: boolean;
  verified_at: string;
  created_at: string;
};

export type OAuthAccount = {
  id: string;
  provider: string;
  provider_subject: string;
  provider_email?: string;
  provider_username?: string;
  provider_avatar_url?: string;
  created_at: string;
};

export type Session = {
  id: string;
  user_id: string;
  client_id: string;
  ip?: string;
  user_agent?: string;
  active: boolean;
  expires_at: string;
  revoked_at?: string;
  last_used_at?: string;
  created_at: string;
};

export type IdentityDetail = {
  user: IdentityUser;
  clients: UserClient[];
  wallets: WalletBinding[];
  accounts: OAuthAccount[];
  sessions: Session[];
};

export type LoginLog = {
  id: string;
  user_id: string;
  client_id: string;
  login_method: string;
  ip?: string;
  user_agent?: string;
  success: boolean;
  failure_code?: string;
  created_at: string;
};

export type SecurityEvent = {
  id: string;
  user_id: string;
  event_type: string;
  target_type?: string;
  target_id?: string;
  ip?: string;
  user_agent?: string;
  success: boolean;
  failure_code?: string;
  created_at: string;
};

export type Client = {
  id: string;
  client_id: string;
  name: string;
  jwt_audience: string;
  allowed_origins: string[];
  allowed_redirect_uris: string[];
  status: string;
  created_at: string;
};

export type ClientCreateInput = {
  client_id: string;
  name: string;
  jwt_audience: string;
  allowed_origins: string[];
  allowed_redirect_uris: string[];
};

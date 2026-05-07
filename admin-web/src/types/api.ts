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

export type SecretStatus = {
  configured: boolean;
  masked: string;
};

export type WebhookSettings = {
  url: string;
  bearer_token?: string;
};

export type SMTPSettings = {
  host: string;
  port: number;
  username: string;
  password?: string;
  from: string;
};

export type AliyunSMSSettings = {
  access_key_id: string;
  access_key_secret?: string;
  sign_name: string;
  template_code: string;
  region_id: string;
  endpoint: string;
};

export type MessageProviderSettings = {
  type: string;
  webhook: WebhookSettings;
  smtp: SMTPSettings;
  aliyun_sms: AliyunSMSSettings;
  headers: Record<string, string>;
};

export type OAuthTenantSettings = {
  host: string;
  client_id: string;
  client_secret?: string;
};

export type OAuthProviderSettings = {
  client_id: string;
  client_secret?: string;
  auth_url: string;
  token_url: string;
  user_info_url: string;
  scopes: string[];
  tenant_credentials: OAuthTenantSettings[];
};

export type RuntimeSettings = {
  http: {
    cors_allowed_origins: string[];
  };
  phone: {
    enabled: boolean;
    provider: MessageProviderSettings;
  };
  email: {
    verification_enabled: boolean;
    provider: MessageProviderSettings;
  };
  oauth: {
    google: OAuthProviderSettings;
    github: OAuthProviderSettings;
  };
};

export type RuntimeSettingsResult = {
  settings: RuntimeSettings;
  secrets: Record<string, SecretStatus>;
  readonly: ReadonlySettings;
};

export type ReadonlySettings = {
  app: {
    name: string;
    env: string;
  };
  http: {
    host: string;
    port: number;
  };
  database: {
    driver: string;
    dsn: string;
    auto_migrate: boolean;
  };
  redis: {
    enabled: boolean;
    addr: string;
    password: string;
    db: number;
  };
  jwt: {
    issuer: string;
    access_token_ttl: string;
    refresh_token_ttl: string;
    private_key_path: string;
    public_key_path: string;
    active_key_id: string;
  };
};

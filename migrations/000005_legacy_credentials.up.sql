CREATE TABLE IF NOT EXISTS legacy_credentials (
  id VARCHAR(64) PRIMARY KEY,
  user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  source VARCHAR(128) NOT NULL,
  hash_type VARCHAR(64) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  salt VARCHAR(255) NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  migrated_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, source)
);

COMMENT ON TABLE legacy_credentials IS '旧系统密码凭证迁移记录，用于旧账号首次登录后无感升级为当前密码哈希';
COMMENT ON COLUMN legacy_credentials.id IS '旧密码凭证记录 ID';
COMMENT ON COLUMN legacy_credentials.user_id IS '统一认证用户 ID，对应 users.id';
COMMENT ON COLUMN legacy_credentials.source IS '旧系统来源标识，例如 case_project、legacy_crm';
COMMENT ON COLUMN legacy_credentials.hash_type IS '旧密码哈希算法，例如 hmac_sha1、sha1、md5';
COMMENT ON COLUMN legacy_credentials.password_hash IS '旧系统保存的密码哈希值';
COMMENT ON COLUMN legacy_credentials.salt IS '旧系统密码盐或 HMAC key；无盐时为空字符串';
COMMENT ON COLUMN legacy_credentials.status IS '迁移状态：active 表示可用于首次登录，migrated 表示已升级，disabled 表示停用';
COMMENT ON COLUMN legacy_credentials.migrated_at IS '旧密码成功校验并升级为当前哈希的时间';
COMMENT ON COLUMN legacy_credentials.created_at IS '记录创建时间';
COMMENT ON COLUMN legacy_credentials.updated_at IS '记录更新时间';

CREATE INDEX IF NOT EXISTS idx_legacy_credentials_user ON legacy_credentials(user_id);
CREATE INDEX IF NOT EXISTS idx_legacy_credentials_status ON legacy_credentials(status);
CREATE INDEX IF NOT EXISTS idx_legacy_credentials_source ON legacy_credentials(source);

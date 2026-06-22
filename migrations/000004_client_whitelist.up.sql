ALTER TABLE clients
  ADD COLUMN IF NOT EXISTS whitelist_enabled BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN clients.whitelist_enabled IS '是否启用应用登录白名单；开启后只有 client_members 中启用的用户可以登录该应用';

CREATE TABLE IF NOT EXISTS client_members (
  id VARCHAR(64) PRIMARY KEY,
  client_id VARCHAR(128) NOT NULL REFERENCES clients(client_id) ON DELETE CASCADE,
  user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role VARCHAR(64) NOT NULL DEFAULT 'member',
  permissions JSONB NOT NULL DEFAULT '[]'::jsonb,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  remark TEXT NOT NULL DEFAULT '',
  created_by VARCHAR(64),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (client_id, user_id)
);

COMMENT ON TABLE client_members IS '应用成员白名单，控制统一身份用户是否允许访问指定业务系统';
COMMENT ON COLUMN client_members.id IS '应用成员授权记录 ID';
COMMENT ON COLUMN client_members.client_id IS '业务系统 client_id，对应 clients.client_id';
COMMENT ON COLUMN client_members.user_id IS '统一身份用户 ID，对应 users.id';
COMMENT ON COLUMN client_members.role IS '该用户在对应业务系统中的基础角色，写入 JWT roles';
COMMENT ON COLUMN client_members.permissions IS '该用户在对应业务系统中的权限标识列表，写入 JWT permissions';
COMMENT ON COLUMN client_members.status IS '授权状态：active 表示允许登录，disabled 表示暂停访问';
COMMENT ON COLUMN client_members.remark IS '管理备注，用于记录授权原因或适用范围';
COMMENT ON COLUMN client_members.created_by IS '创建该授权记录的管理员或系统标识';
COMMENT ON COLUMN client_members.created_at IS '授权记录创建时间';
COMMENT ON COLUMN client_members.updated_at IS '授权记录更新时间';

CREATE INDEX IF NOT EXISTS idx_client_members_client ON client_members(client_id);
CREATE INDEX IF NOT EXISTS idx_client_members_user ON client_members(user_id);
CREATE INDEX IF NOT EXISTS idx_client_members_status ON client_members(status);

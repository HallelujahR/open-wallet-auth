CREATE TABLE IF NOT EXISTS security_events (
  id VARCHAR(64) PRIMARY KEY,
  user_id VARCHAR(64),
  event_type VARCHAR(64) NOT NULL,
  target_type VARCHAR(64),
  target_id VARCHAR(255),
  ip VARCHAR(64),
  user_agent TEXT,
  success BOOLEAN NOT NULL,
  failure_code VARCHAR(128),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_security_events_user_type ON security_events(user_id, event_type);
CREATE INDEX IF NOT EXISTS idx_security_events_created_at ON security_events(created_at);

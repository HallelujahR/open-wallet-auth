DROP TABLE IF EXISTS client_members;

ALTER TABLE clients
  DROP COLUMN IF EXISTS whitelist_enabled;

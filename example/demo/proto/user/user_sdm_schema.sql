CREATE TABLE IF NOT EXISTS pii_users (
  id TEXT,
  ssn BIGINT,
  address TEXT,
  PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS chain_users (
  key TEXT NOT NULL,
  field_name TEXT NOT NULL,
  version BIGSERIAL,
  tx_hash TEXT,
  field_value TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (key, field_name, version)
);

CREATE OR REPLACE VIEW users AS
  SELECT
    p.id,
    p.ssn,
    p.address,
    c_name.field_value AS name
  FROM pii_users p
  LEFT JOIN (SELECT DISTINCT ON (key, field_name) field_value, key FROM chain_users WHERE field_name='name' ORDER BY key, field_name, version DESC) c_name ON p.id = c_name.key
;


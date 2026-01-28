CREATE TABLE IF NOT EXISTS pii_users (
  id TEXT,
  name TEXT,
  email TEXT,
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
    p.name,
    p.email,
    c_hashed_email.field_value AS hashed_email,
    c_address.field_value AS address,
    c_chain_field.field_value AS chain_field
  FROM pii_users p
  LEFT JOIN (SELECT DISTINCT ON (key, field_name) field_value, key FROM chain_users WHERE field_name='hashed_email' ORDER BY key, field_name, version DESC) c_hashed_email ON p.id = c_hashed_email.key
  LEFT JOIN (SELECT DISTINCT ON (key, field_name) field_value, key FROM chain_users WHERE field_name='address' ORDER BY key, field_name, version DESC) c_address ON p.id = c_address.key
  LEFT JOIN (SELECT DISTINCT ON (key, field_name) field_value, key FROM chain_users WHERE field_name='chain_field' ORDER BY key, field_name, version DESC) c_chain_field ON p.id = c_chain_field.key
;


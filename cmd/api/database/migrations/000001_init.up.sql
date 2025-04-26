-- Creating Tables

CREATE TABLE IF NOT EXISTS mediator_did (
  id INT,
  did TEXT,
  added TIMESTAMP,
  PRIMARY KEY (did)
);

CREATE TABLE IF NOT EXISTS mediatees (
  remote_did TEXT,
  group TEXT,
  routing_key TEXT,
  protocol TEXT,
  topic TEXT,
  eventType TEXT,
  recipient_dids SET<TEXT>,
  properties MAP<TEXT, TEXT>,
  added TIMESTAMP,
  PRIMARY KEY (remote_did)
);
CREATE INDEX IF NOT EXISTS ON mediatees (group);
CREATE INDEX IF NOT EXISTS ON mediatees (recipient_dids);

CREATE TABLE IF NOT EXISTS blocked_dids (
  remote_did TEXT,
  added TIMESTAMP,
  PRIMARY KEY (remote_did)
);

CREATE TABLE IF NOT EXISTS secret_types (
  id INT,
  description TEXT,
  PRIMARY KEY (id)
);
-- Inserting into Secret Types Table
INSERT INTO secret_types(id, description) VALUES (1,'SecretTypeJsonWebKey2020');
INSERT INTO secret_types(id, description) VALUES (2,'SecretTypeX25519KeyAgreementKey2019');
INSERT INTO secret_types(id, description) VALUES (3,'SecretTypeEd25519VerificationKey2018');
INSERT INTO secret_types(id, description) VALUES (4,'SecretTypeEcdsaSecp256k1VerificationKey2019');
INSERT INTO secret_types(id, description) VALUES (5,'SecretTypeX25519KeyAgreementKey2020');
INSERT INTO secret_types(id, description) VALUES (6,'SecretTypeEd25519VerificationKey2020');
INSERT INTO secret_types(id, description) VALUES (7,'SecretTypeOther');

CREATE TABLE IF NOT EXISTS secrets (
  id TEXT,
  type INT,
  key TEXT,
  added TIMESTAMP,
  PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS messages (
  id TIMEUUID,
  recipient_did TEXT,
  description TEXT,
  filename TEXT,
  media_type TEXT,
  format TEXT,
  lastmod_time bigint,
  byte_count bigint,
  attachment_data TEXT,
  added TIMESTAMP,
  PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS messages_recipient_did_idx ON messages (recipient_did);

#!/bin/bash

echo "Starting to execute SH script..."

CREATE_KEYSPACE="CREATE KEYSPACE IF NOT EXISTS ${CASSANDRA_KEYSPACE} WITH REPLICATION = {
  'class' : 'SimpleStrategy', 
  'replication_factor' : 1 }; "

USE_KEYSPACE="USE ${CASSANDRA_KEYSPACE}; "

CREATE_CONNECTIONS_TABLE="CREATE TABLE IF NOT EXISTS ${CASSANDRA_KEYSPACE}.mediatees (
  remote_did text, routing_key text, created TIMESTAMP, PRIMARY KEY (remote_did)); "

CREATE_RECIPIENT_TABLE="CREATE TABLE IF NOT EXISTS ${CASSANDRA_KEYSPACE}.recipient_dids (
  recipient_did text,
  remote_did text,
  PRIMARY KEY (recipient_did)
);"

INDEX_RECIPIENT_TABLE="CREATE INDEX IF NOT EXISTS recipient_dids_remote_did_idx ON ${CASSANDRA_KEYSPACE}.recipient_dids (remote_did);"

CREATE_DENY_LIST_TABLE="CREATE TABLE IF NOT EXISTS ${CASSANDRA_KEYSPACE}.deny_list (
  did text, created TIMESTAMP, PRIMARY KEY (did)); "

CREATE_SECRET_TYPES_TABLE="CREATE TABLE IF NOT EXISTS ${CASSANDRA_KEYSPACE}.secret_types 
  (id int, description text, PRIMARY KEY (id)); "

INSERT_INTO_SECRET_TYPES_TABLE="INSERT INTO ${CASSANDRA_KEYSPACE}.secret_types(id, description) VALUES (1,'SecretTypeJsonWebKey2020');
INSERT INTO secret_types(id, description) VALUES (2,'SecretTypeX25519KeyAgreementKey2019');
INSERT INTO secret_types(id, description) VALUES (3,'SecretTypeEd25519VerificationKey2018');
INSERT INTO secret_types(id, description) VALUES (4,'SecretTypeEcdsaSecp256k1VerificationKey2019');
INSERT INTO secret_types(id, description) VALUES (5,'SecretTypeX25519KeyAgreementKey2020');
INSERT INTO secret_types(id, description) VALUES (6,'SecretTypeEd25519VerificationKey2020');
INSERT INTO secret_types(id, description) VALUES (7,'SecretTypeOther'); "

CREATE_SECRETS_TABLE="CREATE TABLE IF NOT EXISTS ${CASSANDRA_KEYSPACE}.secrets 
  (id text, type int, key text,added TIMESTAMP , PRIMARY KEY (id)); "


CREATE_MESSAGES_TABLE="CREATE TABLE IF NOT EXISTS ${CASSANDRA_KEYSPACE}.messages 
  (message_id text, recipient_did text, description text, filename text, media_type text, format text, 
  lastmod_time bigint, byte_count bigint, attachment_data text,added TIMESTAMP PRIMARY KEY (message_id)); "

INDEX_MESSAGES_TABLE="CREATE INDEX IF NOT EXISTS messages_recipient_did_idx ON ${CASSANDRA_KEYSPACE}.messages (recipient_did);"

CQL="$CREATE_KEYSPACE
  $USE_KEYSPACE
  $CREATE_SECRET_TYPES_TABLE
  $INSERT_INTO_SECRET_TYPES_TABLE
  $CREATE_SECRETS_TABLE
  $CREATE_MESSAGES_TABLE
  $INDEX_MESSAGES_TABLE
  $CREATE_DENY_LIST_TABLE
  $CREATE_CONNECTIONS_TABLE
  $CREATE_RECIPIENT_TABLE
  $INDEX_RECIPIENT_TABLE"

while ! cqlsh cassandra_db -u "${CASSANDRA_USERNAME}" -p "${CASSANDRA_PASSWORD}" -e 'describe cluster' ; do
  echo "Waiting for main instance to be ready..."
  sleep 5
done

cqlsh cassandra_db -u "${CASSANDRA_USERNAME}" -p "${CASSANDRA_PASSWORD}" -e "${CQL}"

# for cql_file in ./cql/*.cql;
# do
#  cqlsh cassandra_db -u "${CASSANDRA_USERNAME}" -p "${CASSANDRA_PASSWORD}" -f "${cql_file}" ;
#  echo "Script ""${cql_file}"" executed"
# done

echo "Execution of SH script finished"
echo "Stopping temporary database instance"
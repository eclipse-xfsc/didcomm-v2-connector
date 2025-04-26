package secretsresolver

import (
	"time"

	"github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"
	"github.com/eclipse-xfsc/didcomm-v2-connector/internal/config"

	"github.com/gocql/gocql"
)

type Cassandra struct {
	session  *gocql.Session
	keyspace string
}

func NewCassandra() *Cassandra {
	session, err := newCassandraSession()
	if err != nil {
		config.Logger.Error("NewCassandra", "Error creating session:", err)
		panic("Error creating cassandra session")
	}
	return &Cassandra{
		session:  session,
		keyspace: config.CurrentConfiguration.Database.Keyspace,
	}
}

func (s *Cassandra) GetPlainSecret(secretId string) *didcomm.Secret {
	iter := s.session.Query("SELECT id, type, key FROM "+config.CurrentConfiguration.Database.Keyspace+".secrets WHERE id = ?", secretId).Iter()
	defer iter.Close()
	var secret didcomm.Secret
	var key didcomm.SecretMaterialMultibase
	for iter.Scan(&secret.Id, &secret.Type, &key.PrivateKeyMultibase) {
		if secret.Id == secretId {
			secret.SecretMaterial = key
			return &secret
		}
	}
	return nil
}

// This method should not be used without and cb object which is coming from rust. Otherwise there will popup exceptions during cb.Sucess/Error in cause of Nil Pointer. DONT CALL it with Default didcommGetSecretResult
func (s *Cassandra) GetSecret(secretId string, cb *didcomm.OnGetSecretResult) didcomm.ErrorCode {
	secret := s.GetPlainSecret(secretId)
	if secret == nil {
		errorKind := didcomm.NewErrorKindSecretNotFound()
		err := cb.Error(errorKind, "Secret not found")
		if err != nil {
			return didcomm.ErrorCodeError
		}
	}

	err := cb.Success(secret)
	if err != nil {
		config.Logger.Error("GetSecret", "Error calling callback:", err)
		return didcomm.ErrorCodeError
	}

	return didcomm.ErrorCodeSuccess
}

func (s *Cassandra) FindSecrets(secretIds []string, cb *didcomm.OnFindSecretsResult) didcomm.ErrorCode {
	var secrets []string
	for _, id := range secretIds {
		iter := s.session.Query("SELECT id FROM "+config.CurrentConfiguration.Database.Keyspace+".secrets WHERE id = ?", id).Iter()
		var i string
		for iter.Scan(&i) {
			if i == id {
				secrets = append(secrets, id)
			}
		}
		if err := iter.Close(); err != nil {
			config.Logger.Error("FindSecrets", "Error closing iter:", err)
			return didcomm.ErrorCodeError
		}
	}
	if len(secrets) == len(secretIds) {
		err := cb.Success(secrets)
		if err != nil {
			return didcomm.ErrorCodeError
		}
		return didcomm.ErrorCodeSuccess
	} else {
		errorKind := didcomm.NewErrorKindSecretNotFound()
		err := cb.Error(errorKind, "Secret not found")
		if err != nil {
			return didcomm.ErrorCodeError
		}
		return didcomm.ErrorCodeError
	}
}

func (s *Cassandra) StoreSecret(secret didcomm.Secret) error {
	if err := s.session.Query("INSERT INTO "+config.CurrentConfiguration.Database.Keyspace+".secrets (id, type, key, added) VALUES (?, ?, ?, ?)",
		secret.Id, secret.Type, secret.SecretMaterial.(didcomm.SecretMaterialMultibase).PrivateKeyMultibase, time.Now()).Exec(); err != nil {
		return err
	}
	return nil
}

func newCassandraSession() (*gocql.Session, error) {

	dbConfig := config.CurrentConfiguration.Database

	cluster := gocql.NewCluster(dbConfig.Host)
	cluster.Port = dbConfig.Port
	cluster.Keyspace = dbConfig.Keyspace

	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: dbConfig.User,
		Password: dbConfig.Password,
	}

	session, err := cluster.CreateSession()
	if err != nil {
		config.Logger.Error("Cassandra", "Error creating session:", err)
		return nil, err
	}
	return session, nil
}

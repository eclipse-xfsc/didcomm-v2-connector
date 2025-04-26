package database

import (
	"github.com/eclipse-xfsc/didcomm-v2-connector/internal/config"

	"github.com/gocql/gocql"
	migrate "github.com/golang-migrate/migrate/v4"
	migrateCassandra "github.com/golang-migrate/migrate/v4/database/cassandra"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Migration struct {
	session *gocql.Session
}

func NewMigration() {
	if config.CurrentConfiguration.Database.InMemory {
		config.Logger.Info("No migration necessary")
		return
	}
	// Cassandra
	createKeySpace()
	session, err := newCassandraSession()
	if err != nil {
		config.Logger.Error("NewCassandra", "Error creating session:", err)
		panic("Error creating cassandra session")
	}
	defer session.Close()
	mig := &Migration{session: session}
	mig.migrate()
}

func (mig *Migration) migrate() {
	config.Logger.Info("Cassandra migration")
	driver, err := migrateCassandra.WithInstance(mig.session, &migrateCassandra.Config{
		KeyspaceName:          config.CurrentConfiguration.Database.Keyspace,
		MultiStatementEnabled: true,
	})
	if err != nil {
		config.Logger.Error("NewCassandra", "Error creating driver:", err)
		panic("Error creating cassandra driver")
	}
	instance, err := migrate.NewWithDatabaseInstance("file://database/migrations", "cassandra", driver)
	if err != nil {
		config.Logger.Error("NewCassandra", "Error creating instance:", err)
		panic("Error creating cassandra instance")
	}
	err = instance.Up()
	if err != nil && err != migrate.ErrNoChange {
		config.Logger.Error("NewCassandra", "Error migrating:", err)
		panic("Error migrating cassandra")
	} else {
		config.Logger.Info("Cassandra migration finished")
	}
}

func newCassandraSession() (*gocql.Session, error) {
	logTag := "Database session"

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

		config.Logger.Error(logTag, "Error while creating session", err)
		return nil, err
	}
	config.Logger.Info(logTag, "Initiated", true)
	return session, nil
}

func createKeySpace() {
	dbConfig := config.CurrentConfiguration.Database

	cluster := gocql.NewCluster(dbConfig.Host)
	cluster.Port = dbConfig.Port
	cluster.Keyspace = "system"

	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: dbConfig.User,
		Password: dbConfig.Password,
	}

	session, err := cluster.CreateSession()
	if err != nil {
		config.Logger.Error("Error while creating session", err)
		panic("Error creating cassandra session")
	}
	defer session.Close()
	_ = session.Query("CREATE KEYSPACE IF NOT EXISTS " + dbConfig.Keyspace + " WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1}").Exec()
}

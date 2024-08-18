package db

import (
	"cme/internal/app/constants"
	"cme/internal/app/db/dto"
	"context"
	"fmt"
	"log"

	"github.com/gocql/gocql"
)

// Init : Initializes the Cassandra connection and sets up connection pooling
func Init(ctx context.Context) (session *gocql.Session, err error) {
	// Create a new Cassandra cluster
	cluster := gocql.NewCluster(constants.Config.DatabaseConfig.DB_HOST)

	// Set the port and credentials
	cluster.Port = constants.Config.DatabaseConfig.DB_PORT
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: constants.Config.DatabaseConfig.DB_USER,
		Password: constants.Config.DatabaseConfig.DB_PASSWORD,
	}

	// Set the keyspace if you have one (optional)
	cluster.Keyspace = constants.Config.DatabaseConfig.DB_KEYSPACE // replace with your keyspace name if available
	cluster.Consistency = gocql.Quorum

	// Set the number of connections for pooling
	cluster.NumConns = constants.Config.DatabaseConfig.DB_MAX_OPEN_CONNECTION // Adjust this value as needed for your environment

	// Connect to the cluster
	session, err = cluster.CreateSession()
	if err != nil {
		log.Fatal("Failed to connect to Cassandra:", err)
		return nil, err
	}

	fmt.Println("Connected to Cassandra!")
	dto.RunTableMigration(session)
	// Do not close the session here; return it to be used by the application
	return session, nil
}

type DBService struct {
	Session *gocql.Session
}

// New : Create a new DBService instance
func New(session *gocql.Session) *DBService {
	return &DBService{
		Session: session,
	}
}

// GetSession : Get an instance of Session to connect to the database connection pool
func (d DBService) GetSession() *gocql.Session {
	return d.Session
}

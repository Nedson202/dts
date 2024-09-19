package database

import (
	"github.com/gocql/gocql"
	"github.com/nedson202/dts-go/pkg/logger"
)

type CassandraClient struct {
	Session *gocql.Session
}

func NewCassandraClient(hosts []string, keyspace string) (*CassandraClient, error) {
	logger.Info().Msgf("Connecting to Cassandra with hosts: %v and keyspace: %s", hosts, keyspace)
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum
	session, err := cluster.CreateSession()
	if err != nil {
		logger.Error().Err(err).Msg("Error creating Cassandra session")
		return nil, err
	}
	return &CassandraClient{Session: session}, nil
}

func (c *CassandraClient) Close() {
	c.Session.Close()
	logger.Info().Msg("Cassandra session closed")
}

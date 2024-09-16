package database

import (
	"log"

	"github.com/gocql/gocql"
)

type CassandraClient struct {
	Session *gocql.Session
}

func NewCassandraClient(hosts []string, keyspace string) (*CassandraClient, error) {
	log.Printf("Connecting to Cassandra with hosts: %v and keyspace: %s", hosts, keyspace)
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum
	session, err := cluster.CreateSession()
	if err != nil {
		log.Printf("Error creating Cassandra session: %v", err)
		return nil, err
	}
	return &CassandraClient{Session: session}, nil
}

func (c *CassandraClient) Close() {
	c.Session.Close()
}

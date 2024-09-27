package election

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/nedson202/dts-go/pkg/database"
	"github.com/nedson202/dts-go/pkg/logger"
)

const (
	LeaderKey           = "current-leader"
	LeaderTimeout       = 15 * time.Second // Time after which the leader is considered dead
	LeaderRenewInterval = 5 * time.Second  // Interval at which the leader renews its leadership
)

type Election struct {
	cassandraClient *database.CassandraClient
	instanceID      string
	isLeader        bool
	leader          string
	resultChan      chan bool
}

func NewElection(cassandraClient *database.CassandraClient, instanceID string) *Election {
	return &Election{
		cassandraClient: cassandraClient,
		instanceID:      instanceID,
		isLeader:        false,
		resultChan:      make(chan bool),
	}
}

func (e *Election) RunElection(ctx context.Context) error {
	ticker := time.NewTicker(LeaderRenewInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			close(e.resultChan)
			return ctx.Err()
		case <-ticker.C:
			if e.isLeader {
				// Renew leadership periodically
				if err := e.renewLeadership(ctx); err != nil {
					e.isLeader = false
					e.leader = ""
					logger.Error().Err(err).Msg("Failed to renew leadership")
					e.resultChan <- false
				}
			} else {
				// Check if the current leader is still alive
				leaderAlive, err := e.isLeaderAlive(ctx)
				if err != nil {
					logger.Error().Err(err).Msg("Failed to check if leader is alive")
					continue
				}
				if !leaderAlive {
					logger.Info().Msg("Current leader is dead, removing it")
					if err := e.removeDeadLeader(ctx); err != nil {
						logger.Error().Err(err).Msg("Failed to remove dead leader")
						continue
					}
					logger.Info().Msgf("Trying to become the new leader")
					// Try to become the leader
					if err := e.tryBecomeLeader(ctx); err == nil {
						e.isLeader = true
						e.leader = e.instanceID
						e.resultChan <- true
					} else {
						logger.Error().Err(err).Msg("Failed to become leader")
						e.resultChan <- false
					}

					if e.leader == e.instanceID {
						logger.Info().Msgf("You are the leader %s", e.instanceID)
					} else {
						logger.Info().Msgf("Instance %s is the leader", e.leader)
					}
				}
			}
		}
	}
}

func (e *Election) removeDeadLeader(ctx context.Context) error {
	query := `DELETE FROM leader_election WHERE key = ? IF EXISTS`
	var applied bool
	var err error
	applied, err = e.cassandraClient.Session.Query(query, LeaderKey).WithContext(ctx).MapScanCAS(map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("failed to execute remove dead leader query: %w", err)
	}
	if !applied {
		logger.Info().Msg("Failed to remove dead leader, it may have already been removed")
	}
	return nil
}

func (e *Election) tryBecomeLeader(ctx context.Context) error {
	query := `INSERT INTO leader_election (key, instance_id, last_heartbeat) VALUES (?, ?, ?) IF NOT EXISTS`
	applied, err := e.cassandraClient.Session.Query(query, LeaderKey, e.instanceID, time.Now()).WithContext(ctx).MapScanCAS(map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("failed to execute leader election query: %w", err)
	}
	if !applied {
		return fmt.Errorf("failed to become leader, another instance is already the leader")
	}
	return nil
}

func (e *Election) renewLeadership(ctx context.Context) error {
	query := `UPDATE leader_election SET last_heartbeat = ? WHERE key = ? IF instance_id = ?`
	applied, err := e.cassandraClient.Session.Query(query, time.Now(), LeaderKey, e.instanceID).WithContext(ctx).MapScanCAS(map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("failed to execute renew leadership query: %w", err)
	}
	if !applied {
		return fmt.Errorf("failed to renew leadership, another instance may have become the leader")
	}
	return nil
}

func (e *Election) isLeaderAlive(ctx context.Context) (bool, error) {
	var lastHeartbeat time.Time
	var instanceID string
	query := `SELECT instance_id, last_heartbeat FROM leader_election WHERE key = ?`
	err := e.cassandraClient.Session.Query(query, LeaderKey).WithContext(ctx).Scan(&instanceID, &lastHeartbeat)
	if err != nil {
		if err == gocql.ErrNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to query leader election table: %w", err)
	}

	// Check if the last heartbeat is within the timeout
	if time.Since(lastHeartbeat) > LeaderTimeout {
		return false, nil
	}
	return true, nil
}

func (e *Election) IsLeader() bool {
	return e.isLeader
}

func (e *Election) ResultChan() <-chan bool {
	return e.resultChan
}

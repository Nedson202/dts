package scheduler

import (
	"context"
	"fmt"

	"github.com/nedson202/dts-go/pkg/database"
	"github.com/nedson202/dts-go/pkg/models"
)

type ResourceManager struct {
	cassandraClient *database.CassandraClient
}

func NewResourceManager(cassandraClient *database.CassandraClient) *ResourceManager {
	return &ResourceManager{
		cassandraClient: cassandraClient,
	}
}

func (rm *ResourceManager) AllocateResources(ctx context.Context, required models.Resources) (*models.Resources, error) {
	var available models.Resources
	err := rm.cassandraClient.Session.Query(`
		SELECT cpu, memory, storage FROM available_resources WHERE id = 'global'
	`).Scan(&available.CPU, &available.Memory, &available.Storage)
	if err != nil {
		return nil, fmt.Errorf("failed to get available resources: %v", err)
	}

	if available.CPU < required.CPU || available.Memory < required.Memory || available.Storage < required.Storage {
		return nil, fmt.Errorf("insufficient resources")
	}

	// Allocate resources
	err = rm.cassandraClient.Session.Query(`
		UPDATE available_resources
		SET cpu = cpu - ?, memory = memory - ?, storage = storage - ?
		WHERE id = 'global'
	`, required.CPU, required.Memory, required.Storage).Exec()
	if err != nil {
		return nil, fmt.Errorf("failed to allocate resources: %v", err)
	}

	return &required, nil
}

func (rm *ResourceManager) ReleaseResources(ctx context.Context, resources models.Resources) error {
	err := rm.cassandraClient.Session.Query(`
		UPDATE available_resources
		SET cpu = cpu + ?, memory = memory + ?, storage = storage + ?
		WHERE id = 'global'
	`, resources.CPU, resources.Memory, resources.Storage).Exec()
	if err != nil {
		return fmt.Errorf("failed to release resources: %v", err)
	}
	return nil
}

package repository

import (
	"context"
	"fmt"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *RepoPostgres) UpsertPartitions(partitions []*dao.PartitionInfo) error {
	upsertPartition := `INSERT INTO partitions (
		id, 
		cluster_id,
		name, 
		capacity,
		used_capacity, 
		utilization,
		total_nodes, 
		applications,
		total_containers,
		state, 
		last_state_transition_time) VALUES (@id, @cluster_id, @name, @capacity, @used_capacity, @utilization,
			@total_nodes, @applications, @total_containers, @state, @last_state_transition_time)
	ON CONFLICT (name) DO UPDATE SET
		capacity = EXCLUDED.capacity,
		used_capacity = EXCLUDED.used_capacity,
		utilization = EXCLUDED.utilization,
		total_nodes = EXCLUDED.total_nodes,
		applications = EXCLUDED.applications,
		total_containers = EXCLUDED.total_containers,
		state = EXCLUDED.state,
		last_state_transition_time = EXCLUDED.last_state_transition_time`
	for _, p := range partitions {
		_, err := s.dbpool.Exec(context.Background(), upsertPartition,
			pgx.NamedArgs{
				"id":                         uuid.NewString(),
				"cluster_id":                 p.ClusterID,
				"name":                       p.Name,
				"capacity":                   p.Capacity.Capacity,
				"used_capacity":              p.Capacity.UsedCapacity,
				"utilization":                p.Capacity.Utilization,
				"total_nodes":                p.TotalNodes,
				"applications":               p.Applications,
				"total_containers":           p.TotalContainers,
				"state":                      p.State,
				"last_state_transition_time": p.LastStateTransitionTime,
			})
		if err != nil {
			return fmt.Errorf("could not insert/update partition into DB: %v", err)
		}
	}
	return nil
}

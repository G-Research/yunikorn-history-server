package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid/v2"

	"github.com/G-Research/yunikorn-history-server/internal/model"
)

func (s *PostgresRepository) UpsertPartitions(ctx context.Context, partitions []*dao.PartitionInfo) error {
	upsertSQL := `INSERT INTO partitions (
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
		_, err := s.dbpool.Exec(ctx, upsertSQL,
			pgx.NamedArgs{
				"id":                         ulid.Make().String(),
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

func (s *PostgresRepository) GetAllPartitions(ctx context.Context) ([]*model.PartitionInfo, error) {
	rows, err := s.dbpool.Query(ctx, `SELECT * FROM partitions`)
	if err != nil {
		return nil, fmt.Errorf("could not get partitions from DB: %v", err)
	}
	defer rows.Close()

	var partitions []*model.PartitionInfo
	for rows.Next() {
		var p model.PartitionInfo
		if err := rows.Scan(
			&p.Id,
			&p.ClusterID,
			&p.Name,
			&p.Capacity.Capacity,
			&p.Capacity.UsedCapacity,
			&p.Capacity.Utilization,
			&p.TotalNodes,
			&p.Applications,
			&p.TotalContainers,
			&p.State,
			&p.LastStateTransitionTime,
			&p.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("could not scan partition from DB: %v", err)
		}
		partitions = append(partitions, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read rows: %v", err)
	}
	return partitions, nil
}

func (s *PostgresRepository) GetActivePartitions(ctx context.Context) ([]*model.PartitionInfo, error) {
	rows, err := s.dbpool.Query(ctx, `SELECT * FROM partitions WHERE deleted_at IS NULL`)
	if err != nil {
		return nil, fmt.Errorf("could not get partitions from DB: %v", err)
	}
	defer rows.Close()

	var partitions []*model.PartitionInfo
	for rows.Next() {
		var p model.PartitionInfo
		if err := rows.Scan(
			&p.Id,
			&p.ClusterID,
			&p.Name,
			&p.Capacity.Capacity,
			&p.Capacity.UsedCapacity,
			&p.Capacity.Utilization,
			&p.TotalNodes,
			&p.Applications,
			&p.TotalContainers,
			&p.State,
			&p.LastStateTransitionTime,
			&p.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("could not scan partition from DB: %v", err)
		}
		partitions = append(partitions, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read rows: %v", err)
	}
	return partitions, nil
}

func (s *PostgresRepository) DeletePartitions(ctx context.Context, partitions []*model.PartitionInfo) error {
	partitionIds := make([]string, len(partitions))
	for i, p := range partitions {
		partitionIds[i] = p.Id
	}

	deletedAt := time.Now().Unix()
	query := `UPDATE partitions SET deleted_at = $1 WHERE id = ANY($2)`
	_, err := s.dbpool.Exec(ctx, query, deletedAt, partitionIds)
	if err != nil {
		return fmt.Errorf("could not mark partitions as deleted in DB: %w", err)
	}

	return nil
}

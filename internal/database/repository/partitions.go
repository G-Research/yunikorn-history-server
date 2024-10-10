package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid/v2"

	"github.com/G-Research/yunikorn-history-server/internal/database/sql"

	"github.com/G-Research/yunikorn-history-server/internal/model"
)

type PartitionFilters struct {
	LastStateTransitionTimeStart *time.Time
	LastStateTransitionTimeEnd   *time.Time
	Name                         *string
	ClusterID                    *string
	State                        *string
	Offset                       *int
	Limit                        *int
}

func applyPartitionFilters(builder *sql.Builder, filters PartitionFilters) {
	if filters.LastStateTransitionTimeStart != nil {
		builder.Conditionp("last_state_transition_time", ">=", filters.LastStateTransitionTimeStart.UnixMilli())
	}
	if filters.LastStateTransitionTimeEnd != nil {
		builder.Conditionp("last_state_transition_time", "<=", filters.LastStateTransitionTimeEnd.UnixMilli())
	}
	if filters.Name != nil {
		builder.Conditionp("name", "=", *filters.Name)
	}
	if filters.ClusterID != nil {
		builder.Conditionp("cluster_id", "=", *filters.ClusterID)
	}
	if filters.State != nil {
		builder.Conditionp("state", "=", *filters.State)
	}
	applyLimitAndOffset(builder, filters.Limit, filters.Offset)
}

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

func (s *PostgresRepository) GetAllPartitions(ctx context.Context, filters PartitionFilters) ([]*model.PartitionInfo, error) {
	queryBuilder := sql.NewBuilder().
		SelectAll("partitions", "").
		OrderBy("id", sql.OrderByDescending)
	applyPartitionFilters(queryBuilder, filters)

	query := queryBuilder.Query()
	args := queryBuilder.Args()
	rows, err := s.dbpool.Query(ctx, query, args...)
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

// DeleteInactivePartitions deletes partitions that are not in the list of activePartitions.
func (s *PostgresRepository) DeleteInactivePartitions(ctx context.Context, activePartitions []*dao.PartitionInfo) error {
	partitionNames := make([]string, len(activePartitions))
	for i, p := range activePartitions {
		partitionNames[i] = p.Name
	}
	deletedAt := time.Now().Unix()
	query := `
		UPDATE partitions 
		SET deleted_at = $1 
		WHERE id IN (
			SELECT id 
			FROM partitions 
			WHERE deleted_at IS NULL AND NOT(name = ANY($2))
		)
	`
	_, err := s.dbpool.Exec(ctx, query, deletedAt, partitionNames)
	if err != nil {
		return fmt.Errorf("could not mark partitions as deleted in DB: %w", err)
	}
	return nil
}

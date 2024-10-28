package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

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

func (r *PostgresRepository) InsertPartition(ctx context.Context, partition *model.Partition) error {
	const q = `
INSERT INTO partitions (
	id,
	created_at_nano,
	deleted_at_nano,
	cluster_id,
	name,
	capacity,
	used_capacity,
	utilization,
	total_nodes,
	applications,
	total_containers,
	state,
	last_state_transition_time
) VALUES (
	@id,
	@created_at_nano,
	@deleted_at_nano,
	@cluster_id,
	@name,
	@capacity,
	@used_capacity,
	@utilization,
	@total_nodes,
	@applications,
	@total_containers,
	@state,
	@last_state_transition_time
)`
	_, err := r.dbpool.Exec(ctx, q,
		pgx.NamedArgs{
			"id":                         partition.ID,
			"created_at_nano":            partition.CreatedAtNano,
			"deleted_at_nano":            partition.DeletedAtNano,
			"cluster_id":                 partition.ClusterID,
			"name":                       partition.Name,
			"capacity":                   partition.Capacity.Capacity,
			"used_capacity":              partition.Capacity.UsedCapacity,
			"utilization":                partition.Capacity.Utilization,
			"total_nodes":                partition.TotalNodes,
			"applications":               partition.Applications,
			"total_containers":           partition.TotalContainers,
			"state":                      partition.State,
			"last_state_transition_time": partition.LastStateTransitionTime,
		},
	)
	if err != nil {
		return fmt.Errorf("could not insert partition into DB: %v", err)
	}

	return nil
}

func (s *PostgresRepository) GetAllPartitions(ctx context.Context, filters PartitionFilters) ([]*model.Partition, error) {
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

	var partitions []*model.Partition
	for rows.Next() {
		var p model.Partition
		if err := rows.Scan(
			&p.ID,
			&p.CreatedAtNano,
			&p.DeletedAtNano,
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

func (r *PostgresRepository) UpdatePartition(ctx context.Context, partition *model.Partition) error {
	const q = `
UPDATE partitions
SET
	deleted_at_nano = @deleted_at_nano,
	cluster_id = @cluster_id,
	name = @name,
	capacity = @capacity,
	used_capacity = @used_capacity,
	utilization = @utilization,
	total_nodes = @total_nodes,
	applications = @applications,
	total_containers = @total_containers,
	state = @state,
	last_state_transition_time = @last_state_transition_time
WHERE id = @id`

	res, err := r.dbpool.Exec(ctx, q,
		pgx.NamedArgs{
			"id":                         partition.ID,
			"deleted_at_nano":            partition.DeletedAtNano,
			"cluster_id":                 partition.ClusterID,
			"name":                       partition.Name,
			"capacity":                   partition.Capacity.Capacity,
			"used_capacity":              partition.Capacity.UsedCapacity,
			"utilization":                partition.Capacity.Utilization,
			"total_nodes":                partition.TotalNodes,
			"applications":               partition.Applications,
			"total_containers":           partition.TotalContainers,
			"state":                      partition.State,
			"last_state_transition_time": partition.LastStateTransitionTime,
		},
	)
	if err != nil {
		return fmt.Errorf("could not update partition in DB: %v", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("failed to update partition %q: no rows affected", partition.ID)
	}

	return nil
}

func (s *PostgresRepository) DeletePartitionsNotInIDs(ctx context.Context, ids []string, deletedAtNano int64) error {
	const q = `
UPDATE partitions
SET deleted_at_nano = @deleted_at_nano
WHERE deleted_at_nano IS NULL AND NOT (id = ANY(@ids))`

	_, err := s.dbpool.Exec(
		ctx,
		q,
		pgx.NamedArgs{
			"ids":             ids,
			"deleted_at_nano": deletedAtNano,
		},
	)
	if err != nil {
		return fmt.Errorf("could not delete partitions from DB: %v", err)
	}

	return nil
}

func (s *PostgresRepository) GetPartitionByID(ctx context.Context, id string) (*model.Partition, error) {
	const q = `
SELECT
	id,
	created_at_nano,
	deleted_at_nano,
	cluster_id,
	name,
	capacity,
	used_capacity,
	utilization,
	total_nodes,
	applications,
	total_containers,
	state,
	last_state_transition_time
FROM partitions
WHERE id = @id`

	row := s.dbpool.QueryRow(ctx, q, id)
	var p model.Partition
	if err := row.Scan(
		&p.ID,
		&p.CreatedAtNano,
		&p.DeletedAtNano,
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
	); err != nil {
		return nil, fmt.Errorf("could not get partition from DB: %v", err)
	}

	return &p, nil
}

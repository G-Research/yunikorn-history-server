package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/G-Research/unicorn-history-server/internal/model"
)

func (s *PostgresRepository) InsertQueue(ctx context.Context, q *model.Queue) error {
	insertSQL := `INSERT INTO queues (
		id, created_at_nano, queue_name, parent_id, parent, status, partition_id, pending_resource, max_resource,
		guaranteed_resource, allocated_resource, preempting_resource, head_room, is_leaf, is_managed,
		properties, template_info, abs_used_capacity, max_running_apps, running_apps,
		current_priority, allocating_accepted_apps)
		VALUES (@id, @created_at_nano, @queue_name, @parent_id, @parent, @status, @partition_id, @pending_resource, @max_resource,
		@guaranteed_resource, @allocated_resource, @preempting_resource, @head_room, @is_leaf, @is_managed,
		@properties, @template_info, @abs_used_capacity, @max_running_apps, @running_apps,
		@current_priority, @allocating_accepted_apps)`

	_, err := s.dbpool.Exec(ctx, insertSQL,
		pgx.NamedArgs{
			"id":                       q.ID,
			"created_at_nano":          q.CreatedAtNano,
			"queue_name":               q.QueueName,
			"parent_id":                q.ParentID,
			"parent":                   q.Parent,
			"status":                   q.Status,
			"partition_id":             q.PartitionID,
			"pending_resource":         q.PendingResource,
			"max_resource":             q.MaxResource,
			"guaranteed_resource":      q.GuaranteedResource,
			"allocated_resource":       q.AllocatedResource,
			"preempting_resource":      q.PreemptingResource,
			"head_room":                q.HeadRoom,
			"is_leaf":                  q.IsLeaf,
			"is_managed":               q.IsManaged,
			"properties":               q.Properties,
			"template_info":            q.TemplateInfo,
			"abs_used_capacity":        q.AbsUsedCapacity,
			"max_running_apps":         q.MaxRunningApps,
			"running_apps":             q.RunningApps,
			"current_priority":         q.CurrentPriority,
			"allocating_accepted_apps": q.AllocatingAcceptedApps,
		})

	return err
}

func (s *PostgresRepository) UpdateQueue(ctx context.Context, queue *model.Queue) error {
	updateSQL := `
    UPDATE queues SET
		deleted_at_nano = @deleted_at_nano,
        status = @status,
		parent_id = @parent_id,
        parent = @parent,
        partition_id = @partition_id,
        pending_resource = @pending_resource,
        max_resource = @max_resource,
        guaranteed_resource = @guaranteed_resource,
        allocated_resource = @allocated_resource,
        preempting_resource = @preempting_resource,
        head_room = @head_room,
        is_leaf = @is_leaf,
        is_managed = @is_managed,
        properties = @properties,
        template_info = @template_info,
        abs_used_capacity = @abs_used_capacity,
        max_running_apps = @max_running_apps,
        running_apps = @running_apps,
        current_priority = @current_priority,
        allocating_accepted_apps = @allocating_accepted_apps
    WHERE id = @id`
	result, err := s.dbpool.Exec(ctx, updateSQL,
		pgx.NamedArgs{
			"id":                       queue.ID,
			"deleted_at_nano":          queue.DeletedAtNano,
			"queue_name":               queue.QueueName,
			"parent_id":                queue.ParentID,
			"parent":                   queue.Parent,
			"status":                   queue.Status,
			"partition_id":             queue.PartitionID,
			"pending_resource":         queue.PendingResource,
			"max_resource":             queue.MaxResource,
			"guaranteed_resource":      queue.GuaranteedResource,
			"allocated_resource":       queue.AllocatedResource,
			"preempting_resource":      queue.PreemptingResource,
			"head_room":                queue.HeadRoom,
			"is_leaf":                  queue.IsLeaf,
			"is_managed":               queue.IsManaged,
			"properties":               queue.Properties,
			"template_info":            queue.TemplateInfo,
			"abs_used_capacity":        queue.AbsUsedCapacity,
			"max_running_apps":         queue.MaxRunningApps,
			"running_apps":             queue.RunningApps,
			"current_priority":         queue.CurrentPriority,
			"allocating_accepted_apps": queue.AllocatingAcceptedApps,
		},
	)
	if err != nil {
		return fmt.Errorf("could not update queue in DB: %v", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("queue not found: %s", queue.QueueName)
	}

	return nil
}

func (s *PostgresRepository) GetAllQueues(ctx context.Context) ([]*model.Queue, error) {
	const q = `
SELECT
    id,
    created_at_nano,
    deleted_at_nano,
    queue_name,
	parent_id,
    parent,
    status,
    partition_id,
    pending_resource,
    max_resource,
    guaranteed_resource,
    allocated_resource,
    preempting_resource,
    head_room,
    is_leaf,
    is_managed,
    properties,
    template_info,
    abs_used_capacity,
    max_running_apps,
    running_apps,
    current_priority,
    allocating_accepted_apps
FROM queues
ORDER BY id DESC
		`
	rows, err := s.dbpool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("could not get queues from DB: %v", err)
	}
	defer rows.Close()

	var queues []*model.Queue
	for rows.Next() {
		var q model.Queue
		err = rows.Scan(
			&q.ID,
			&q.CreatedAtNano,
			&q.DeletedAtNano,
			&q.QueueName,
			&q.ParentID,
			&q.Parent,
			&q.Status,
			&q.PartitionID,
			&q.PendingResource,
			&q.MaxResource,
			&q.GuaranteedResource,
			&q.AllocatedResource,
			&q.PreemptingResource,
			&q.HeadRoom,
			&q.IsLeaf,
			&q.IsManaged,
			&q.Properties,
			&q.TemplateInfo,
			&q.AbsUsedCapacity,
			&q.MaxRunningApps,
			&q.RunningApps,
			&q.CurrentPriority,
			&q.AllocatingAcceptedApps,
		)
		if err != nil {
			return nil, fmt.Errorf("could not scan queue from DB: %v", err)
		}
		queues = append(queues, &q)
	}
	return queues, nil
}

func (s *PostgresRepository) GetQueue(ctx context.Context, queueID string) (*model.Queue, error) {
	const q = `
SELECT
    id,
    created_at_nano,
    deleted_at_nano,
    queue_name,
	parent_id,
    parent,
    status,
    partition_id,
    pending_resource,
    max_resource,
    guaranteed_resource,
    allocated_resource,
    preempting_resource,
    head_room,
    is_leaf,
    is_managed,
    properties,
    template_info,
    abs_used_capacity,
    max_running_apps,
    running_apps,
    current_priority,
    allocating_accepted_apps
FROM queues
WHERE id = @id
ORDER BY id DESC
LIMIT 1
`
	var queue model.Queue
	err := s.dbpool.QueryRow(
		ctx,
		q,
		&pgx.NamedArgs{
			"id": queueID,
		},
	).Scan(
		&queue.ID,
		&queue.CreatedAtNano,
		&queue.DeletedAtNano,
		&queue.QueueName,
		&queue.ParentID,
		&queue.Parent,
		&queue.Status,
		&queue.PartitionID,
		&queue.PendingResource,
		&queue.MaxResource,
		&queue.GuaranteedResource,
		&queue.AllocatedResource,
		&queue.PreemptingResource,
		&queue.HeadRoom,
		&queue.IsLeaf,
		&queue.IsManaged,
		&queue.Properties,
		&queue.TemplateInfo,
		&queue.AbsUsedCapacity,
		&queue.MaxRunningApps,
		&queue.RunningApps,
		&queue.CurrentPriority,
		&queue.AllocatingAcceptedApps,
	)
	if err != nil {
		return nil, fmt.Errorf("could not get queue from DB: %v", err)
	}

	return &queue, nil
}

func (s *PostgresRepository) GetQueuesInPartition(ctx context.Context, partitionID string) ([]*model.Queue, error) {
	const q = `
SELECT
    id,
    created_at_nano,
    deleted_at_nano,
    queue_name,
    status,
    partition_id,
    parent,
    pending_resource,
    max_resource,
    guaranteed_resource,
    allocated_resource,
    preempting_resource,
    head_room,
    is_leaf,
    is_managed,
    properties,
	parent_id,
    template_info,
    abs_used_capacity,
    max_running_apps,
    running_apps,
    current_priority,
    allocating_accepted_apps
FROM queues
WHERE partition_id = @partition_id
ORDER BY id DESC
`
	var queues []*model.Queue
	rows, err := s.dbpool.Query(
		ctx,
		q,
		&pgx.NamedArgs{"partition_id": partitionID},
	)
	if err != nil {
		return nil, fmt.Errorf("could not get queue from DB: %v", err)
	}

	for rows.Next() {
		var queue model.Queue
		if err := rows.Scan(
			&queue.ID,
			&queue.CreatedAtNano,
			&queue.DeletedAtNano,
			&queue.QueueName,
			&queue.Status,
			&queue.PartitionID,
			&queue.Parent,
			&queue.PendingResource,
			&queue.MaxResource,
			&queue.GuaranteedResource,
			&queue.AllocatedResource,
			&queue.PreemptingResource,
			&queue.HeadRoom,
			&queue.IsLeaf,
			&queue.IsManaged,
			&queue.Properties,
			&queue.ParentID,
			&queue.TemplateInfo,
			&queue.AbsUsedCapacity,
			&queue.MaxRunningApps,
			&queue.RunningApps,
			&queue.CurrentPriority,
			&queue.AllocatingAcceptedApps,
		); err != nil {
			return nil, fmt.Errorf("could not get queue from DB: %v", err)
		}

		queues = append(queues, &queue)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("could not get queue from DB: %v", err)
	}

	return queues, nil
}

func (s *PostgresRepository) DeleteQueuesNotInIDs(ctx context.Context, partitionID string, ids []string, deletedAtNano int64) error {
	const q = `
UPDATE queues
SET deleted_at_nano = @deleted_at_nano
WHERE deleted_at_nano IS NULL AND NOT (id = ANY(@ids)) AND partition_id = @partition_id
		`

	_, err := s.dbpool.Exec(
		ctx,
		q,
		pgx.NamedArgs{
			"ids":             ids,
			"deleted_at_nano": deletedAtNano,
			"partition_id":    partitionID,
		},
	)

	return err
}

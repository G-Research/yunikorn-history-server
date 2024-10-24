package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/G-Research/yunikorn-history-server/internal/model"
)

func (s *PostgresRepository) InsertQueue(ctx context.Context, q *model.Queue) error {
	insertSQL := `INSERT INTO queues (
		id, created_at_nano, queue_name, parent_id, status, partition, pending_resource, max_resource,
		guaranteed_resource, allocated_resource, preempting_resource, head_room, is_leaf, is_managed,
		properties, parent, template_info, abs_used_capacity, max_running_apps, running_apps,
		current_priority, allocating_accepted_apps)
		VALUES (@id, @created_at_nano, @queue_name, (
			SELECT CASE WHEN CAST(@parent AS TEXT) IS NOT NULL THEN
				(SELECT id FROM queues WHERE partition = @partition AND queue_name = @queue_name ORDER BY ID DESC LIMIT 1)
			ELSE NULL
			END
		), @status, @partition, @pending_resource, @max_resource,
		@guaranteed_resource, @allocated_resource, @preempting_resource, @head_room, @is_leaf, @is_managed,
		@properties, @parent, @template_info, @abs_used_capacity, @max_running_apps, @running_apps,
		@current_priority, @allocating_accepted_apps)
		RETURNING id`

	_, err := s.dbpool.Exec(ctx, insertSQL,
		pgx.NamedArgs{
			"id":                       q.ID,
			"created_at_nano":          q.CreatedAtNano,
			"queue_name":               q.QueueName,
			"status":                   q.Status,
			"partition":                q.Partition,
			"pending_resource":         q.PendingResource,
			"max_resource":             q.MaxResource,
			"guaranteed_resource":      q.GuaranteedResource,
			"allocated_resource":       q.AllocatedResource,
			"preempting_resource":      q.PreemptingResource,
			"head_room":                q.HeadRoom,
			"is_leaf":                  q.IsLeaf,
			"is_managed":               q.IsManaged,
			"properties":               q.Properties,
			"parent":                   q.Parent,
			"template_info":            q.TemplateInfo,
			"abs_used_capacity":        q.AbsUsedCapacity,
			"max_running_apps":         q.MaxRunningApps,
			"running_apps":             q.RunningApps,
			"current_priority":         q.CurrentPriority,
			"allocating_accepted_apps": q.AllocatingAcceptedApps,
		})

	return err
}

// UpdateQueue updates the queue based on the queue_name and partition.
// If the queue has children, the function will recursively update them.
// If provided child queue does not exist, the function will add it.
// The function returns an error if the update operation fails.
func (s *PostgresRepository) UpdateQueue(ctx context.Context, queue *model.Queue) error {
	updateSQL := `
    UPDATE queues SET
		deleted_at_nano = @deleted_at_nano,
        status = @status,
        partition = @partition,
        pending_resource = @pending_resource,
        max_resource = @max_resource,
        guaranteed_resource = @guaranteed_resource,
        allocated_resource = @allocated_resource,
        preempting_resource = @preempting_resource,
        head_room = @head_room,
        is_leaf = @is_leaf,
        is_managed = @is_managed,
        properties = @properties,
        parent = @parent,
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
			"status":                   queue.Status,
			"partition":                queue.Partition,
			"pending_resource":         queue.PendingResource,
			"max_resource":             queue.MaxResource,
			"guaranteed_resource":      queue.GuaranteedResource,
			"allocated_resource":       queue.AllocatedResource,
			"preempting_resource":      queue.PreemptingResource,
			"head_room":                queue.HeadRoom,
			"is_leaf":                  queue.IsLeaf,
			"is_managed":               queue.IsManaged,
			"properties":               queue.Properties,
			"parent":                   queue.Parent,
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

// GetAllQueues returns all queues from the database as a flat list
// child queues are not nested in the parent queue.Children field
func (s *PostgresRepository) GetAllQueues(ctx context.Context) ([]*model.Queue, error) {
	const q = `
SELECT
    id,
    created_at_nano,
    deleted_at_nano,
    queue_name,
	parent_id,
    status,
    partition,
    pending_resource,
    max_resource,
    guaranteed_resource,
    allocated_resource,
    preempting_resource,
    head_room,
    is_leaf,
    is_managed,
    properties,
    parent,
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
			&q.Status,
			&q.Partition,
			&q.PendingResource,
			&q.MaxResource,
			&q.GuaranteedResource,
			&q.AllocatedResource,
			&q.PreemptingResource,
			&q.HeadRoom,
			&q.IsLeaf,
			&q.IsManaged,
			&q.Properties,
			&q.Parent,
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

func (s *PostgresRepository) GetQueueInPartition(ctx context.Context, partition, queueName string) (*model.Queue, error) {
	const q = `
SELECT
    id,
    created_at_nano,
    deleted_at_nano,
    queue_name,
	parent_id,
    status,
    partition,
    pending_resource,
    max_resource,
    guaranteed_resource,
    allocated_resource,
    preempting_resource,
    head_room,
    is_leaf,
    is_managed,
    properties,
    parent,
    template_info,
    abs_used_capacity,
    max_running_apps,
    running_apps,
    current_priority,
    allocating_accepted_apps
FROM queues
WHERE queue_name = @queue_name AND partition = @partition
ORDER BY id DESC
LIMIT 1
`
	var queue model.Queue
	err := s.dbpool.QueryRow(
		ctx,
		q,
		&pgx.NamedArgs{"queue_name": queueName, "partition": partition},
	).Scan(
		&queue.ID,
		&queue.CreatedAtNano,
		&queue.DeletedAtNano,
		&queue.QueueName,
		&queue.ParentID,
		&queue.Status,
		&queue.Partition,
		&queue.PendingResource,
		&queue.MaxResource,
		&queue.GuaranteedResource,
		&queue.AllocatedResource,
		&queue.PreemptingResource,
		&queue.HeadRoom,
		&queue.IsLeaf,
		&queue.IsManaged,
		&queue.Properties,
		&queue.Parent,
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

func (s *PostgresRepository) GetQueuesInPartition(ctx context.Context, partition string) ([]*model.Queue, error) {
	const q = `
SELECT DISTINCT ON (queue_name)
    id,
    created_at_nano,
    deleted_at_nano,
    queue_name,
    status,
    partition,
    pending_resource,
    max_resource,
    guaranteed_resource,
    allocated_resource,
    preempting_resource,
    head_room,
    is_leaf,
    is_managed,
    properties,
    parent,
    template_info,
    abs_used_capacity,
    max_running_apps,
    running_apps,
    current_priority,
    allocating_accepted_apps
FROM queues
WHERE partition = @partition
ORDER BY queue_name, id DESC
`
	var queues []*model.Queue
	rows, err := s.dbpool.Query(
		ctx,
		q,
		&pgx.NamedArgs{"partition": partition},
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
			&queue.Partition,
			&queue.PendingResource,
			&queue.MaxResource,
			&queue.GuaranteedResource,
			&queue.AllocatedResource,
			&queue.PreemptingResource,
			&queue.HeadRoom,
			&queue.IsLeaf,
			&queue.IsManaged,
			&queue.Properties,
			&queue.Parent,
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

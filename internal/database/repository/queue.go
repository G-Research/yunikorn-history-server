package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid/v2"

	"github.com/G-Research/yunikorn-history-server/internal/model"
	"github.com/G-Research/yunikorn-history-server/internal/util"
)

func (s *PostgresRepository) UpsertQueues(ctx context.Context, queues []*dao.PartitionQueueDAOInfo) error {
	upsertSQL := `INSERT INTO queues (
		id,
        parent_id,
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
		allocating_accepted_apps,
    	created_at) VALUES (@id, @parent_id, @queue_name, @status, @partition, @pending_resource, @max_resource,
		@guaranteed_resource, @allocated_resource, @preempting_resource, @head_room, @is_leaf, @is_managed, @properties,
		@parent, @template_info, @abs_used_capacity, @max_running_apps, @running_apps,
		@current_priority, @allocating_accepted_apps, @created_at)
	ON CONFLICT (partition, queue_name) DO UPDATE SET
		status = EXCLUDED.status,
		pending_resource = EXCLUDED.pending_resource,
		max_resource = EXCLUDED.max_resource,
		guaranteed_resource = EXCLUDED.guaranteed_resource,
		allocated_resource = EXCLUDED.allocated_resource,
		preempting_resource = EXCLUDED.preempting_resource,
		head_room = EXCLUDED.head_room,
		is_leaf = EXCLUDED.is_leaf,
		is_managed = EXCLUDED.is_managed,
		max_running_apps = EXCLUDED.max_running_apps,
		running_apps = EXCLUDED.running_apps`
	for _, q := range queues {
		parentId, err := s.getQueueID(ctx, q.Parent, q.Partition)
		if err != nil {
			return fmt.Errorf("could not get parent queue from DB: %v", err)
		}
		_, err = s.dbpool.Exec(ctx, upsertSQL,
			pgx.NamedArgs{
				"id":                       ulid.Make().String(),
				"parent_id":                parentId,
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
				"created_at":               time.Now().Unix(),
			})
		if err != nil {
			return fmt.Errorf("could not insert/update queue into DB: %v", err)
		}
	}
	return nil
}

func (s *PostgresRepository) AddQueues(ctx context.Context, parentId *string, queues []*dao.PartitionQueueDAOInfo) error {
	insertSQL := `INSERT INTO queues (
		parent_id, queue_name, status, partition, pending_resource, max_resource,
		guaranteed_resource, allocated_resource, preempting_resource, head_room, is_leaf, is_managed,
		properties, parent, template_info, abs_used_capacity, max_running_apps, running_apps,
		current_priority, allocating_accepted_apps, created_at)
		VALUES (@parent_id, @queue_name, @status, @partition, @pending_resource, @max_resource,
		@guaranteed_resource, @allocated_resource, @preempting_resource, @head_room, @is_leaf, @is_managed,
		@properties, @parent, @template_info, @abs_used_capacity, @max_running_apps, @running_apps,
		@current_priority, @allocating_accepted_apps, @created_at)
		RETURNING id`

	for _, q := range queues {
		var err error
		if parentId == nil {
			parentId, err = s.getQueueID(ctx, q.Parent, q.Partition)
			if err != nil {
				return fmt.Errorf("could not get parent queue from DB: %v", err)
			}
		}
		if q.Partition == "" {
			return fmt.Errorf("partition is required for queue %s", q.QueueName)
		}

		row := s.dbpool.QueryRow(ctx, insertSQL,
			pgx.NamedArgs{
				"parent_id":                parentId,
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
				"created_at":               time.Now().Unix(),
			})

		var id string
		err = row.Scan(&id)
		if err != nil {
			return fmt.Errorf("could not add queue %s into DB: %v", q.QueueName, err)
		}

		if len(q.Children) > 0 {
			var children []*dao.PartitionQueueDAOInfo
			for _, child := range q.Children {
				// add parent partition to the child
				// yunikorn does not provide partition for child queues
				child.Partition = q.Partition
				children = append(children, &child)
			}
			err = s.AddQueues(ctx, &id, children)
			if err != nil {
				return fmt.Errorf("could not add one or more children of queue %s into DB: %v", q.QueueName, err)
			}
		}
	}
	return nil
}

// UpdateQueue updates the queue based on the queue_name and partition.
// If the queue has children, the function will recursively update them.
// If provided child queue does not exist, the function will add it.
// The function returns an error if the update operation fails.
func (s *PostgresRepository) UpdateQueue(ctx context.Context, queue *dao.PartitionQueueDAOInfo) error {
	if queue.Partition == "" {
		return fmt.Errorf("partition is required for queue %s", queue.QueueName)
	}

	updateSQL := `
    UPDATE queues SET
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
    WHERE queue_name = @queue_name AND partition = @partition AND deleted_at IS NULL
`
	result, err := s.dbpool.Exec(ctx, updateSQL,
		pgx.NamedArgs{
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

	// If there are children, recursively update them
	if len(queue.Children) > 0 {
		for _, child := range queue.Children {
			// add parent partition to the child
			// yunikorn does not provide partition for child queues
			child.Partition = queue.Partition
			err := s.UpdateQueue(ctx, util.ToPtr(child))
			// if the child queue does not exist, we should add it
			if err != nil {
				err := s.AddQueues(ctx, nil, []*dao.PartitionQueueDAOInfo{&child})
				if err != nil {
					return fmt.Errorf("could not add child queue %s into DB: %v", child.QueueName, err)
				}
			}
		}
	}
	return nil
}

// GetAllQueues returns all queues from the database as a flat list
// child queues are not nested in the parent queue.Children field
func (s *PostgresRepository) GetAllQueues(ctx context.Context) ([]*model.PartitionQueueDAOInfo, error) {
	var queues []*model.PartitionQueueDAOInfo
	rows, err := s.dbpool.Query(ctx, "SELECT * FROM queues")
	if err != nil {
		return nil, fmt.Errorf("could not get queues from DB: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var q model.PartitionQueueDAOInfo
		err = rows.Scan(
			&q.Id,
			&q.ParentId,
			&q.CreatedAt,
			&q.DeletedAt,
			&q.QueueName,
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

// GetQueuesPerPartition returns all the queues associated with the partition as a flat list
// It is left to the caller to build the hierarchy of the queues
func (s *PostgresRepository) GetQueuesPerPartition(
	ctx context.Context,
	parition string,
) ([]*model.PartitionQueueDAOInfo, error) {
	selectSQL := `SELECT * FROM queues WHERE partition = $1`

	rows, err := s.dbpool.Query(ctx, selectSQL, parition)
	if err != nil {
		return nil, fmt.Errorf("could not get queues from DB: %v", err)
	}
	defer rows.Close()

	var queues []*model.PartitionQueueDAOInfo
	for rows.Next() {
		var q model.PartitionQueueDAOInfo
		err = rows.Scan(
			&q.Id,
			&q.ParentId,
			&q.CreatedAt,
			&q.DeletedAt,
			&q.QueueName,
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

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("could not get queues from DB: %v", err)
	}

	return queues, nil
}

// GetQueue the queue with the given name and partition
// child queues are nested in the queue.Children field
func (s *PostgresRepository) GetQueue(ctx context.Context, partition, queueName string) (*model.PartitionQueueDAOInfo, error) {
	selectSQL := `
		WITH RECURSIVE generation AS (
			-- Start with the specific queue based on queue_name and partition
			SELECT *,
				   0 AS generation_number
			FROM queues
			WHERE queue_name = $1
			  AND partition = $2
              -- Only select alive queues
              AND deleted_at IS NULL
			UNION ALL

			-- Recursively fetch all child queues
			SELECT child.*,
				   g.generation_number + 1 AS generation_number
			FROM queues child
			JOIN generation g
			ON g.id = child.parent_id
		)
		SELECT * FROM generation ORDER BY generation_number;
	`
	rows, err := s.dbpool.Query(ctx, selectSQL, queueName, partition)
	if err != nil {
		return nil, fmt.Errorf("could not get queues from DB: %v", err)
	}
	defer rows.Close()

	// Initialize map to track parent-child relationships
	childrenMap := make(map[string][]*model.PartitionQueueDAOInfo)
	var rootQueue *model.PartitionQueueDAOInfo

	for rows.Next() {
		var q model.PartitionQueueDAOInfo
		var generationNumber int
		err = rows.Scan(
			&q.Id,
			&q.ParentId,
			&q.CreatedAt,
			&q.DeletedAt,
			&q.QueueName,
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
			&generationNumber,
		)
		if err != nil {
			return nil, fmt.Errorf("could not scan queue from DB: %v", err)
		}

		// Track the root queue for the current query
		if rootQueue == nil && generationNumber == 0 {
			rootQueue = &q
		} else if q.ParentId != nil {
			// Otherwise, add the queue to the children map
			childrenMap[*q.ParentId] = append(childrenMap[*q.ParentId], &q)
		}
	}

	if rootQueue == nil {
		return nil, fmt.Errorf("queue not found: %s", queueName)
	}
	// Recursively populate the children for the root queue
	rootQueue.Children = getChildrenFromMap(rootQueue.Id, childrenMap)

	return rootQueue, nil
}

func (s *PostgresRepository) getQueueID(ctx context.Context, queueName string, partition string) (*string, error) {
	if queueName == "" {
		return nil, nil
	}
	const queueIDSQL = "SELECT id FROM queues WHERE queue_name = $1 AND partition = $2 AND deleted_at IS NULL"
	var id string
	err := s.dbpool.QueryRow(ctx, queueIDSQL, queueName, partition).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("could not get queueName queue from DB: %v", err)
	}
	return &id, nil
}

func getChildrenFromMap(queueID string, childrenMap map[string][]*model.PartitionQueueDAOInfo) []*model.PartitionQueueDAOInfo {
	children := childrenMap[queueID]
	var childrenResult []*model.PartitionQueueDAOInfo
	for _, child := range children {
		child.Children = getChildrenFromMap(child.Id, childrenMap)
		childrenResult = append(childrenResult, child)
	}
	return childrenResult
}

// DeleteQueues marks the provided queues and their children as deleted.
// The function works recursively, so if a queue has children, the function will
// call itself with the children queues.
func (s *PostgresRepository) DeleteQueues(ctx context.Context, queues []*model.PartitionQueueDAOInfo) error {
	deleteSQL := `UPDATE queues SET deleted_at = $1 WHERE id = $2`
	for _, q := range queues {

		// If there are children, recursively delete them
		if len(q.Children) > 0 {
			err := s.DeleteQueues(ctx, q.Children)
			if err != nil {
				return err
			}
		}

		// Delete the current queue
		_, err := s.dbpool.Exec(ctx, deleteSQL, time.Now().Unix(), q.Id)
		if err != nil {
			return fmt.Errorf("could not delete queue from DB: %v", err)
		}
	}
	return nil
}

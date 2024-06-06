package repository

import (
	"context"
	"fmt"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *RepoPostgres) UpsertQueues(ctx context.Context, queues []dao.PartitionQueueDAOInfo) error {
	upsertSQL := `INSERT INTO queues (
		id,
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
		children,
		children_names,
		abs_used_capacity,
		max_running_apps,
		running_apps,
		current_priority,
		allocating_accepted_apps) VALUES (@id, @queue_name, @status, @partition, @pending_resource, @max_resource,
		@guaranteed_resource, @allocated_resource, @preempting_resource, @head_room, @is_leaf, @is_managed, @properties,
		@parent, @template_info, @children, @children_names, @abs_used_capacity, @max_running_apps, @running_apps,
		@current_priority, @allocating_accepted_apps)
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
		children = EXCLUDED.children,
		children_names = EXCLUDED.children_names,
		max_running_apps = EXCLUDED.max_running_apps,
		running_apps = EXCLUDED.running_apps`
	for _, q := range queues {
		_, err := s.dbpool.Exec(ctx, upsertSQL,
			pgx.NamedArgs{
				"id":                       uuid.NewString(),
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
				"children":                 q.Children,
				"children_names":           q.ChildrenNames,
				"abs_used_capacity":        q.AbsUsedCapacity,
				"max_running_apps":         q.MaxRunningApps,
				"running_apps":             q.RunningApps,
				"current_priority":         q.CurrentPriority,
				"allocating_accepted_apps": q.AllocatingAcceptedApps,
			})
		if err != nil {
			return fmt.Errorf("could not insert/update queue into DB: %v", err)
		}
	}
	return nil
}

func (s *RepoPostgres) GetAllQueues(ctx context.Context) ([]*dao.PartitionQueueDAOInfo, error) {
	var queues []*dao.PartitionQueueDAOInfo
	rows, err := s.dbpool.Query(ctx, "SELECT * FROM queues")
	if err != nil {
		return nil, fmt.Errorf("could not get queues from DB: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var q dao.PartitionQueueDAOInfo
		var id string
		err = rows.Scan(
			&id,
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
			&q.Children,
			&q.ChildrenNames,
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

func (s RepoPostgres) GetQueuesPerPartition(ctx context.Context, parition string) (
	[]*dao.PartitionQueueDAOInfo, error) {

	selectSQL := `SELECT * FROM queues WHERE partition = $1`

	var queues []*dao.PartitionQueueDAOInfo

	rows, err := s.dbpool.Query(ctx, selectSQL, parition)
	if err != nil {
		return nil, fmt.Errorf("could not get queues from DB: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var q dao.PartitionQueueDAOInfo
		var id string
		err = rows.Scan(
			&id,
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
			&q.Children,
			&q.ChildrenNames,
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

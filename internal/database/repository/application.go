package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/G-Research/yunikorn-history-server/internal/model"
	"github.com/G-Research/yunikorn-history-server/internal/util"

	"github.com/G-Research/yunikorn-history-server/internal/database/sql"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/jackc/pgx/v5"
)

type ApplicationFilters struct {
	SubmissionStartTime *time.Time
	SubmissionEndTime   *time.Time
	FinishedStartTime   *time.Time
	FinishedEndTime     *time.Time
	User                *string
	Groups              []string
	Offset              *int
	Limit               *int
}

// applyApplicationFilters adds application filters to the sql query using positional arguments and
// returns the arguments in the same order.
func applyApplicationFilters(builder *sql.Builder, filters ApplicationFilters) {
	if filters.SubmissionStartTime != nil {
		builder.Conditionp("submission_time", ">=", filters.SubmissionStartTime.UnixMilli())
	}
	if filters.SubmissionEndTime != nil {
		builder.Conditionp("submission_time", "<=", filters.SubmissionEndTime.UnixMilli())
	}
	if filters.FinishedStartTime != nil {
		builder.Conditionp("finished_time", ">=", filters.FinishedStartTime.UnixMilli())
	}
	if filters.FinishedEndTime != nil {
		builder.Conditionp("finished_time", "<=", filters.FinishedEndTime.UnixMilli())
	}
	if len(filters.Groups) > 0 {
		builder.Conditionf("groups && ARRAY[%s]", util.SliceToCommaSeparated(filters.Groups, true))
	}
	if filters.User != nil {
		builder.Conditionp("\"user\"", "=", *filters.User)
	}
	applyLimitAndOffset(builder, filters.Limit, filters.Offset)
}

func (s *PostgresRepository) UpsertApplications(ctx context.Context, apps []*dao.ApplicationDAOInfo) error {
	upsertSQL := `INSERT INTO applications (id, app_id, used_resource, max_used_resource, pending_resource,
			partition, queue_name, queue_id, submission_time, finished_time, requests, allocations, state,
			"user", groups, rejected_message, state_log, place_holder_data, has_reserved, reservations,
			max_request_priority)
			VALUES (@id, @app_id,@used_resource, @max_used_resource, @pending_resource, @partition, @queue_name, @queue_id,
			@submission_time, @finished_time, @requests, @allocations, @state, @user, @groups,
			@rejected_message, @state_log, @place_holder_data, @has_reserved, @reservations, @max_request_priority)
		ON CONFLICT (partition, queue_name, app_id) DO UPDATE SET
			used_resource = COALESCE(EXCLUDED.used_resource, applications.used_resource),
			max_used_resource = COALESCE(EXCLUDED.max_used_resource, applications.max_used_resource),
			pending_resource = COALESCE(EXCLUDED.pending_resource, applications.pending_resource),
			finished_time = COALESCE(EXCLUDED.finished_time, applications.finished_time),
			requests = COALESCE(EXCLUDED.requests, applications.requests),
			allocations = COALESCE(EXCLUDED.allocations, applications.allocations),
			state = COALESCE(EXCLUDED.state, applications.state),
			rejected_message = COALESCE(EXCLUDED.rejected_message, applications.rejected_message),
			state_log = COALESCE(EXCLUDED.state_log, applications.state_log),
			place_holder_data = COALESCE(EXCLUDED.place_holder_data, applications.place_holder_data),
			has_reserved = COALESCE(EXCLUDED.has_reserved, applications.has_reserved),
			reservations = COALESCE(EXCLUDED.reservations, applications.reservations),
			max_request_priority = COALESCE(EXCLUDED.max_request_priority, applications.max_request_priority)`

	for _, a := range apps {
		queueId, err := s.getQueueID(ctx, a.QueueName, a.Partition)
		if err != nil {
			return fmt.Errorf("could not get queue_id from DB: %v", err)
		}
		_, err = s.dbpool.Exec(ctx, upsertSQL,
			pgx.NamedArgs{
				"id":                   ulid.Make().String(),
				"app_id":               a.ApplicationID,
				"used_resource":        a.UsedResource,
				"max_used_resource":    a.MaxUsedResource,
				"pending_resource":     a.PendingResource,
				"partition":            a.Partition,
				"queue_name":           a.QueueName,
				"queue_id":             queueId,
				"submission_time":      a.SubmissionTime,
				"finished_time":        a.FinishedTime,
				"requests":             a.Requests,
				"allocations":          a.Allocations,
				"state":                a.State,
				"user":                 a.User,
				"groups":               a.Groups,
				"rejected_message":     a.RejectedMessage,
				"state_log":            a.StateLog,
				"place_holder_data":    a.PlaceholderData,
				"has_reserved":         a.HasReserved,
				"reservations":         a.Reservations,
				"max_request_priority": a.MaxRequestPriority,
			})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresRepository) GetAllApplications(ctx context.Context, filters ApplicationFilters) ([]*model.ApplicationDAOInfo, error) {
	queryBuilder := sql.NewBuilder().SelectAll("applications", "a").OrderBy("a.submission_time", sql.OrderByDescending)
	applyApplicationFilters(queryBuilder, filters)

	var apps []*model.ApplicationDAOInfo

	query := queryBuilder.Query()
	args := queryBuilder.Args()
	rows, err := s.dbpool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("could not get applications from DB: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var app model.ApplicationDAOInfo
		var id string
		err := rows.Scan(&id, &app.ApplicationID, &app.UsedResource, &app.MaxUsedResource, &app.PendingResource,
			&app.Partition, &app.QueueName, &app.QueueID, &app.SubmissionTime, &app.FinishedTime, &app.Requests, &app.Allocations,
			&app.State, &app.User, &app.Groups, &app.RejectedMessage, &app.StateLog, &app.PlaceholderData,
			&app.HasReserved, &app.Reservations, &app.MaxRequestPriority)
		if err != nil {
			return nil, fmt.Errorf("could not scan application from DB: %v", err)
		}
		apps = append(apps, &app)
	}
	return apps, nil
}

func (s *PostgresRepository) GetAppsPerPartitionPerQueue(ctx context.Context, partition, queue string, filters ApplicationFilters) (
	[]*model.ApplicationDAOInfo, error) {
	queryBuilder := sql.NewBuilder().
		SelectAll("applications", "").
		Conditionp("queue_name", "=", queue).
		Conditionp("partition", "=", partition).
		OrderBy("submission_time", sql.OrderByDescending)
	applyApplicationFilters(queryBuilder, filters)

	var apps []*model.ApplicationDAOInfo

	query := queryBuilder.Query()
	args := queryBuilder.Args()
	rows, err := s.dbpool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("could not get applications from DB: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var app model.ApplicationDAOInfo
		var id string
		err := rows.Scan(&id, &app.ApplicationID, &app.UsedResource, &app.MaxUsedResource, &app.PendingResource,
			&app.Partition, &app.QueueName, &app.QueueID, &app.SubmissionTime, &app.FinishedTime, &app.Requests, &app.Allocations,
			&app.State, &app.User, &app.Groups, &app.RejectedMessage, &app.StateLog, &app.PlaceholderData,
			&app.HasReserved, &app.Reservations, &app.MaxRequestPriority)
		if err != nil {
			return nil, fmt.Errorf("could not scan application from DB: %v", err)
		}
		apps = append(apps, &app)
	}
	return apps, nil
}

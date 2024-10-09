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

func (s *PostgresRepository) InsertApplication(ctx context.Context, app *model.Application) error {
	const q = `
INSERT INTO applications
(id, created_at, deleted_at, app_id, used_resource, max_used_resource, pending_resource, partition, queue_name, submission_time, finished_time, requests, allocations, state, "user", groups, rejected_message, state_log, place_holder_data, has_reserved, reservations, max_request_priority)
VALUES
(@id, @created_at, @deleted_at, @app_id, @used_resource, @max_used_resource, @pending_resource, @partition, @queue_name, @submission_time, @finished_time, @requests, @allocations, @state, @user, @groups, @rejected_message, @state_log, @place_holder_data, @has_reserved, @reservations, @max_request_priority)
	`

	_, err := s.dbpool.Exec(
		ctx,
		q,
		pgx.NamedArgs{
			"id":                   app.ID,
			"created_at":           app.CreatedAt,
			"deleted_at":           app.DeletedAt,
			"app_id":               app.ApplicationID,
			"used_resource":        app.UsedResource,
			"max_used_resource":    app.MaxUsedResource,
			"pending_resource":     app.PendingResource,
			"partition":            app.Partition,
			"queue_name":           app.QueueName,
			"submission_time":      app.SubmissionTime,
			"finished_time":        app.FinishedTime,
			"requests":             app.Requests,
			"allocations":          app.Allocations,
			"state":                app.State,
			"user":                 app.User,
			"groups":               app.Groups,
			"rejected_message":     app.RejectedMessage,
			"state_log":            app.StateLog,
			"place_holder_data":    app.PlaceholderData,
			"has_reserved":         app.HasReserved,
			"reservations":         app.Reservations,
			"max_request_priority": app.MaxRequestPriority,
		})
	return err
}

func (s *PostgresRepository) GetActiveApplicationByApplicationID(ctx context.Context, appID string) (*model.Application, error) {
	const q = `
SELECT id, created_at, deleted_at, app_id, used_resource, max_used_resource, pending_resource, partition, queue_name, submission_time, finished_time, requests, allocations, state, "user", groups, rejected_message, state_log, place_holder_data, has_reserved, reservations, max_request_priority
FROM applications
WHERE app_id = @app_id AND deleted_at IS NULL
	`

	var app model.Application
	row := s.dbpool.QueryRow(
		ctx,
		q,
		pgx.NamedArgs{
			"app_id": appID,
		},
	)

	if err := row.Scan(
		&app.ID,
		&app.CreatedAt,
		&app.DeletedAt,
		&app.ApplicationID,
		&app.UsedResource,
		&app.MaxUsedResource,
		&app.PendingResource,
		&app.Partition,
		&app.QueueName,
		&app.SubmissionTime,
		&app.FinishedTime,
		&app.Requests,
		&app.Allocations,
		&app.State,
		&app.User,
		&app.Groups,
		&app.RejectedMessage,
		&app.StateLog,
		&app.PlaceholderData,
		&app.HasReserved,
		&app.Reservations,
		&app.MaxRequestPriority,
	); err != nil {
		return nil, err
	}

	return &app, nil
}

func (s *PostgresRepository) UpdateApplication(ctx context.Context, app *model.Application) error {
	const q = `
UPDATE applications
SET
	deleted_at = @deleted_at,
	used_resource = @used_resource,
	max_used_resource = @max_used_resource,
	pending_resource = @pending_resource,
	finished_time = @finished_time,
	requests = @requests,
	allocations = @allocations,
	state = @state,
	rejected_message = @rejected_message,
	state_log = @state_log,
	place_holder_data = @place_holder_data,
	has_reserved = @has_reserved,
	reservations = @reservations,
	max_request_priority = @max_request_priority
WHERE id = @id
	`

	res, err := s.dbpool.Exec(
		ctx,
		q,
		pgx.NamedArgs{
			"id":                   app.ID,
			"deleted_at":           app.DeletedAt,
			"used_resource":        app.UsedResource,
			"max_used_resource":    app.MaxUsedResource,
			"pending_resource":     app.PendingResource,
			"finished_time":        app.FinishedTime,
			"requests":             app.Requests,
			"allocations":          app.Allocations,
			"state":                app.State,
			"rejected_message":     app.RejectedMessage,
			"state_log":            app.StateLog,
			"place_holder_data":    app.PlaceholderData,
			"has_reserved":         app.HasReserved,
			"reservations":         app.Reservations,
			"max_request_priority": app.MaxRequestPriority,
		})
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("application with id %s not found", app.ID)
	}

	return nil
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

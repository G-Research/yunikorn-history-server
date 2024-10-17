package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/G-Research/yunikorn-history-server/internal/database/sql"
	"github.com/G-Research/yunikorn-history-server/internal/model"
	"github.com/G-Research/yunikorn-history-server/internal/util"
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
(
	id,
	created_at_nano,
	deleted_at_nano,
	app_id,
	used_resource,
	max_used_resource,
	pending_resource,
	partition,
	queue_name,
	submission_time,
	finished_time,
	requests,
	allocations,
	state,
	"user",
	groups,
	rejected_message,
	state_log,
	place_holder_data,
	has_reserved,
	reservations,
	max_request_priority
)
VALUES
(
	@id,
	@created_at_nano,
	@deleted_at_nano,
	@app_id,
	@used_resource,
	@max_used_resource,
	@pending_resource,
	@partition,
	@queue_name,
	@submission_time,
	@finished_time,
	@requests,
	@allocations,
	@state,
	@user,
	@groups,
	@rejected_message,
	@state_log,
	@place_holder_data,
	@has_reserved,
	@reservations,
	@max_request_priority
)
	`

	_, err := s.dbpool.Exec(
		ctx,
		q,
		pgx.NamedArgs{
			"id":                   app.ID,
			"created_at_nano":      app.CreatedAtNano,
			"deleted_at_nano":      app.DeletedAtNano,
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

func (s *PostgresRepository) GetLatestApplicationByApplicationID(ctx context.Context, appID string) (*model.Application, error) {
	const q = `
SELECT
	id,
	created_at_nano,
	deleted_at_nano,
	app_id,
	used_resource,
	max_used_resource,
	pending_resource,
	partition,
	queue_name,
	submission_time,
	finished_time,
	requests,
	allocations,
	state,
	"user",
	groups,
	rejected_message,
	state_log,
	place_holder_data,
	has_reserved,
	reservations,
	max_request_priority
FROM
	applications
WHERE
	app_id = @app_id
ORDER BY id DESC
LIMIT 1
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
		&app.CreatedAtNano,
		&app.DeletedAtNano,
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

func (s *PostgresRepository) GetLatestApplicationsByApplicationID(ctx context.Context) ([]*model.Application, error) {
	const q = `
SELECT DISTINCT ON (app_id)
	id,
	created_at_nano,
	deleted_at_nano,
	app_id,
	used_resource,
	max_used_resource,
	pending_resource,
	partition,
	queue_name,
	submission_time,
	finished_time,
	requests,
	allocations,
	state,
	"user",
	groups,
	rejected_message,
	state_log,
	place_holder_data,
	has_reserved,
	reservations,
	max_request_priority
FROM
	applications
ORDER BY app_id, id DESC`

	rows, err := s.dbpool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("could not get applications from DB: %v", err)
	}
	defer rows.Close()

	var apps []*model.Application
	for rows.Next() {
		var app model.Application
		if err := rows.Scan(
			&app.ID,
			&app.CreatedAtNano,
			&app.DeletedAtNano,
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
			return nil, fmt.Errorf("could not scan application from DB: %v", err)
		}
		apps = append(apps, &app)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read rows: %v", err)
	}

	return apps, nil
}

func (s *PostgresRepository) UpdateApplication(ctx context.Context, app *model.Application) error {
	const q = `
UPDATE applications
SET
	deleted_at_nano = @deleted_at_nano,
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
			"deleted_at_nano":      app.DeletedAtNano,
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
		},
	)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("application with id %s not found", app.ID)
	}

	return nil
}

func (s *PostgresRepository) GetAllApplications(ctx context.Context, filters ApplicationFilters) ([]*model.Application, error) {
	queryBuilder := sql.NewBuilder().SelectAll("applications", "a").OrderBy("a.submission_time", sql.OrderByDescending)
	applyApplicationFilters(queryBuilder, filters)

	var apps []*model.Application

	query := queryBuilder.Query()
	args := queryBuilder.Args()
	rows, err := s.dbpool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("could not get applications from DB: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var app model.Application
		err := rows.Scan(
			&app.ID,
			&app.CreatedAtNano,
			&app.DeletedAtNano,
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
		)
		if err != nil {
			return nil, fmt.Errorf("could not scan application from DB: %v", err)
		}
		apps = append(apps, &app)
	}
	return apps, nil
}

//nolint:all
func (s *PostgresRepository) GetAppsPerPartitionPerQueue(ctx context.Context, partition, queue string, filters ApplicationFilters) ([]*model.Application, error) {
	queryBuilder := sql.NewBuilder().
		SelectAll("applications", "").
		Conditionp("queue_name", "=", queue).
		Conditionp("partition", "=", partition).
		OrderBy("submission_time", sql.OrderByDescending)
	applyApplicationFilters(queryBuilder, filters)

	var apps []*model.Application

	query := queryBuilder.Query()
	args := queryBuilder.Args()
	rows, err := s.dbpool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("could not get applications from DB: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var app model.Application

		err := rows.Scan(
			&app.ID,
			&app.CreatedAtNano,
			&app.DeletedAtNano,
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
		)
		if err != nil {
			return nil, fmt.Errorf("could not scan application from DB: %v", err)
		}
		apps = append(apps, &app)
	}
	return apps, nil
}

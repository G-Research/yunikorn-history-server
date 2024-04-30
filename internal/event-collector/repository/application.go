package repository

import (
	"context"
	"fmt"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *RepoPostgres) UpsertApplications(apps []*dao.ApplicationDAOInfo) error {
	insertSQL := `INSERT INTO applications (id, app_id, used_resource, max_used_resource, pending_resource,
			partition, queue_name, submission_time, finished_time, requests, allocations, state,
			"user", groups, rejected_message, state_log, place_holder_data, has_reserved, reservations,
			max_request_priority)
			VALUES (@id, @app_id,@used_resource, @max_used_resource, @pending_resource, @partition, @queue_name,
			@submission_time, @finished_time, @requests, @allocations, @state, @user, @groups,
			@rejected_message, @state_log, @place_holder_data, @has_reserved, @reservations, @max_request_priority)
		ON CONFLICT (partition, queue_name, app_id) DO UPDATE SET
			used_resource = EXCLUDED.used_resource,
			max_used_resource = EXCLUDED.max_used_resource,
			pending_resource = EXCLUDED.pending_resource,
			finished_time = EXCLUDED.finished_time,
			requests = EXCLUDED.requests,
			allocations = EXCLUDED.allocations,
			state = EXCLUDED.state,
			rejected_message = EXCLUDED.rejected_message,
			state_log = EXCLUDED.state_log,
			place_holder_data = EXCLUDED.place_holder_data,
			has_reserved = EXCLUDED.has_reserved,
			reservations = EXCLUDED.reservations,
			max_request_priority = EXCLUDED.max_request_priority`

	for _, a := range apps {
		_, err := s.dbpool.Exec(context.Background(), insertSQL,
			pgx.NamedArgs{
				"id":                   uuid.NewString(),
				"app_id":               a.ApplicationID,
				"used_resource":        a.UsedResource,
				"max_used_resource":    a.MaxUsedResource,
				"pending_resource":     a.PendingResource,
				"partition":            a.Partition,
				"queue_name":           a.QueueName,
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
			return fmt.Errorf("could not insert application into DB: %v", err)
		}
	}
	return nil
}

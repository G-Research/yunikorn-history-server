package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid/v2"

	"github.com/G-Research/yunikorn-history-server/internal/database/sql"
)

type HistoryFilters struct {
	TimestampStart *time.Time
	TimestampEnd   *time.Time
	Offset         *int
	Limit          *int
}

func applyHistoryFilters(builder *sql.Builder, filters HistoryFilters) {
	if filters.TimestampStart != nil {
		builder.Conditionp("timestamp", ">=", filters.TimestampStart.UnixMilli())
	}
	if filters.TimestampEnd != nil {
		builder.Conditionp("timestamp", "<=", filters.TimestampEnd.UnixMilli())
	}
	applyLimitAndOffset(builder, filters.Limit, filters.Offset)
}

func (s *PostgresRepository) UpdateHistory(
	ctx context.Context,
	apps []*dao.ApplicationHistoryDAOInfo,
	containers []*dao.ContainerHistoryDAOInfo,
) error {

	appSQL := `INSERT INTO history (id, history_type, total_number, timestamp)
		VALUES (@id, 'application', @total_number, @timestamp)`
	containerSQL := `INSERT INTO history (id, history_type, total_number, timestamp)
		VALUES (@id, 'container', @total_number, @timestamp)`

	for _, app := range apps {
		_, err := s.dbpool.Exec(ctx, appSQL,
			pgx.NamedArgs{
				"id":           ulid.Make().String(),
				"total_number": app.TotalApplications,
				"timestamp":    app.Timestamp,
			})
		if err != nil {
			return fmt.Errorf("could not update applications history into DB: %v", err)
		}
	}
	for _, container := range containers {
		_, err := s.dbpool.Exec(ctx, containerSQL,
			pgx.NamedArgs{
				"id":           ulid.Make().String(),
				"total_number": container.TotalContainers,
				"timestamp":    container.Timestamp,
			})
		if err != nil {
			return fmt.Errorf("could not update containers history into DB: %v", err)
		}
	}
	return nil
}

func (s *PostgresRepository) GetApplicationsHistory(ctx context.Context, filters HistoryFilters) ([]*dao.ApplicationHistoryDAOInfo, error) {
	queryBuilder := sql.NewBuilder().
		SelectAll("history", "").
		Conditionp("history_type", "=", "application").
		OrderBy("timestamp", sql.OrderByDescending)
	applyHistoryFilters(queryBuilder, filters)

	var apps []*dao.ApplicationHistoryDAOInfo

	query := queryBuilder.Query()
	args := queryBuilder.Args()
	rows, err := s.dbpool.Query(ctx, query, args...)

	if err != nil {
		return nil, fmt.Errorf("could not get applications history from DB: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var app dao.ApplicationHistoryDAOInfo
		var id string
		err := rows.Scan(&id, nil, &app.TotalApplications, &app.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("could not scan applications history from DB: %v", err)
		}
		apps = append(apps, &app)
	}
	return apps, nil
}

func (s *PostgresRepository) GetContainersHistory(ctx context.Context, filters HistoryFilters) ([]*dao.ContainerHistoryDAOInfo, error) {
	queryBuilder := sql.NewBuilder().
		SelectAll("history", "").
		Conditionp("history_type", "=", "container").
		OrderBy("timestamp", sql.OrderByDescending)
	applyHistoryFilters(queryBuilder, filters)

	var containers []*dao.ContainerHistoryDAOInfo

	query := queryBuilder.Query()
	args := queryBuilder.Args()
	rows, err := s.dbpool.Query(ctx, query, args...)

	if err != nil {
		return nil, fmt.Errorf("could not get container history from DB: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var container dao.ContainerHistoryDAOInfo
		var id string
		err := rows.Scan(&id, nil, &container.TotalContainers, &container.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("could not scan contaienrs history from DB: %v", err)
		}
		containers = append(containers, &container)
	}
	return containers, nil
}

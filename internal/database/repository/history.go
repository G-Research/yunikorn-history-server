package repository

import (
	"context"
	"fmt"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid/v2"
)

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

func (s *PostgresRepository) GetApplicationsHistory(ctx context.Context) ([]*dao.ApplicationHistoryDAOInfo, error) {
	selectSQL := `SELECT * FROM history WHERE history_type = 'application'`
	rows, err := s.dbpool.Query(ctx, selectSQL)
	if err != nil {
		return nil, fmt.Errorf("could not get applications history from DB: %v", err)
	}
	defer rows.Close()

	var apps []*dao.ApplicationHistoryDAOInfo
	for rows.Next() {
		app := &dao.ApplicationHistoryDAOInfo{}
		var id string
		err := rows.Scan(&id, nil, &app.TotalApplications, &app.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("could not scan applications history from DB: %v", err)
		}
		apps = append(apps, app)
	}
	return apps, nil
}

func (s *PostgresRepository) GetContainersHistory(ctx context.Context) ([]*dao.ContainerHistoryDAOInfo, error) {
	selectSQL := `SELECT * FROM history WHERE history_type = 'container'`
	rows, err := s.dbpool.Query(ctx, selectSQL)
	if err != nil {
		return nil, fmt.Errorf("could not get containers history from DB: %v", err)
	}
	defer rows.Close()

	var containers []*dao.ContainerHistoryDAOInfo
	for rows.Next() {
		container := &dao.ContainerHistoryDAOInfo{}
		var id string
		err := rows.Scan(&id, nil, &container.TotalContainers, &container.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("could not scan contaienrs history from DB: %v", err)
		}
		containers = append(containers, container)
	}
	return containers, nil
}

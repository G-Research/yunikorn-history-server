package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	_ "github.com/jackc/pgx/v5"
)

func (s *PostgresRepository) GetResourceUsage(ctx context.Context, partition string) ([]*dao.UserResourceUsageDAOInfo, error) {
	appQuery := `SELECT distinct "user", groups FROM applications WHERE partition = $1`
	rows, err := s.dbpool.Query(ctx, appQuery, partition)
	if err != nil {
		return nil, fmt.Errorf("could not get applications from DB: %v", err)
	}
	defer rows.Close()

	var userResourceUsages []*dao.UserResourceUsageDAOInfo
	for rows.Next() {
		var user string
		var groups []string
		err = rows.Scan(&user, &groups)
		if err != nil {
			return nil, fmt.Errorf("could not scan applications from DB: %v", err)
		}

		// get all queues for the user
		queuesQuery := `SELECT distinct queue_name FROM applications WHERE partition = $1 AND "user" = $2`
		queuesRows, err := s.dbpool.Query(ctx, queuesQuery, partition, user)
		if err != nil {
			return nil, fmt.Errorf("could not get queues from DB: %v", err)
		}
		defer queuesRows.Close()

		var queueName string
		for queuesRows.Next() {
			err = queuesRows.Scan(&queueName)
			if err != nil {
				return nil, fmt.Errorf("could not scan queues from DB: %v", err)
			}

			ru, err := s.getResourceUsageByQueue(ctx, queueName)
			if err != nil {
				return nil, err
			}

			userResourceUsages = append(userResourceUsages, &dao.UserResourceUsageDAOInfo{
				UserName: user,
				Groups:   convertGroupsToMap(groups),
				Queues:   ru,
			})
		}
	}
	return userResourceUsages, nil

}

func (s *PostgresRepository) getResourceUsageByQueue(ctx context.Context, queueName string) (*dao.ResourceUsageDAOInfo, error) {
	ru := &dao.ResourceUsageDAOInfo{}
	query := `SELECT queue_name, abs_used_capacity, max_resource, max_running_apps, children FROM queues WHERE queue_name = $1`
	row := s.dbpool.QueryRow(ctx, query, queueName)

	err := row.Scan(&ru.QueuePath, &ru.ResourceUsage, &ru.MaxResources, &ru.MaxApplications, &ru.Children)
	if err != nil {
		return nil, fmt.Errorf("could not get queue from DB: %v", err)
	}
	return ru, nil
}

func convertGroupsToMap(groups []string) map[string]string {
	groupMap := make(map[string]string)
	for _, group := range groups {
		parts := strings.Split(group, ":")
		if len(parts) > 1 {
			groupMap[parts[0]] = parts[1]
		}
	}
	return groupMap
}

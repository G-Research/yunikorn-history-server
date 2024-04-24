package repository

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"richscott/yhs/internal/event-collector/config"
)

type RepoPostgres struct {
	config *config.ECConfig
	dbpool *pgxpool.Pool
}

func NewECRepo(cfg *config.ECConfig) (error, *RepoPostgres) {
	poolCfg, err := pgxpool.ParseConfig(CreateConnectionString(cfg.PostgresConfig.Connection))
	if err != nil {
		return errors.Wrap(err, "cannot parse Postgres connection config"), nil
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return errors.Wrap(err, "cannot create Postgres connection pool"), nil
	}

	return nil, &RepoPostgres{config: cfg, dbpool: pool}
}

// Set up the DB for use, create tables
func (s *RepoPostgres) Setup(ctx context.Context) {
	setupStmts := []string{
		// For yunikorn-core/pkg/webservice/dao/ApplicationDAOInfo struct
		`DROP TABLE IF EXISTS applications`,
		`CREATE TABLE applications(
			id UUID,
			used_resource JSONB NOT NULL,
			max_used_resource JSONB NOT NULL,
			pending_resource JSONB NOT NULL,
			partition TEXT NOT NULL,
			queue_name TEXT NOT NULL,
			submission_time BIGINT NOT NULL,
			finished_time BIGINT,
			requests JSONB NOT NULL,
			allocations JSONB NOT NULL,
			state TEXT,
			"user" TEXT,
			groups TEXT[],
			rejected_message TEXT,
			state_log JSONB NOT NULL,
			place_holder_data JSONB NOT NULL,
			has_reserved BOOLEAN,
			reservations TEXT[],
			max_request_priority INTEGER,
			UNIQUE (id),
			PRIMARY KEY (Id))`,
	}

	for _, stmt := range setupStmts {
		_, err := s.dbpool.Exec(ctx, stmt)
		if err != nil {
			panic(err)
		}
	}
}

func CreateConnectionString(values map[string]string) string {
	// https://www.postgresql.org/docs//libpq-connect.html#id-1.7.3.8.3.5
	pairs := []string{}

	replacer := strings.NewReplacer(`\`, `\\`, `'`, `\'`)
	for k, v := range values {
		pairs = append(pairs, k+"='"+replacer.Replace(v)+"'")
	}

	result := strings.Join(pairs, " ")

	fmt.Fprintf(os.Stderr, "connection string is %s\n", result)
	return result
}

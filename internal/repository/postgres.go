package repository

import (
	"context"
	"strings"

	"richscott/yhs/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type RepoPostgres struct {
	config *config.ECConfig
	dbpool *pgxpool.Pool
}

func NewECRepo(ctx context.Context, cfg *config.ECConfig) (*RepoPostgres, error) {
	poolCfg, err := pgxpool.ParseConfig(CreateConnectionString(cfg.PostgresConfig.Connection))
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse Postgres connection config")
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create Postgres connection pool")
	}

	return &RepoPostgres{config: cfg, dbpool: pool}, nil
}

// Set up the DB for use, create tables
func (s *RepoPostgres) Setup(ctx context.Context) {
	setupStmts := []string{
		`DROP TABLE IF EXISTS partitions`,
		`CREATE TABLE partitions(
			id UUID,
			cluster_id TEXT NOT NULL,
			name TEXT NOT NULL,
			capacity JSONB NOT NULL,
			used_capacity JSONB NOT NULL,
			utilization JSONB NOT NULL,
			total_nodes INTEGER,
			applications JSONB NOT NULL,
			total_containers INTEGER,
			state TEXT,
			last_state_transition_time BIGINT,
			UNIQUE (id),
			UNIQUE (name),
			PRIMARY KEY (id))`,
		// For yunikorn-core/pkg/webservice/dao/ApplicationDAOInfo struct
		`DROP TABLE IF EXISTS applications`,
		`CREATE TABLE applications(
			id UUID,
			app_id TEXT NOT NULL,
			used_resource JSONB,
			max_used_resource JSONB,
			pending_resource JSONB,
			partition TEXT NOT NULL,
			queue_name TEXT NOT NULL,
			submission_time BIGINT NOT NULL,
			finished_time BIGINT,
			requests JSONB,
			allocations JSONB,
			state TEXT,
			"user" TEXT,
			groups TEXT[],
			rejected_message TEXT,
			state_log JSONB,
			place_holder_data JSONB,
			has_reserved BOOLEAN,
			reservations TEXT[],
			max_request_priority INTEGER,
			UNIQUE (id),
			PRIMARY KEY (id))`,
		`CREATE UNIQUE INDEX idx_partition_queue_app_id ON applications (partition, queue_name, app_id)`,
		// for yunikorn-core/pkg/webservice/dao/PartitionQueueDAOInfo struct
		`DROP TABLE IF EXISTS queues`,
		`CREATE TABLE queues(
			id UUID,
			queue_name TEXT NOT NULL,
			status TEXT,
			partition TEXT NOT NULL,
			pending_resource JSONB,
			max_resource JSONB,
			guaranteed_resource JSONB ,
			allocated_resource JSONB ,
			preempting_resource JSONB ,
			head_room JSONB,
			is_leaf BOOLEAN,
			is_managed BOOLEAN,
			properties JSONB,
			parent TEXT,
			template_info JSONB,
			children JSONB,
			children_names TEXT[],
			abs_used_capacity JSONB,
			max_running_apps INTEGER,
			running_apps INTEGER NOT NULL,
			current_priority INTEGER,
			allocating_accepted_apps TEXT[],
			UNIQUE (id),
			PRIMARY KEY (id))`,
		`CREATE UNIQUE INDEX idx_partition_queue_name ON queues (partition, queue_name)`,
		// for yunikorn-core/pkg/webservice/dao/NodeDAOInfo struct
		`DROP TABLE IF EXISTS nodes`,
		`CREATE TABLE nodes(
			id UUID,
			node_id TEXT NOT NULL,
			partition TEXT NOT NULL,
			host_name TEXT NOT NULL,
			rack_name TEXT,
			attributes JSONB,
			capacity JSONB NOT NULL,
			allocated JSONB NOT NULL,
			occupied JSONB NOT NULL,
			available JSONB NOT NULL,
			utilized JSONB NOT NULL,
			allocations JSONB,
			schedulable BOOLEAN,
			is_reserved BOOLEAN,
			reservations TEXT[],
			UNIQUE (id),
			UNIQUE (node_id),
			PRIMARY KEY (id))`,
		// for yunikorn-core/pkg/webservice/dao/PartitionNodesUtilDAOInfo struct
		`DROP TABLE IF EXISTS partition_nodes_util`,
		`CREATE TABLE partition_nodes_util(
			id UUID,
			cluster_id TEXT NOT NULL,
			partition TEXT NOT NULL,
			nodes_util_list JSONB,
			UNIQUE (id),
			PRIMARY KEY (id))`,
		// for yunikorn-core/pkg/webservice/dao/ContainerHistoryDAOInfo and yunikorn-core/pkg/webservice/dao/ApplicationHistoryDAOInfo
		`DROP TABLE IF EXISTS history`,
		`DROP TYPE IF EXISTS history_type`,
		// NOTE(mo-fatah): Is this the best way to do this?
		`CREATE TYPE history_type AS ENUM ('container', 'application')`,
		`CREATE TABLE history(
			id UUID,
			history_type history_type NOT NULL,
			total_number BIGINT NOT NULL,
			timestamp BIGINT NOT NULL,
			UNIQUE (id),
			PRIMARY KEY (id))`,
	}

	for _, stmt := range setupStmts {
		_, err := s.dbpool.Exec(ctx, stmt)
		if err != nil {
			panic(err)
		}
	}
}

func CreateConnectionString(values map[string]string) string {
	pairs := []string{}

	replacer := strings.NewReplacer(`\`, `\\`, `'`, `\'`)
	for k, v := range values {
		pairs = append(pairs, k+"='"+replacer.Replace(v)+"'")
	}

	return strings.Join(pairs, " ")
}

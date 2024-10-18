package repository

import (
	"context"
	"fmt"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid/v2"

	"github.com/G-Research/yunikorn-history-server/internal/database/sql"
	"github.com/G-Research/yunikorn-history-server/internal/model"
)

type NodeFilters struct {
	NodeId      *string
	HostName    *string
	RackName    *string
	Schedulable *bool
	IsReserved  *bool
	Offset      *int
	Limit       *int
}

type NodesUtilFilters struct {
	ClusterID *string
	Partition *string
	Offset    *int
	Limit     *int
}

func applyNodeUtilFilters(builder *sql.Builder, filters NodesUtilFilters) {
	if filters.ClusterID != nil {
		builder.Conditionp("cluster_id", "=", *filters.ClusterID)
	}
	if filters.Partition != nil {
		builder.Conditionp("partition", "=", *filters.Partition)
	}
	applyLimitAndOffset(builder, filters.Limit, filters.Offset)
}

func applyNodeFilters(builder *sql.Builder, filters NodeFilters) {
	if filters.NodeId != nil {
		builder.Conditionp("node_id", "=", *filters.NodeId)
	}
	if filters.HostName != nil {
		builder.Conditionp("host_name", "=", *filters.HostName)
	}
	if filters.RackName != nil {
		builder.Conditionp("rack_name", "=", *filters.RackName)
	}
	if filters.Schedulable != nil {
		builder.Conditionp("schedulable", "=", *filters.Schedulable)
	}
	if filters.IsReserved != nil {
		builder.Conditionp("is_reserved", "=", *filters.IsReserved)
	}
	applyLimitAndOffset(builder, filters.Limit, filters.Offset)
}

func (s *PostgresRepository) UpsertNodes(ctx context.Context, nodes []*dao.NodeDAOInfo, partition string) error {
	upsertSQL := `INSERT INTO nodes (id, node_id, partition, host_name, rack_name, attributes, capacity, allocated,
		occupied, available, utilized, allocations, schedulable, is_reserved, reservations )
		VALUES (@id, @node_id, @partition, @host_name, @rack_name, @attributes, @capacity, @allocated,
		@occupied, @available, @utilized, @allocations, @schedulable, @is_reserved, @reservations)
	ON CONFLICT (node_id) DO UPDATE SET
		capacity = EXCLUDED.capacity,
		allocated = EXCLUDED.allocated,
		occupied = EXCLUDED.occupied,
		available = EXCLUDED.available,
		utilized = EXCLUDED.utilized,
		allocations = EXCLUDED.allocations,
		schedulable = EXCLUDED.schedulable,
		is_reserved = EXCLUDED.is_reserved,
		reservations = EXCLUDED.reservations`

	for _, n := range nodes {
		_, err := s.dbpool.Exec(ctx, upsertSQL,
			pgx.NamedArgs{
				"id":           ulid.Make().String(),
				"node_id":      n.NodeID,
				"partition":    partition,
				"host_name":    n.HostName,
				"rack_name":    n.RackName,
				"attributes":   n.Attributes,
				"capacity":     n.Capacity,
				"allocated":    n.Allocated,
				"occupied":     n.Occupied,
				"available":    n.Available,
				"utilized":     n.Utilized,
				"allocations":  n.Allocations,
				"schedulable":  n.Schedulable,
				"is_reserved":  n.IsReserved,
				"reservations": n.Reservations,
			})
		if err != nil {
			return fmt.Errorf("could not insert application into DB: %v", err)
		}
	}
	return nil
}

func (s *PostgresRepository) InsertNodesUtil(
	ctx context.Context,
	nu *model.NodesUtil,
) error {
	const q = `
INSERT INTO partition_nodes_util 
(id,  created_at_nano, deleted_at_nano, cluster_id, partition, nodes_util_list)
VALUES
(@id, @created_at_nano, @deleted_at_nano, @cluster_id, @partition, @nodes_util_list)
`

	_, err := s.dbpool.Exec(
		ctx,
		q,
		pgx.NamedArgs{
			"id":              nu.ID,
			"created_at_nano": nu.CreatedAtNano,
			"deleted_at_nano": nu.DeletedAtNano,
			"cluster_id":      nu.ClusterID,
			"partition":       nu.Partition,
			"nodes_util_list": nu.NodesUtilList,
		})
	if err != nil {
		return fmt.Errorf("could not insert node utilizations into DB: %w", err)
	}

	return nil
}

func (s *PostgresRepository) GetNodesUtils(
	ctx context.Context,
	filters NodesUtilFilters,
) ([]*model.NodesUtil, error) {
	queryBuilder := sql.NewBuilder().
		SelectAll("partition_nodes_util", "").
		OrderBy("id", sql.OrderByDescending)

	applyNodeUtilFilters(queryBuilder, filters)

	var nodesUtils []*model.NodesUtil

	query := queryBuilder.Query()
	args := queryBuilder.Args()
	rows, err := s.dbpool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("could not get node utilizations from DB: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var nu model.NodesUtil
		err := rows.Scan(&nu.ID, &nu.CreatedAtNano, &nu.DeletedAtNano, &nu.ClusterID, &nu.Partition, &nu.NodesUtilList)
		if err != nil {
			return nil, fmt.Errorf("could not scan node utilizations from DB: %v", err)
		}
		nodesUtils = append(nodesUtils, &nu)
	}
	return nodesUtils, nil
}

func (s *PostgresRepository) GetNodesPerPartition(ctx context.Context, partition string, filters NodeFilters) ([]*dao.NodeDAOInfo, error) {
	queryBuilder := sql.NewBuilder().
		SelectAll("nodes", "").
		Conditionp("partition", "=", partition).
		OrderBy("node_id", sql.OrderByDescending)
	applyNodeFilters(queryBuilder, filters)

	var nodes []*dao.NodeDAOInfo

	query := queryBuilder.Query()
	args := queryBuilder.Args()
	rows, err := s.dbpool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("could not get nodes from DB: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var n dao.NodeDAOInfo
		var id string
		err := rows.Scan(&id, &n.NodeID, nil, &n.HostName, &n.RackName, &n.Attributes, &n.Capacity,
			&n.Allocated, &n.Occupied, &n.Available, &n.Utilized, &n.Allocations, &n.Schedulable,
			&n.IsReserved, &n.Reservations)
		if err != nil {
			return nil, fmt.Errorf("could not scan node: %v", err)
		}
		nodes = append(nodes, &n)
	}
	return nodes, nil
}

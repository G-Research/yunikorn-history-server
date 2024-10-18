package repository

import (
	"context"
	"fmt"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/G-Research/yunikorn-history-server/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid/v2"

	"github.com/G-Research/yunikorn-history-server/internal/database/sql"
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

type NodeUtilFilters struct {
	ClusterID *string
	Partition *string
	Offset    *int
	Limit     *int
}

func applyNodeUtilFilters(builder *sql.Builder, filters NodeUtilFilters) {
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

func (s *PostgresRepository) InsertNode(ctx context.Context, node *model.Node) error {
	const q = `
INSERT INTO nodes (
	id,
    created_at_nano,
	deleted_at_nano,
	node_id,
	partition,
	host_name,
	rack_name,
	attributes,
	capacity,
	allocated,
	occupied,
	available,
	utilized,
	allocations,
	schedulable,
	is_reserved,
	reservations
) VALUES (
	@id,
	@created_at_nano,
	@deleted_at_nano,
	@node_id,
	@partition,
	@host_name,
	@rack_name,
	@attributes,
	@capacity,
	@allocated,
	@occupied,
	@available,
	@utilized,
	@allocations,
	@schedulable,
	@is_reserved,
	@reservations
)`

	_, err := s.dbpool.Exec(ctx, q,
		pgx.NamedArgs{
			"id":              node.ID,
			"created_at_nano": node.CreatedAtNano,
			"deleted_at_nano": node.DeletedAtNano,
			"node_id":         node.NodeID,
			"partition":       node.Partition,
			"host_name":       node.HostName,
			"rack_name":       node.RackName,
			"attributes":      node.Attributes,
			"capacity":        node.Capacity,
			"allocated":       node.Allocated,
			"occupied":        node.Occupied,
			"available":       node.Available,
			"utilized":        node.Utilized,
			"allocations":     node.Allocations,
			"schedulable":     node.Schedulable,
			"is_reserved":     node.IsReserved,
			"reservations":    node.Reservations,
		})
	if err != nil {
		return fmt.Errorf("could not insert node into DB: %v", err)
	}
	return nil
}

func (s *PostgresRepository) UpdateNode(ctx context.Context, node *model.Node) error {
	const q = `
UPDATE nodes
SET
	deleted_at_nano = @deleted_at_nano,
	node_id = @node_id,
	partition = @partition,
	host_name = @host_name,
	rack_name = @rack_name,
	attributes = @attributes,
	capacity = @capacity,
	allocated = @allocated,
	occupied = @occupied,
	available = @available,
	utilized = @utilized,
	allocations = @allocations,
	schedulable = @schedulable,
	is_reserved = @is_reserved,
	reservations = @reservations
WHERE id = @id`

	res, err := s.dbpool.Exec(
		ctx,
		q,
		pgx.NamedArgs{
			"id":              node.ID,
			"deleted_at_nano": node.DeletedAtNano,
			"node_id":         node.NodeID,
			"partition":       node.Partition,
			"host_name":       node.HostName,
			"rack_name":       node.RackName,
			"attributes":      node.Attributes,
			"capacity":        node.Capacity,
			"allocated":       node.Allocated,
			"occupied":        node.Occupied,
			"available":       node.Available,
			"utilized":        node.Utilized,
			"allocations":     node.Allocations,
			"schedulable":     node.Schedulable,
			"is_reserved":     node.IsReserved,
			"reservations":    node.Reservations,
		})
	if err != nil {
		return fmt.Errorf("could not update node in DB: %v", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("failed to update node %q: no rows affected", node.ID)
	}

	return nil
}

func (s *PostgresRepository) GetLatestNodeByID(ctx context.Context, nodeId, partition string) (*model.Node, error) {
	const q = `
SELECT
	id,
    created_at_nano,
	deleted_at_nano,
	node_id,
	partition,
	host_name,
	rack_name,
	attributes,
	capacity,
	allocated,
	occupied,
	available,
	utilized,
	allocations,
	schedulable,
	is_reserved,
	reservations
FROM nodes
WHERE node_id = $1 AND partition = $2
ORDER BY node_id, id DESC
LIMIT 1`

	var node model.Node
	row := s.dbpool.QueryRow(ctx, q, nodeId, partition)
	if err := row.Scan(
		&node.ID,
		&node.CreatedAtNano,
		&node.DeletedAtNano,
		&node.NodeID,
		&node.Partition,
		&node.HostName,
		&node.RackName,
		&node.Attributes,
		&node.Capacity,
		&node.Allocated,
		&node.Occupied,
		&node.Available,
		&node.Utilized,
		&node.Allocations,
		&node.Schedulable,
		&node.IsReserved,
		&node.Reservations,
	); err != nil {
		return nil, fmt.Errorf("could not get node from DB: %v", err)
	}
	return &node, nil
}

func (s *PostgresRepository) GetLatestNodesByID(ctx context.Context, partition string) ([]*model.Node, error) {
	const q = `
SELECT DISTINCT ON (node_id)
	id,
    created_at_nano,
	deleted_at_nano,
	node_id,
	partition,
	host_name,
	rack_name,
	attributes,
	capacity,
	allocated,
	occupied,
	available,
	utilized,
	allocations,
	schedulable,
	is_reserved,
	reservations
FROM nodes
WHERE partition = $1
ORDER BY node_id, id DESC`

	rows, err := s.dbpool.Query(ctx, q, partition)
	if err != nil {
		return nil, fmt.Errorf("could not get nodes from DB: %v", err)
	}
	defer rows.Close()

	var nodes []*model.Node
	for rows.Next() {
		var node model.Node
		if err := rows.Scan(
			&node.ID,
			&node.CreatedAtNano,
			&node.DeletedAtNano,
			&node.NodeID,
			&node.Partition,
			&node.HostName,
			&node.RackName,
			&node.Attributes,
			&node.Capacity,
			&node.Allocated,
			&node.Occupied,
			&node.Available,
			&node.Utilized,
			&node.Allocations,
			&node.Schedulable,
			&node.IsReserved,
			&node.Reservations,
		); err != nil {
			return nil, fmt.Errorf("could not scan node from DB: %v", err)
		}
		nodes = append(nodes, &node)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read rows: %v", err)
	}
	return nodes, nil
}

func (s *PostgresRepository) InsertNodeUtilizations(
	ctx context.Context,
	nus []*dao.PartitionNodesUtilDAOInfo,
) error {
	insertSQL := `INSERT INTO partition_nodes_util (id, cluster_id, partition, nodes_util_list)
		VALUES (@id, @cluster_id, @partition, @nodes_util_list)`

	for _, nu := range nus {
		_, err := s.dbpool.Exec(ctx, insertSQL,
			pgx.NamedArgs{
				"id":              ulid.Make().String(),
				"cluster_id":      nu.ClusterID,
				"partition":       nu.Partition,
				"nodes_util_list": nu.NodesUtilList,
			})
		if err != nil {
			return fmt.Errorf("could not insert node utilizations into DB: %v", err)
		}

	}
	return nil
}

func (s *PostgresRepository) GetNodeUtilizations(
	ctx context.Context,
	filters NodeUtilFilters,
) ([]*dao.PartitionNodesUtilDAOInfo, error) {
	queryBuilder := sql.NewBuilder().
		SelectAll("partition_nodes_util", "").
		OrderBy("id", sql.OrderByDescending)

	applyNodeUtilFilters(queryBuilder, filters)

	var nodesUtil []*dao.PartitionNodesUtilDAOInfo

	query := queryBuilder.Query()
	args := queryBuilder.Args()
	rows, err := s.dbpool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("could not get node utilizations from DB: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var nu dao.PartitionNodesUtilDAOInfo
		var id string
		err := rows.Scan(&id, &nu.ClusterID, &nu.Partition, &nu.NodesUtilList)
		if err != nil {
			return nil, fmt.Errorf("could not scan node utilizations from DB: %v", err)
		}
		nodesUtil = append(nodesUtil, &nu)
	}
	return nodesUtil, nil
}

func (s *PostgresRepository) GetNodesPerPartition(ctx context.Context, partition string, filters NodeFilters) ([]*model.Node, error) {
	queryBuilder := sql.NewBuilder().
		SelectAll("nodes", "").
		Conditionp("partition", "=", partition).
		OrderBy("node_id", sql.OrderByDescending)
	applyNodeFilters(queryBuilder, filters)

	var nodes []*model.Node

	query := queryBuilder.Query()
	args := queryBuilder.Args()
	rows, err := s.dbpool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("could not get nodes from DB: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var n model.Node
		err := rows.Scan(
			&n.ID,
			&n.CreatedAtNano,
			&n.DeletedAtNano,
			&n.NodeID,
			&n.Partition,
			&n.HostName,
			&n.RackName,
			&n.Attributes,
			&n.Capacity,
			&n.Allocated,
			&n.Occupied,
			&n.Available,
			&n.Utilized,
			&n.Allocations,
			&n.Schedulable,
			&n.IsReserved,
			&n.Reservations,
		)
		if err != nil {
			return nil, fmt.Errorf("could not scan node: %v", err)
		}
		nodes = append(nodes, &n)
	}
	return nodes, nil
}

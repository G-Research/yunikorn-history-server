package repository

import (
	"context"
	"fmt"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid/v2"
)

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

func (s *PostgresRepository) GetNodeUtilizations(ctx context.Context) ([]*dao.PartitionNodesUtilDAOInfo, error) {
	selectSQL := `SELECT * FROM partition_nodes_util`

	rows, err := s.dbpool.Query(ctx, selectSQL)
	if err != nil {
		return nil, fmt.Errorf("could not get node utilizations from DB: %v", err)
	}
	defer rows.Close()

	var nodesUtil []*dao.PartitionNodesUtilDAOInfo
	for rows.Next() {
		nu := &dao.PartitionNodesUtilDAOInfo{}
		var id string
		err := rows.Scan(&id, &nu.ClusterID, &nu.Partition, &nu.NodesUtilList)
		if err != nil {
			return nil, fmt.Errorf("could not scan node utilizations from DB: %v", err)
		}
		nodesUtil = append(nodesUtil, nu)
	}
	return nodesUtil, nil
}

func (s *PostgresRepository) GetNodesPerPartition(ctx context.Context, partition string) ([]*dao.NodeDAOInfo, error) {
	selectSQL := "SELECT * FROM nodes WHERE partition = $1"

	nodes := []*dao.NodeDAOInfo{}

	rows, err := s.dbpool.Query(ctx, selectSQL, partition)
	if err != nil {
		return nil, fmt.Errorf("could not get nodes from DB: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		n := dao.NodeDAOInfo{}
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

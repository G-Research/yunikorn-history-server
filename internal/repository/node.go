package repository

import (
	"context"
	"fmt"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *RepoPostgres) UpsertNodes(nodes []*dao.NodeDAOInfo) error {
	upsertNode := `INSERT INTO nodes (id, node_id, host_name, rack_name, attributes, capacity, allocated,
		occupied, available, utilized, allocations, schedulable, is_reserved, reservations )
		VALUES (@id, @node_id, @host_name, @rack_name, @attributes, @capacity, @allocated,
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
		_, err := s.dbpool.Exec(context.Background(), upsertNode,
			pgx.NamedArgs{
				"id":           uuid.NewString(),
				"node_id":      n.NodeID,
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

func (s *RepoPostgres) InsertNodeUtilizations(u uuid.UUID, nus *[]dao.PartitionNodesUtilDAOInfo) error {
	insertSQL := `INSERT INTO partition_nodes_util (id, cluster_id, partition, nodes_util_list)
		VALUES (@id, @cluster_id, @partition, @nodes_util_list)`

	for _, nu := range *nus {
		_, err := s.dbpool.Exec(context.Background(), insertSQL,
			pgx.NamedArgs{
				"id":              u.String(),
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

func (s RepoPostgres) GetNodesPerPartition(partition string) ([]*dao.NodeDAOInfo, error) {
	selectStmt := "SELECT * FROM nodes WHERE partition = $1"

	nodes := []*dao.NodeDAOInfo{}

	rows, err := s.dbpool.Query(context.Background(), selectStmt, partition)
	if err != nil {
		return nil, fmt.Errorf("could not get nodes from DB: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		n := dao.NodeDAOInfo{}
		var id string
		err := rows.Scan(&id, &n.NodeID, &n.HostName, &n.RackName, &n.Attributes, &n.Capacity, &n.Allocated, &n.Occupied, &n.Available, &n.Utilized, &n.Allocations, &n.Schedulable, &n.IsReserved, &n.Reservations)
		if err != nil {
			return nil, fmt.Errorf("could not scan node: %v", err)
		}
		nodes = append(nodes, &n)
	}
	return nodes, nil
}

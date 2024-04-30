package repository

import (
	"context"
	"fmt"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *RepoPostgres) UpsertNodes(nodes []*dao.NodeDAOInfo) error {
	insertSQL := `INSERT INTO nodes (id, node_id, host_name, rack_name, attributes, capacity, allocated,
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
		_, err := s.dbpool.Exec(context.Background(), insertSQL,
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

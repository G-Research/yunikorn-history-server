package yunikorn

import (
	"context"
	"time"

	"github.com/G-Research/unicorn-history-server/internal/database/repository"
	"github.com/G-Research/unicorn-history-server/internal/model"
	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SyncNodesTestSuite struct {
	suite.Suite
	pool *pgxpool.Pool
	repo *repository.PostgresRepository
}

func (ss *SyncNodesTestSuite) SetupSuite() {
	require.NotNil(ss.T(), ss.pool)
	repo, err := repository.NewPostgresRepository(ss.pool)
	require.NoError(ss.T(), err)
	ss.repo = repo
}

func (ss *SyncNodesTestSuite) TearDownSuite() {
	ss.pool.Close()
}

func (ss *SyncNodesTestSuite) TestSyncNodes() {

	ctx := context.Background()
	nowNano := time.Now().UnixNano()
	partitionID := ulid.Make().String()

	tests := []struct {
		name          string
		stateNodes    []*dao.NodesDAOInfo
		existingNodes []*model.Node
		expectedNodes []*model.Node
		wantErr       bool
	}{
		{
			name: "Sync nodes with no existing nodes",
			stateNodes: []*dao.NodesDAOInfo{
				{
					PartitionName: "default",
					Nodes: []*dao.NodeDAOInfo{
						{ID: "1", NodeID: "node-1", PartitionID: partitionID, HostName: "host-1"},
						{ID: "2", NodeID: "node-2", PartitionID: partitionID, HostName: "host-2"},
					},
				},
			},
			existingNodes: nil,
			expectedNodes: []*model.Node{
				{NodeDAOInfo: dao.NodeDAOInfo{ID: "2", NodeID: "node-2", HostName: "host-2"}},
				{NodeDAOInfo: dao.NodeDAOInfo{ID: "1", NodeID: "node-1", HostName: "host-1"}},
			},
			wantErr: false,
		},
		{
			name: "Sync nodes with existing nodes in DB",
			stateNodes: []*dao.NodesDAOInfo{
				{
					PartitionName: "default",
					Nodes: []*dao.NodeDAOInfo{
						{ID: "2", NodeID: "node-2", PartitionID: partitionID, HostName: "host-2-updated"},
					},
				},
			},
			existingNodes: []*model.Node{
				{NodeDAOInfo: dao.NodeDAOInfo{ID: "1", NodeID: "node-1", HostName: "host-1", PartitionID: partitionID}},
				{NodeDAOInfo: dao.NodeDAOInfo{ID: "2", NodeID: "node-2", HostName: "host-2", PartitionID: partitionID}},
			},
			expectedNodes: []*model.Node{
				{NodeDAOInfo: dao.NodeDAOInfo{ID: "2", NodeID: "node-2", HostName: "host-2-updated", PartitionID: partitionID}}, // updated
				{
					Metadata:    model.Metadata{DeletedAtNano: &nowNano}, // deleted
					NodeDAOInfo: dao.NodeDAOInfo{ID: "1", NodeID: "node-1", HostName: "host-1", PartitionID: partitionID},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		ss.Run(tt.name, func() {
			// clean up the table after the test
			ss.T().Cleanup(func() {
				_, err := ss.pool.Exec(ctx, "DELETE FROM nodes")
				require.NoError(ss.T(), err)
			})

			for _, node := range tt.existingNodes {
				err := ss.repo.InsertNode(ctx, node)
				require.NoError(ss.T(), err)
			}

			s := NewService(ss.repo, nil, nil)

			err := s.syncNodes(ctx, tt.stateNodes)
			if tt.wantErr {
				require.Error(ss.T(), err)
				return
			}
			require.NoError(ss.T(), err)

			nodesInDB, err := s.repo.GetNodesPerPartition(ctx, partitionID, repository.NodeFilters{})
			require.NoError(ss.T(), err)
			for i, target := range tt.expectedNodes {
				require.Equal(ss.T(), target.ID, nodesInDB[i].ID)
				require.Equal(ss.T(), target.NodeID, nodesInDB[i].NodeID)
				require.Equal(ss.T(), target.HostName, nodesInDB[i].HostName)
				if target.DeletedAtNano != nil {
					require.NotNil(ss.T(), nodesInDB[i].DeletedAtNano)
				}
			}
		})
	}
}

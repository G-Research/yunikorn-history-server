package yunikorn

import (
	"context"
	"time"

	"github.com/G-Research/unicorn-history-server/internal/database/repository"
	"github.com/G-Research/unicorn-history-server/internal/model"
	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SyncApplicationsTestSuite struct {
	suite.Suite
	pool *pgxpool.Pool
	repo *repository.PostgresRepository
}

func (ss *SyncApplicationsTestSuite) SetupSuite() {
	require.NotNil(ss.T(), ss.pool)
	repo, err := repository.NewPostgresRepository(ss.pool)
	require.NoError(ss.T(), err)
	ss.repo = repo
}

func (ss *SyncApplicationsTestSuite) TearDownSuite() {
	ss.pool.Close()
}

func (ss *SyncApplicationsTestSuite) TestSyncApplications() {
	ctx := context.Background()
	now := time.Now().UnixNano()

	tests := []struct {
		name                 string
		stateApplications    []*dao.ApplicationDAOInfo
		existingApplications []*model.Application
		expectedLive         []*model.Application
		expectedDeleted      []*model.Application
		wantErr              bool
	}{
		{
			name: "Sync applications with no existing applications in DB",
			stateApplications: []*dao.ApplicationDAOInfo{
				{ID: "1", ApplicationID: "app-1"},
				{ID: "2", ApplicationID: "app-2"},
			},
			existingApplications: nil,
			expectedLive: []*model.Application{
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					ApplicationDAOInfo: dao.ApplicationDAOInfo{
						ID:            "1",
						ApplicationID: "app-1",
					},
				},
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					ApplicationDAOInfo: dao.ApplicationDAOInfo{
						ID:            "2",
						ApplicationID: "app-2",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Should mark application as deleted in DB",
			stateApplications: []*dao.ApplicationDAOInfo{
				{ID: "1", ApplicationID: "app-1"},
			},
			existingApplications: []*model.Application{
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					ApplicationDAOInfo: dao.ApplicationDAOInfo{
						ID:            "1",
						ApplicationID: "app-1",
					},
				},
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					ApplicationDAOInfo: dao.ApplicationDAOInfo{
						ID:            "2",
						ApplicationID: "app-2",
					},
				},
			},
			expectedLive: []*model.Application{
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					ApplicationDAOInfo: dao.ApplicationDAOInfo{
						ID:            "1",
						ApplicationID: "app-1",
					},
				},
			},
			expectedDeleted: []*model.Application{
				{
					Metadata: model.Metadata{
						CreatedAtNano: now,
					},
					ApplicationDAOInfo: dao.ApplicationDAOInfo{
						ID:            "2",
						ApplicationID: "app-2",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		ss.Run(tt.name, func() {
			// clean up the table after the test
			ss.T().Cleanup(func() {
				_, err := ss.pool.Exec(ctx, "DELETE FROM applications")
				require.NoError(ss.T(), err)
			})

			for _, app := range tt.existingApplications {
				err := ss.repo.InsertApplication(ctx, app)
				require.NoError(ss.T(), err)
			}

			s := NewService(ss.repo, nil, nil)

			err := s.syncApplications(ctx, tt.stateApplications)
			if tt.wantErr {
				require.Error(ss.T(), err)
				return
			}
			require.NoError(ss.T(), err)

			applicationsInDB, err := s.repo.GetAllApplications(
				ctx,
				repository.ApplicationFilters{},
			)
			require.NoError(ss.T(), err)

			require.Equal(ss.T(), len(tt.expectedLive)+len(tt.expectedDeleted), len(applicationsInDB))

			lookup := make(map[string]model.Application)
			for _, app := range applicationsInDB {
				lookup[app.ID] = *app
			}

			for _, target := range tt.expectedLive {
				state, ok := lookup[target.ID]
				require.True(ss.T(), ok)
				assert.NotEmpty(ss.T(), state.ID)
				assert.Greater(ss.T(), state.Metadata.CreatedAtNano, int64(0))
				assert.Nil(ss.T(), state.Metadata.DeletedAtNano)
			}

			for _, target := range tt.expectedDeleted {
				state, ok := lookup[target.ID]
				require.True(ss.T(), ok)
				assert.NotEmpty(ss.T(), state.ID)
				assert.Greater(ss.T(), state.Metadata.CreatedAtNano, int64(0))
				assert.NotNil(ss.T(), state.Metadata.DeletedAtNano)
			}
		})
	}
}

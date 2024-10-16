package yunikorn

import (
	"context"
	"errors"
	"testing"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/G-Research/yunikorn-history-server/internal/database/repository"
	"github.com/G-Research/yunikorn-history-server/internal/model"
)

func TestSync_findQueueDeleteCandidates(t *testing.T) {
	tests := []struct {
		name           string
		apiQueues      []*dao.PartitionQueueDAOInfo
		dbQueues       []*model.PartitionQueueDAOInfo
		expectedDelete []string
		expectedErr    error
	}{
		{
			name: "Single queue in DB not present in API",
			apiQueues: []*dao.PartitionQueueDAOInfo{
				{QueueName: "queue1"},
			},
			dbQueues: []*model.PartitionQueueDAOInfo{
				{PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{QueueName: "queue1"}},
				{PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{QueueName: "queue3"}},
			},
			expectedDelete: []string{"queue3"},
			expectedErr:    nil,
		},
		{
			name: "Multiple queues, no delete candidates",
			apiQueues: []*dao.PartitionQueueDAOInfo{
				{QueueName: "queue1"},
				{QueueName: "queue2"},
			},
			dbQueues: []*model.PartitionQueueDAOInfo{
				{PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{QueueName: "queue1"}},
				{PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{QueueName: "queue2"}},
			},
			expectedDelete: nil,
			expectedErr:    nil,
		},
		{
			name: "Multiple delete candidates in DB",
			apiQueues: []*dao.PartitionQueueDAOInfo{
				{QueueName: "queue1"},
			},
			dbQueues: []*model.PartitionQueueDAOInfo{
				{PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{QueueName: "queue1"}},
				{PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{QueueName: "queue2"}},
				{PartitionQueueDAOInfo: dao.PartitionQueueDAOInfo{QueueName: "queue3"}},
			},
			expectedDelete: []string{"queue2", "queue3"},
			expectedErr:    nil,
		},
		{
			name:           "No queues in API or DB",
			apiQueues:      []*dao.PartitionQueueDAOInfo{},
			dbQueues:       []*model.PartitionQueueDAOInfo{},
			expectedDelete: nil,
			expectedErr:    nil,
		},
		{
			name: "DB returns error",
			apiQueues: []*dao.PartitionQueueDAOInfo{
				{QueueName: "queue1"},
			},
			dbQueues:       nil, // Simulate an error from the DB
			expectedDelete: nil,
			expectedErr:    errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			partition := &dao.PartitionInfo{Name: "testPartition"}

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockRepo := repository.NewMockRepository(mockCtrl)

			if tt.expectedErr != nil {
				mockRepo.EXPECT().GetQueuesPerPartition(ctx, partition.Name).Return(nil, tt.expectedErr)
			} else {
				mockRepo.EXPECT().GetQueuesPerPartition(ctx, partition.Name).Return(tt.dbQueues, nil)
			}

			s := &Service{
				repo: mockRepo,
			}

			deleteCandidates, err := s.findQueueDeleteCandidates(ctx, partition, tt.apiQueues)

			if tt.expectedErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, len(tt.expectedDelete), len(deleteCandidates))

			// Check the QueueName of delete candidates
			for i, expectedQueueName := range tt.expectedDelete {
				require.Equal(t, expectedQueueName, deleteCandidates[i].QueueName)
			}
		})
	}
}

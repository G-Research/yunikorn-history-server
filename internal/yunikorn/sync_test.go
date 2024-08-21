package yunikorn

import (
	"testing"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/stretchr/testify/assert"
)

func TestFlattenQueues(t *testing.T) {
	queues := []*dao.PartitionQueueDAOInfo{
		{
			QueueName: "root",
			Partition: "default",
			Children: []dao.PartitionQueueDAOInfo{
				{
					QueueName: "org",
					Children: []dao.PartitionQueueDAOInfo{
						{
							QueueName: "eng",
							Children: []dao.PartitionQueueDAOInfo{
								{
									QueueName: "test",
								},
								{
									QueueName: "prod",
								},
							},
						},
						{
							QueueName: "sales",
							Children: []dao.PartitionQueueDAOInfo{
								{
									QueueName: "test",
								},
								{
									QueueName: "prod",
								},
							},
						},
					},
				},
				{
					QueueName: "system",
				},
			},
		},
	}
	flattenedQueues := flattenQueues(queues)

	assert.Equal(t, 9, len(flattenedQueues))
	for _, q := range flattenedQueues {
		assert.Equal(t, "default", q.Partition)
	}
}

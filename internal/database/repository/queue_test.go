package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/yunikorn-history-server/internal/model"
)

func TestGetChildrenFromMap(t *testing.T) {

	childrenMap := make(map[string][]*model.PartitionQueueDAOInfo)
	child1 := &model.PartitionQueueDAOInfo{ID: "child1"}
	child2 := &model.PartitionQueueDAOInfo{ID: "child2"}
	child3 := &model.PartitionQueueDAOInfo{ID: "child3"}

	childrenMap["parent1"] = []*model.PartitionQueueDAOInfo{child1, child2}
	childrenMap["parent2"] = []*model.PartitionQueueDAOInfo{child3}

	result := getChildrenFromMap("parent1", childrenMap)
	assert.Equal(t, 2, len(result))

	result = getChildrenFromMap("parent2", childrenMap)
	assert.Equal(t, 1, len(result))

	result = getChildrenFromMap("parent3", childrenMap)
	assert.Equal(t, 0, len(result)) // parent3 does not exist
}

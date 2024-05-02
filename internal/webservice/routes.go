package webservice

const (
	PARTITIONS                   = "/ws/v1/partitions"
	QUEUES_PER_PARTITION         = "/ws/v1/partition/:partition_name/queues"
	APPS_PER_PARTITION_PER_QUEUE = "/ws/v1/partition/:partition_name/queue/:queue_name/applications"
)

package webservice

const (
	CLUSTERS                     = "/ws/v1/clusters"
	PARTITIONS                   = "/ws/v1/partitions"
	QUEUES_PER_PARTITION         = "/ws/v1/partition/:partition_name/queues"
	APPS_PER_PARTITION_PER_QUEUE = "/ws/v1/partition/:partition_name/queue/:queue_name/applications"
	APPS_HISTORY                 = "/ws/v1/history/apps"
	CONTAINERS_HISTORY           = "/ws/v1/history/containers"
	NODES_PER_PARTITION          = "/ws/v1/partition/:partition_name/nodes"
	NODE_UTILIZATION             = "/ws/v1/scheduler/node-utilizations"
	SCHEDULER_HEALTHCHECK        = "/ws/v1/scheduler/healthcheck"
)

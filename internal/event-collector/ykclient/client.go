package ykclient

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"richscott/yhs/internal/event-collector/repository"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/apache/yunikorn-scheduler-interface/lib/go/si"
	"github.com/google/uuid"
)

var (
	streamEndPt           = "/ws/v1/events/stream"
	appHistoryEndPt       = "/ws/v1/history/apps"
	containerHistoryEndPt = "/ws/v1/history/containers"
	partitionsEndPt       = "/ws/v1/partitions"
	queuesEndPt           = func(partitionName string) string {
		return fmt.Sprintf("/ws/v1/partition/%s/queues", partitionName)
	}
	queueEndPt = func(partitionName, queueName string) string {
		return fmt.Sprintf("/ws/v1/partition/%s/queue/%s", partitionName, queueName)
	}
	nodesEndPt = "/ws/v1/nodes"
	nodeEndPt  = func(partitionName, nodeID string) string {
		return fmt.Sprintf("/ws/v1/partition/%s/node/%s", partitionName, nodeID)
	}
	applicationsEndPt = func(partitionName, queueName string) string {
		return fmt.Sprintf("/ws/v1/partition/%s/queue/%s/applications", partitionName, queueName)
	}
	applicationEndPt = func(partitionName, queueName, appID string) string {
		return fmt.Sprintf("/ws/v1/partition/%s/queue/%s/application/%s", partitionName, queueName, appID)
	}
)

type Client struct {
	httpProto string
	ykHost    string
	ykPort    int
	repo      *repository.RepoPostgres
}

func NewClient(httpProto string, ykHost string, ykPort int, repo *repository.RepoPostgres) *Client {
	return &Client{
		httpProto: httpProto,
		ykHost:    ykHost,
		ykPort:    ykPort,
		repo:      repo,
	}
}

func (c *Client) Run() error {
	err := c.loadUpCurrentClusterState()
	if err != nil {
		panic(err)
	}
	streamURL := c.endPointURL(streamEndPt)
	resp, err := http.Get(streamURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not request from %s: %v", streamURL, err)
		os.Exit(1)
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not read from http stream: %v", err)
			break
		}

		ev := si.EventRecord{}
		err = json.Unmarshal(line, &ev)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not unmarshal event from stream: %v\n", err)
			break
		}

		if ev.Type == si.EventRecord_APP {
			fmt.Printf("Application\n")
			fmt.Printf("---------\n")
			fmt.Printf("Type         : %s\n", si.EventRecord_Type_name[int32(ev.Type)])
			fmt.Printf("ObjectId     : %s\n", ev.ObjectID)
			fmt.Printf("Message      : %s\n", ev.Message)
			fmt.Printf("Change Type  : %s\n", ev.EventChangeType)
			fmt.Printf("Change Detail: %s\n", ev.EventChangeDetail)
			fmt.Printf("Reference ID:  %s\n", ev.ReferenceID)
		}
	}
	return nil
}

func (c *Client) loadUpCurrentClusterState() error {
	partitions, err := c.loadCurrentPartitions()
	_, err = c.loadCurrentQueues(partitions)
	// ... other resources
	return err
}

func (c *Client) loadCurrentPartitions() ([]*dao.PartitionInfo, error) {
	partitions := []*dao.PartitionInfo{}
	url := c.endPointURL(partitionsEndPt)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not get partitions from %s: %v", url, err)
	}
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("could not read partitions from response: %v", err)
		}
		if len(line) == 0 {
			break
		}
		pi := []*dao.PartitionInfo{}
		err = json.Unmarshal(line, &pi)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal partition from response: %v", err)
		}
		partitions = append(partitions, pi...)
	}
	// Insert partitions into DB
	for _, p := range partitions {
		// Insert partition into DB
		_, err := c.repo.Dbpool.Exec(context.Background(), "INSERT INTO partitions (id, cluster_id, name, capacity, used_capacity, utilization, total_nodes, applications, total_containers, state, last_state_transition_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
			uuid.New(), p.ClusterID, p.Name, p.Capacity.Capacity, p.Capacity.UsedCapacity, p.Capacity.Utilization, p.TotalNodes, p.Applications, p.TotalContainers, p.State, p.LastStateTransitionTime)
		if err != nil {
			return nil, fmt.Errorf("could not insert partition into DB: %v", err)
		}
	}
	return partitions, nil
}

func (c *Client) loadCurrentQueues(partitions []*dao.PartitionInfo) (map[string][]*dao.PartitionQueueDAOInfo, error) {
	queuesByPartition := map[string][]*dao.PartitionQueueDAOInfo{}
	for _, p := range partitions {
		url := c.endPointURL(queuesEndPt(p.Name))

		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("could not get partitions from %s: %v", url, err)
		}
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				return nil, fmt.Errorf("could not read queues from response: %v", err)
			}
			if len(line) == 0 {
				break
			}
			qi := dao.PartitionQueueDAOInfo{}
			err = json.Unmarshal(line, &qi)
			if err != nil {
				return nil, fmt.Errorf("could not unmarshal queue from response: %v", err)
			}
			queuesByPartition[p.Name] = append(queuesByPartition[p.Name], &qi)
		}
	}
	for _, queues := range queuesByPartition {
		for _, q := range queues {
			_, err := c.repo.Dbpool.Exec(context.Background(), "INSERT INTO queues (id, queue_name, status, partition, pending_resource, max_resource, guaranteed_resource, allocated_resource, preempting_resource, head_room, is_leaf, is_managed, properties, parent, template_info, children, children_names, abs_used_capacity, max_running_apps, running_apps, current_priority, allocating_accepted_apps) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)",
				uuid.New(), q.QueueName, q.Status, q.Partition, q.PendingResource, q.MaxResource, q.GuaranteedResource, q.AllocatedResource, q.PreemptingResource, q.HeadRoom, q.IsLeaf, q.IsManaged, q.Properties, q.Parent, q.TemplateInfo, q.Children, q.ChildrenNames, q.AbsUsedCapacity, q.MaxRunningApps, q.RunningApps, q.CurrentPriority, q.AllocatingAcceptedApps)
			if err != nil {
				return nil, fmt.Errorf("could not insert queue into DB: %v", err)
			}
		}

	}
	return queuesByPartition, nil
}

func (c *Client) endPointURL(endPt string) string {
	return fmt.Sprintf("%s://%s:%d%s", c.httpProto, c.ykHost, c.ykPort, endPt)
}

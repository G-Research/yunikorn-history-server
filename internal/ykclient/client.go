package ykclient

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"richscott/yhs/internal/repository"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/apache/yunikorn-scheduler-interface/lib/go/si"
)

var (
	streamEndPt = "/ws/v1/events/stream"
	// appHistoryEndPt       = "/ws/v1/history/apps"
	// containerHistoryEndPt = "/ws/v1/history/containers"
	partitionsEndPt     = "/ws/v1/partitions"
	partitionNodesEndPt = func(partitionName string) string {
		return fmt.Sprintf("/ws/v1/partition/%s/nodes", partitionName)
	}
	queuesEndPt = func(partitionName string) string {
		return fmt.Sprintf("/ws/v1/partition/%s/queues", partitionName)
	}
	// queueEndPt = func(partitionName, queueName string) string {
	// 	return fmt.Sprintf("/ws/v1/partition/%s/queue/%s", partitionName, queueName)
	// }
	// nodesEndPt = "/ws/v1/nodes"
	// nodeEndPt  = func(partitionName, nodeID string) string {
	// 	return fmt.Sprintf("/ws/v1/partition/%s/node/%s", partitionName, nodeID)
	// }
	applicationsEndPt = func(partitionName, queueName string) string {
		return fmt.Sprintf("/ws/v1/partition/%s/queue/%s.%s/applications", partitionName, queueName, partitionName)
	}
	// applicationEndPt = func(partitionName, queueName, appID string) string {
	// 	return fmt.Sprintf("/ws/v1/partition/%s/queue/%s/application/%s", partitionName, queueName, appID)
	// }
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
	if err != nil {
		return err
	}

	for _, part := range partitions {
		_, err := c.loadCurrentPartitionNodes(part.Name)
		if err != nil {
			return err
		}
	}

	partitionQueues, err := c.loadCurrentQueues(partitions)
	if err != nil {
		return err
	}

	for _, q := range partitionQueues {
		fmt.Printf("loading applications for partition %s, queue %s\n", q.Partition, q.QueueName)
		_, err := c.loadCurrentApplications(q.Partition, q.QueueName)
		if err != nil {
			return err
		}
	}

	return nil
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
	if err = c.repo.UpsertPartitions(partitions); err != nil {
		return nil, err
	}
	return partitions, nil
}

func (c *Client) loadCurrentQueues(partitions []*dao.PartitionInfo) ([]*dao.PartitionQueueDAOInfo, error) {
	queues := []*dao.PartitionQueueDAOInfo{}
	for _, p := range partitions {
		url := c.endPointURL(queuesEndPt(p.Name))

		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("could not get queues from %s: %v", url, err)
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
			queues = append(queues, &qi)
		}
	}
	if err := c.repo.UpsertQueues(queues); err != nil {
		return nil, err
	}
	return queues, nil
}

func (c *Client) loadCurrentApplications(partitionName, queueName string) ([]*dao.ApplicationDAOInfo, error) {
	apps := []*dao.ApplicationDAOInfo{}
	url := c.endPointURL(applicationsEndPt(partitionName, queueName))

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not get applications from %s: %v", url, err)
	}
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("could not read applications from response: %v", err)
		}
		if len(line) == 0 {
			break
		}
		responseApps := []dao.ApplicationDAOInfo{}
		err = json.Unmarshal(line, &responseApps)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal application from response: %v", err)
		}

		for _, a := range responseApps {
			apps = append(apps, &a)
		}
	}
	if err := c.repo.UpsertApplications(apps); err != nil {
		return nil, err
	}
	return apps, nil
}

func (c *Client) loadCurrentPartitionNodes(partitionName string) ([]*dao.NodeDAOInfo, error) {
	nodes := []*dao.NodeDAOInfo{}
	url := c.endPointURL(partitionNodesEndPt(partitionName))

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not get partition nodes from %s: %v", url, err)
	}
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("could not read partition nodes from response: %v", err)
		}
		if len(line) == 0 {
			break
		}
		responseNodes := []dao.NodeDAOInfo{}
		err = json.Unmarshal(line, &responseNodes)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal partition nodes from response: %v", err)
		}

		for _, n := range responseNodes {
			nodes = append(nodes, &n)
		}
	}
	if err := c.repo.UpsertNodes(nodes); err != nil {
		return nil, err
	}
	return nodes, nil
}

func (c *Client) endPointURL(endPt string) string {
	return fmt.Sprintf("%s://%s:%d%s", c.httpProto, c.ykHost, c.ykPort, endPt)
}

package ykclient

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/G-Research/yunikorn-history-server/internal/repository"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/apache/yunikorn-scheduler-interface/lib/go/si"
	"github.com/google/uuid"
)

var (
	streamEndPt = "/ws/v1/events/stream"
	// appHistoryEndPt       = "/ws/v1/history/apps"
	// containerHistoryEndPt = "/ws/v1/history/containers"
	partitionsEndPt        = "/ws/v1/partitions"
	appsHistoryEndPt       = "/ws/v1/history/apps"
	containersHistoryEndPt = "/ws/v1/history/containers"
	partitionNodesEndPt    = func(partitionName string) string {
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
	nodeUtilEndPt     = "/ws/v1/scheduler/node-utilizations"
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

func (c *Client) Run(ctx context.Context) {
	err := c.loadUpCurrentClusterState(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not load current cluster state: %v\n", err)
	}
	streamURL := c.endPointURL(streamEndPt)
	resp, err := http.Get(streamURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not request from %s: %v", streamURL, err)
	}

	reader := bufio.NewReader(resp.Body)
	go func() {
		fmt.Println("Starting YuniKorn event stream client")
		for {
			select {
			case <-ctx.Done():
				// TODO: add logging here to indicate that the client is shutting down
				return
			default:
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
		}
	}()
}

func (c *Client) loadUpCurrentClusterState(ctx context.Context) error {
	partitions, err := c.loadCurrentPartitions(ctx)
	if err != nil {
		return err
	}

	for _, part := range partitions {
		_, err := c.loadCurrentPartitionNodes(ctx, part.Name)
		if err != nil {
			return err
		}
	}

	partitionQueues, err := c.loadCurrentQueues(ctx, partitions)
	if err != nil {
		return err
	}

	for _, q := range partitionQueues {
		fmt.Printf("loading applications for partition %s, queue %s\n", q.Partition, q.QueueName)
		_, err := c.loadCurrentApplications(ctx, q.Partition, q.QueueName)
		if err != nil {
			return err
		}
	}

	_, err = c.loadCurrentNodeUtil(ctx)
	if err != nil {
		return err
	}
	err = c.loadClusterHistory(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) loadCurrentPartitions(ctx context.Context) ([]*dao.PartitionInfo, error) {
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
	if err = c.repo.UpsertPartitions(ctx, partitions); err != nil {
		return nil, err
	}
	return partitions, nil
}

func (c *Client) loadCurrentQueues(ctx context.Context, partitions []*dao.PartitionInfo) ([]*dao.PartitionQueueDAOInfo, error) {
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
	if err := c.repo.UpsertQueues(ctx, queues); err != nil {
		return nil, err
	}
	return queues, nil
}

func (c *Client) loadCurrentApplications(ctx context.Context, partitionName, queueName string) ([]*dao.ApplicationDAOInfo, error) {
	apps := []*dao.ApplicationDAOInfo{}
	url := c.endPointURL(applicationsEndPt(partitionName, queueName))

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not get applications from %s: %v", url, err)
	}

	if resp.StatusCode != 200 {
		errBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("could not read body of response error: %v", err)
		}

		return nil, fmt.Errorf("could not get applications from %s: %s", url, string(errBody))
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
	if err := c.repo.UpsertApplications(ctx, apps); err != nil {
		return nil, err
	}
	return apps, nil
}

func (c *Client) loadCurrentPartitionNodes(ctx context.Context, partitionName string) ([]*dao.NodeDAOInfo, error) {
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
	if err := c.repo.UpsertNodes(ctx, nodes, partitionName); err != nil {
		return nil, err
	}
	return nodes, nil
}

func (c *Client) loadCurrentNodeUtil(ctx context.Context) (*[]dao.PartitionNodesUtilDAOInfo, error) {
	url := c.endPointURL(nodeUtilEndPt)
	nus := []dao.PartitionNodesUtilDAOInfo{}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not get node utilizations from %s: %v", url, err)
	}
	reader := bufio.NewReader(resp.Body)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("could not read node utilizations from response: %v", err)
	}
	if len(line) == 0 {
		return nil, fmt.Errorf("could not read node utilizations from response: %v", err)
	}
	err = json.Unmarshal(line, &nus)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal node utilizations from response: %v", err)
	}

	if err := c.repo.InsertNodeUtilizations(ctx, uuid.New(), &nus); err != nil {
		return nil, err
	}

	return &nus, nil
}

func (c *Client) loadClusterHistory(ctx context.Context) error {
	appsHistory, err := c.loadAppsHistory()
	if err != nil {
		return err
	}
	containersHistory, err := c.loadContainersHistory()
	if err != nil {
		return err
	}
	return c.repo.UpdateHistory(ctx, appsHistory, containersHistory)
}

func (c *Client) loadAppsHistory() ([]*dao.ApplicationHistoryDAOInfo, error) {
	appsHistory := []*dao.ApplicationHistoryDAOInfo{}
	url := c.endPointURL(appsHistoryEndPt)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not get applications history from %s: %v", url, err)
	}
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("could not read applications history from response: %v", err)
		}
		if len(line) == 0 {
			break
		}
		responseAppsHistory := []dao.ApplicationHistoryDAOInfo{}
		err = json.Unmarshal(line, &responseAppsHistory)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal applications history from response: %v", err)
		}

		for _, a := range responseAppsHistory {
			appsHistory = append(appsHistory, &a)
		}
	}
	return appsHistory, nil
}

func (c *Client) loadContainersHistory() ([]*dao.ContainerHistoryDAOInfo, error) {
	containersHistory := []*dao.ContainerHistoryDAOInfo{}
	url := c.endPointURL(containersHistoryEndPt)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not get containers history from %s: %v", url, err)
	}
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("could not read containers history from response: %v", err)
		}
		if len(line) == 0 {
			break
		}
		responseContainersHistory := []dao.ContainerHistoryDAOInfo{}
		err = json.Unmarshal(line, &responseContainersHistory)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal containers history from response: %v", err)
		}

		for _, c := range responseContainersHistory {
			containersHistory = append(containersHistory, &c)
		}
	}
	return containersHistory, nil
}

func (c *Client) endPointURL(endPt string) string {
	return fmt.Sprintf("%s://%s:%d%s", c.httpProto, c.ykHost, c.ykPort, endPt)
}

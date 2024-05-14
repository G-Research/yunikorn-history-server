package ykclient

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
)

var (
	streamEndPt            = "/ws/v1/events/stream"
	partitionsEndPt        = "/ws/v1/partitions"
	appsHistoryEndPt       = "/ws/v1/history/apps"
	containersHistoryEndPt = "/ws/v1/history/containers"
	partitionNodesEndPt    = func(partitionName string) string {
		return fmt.Sprintf("/ws/v1/partition/%s/nodes", partitionName)
	}
	queuesEndPt = func(partitionName string) string {
		return fmt.Sprintf("/ws/v1/partition/%s/queues", partitionName)
	}
	nodeUtilEndPt           = "/ws/v1/scheduler/node-utilizations"
	applicationsByPartEndPt = func(partitionName string) string {
		return fmt.Sprintf("/ws/v1/partition/%s/applications/active", partitionName)
	}
	applicationsByQueueEndPt = func(partitionName, queueName string) string {
		return fmt.Sprintf("/ws/v1/partition/%s/queue/%s/applications", partitionName, queueName)
	}
	applicationByPartEndPt = func(partitionName, appID string) string {
		return fmt.Sprintf("/ws/v1/partition/%s/application/%s", partitionName, appID)
	}
	applicationByQueueEndPt = func(partitionName, queueName, appID string) string {
		return fmt.Sprintf("/ws/v1/partition/%s/queue/%s/application/%s", partitionName, queueName, appID)
	}
)

func (c *Client) GetPartitions(ctx context.Context) ([]*dao.PartitionInfo, error) {
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
	return partitions, nil
}

func (c *Client) GetPartitionQueues(ctx context.Context, partitionName string) ([]dao.PartitionQueueDAOInfo, error) {
	queues := []dao.PartitionQueueDAOInfo{}
	url := c.endPointURL(queuesEndPt(partitionName))

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
		queues = append(queues, qi)
	}
	return queues, nil
}

func (c *Client) GetApplications(ctx context.Context, partitionName, queueName string) ([]*dao.ApplicationDAOInfo, error) {
	var url string
	if partitionName == "" {
		partitionName = "default"
	}
	if queueName == "" {
		url = c.endPointURL(applicationsByPartEndPt(partitionName))
	} else {
		url = c.endPointURL(applicationsByQueueEndPt(partitionName, queueName))
	}
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
	apps := []*dao.ApplicationDAOInfo{}

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

	return apps, nil
}

func (c *Client) GetApplication(ctx context.Context, partitionName, queueName, appID string) (*dao.ApplicationDAOInfo, error) {
	var url string
	if partitionName == "" {
		partitionName = "default"
	}
	if queueName == "" {
		url = c.endPointURL(applicationByPartEndPt(partitionName, appID))
	} else {
		url = c.endPointURL(applicationByQueueEndPt(partitionName, queueName, appID))
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not get application from %s: %v", url, err)
	}

	if resp.StatusCode != 200 {
		errBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("could not read body of response error: %v", err)
		}
		return nil, fmt.Errorf("could not get application from %s: %s", url, string(errBody))
	}

	var app dao.ApplicationDAOInfo
	err = json.NewDecoder(resp.Body).Decode(&app)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal application from response: %v", err)
	}

	return &app, nil
}

func (c *Client) GetPartitionNodes(ctx context.Context, partitionName string) ([]*dao.NodeDAOInfo, error) {
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
		responseNodes := []*dao.NodeDAOInfo{}
		err = json.Unmarshal(line, &responseNodes)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal partition nodes from response: %v", err)
		}

		nodes = append(nodes, responseNodes...)
	}
	return nodes, nil
}

func (c *Client) GetNodeUtil(ctx context.Context) (*[]dao.PartitionNodesUtilDAOInfo, error) {
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
	return &nus, nil
}

func (c *Client) GetAppsHistory(ctx context.Context) ([]*dao.ApplicationHistoryDAOInfo, error) {
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
		responseAppsHistory := []*dao.ApplicationHistoryDAOInfo{}
		err = json.Unmarshal(line, &responseAppsHistory)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal applications history from response: %v", err)
		}
		appsHistory = append(appsHistory, responseAppsHistory...)
	}
	return appsHistory, nil
}

func (c *Client) GetContainersHistory(ctx context.Context) ([]*dao.ContainerHistoryDAOInfo, error) {
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

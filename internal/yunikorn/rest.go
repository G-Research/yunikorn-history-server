package yunikorn

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/G-Research/yunikorn-history-server/internal/config"

	"github.com/G-Research/yunikorn-history-server/internal/log"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
)

const (
	endpointStream            = "/ws/v1/events/stream"
	endpointPartitions        = "/ws/v1/partitions"
	endpointAppsHistory       = "/ws/v1/history/apps"
	endpointContainersHistory = "/ws/v1/history/containers"
	endpointNodeUtil          = "/ws/v1/scheduler/node-utilizations"
	endpointHealthcheck       = "/ws/v1/scheduler/healthcheck"
)

// RESTClient implements the Client interface which defines functions to interact with the Yunikorn REST API
type RESTClient struct {
	protocol string
	host     string
	port     int
}

func NewRESTClient(cfg *config.YunikornConfig) *RESTClient {
	protocol := "http"
	if cfg.Secure {
		protocol = "https"
	}
	return &RESTClient{
		protocol: protocol,
		host:     cfg.Host,
		port:     cfg.Port,
	}
}

func (c *RESTClient) GetPartitions(ctx context.Context) ([]*dao.PartitionInfo, error) {
	resp, err := c.get(ctx, endpointPartitions)
	if err != nil {
		return nil, err
	}
	defer closeBody(ctx, resp)

	if resp.StatusCode != 200 {
		return nil, handleNonOKResponse(ctx, resp)
	}

	var partitions []*dao.PartitionInfo
	if err = unmarshallBody(ctx, resp, &partitions); err != nil {
		return nil, err
	}

	return partitions, nil
}

func (c *RESTClient) GetPartitionQueues(ctx context.Context, partitionName string) (*dao.PartitionQueueDAOInfo, error) {
	resp, err := c.get(ctx, endpointQueues(partitionName))
	if err != nil {
		return nil, err
	}
	defer closeBody(ctx, resp)

	if resp.StatusCode != 200 {
		return nil, handleNonOKResponse(ctx, resp)
	}

	var queues *dao.PartitionQueueDAOInfo
	if err = unmarshallBody(ctx, resp, &queues); err != nil {
		return nil, err
	}

	return queues, nil
}

func (c *RESTClient) GetPartitionQueue(ctx context.Context, partitionName, queueName string) (*dao.PartitionQueueDAOInfo, error) {
	resp, err := c.get(ctx, endpointQueue(partitionName, queueName))
	if err != nil {
		return nil, err
	}
	defer closeBody(ctx, resp)

	if resp.StatusCode != 200 {
		return nil, handleNonOKResponse(ctx, resp)
	}

	var queues *dao.PartitionQueueDAOInfo
	if err = unmarshallBody(ctx, resp, &queues); err != nil {
		return nil, err
	}

	return queues, nil
}

func (c *RESTClient) GetApplications(ctx context.Context, partitionName, queueName string) (
	[]*dao.ApplicationDAOInfo, error,
) {
	if partitionName == "" {
		partitionName = "default"
	}
	var endpoint string
	if queueName == "" {
		endpoint = endpointApplicationsByPartition(partitionName)
	} else {
		endpoint = endpointApplicationsByQueue(partitionName, queueName)
	}

	resp, err := c.get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer closeBody(ctx, resp)

	if resp.StatusCode != 200 {
		return nil, handleNonOKResponse(ctx, resp)
	}

	var apps []*dao.ApplicationDAOInfo
	if err = unmarshallBody(ctx, resp, &apps); err != nil {
		return nil, err
	}

	return apps, nil
}

// GetApplication calls the Yunikorn Scheduler API to get the application information for the given
// partition, queue and appID. If partitionName is empty, it defaults to "default".
// If queueName is empty, it gets the application from the partition level
func (c *RESTClient) GetApplication(
	ctx context.Context,
	partitionName, queueName, appID string,
) (*dao.ApplicationDAOInfo, error) {
	if partitionName == "" {
		partitionName = "default"
	}
	endpoint := endpointApplicationByPartition(partitionName, appID)
	if queueName != "" {
		endpoint = endpointApplicationByQueue(partitionName, queueName, appID)
	}

	resp, err := c.get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer closeBody(ctx, resp)

	if resp.StatusCode != 200 {
		return nil, handleNonOKResponse(ctx, resp)
	}

	var app dao.ApplicationDAOInfo
	if err = unmarshallBody(ctx, resp, &app); err != nil {
		return nil, err
	}

	return &app, nil
}

func (c *RESTClient) GetPartitionNodes(ctx context.Context, partitionName string) ([]*dao.NodeDAOInfo, error) {
	resp, err := c.get(ctx, endpointPartitionNodes(partitionName))
	if err != nil {
		return nil, err
	}
	defer closeBody(ctx, resp)

	if resp.StatusCode != 200 {
		return nil, handleNonOKResponse(ctx, resp)
	}

	var nodes []*dao.NodeDAOInfo
	if err = unmarshallBody(ctx, resp, &nodes); err != nil {
		return nil, err
	}

	return nodes, nil
}

func (c *RESTClient) GetNodeUtil(ctx context.Context) ([]*dao.PartitionNodesUtilDAOInfo, error) {
	resp, err := c.get(ctx, endpointNodeUtil)
	if err != nil {
		return nil, err
	}
	defer closeBody(ctx, resp)

	if resp.StatusCode != 200 {
		return nil, handleNonOKResponse(ctx, resp)
	}

	var nus []*dao.PartitionNodesUtilDAOInfo
	if err = unmarshallBody(ctx, resp, &nus); err != nil {
		return nil, err
	}

	return nus, nil
}

func (c *RESTClient) GetAppsHistory(ctx context.Context) ([]*dao.ApplicationHistoryDAOInfo, error) {
	resp, err := c.get(ctx, endpointAppsHistory)
	if err != nil {
		return nil, err
	}
	defer closeBody(ctx, resp)

	if resp.StatusCode != 200 {
		return nil, handleNonOKResponse(ctx, resp)
	}

	var appsHistory []*dao.ApplicationHistoryDAOInfo
	if err = unmarshallBody(ctx, resp, &appsHistory); err != nil {
		return nil, err
	}

	return appsHistory, nil
}

func (c *RESTClient) GetContainersHistory(ctx context.Context) ([]*dao.ContainerHistoryDAOInfo, error) {
	resp, err := c.get(ctx, endpointContainersHistory)
	if err != nil {
		return nil, err
	}
	defer closeBody(ctx, resp)

	if resp.StatusCode != 200 {
		return nil, handleNonOKResponse(ctx, resp)
	}

	var containersHistory []*dao.ContainerHistoryDAOInfo
	if err = unmarshallBody(ctx, resp, &containersHistory); err != nil {
		return nil, err
	}

	return containersHistory, nil
}

func (c *RESTClient) Healthcheck(ctx context.Context) (*dao.SchedulerHealthDAOInfo, error) {
	resp, err := c.get(ctx, endpointHealthcheck)
	if err != nil {
		return nil, err
	}
	defer closeBody(ctx, resp)

	if resp.StatusCode != 200 {
		return nil, handleNonOKResponse(ctx, resp)
	}

	var schedulerHealth dao.SchedulerHealthDAOInfo
	if err = unmarshallBody(ctx, resp, &schedulerHealth); err != nil {
		return nil, err
	}

	return &schedulerHealth, nil
}

func (c *RESTClient) GetEventStream(ctx context.Context) (*http.Response, error) {
	resp, err := c.get(ctx, endpointStream)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, handleNonOKResponse(ctx, resp)
	}

	return resp, nil
}

// get makes a GET request to the given URL and returns the response
func (c *RESTClient) get(ctx context.Context, endpoint string) (*http.Response, error) {
	url := c.url(endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *RESTClient) url(endpoint string) string {
	return fmt.Sprintf("%s://%s:%d%s", c.protocol, c.host, c.port, endpoint)
}

func handleNonOKResponse(ctx context.Context, resp *http.Response) error {
	logger := log.FromContext(ctx)

	errBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	logger.Errorw(
		"yunikorn api returned non-OK status code",
		"endpoint", resp.Request.URL.Path, "statusCode", resp.StatusCode, "body", string(errBody),
	)
	return fmt.Errorf("yunicorn api returned non-OK status code: %d", resp.StatusCode)
}

func unmarshallBody(ctx context.Context, resp *http.Response, v any) error {
	logger := log.FromContext(ctx)

	err := json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		logger.Errorw(
			"error unmarshalling response body",
			"endpoint", resp.Request.URL.Path, "statusCode", resp.StatusCode,
		)
		return err
	}
	return nil
}

func closeBody(ctx context.Context, resp *http.Response) {
	logger := log.FromContext(ctx)

	err := resp.Body.Close()
	if err != nil {
		logger.Errorw(
			"could not close response body",
			"endpoint", resp.Request.URL.Path, "statusCode", resp.StatusCode,
		)
	}
}

var _ Client = &RESTClient{}

func endpointPartitionNodes(partitionName string) string {
	return fmt.Sprintf("/ws/v1/partition/%s/nodes", partitionName)
}

func endpointQueues(partitionName string) string {
	return fmt.Sprintf("/ws/v1/partition/%s/queues", partitionName)
}

func endpointQueue(partitionName, queueName string) string {
	return fmt.Sprintf("/ws/v1/partition/%s/queue/%s", partitionName, queueName)
}

func endpointApplicationsByPartition(partitionName string) string {
	return fmt.Sprintf("/ws/v1/partition/%s/applications/active", partitionName)
}

func endpointApplicationsByQueue(partitionName, queueName string) string {
	return fmt.Sprintf("/ws/v1/partition/%s/queue/%s/applications", partitionName, queueName)
}

func endpointApplicationByPartition(partitionName, appID string) string {
	return fmt.Sprintf("/ws/v1/partition/%s/application/%s", partitionName, appID)
}

func endpointApplicationByQueue(partitionName, queueName, appID string) string {
	return fmt.Sprintf("/ws/v1/partition/%s/queue/%s/application/%s", partitionName, queueName, appID)
}

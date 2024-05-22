package ykclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/G-Research/yunikorn-history-server/internal/config"
	"github.com/G-Research/yunikorn-history-server/internal/repository"
	"github.com/google/uuid"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/apache/yunikorn-scheduler-interface/lib/go/si"
)

type Client struct {
	httpProto string
	ykHost    string
	ykPort    int
	repo      *repository.RepoPostgres
	appMap    map[string]*dao.ApplicationDAOInfo
}

func NewClient(httpProto string, ykHost string, ykPort int, repo *repository.RepoPostgres) *Client {
	return &Client{
		httpProto: httpProto,
		ykHost:    ykHost,
		ykPort:    ykPort,
		repo:      repo,
		appMap:    make(map[string]*dao.ApplicationDAOInfo),
	}
}

func (c *Client) Run(ctx context.Context) {
	go c.startup(ctx)
	streamURL := c.endPointURL(streamEndPt)

	evCounts := ctx.Value(config.EventCounts).(config.EventTypeCounts)
	if evCounts == nil {
		fmt.Fprintf(os.Stderr, "could not get eventCounts map from context\n")
		return
	}

	go func() {
		fmt.Println("Starting YuniKorn event stream client")
		c.FetchEventStream(ctx, streamURL, evCounts)
	}()
}

func (c *Client) FetchEventStream(ctx context.Context, streamURL string, evCounts config.EventTypeCounts) {
	ctx, cancel := context.WithCancel(ctx)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, streamURL, bytes.NewBuffer([]byte{}))
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not request from %s: %v", streamURL, err)
		return
	}
	defer func() {
		resp.Body.Close()
	}()

	reader := bufio.NewReader(resp.Body)

	for {
		select {
		case <-ctx.Done():
			// TODO: add logging here to indicate that the client is shutting down
			cancel()
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
			// TODO: This is Okayish for small number of events, but for large number of events this will be a bottleneck
			// We should consider using a channel? or a pool of workers? or a different queuing system ? to handle events.
			c.handleEvent(ctx, &ev)

			evKey := config.EventTypeKey{Type: ev.Type, ChangeType: ev.EventChangeType}
			if count, exists := evCounts[evKey]; exists {
				evCounts[evKey] = count + 1
			} else {
				evCounts[evKey] = 1
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
				fmt.Printf("Resource    : %+v\n", ev.Resource)
			}
		}
	}
}

// startup performs all necessary steps to load up the current state of the cluster
func (c *Client) startup(ctx context.Context) {
	// Get partitions from YuniKorn API and upsert into DB
	partitions, err := c.GetPartitions(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get partitions: %v\n", err)
	}
	if err = c.repo.UpsertPartitions(ctx, partitions); err != nil {
		fmt.Fprintf(os.Stderr, "could not upsert partitions: %v\n", err)
	}

	// Get partition queues from YuniKorn API and upsert into DB
	queues := []dao.PartitionQueueDAOInfo{}
	for _, part := range partitions {
		qs, err := c.GetPartitionQueues(ctx, part.Name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not get queues for partition %s: %v\n", part.Name, err)
		} else if len(qs) > 0 {
			queues = append(queues, qs...)
		}
	}
	queues = flattenQueues(queues)
	if err = c.repo.UpsertQueues(ctx, queues); err != nil {
		fmt.Fprintf(os.Stderr, "could not upsert queues: %v\n", err)
	}

	// Get partition nodes from YuniKorn API and upsert into DB
	for _, part := range partitions {
		nodes, err := c.GetPartitionNodes(ctx, part.Name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not get nodes for partition %s: %v\n", part.Name, err)
		} else {
			if err = c.repo.UpsertNodes(ctx, nodes, part.Name); err != nil {
				fmt.Fprintf(os.Stderr, "could not upsert nodes: %v\n", err)
			}
		}
	}

	// Get applications from YuniKorn API and upsert into DB
	apps := []*dao.ApplicationDAOInfo{}
	for _, q := range queues {
		queueApps, err := c.GetApplications(ctx, q.Partition, q.QueueName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not get applications for partition %s, queue %s: %v\n", q.Partition, q.QueueName, err)
		} else {
			apps = append(apps, queueApps...)
		}
	}
	if err = c.repo.UpsertApplications(ctx, apps); err != nil {
		fmt.Fprintf(os.Stderr, "could not upsert applications: %v\n", err)
	}

	// Get node utilizations from YuniKorn API and insert into DB
	nus, err := c.GetNodeUtil(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get node utilizations: %v\n", err)
	}
	if err := c.repo.InsertNodeUtilizations(ctx, uuid.New(), nus); err != nil {
		fmt.Fprintf(os.Stderr, "could not insert node utilizations: %v\n", err)
	}

	// Get apps history and containers history from YuniKorn API and update history in DB
	appsHistory, err := c.GetAppsHistory(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get apps history: %v\n", err)
	}
	containersHistory, err := c.GetContainersHistory(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get containers history: %v\n", err)
	}
	if err = c.repo.UpdateHistory(ctx, appsHistory, containersHistory); err != nil {
		fmt.Fprintf(os.Stderr, "could not update history: %v\n", err)
	}
}

// Usually the returned queues are a signle heirarchical structure, the root queue,
// and all other queues are children queues.
// flattenQueues returns a list of all queues in the hierarchy in a flat array
func flattenQueues(qs []dao.PartitionQueueDAOInfo) []dao.PartitionQueueDAOInfo {
	var queues []dao.PartitionQueueDAOInfo
	for _, q := range qs {
		queues = append(queues, q)
		if len(q.Children) > 0 {
			queues = append(queues, flattenQueues(q.Children)...)
		}
	}
	return queues
}

func (c *Client) endPointURL(endPt string) string {
	return fmt.Sprintf("%s://%s:%d%s", c.httpProto, c.ykHost, c.ykPort, endPt)
}

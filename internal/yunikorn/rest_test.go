package yunikorn

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"testing"

	testconfig "github.com/G-Research/yunikorn-history-server/test/config"

	"github.com/G-Research/yunikorn-history-server/internal/config"

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRESTClient_GetPartitions_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}

	client := NewRESTClient(testconfig.GetTestYunikornConfig())

	partitions, err := client.GetPartitions(context.Background())
	if err != nil {
		t.Fatalf("error getting partitions: %v", err)
	}

	fmt.Println(partitions)
}

func TestRESTClient_GetApplication(t *testing.T) {
	tests := []struct {
		name      string
		partName  string
		queueName string
		appId     string
		expected  dao.ApplicationDAOInfo
	}{
		{
			name:      "All Fields Specified",
			partName:  "specialPartition",
			queueName: "myQueue",
			appId:     "app1",
			expected: dao.ApplicationDAOInfo{
				ApplicationID: "app1",
				Partition:     "specialPartition",
				QueueName:     "myQueue",
			},
		},
		{
			name:      "Empty Partition Name",
			partName:  "",
			queueName: "default",
			appId:     "app1",
			expected: dao.ApplicationDAOInfo{
				ApplicationID: "app1",
				Partition:     "default",
				QueueName:     "default",
			},
		},
		{
			name:      "Empty Queue Name",
			partName:  "myPartition",
			queueName: "",
			appId:     "app1",
			expected: dao.ApplicationDAOInfo{
				ApplicationID: "app1",
				Partition:     "myPartition",
				QueueName:     "myPartition.default",
			},
		},
	}

	partQueueAppRe := regexp.MustCompile(`/ws/v1/partition/(\w+)/queue/(\w+)/application/(\w+)`)
	partAppRe := regexp.MustCompile(`/ws/v1/partition/(\w+)/application/(\w+)`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				responseApp := dao.ApplicationDAOInfo{}

				matches := partQueueAppRe.FindStringSubmatch(r.URL.Path)
				if matches != nil {
					responseApp = dao.ApplicationDAOInfo{
						Partition:     matches[1],
						QueueName:     matches[2],
						ApplicationID: matches[3],
					}
				} else {
					matches = partAppRe.FindStringSubmatch(r.URL.Path)
					if matches == nil {
						http.Error(w, errors.New("invalid request path").Error(), http.StatusNotFound)
						return
					}

					responseApp = dao.ApplicationDAOInfo{
						Partition:     matches[1],
						QueueName:     fmt.Sprintf("%s.%s", matches[1], "default"),
						ApplicationID: matches[2],
					}
				}
				err := json.NewEncoder(w).Encode(&responseApp)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}

			}))

			defer ts.Close()
			client := NewRESTClient(getMockServerYunikornConfig(t, ts.URL))

			app, err := client.GetApplication(context.Background(), tt.partName, tt.queueName, tt.appId)
			require.NoError(t, err)
			assert.NotNil(t, app)
			assert.Equal(t, tt.expected.Partition, app.Partition)
			assert.Equal(t, tt.expected.QueueName, app.QueueName)
			assert.Equal(t, tt.expected.ApplicationID, app.ApplicationID)
		})
	}
}

func TestRESTClient_GetApplications(t *testing.T) {
	tests := []struct {
		name      string
		partName  string
		queueName string
		expected  []*dao.ApplicationDAOInfo
	}{
		{
			name:      "All Fields Specified",
			partName:  "specialPartition",
			queueName: "myQueue",
			expected: []*dao.ApplicationDAOInfo{
				{
					ApplicationID: "app1",
					Partition:     "specialPartition",
					QueueName:     "myQueue",
				},
			},
		},
		{
			name:      "Empty Partition Name",
			partName:  "",
			queueName: "myQueue",
			expected: []*dao.ApplicationDAOInfo{
				{
					ApplicationID: "app1",
					Partition:     "default",
					QueueName:     "myQueue",
				},
			},
		},
		{
			name:      "Empty Queue Name",
			partName:  "myPartition",
			queueName: "",
			expected: []*dao.ApplicationDAOInfo{
				{
					ApplicationID: "app1",
					Partition:     "myPartition",
					QueueName:     "myPartition.default",
				},
			},
		},
	}

	partQueueAppRe := regexp.MustCompile(`/ws/v1/partition/(\w+)/queue/(\w+)/applications`)
	partAppRe := regexp.MustCompile(`/ws/v1/partition/(\w+)/applications/active`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var responseApp []*dao.ApplicationDAOInfo

				matches := partQueueAppRe.FindStringSubmatch(r.URL.Path)
				if matches != nil {
					responseApp = []*dao.ApplicationDAOInfo{
						{
							Partition:     matches[1],
							QueueName:     matches[2],
							ApplicationID: "app1",
						},
					}
				} else {
					matches = partAppRe.FindStringSubmatch(r.URL.Path)
					if matches == nil {
						http.Error(w, errors.New("invalid request path").Error(), http.StatusNotFound)
						return
					}

					responseApp = []*dao.ApplicationDAOInfo{
						{
							Partition:     matches[1],
							QueueName:     fmt.Sprintf("%s.%s", matches[1], "default"),
							ApplicationID: "app1",
						},
					}

				}
				err := json.NewEncoder(w).Encode(&responseApp)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}

			}))

			defer ts.Close()
			client := NewRESTClient(getMockServerYunikornConfig(t, ts.URL))

			apps, err := client.GetApplications(context.Background(), tt.partName, tt.queueName)
			require.NoError(t, err)
			assert.NotNil(t, apps)
			for i := range tt.expected {
				assert.Equal(t, *tt.expected[i], *apps[i])
			}
		})
	}
}

func TestRESTClient_GetPartitions(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() *httptest.Server
		expected       []*dao.PartitionInfo
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name: "200 OK Response",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := []*dao.PartitionInfo{
						{
							Name: "partition1",
						},
						{
							Name: "partition2",
						},
					}
					writeResponse(t, w, response)
				}))
			},
			expected: []*dao.PartitionInfo{
				{
					Name: "partition1",
				},
				{
					Name: "partition2",
				},
			},
			wantErr: false,
		},
		{
			name: "Server Error",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "server error", http.StatusInternalServerError)
				}))
			},
			expected:       nil,
			wantErr:        true,
			expectedErrMsg: "yunicorn api returned non-OK status code: 500",
		},
		{
			name: "Unexpected JSON",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					type unexpected struct {
						Unexpected string `json:"unexpected"`
					}
					writeResponse(t, w, &unexpected{Unexpected: "unexpected"})
				}))
			},
			expected:       nil,
			wantErr:        true,
			expectedErrMsg: "json: cannot unmarshal object into Go value of type []*dao.PartitionInfo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := tt.setup()
			defer ts.Close()

			client := NewRESTClient(getMockServerYunikornConfig(t, ts.URL))

			partitions, err := client.GetPartitions(context.Background())
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErrMsg, err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, partitions)
			}
		})
	}
}

func TestRESTClient_GetPartitionQueues(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() *httptest.Server
		expected       *dao.PartitionQueueDAOInfo
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name: "200 OK Response",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := dao.PartitionQueueDAOInfo{

						QueueName: "queue1",
					}
					writeResponse(t, w, response)
				}))
			},
			expected: &dao.PartitionQueueDAOInfo{
				QueueName: "queue1",
			},

			wantErr: false,
		},
		{
			name: "Server Error",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "server error", http.StatusInternalServerError)
				}))
			},
			expected:       nil,
			wantErr:        true,
			expectedErrMsg: "yunicorn api returned non-OK status code: 500",
		},
		{
			name: "Unexpected JSON",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					type unexpected struct {
						Unexpected string `json:"unexpected"`
					}
					writeResponse(t, w, &[]unexpected{{Unexpected: "unexpected"}})
				}))
			},
			expected:       nil,
			wantErr:        true,
			expectedErrMsg: "json: cannot unmarshal array into Go value of type dao.PartitionQueueDAOInfo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := tt.setup()
			defer ts.Close()

			client := NewRESTClient(getMockServerYunikornConfig(t, ts.URL))

			queues, err := client.GetPartitionQueues(context.Background(), "testPartition")
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErrMsg, err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, queues)
			}
		})
	}
}

func TestRESTClient_GetPartitionNodes(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() *httptest.Server
		expected       []*dao.NodeDAOInfo
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name: "200 OK Response",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := []*dao.NodeDAOInfo{
						{
							NodeID: "node1",
						},
						{
							NodeID: "node2",
						},
					}
					writeResponse(t, w, response)
				}))
			},
			expected: []*dao.NodeDAOInfo{
				{
					NodeID: "node1",
				},
				{
					NodeID: "node2",
				},
			},
			wantErr: false,
		},
		{
			name: "Server Error",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "server error", http.StatusInternalServerError)
				}))
			},
			expected:       nil,
			wantErr:        true,
			expectedErrMsg: "yunicorn api returned non-OK status code: 500",
		},
		{
			name: "Unexpected JSON",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					type unexpected struct {
						Unexpected string `json:"unexpected"`
					}
					writeResponse(t, w, &unexpected{Unexpected: "unexpected"})
				}))
			},
			expected:       nil,
			wantErr:        true,
			expectedErrMsg: "json: cannot unmarshal object into Go value of type []*dao.NodeDAOInfo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := tt.setup()
			defer ts.Close()

			client := NewRESTClient(getMockServerYunikornConfig(t, ts.URL))

			nodes, err := client.GetPartitionNodes(context.Background(), "testPartition")
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErrMsg, err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, nodes)
			}
		})
	}
}

func TestRESTClient_GetNodeUtil(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() *httptest.Server
		expected       []*dao.PartitionNodesUtilDAOInfo
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name: "200 OK Response",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := []*dao.PartitionNodesUtilDAOInfo{
						{
							NodesUtilList: []*dao.NodesUtilDAOInfo{
								{
									ResourceType: "vcore",
									NodesUtil: []*dao.NodeUtilDAOInfo{
										{
											BucketName: "0-10%",
											NumOfNodes: 1,
											NodeNames:  []string{"aethergpu"},
										},
										{
											BucketName: "10-20%",
											NumOfNodes: 2,
											NodeNames:  []string{"primary-node", "second-node"},
										},
									},
								},
							},
						},
					}
					writeResponse(t, w, response)
				}))
			},
			expected: []*dao.PartitionNodesUtilDAOInfo{
				{
					NodesUtilList: []*dao.NodesUtilDAOInfo{
						{
							ResourceType: "vcore",
							NodesUtil: []*dao.NodeUtilDAOInfo{
								{
									BucketName: "0-10%",
									NumOfNodes: 1,
									NodeNames:  []string{"aethergpu"},
								},
								{
									BucketName: "10-20%",
									NumOfNodes: 2,
									NodeNames:  []string{"primary-node", "second-node"},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Server Error",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "server error", http.StatusInternalServerError)
				}))
			},
			expected:       nil,
			wantErr:        true,
			expectedErrMsg: "yunicorn api returned non-OK status code: 500",
		},
		{
			name: "Unexpected JSON",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					type unexpected struct {
						Unexpected string `json:"unexpected"`
					}
					writeResponse(t, w, &unexpected{Unexpected: "unexpected"})
				}))
			},
			expected:       nil,
			wantErr:        true,
			expectedErrMsg: "json: cannot unmarshal object into Go value of type []*dao.PartitionNodesUtilDAOInfo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := tt.setup()
			defer ts.Close()

			client := NewRESTClient(getMockServerYunikornConfig(t, ts.URL))

			nodeUtil, err := client.GetNodeUtil(context.Background())
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErrMsg, err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, nodeUtil)
			}
		})
	}
}

func TestRESTClient_GetAppsHistory(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() *httptest.Server
		expected       []*dao.ApplicationHistoryDAOInfo
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name: "200 OK Response",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := []*dao.ApplicationHistoryDAOInfo{
						{
							Timestamp:         1595939966153460000,
							TotalApplications: "1",
						},
						{
							Timestamp:         1595940206155187000,
							TotalApplications: "2",
						},
					}
					writeResponse(t, w, response)
				}))
			},
			expected: []*dao.ApplicationHistoryDAOInfo{
				{
					Timestamp:         1595939966153460000,
					TotalApplications: "1",
				},
				{
					Timestamp:         1595940206155187000,
					TotalApplications: "2",
				},
			},
			wantErr: false,
		},
		{
			name: "Server Error",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "server error", http.StatusInternalServerError)
				}))
			},
			expected:       nil,
			wantErr:        true,
			expectedErrMsg: "yunicorn api returned non-OK status code: 500",
		},
		{
			name: "Unexpected JSON",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					type unexpected struct {
						Unexpected string `json:"unexpected"`
					}
					writeResponse(t, w, &unexpected{Unexpected: "unexpected"})
				}))
			},
			expected:       nil,
			wantErr:        true,
			expectedErrMsg: "json: cannot unmarshal object into Go value of type []*dao.ApplicationHistoryDAOInfo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := tt.setup()
			defer ts.Close()

			client := NewRESTClient(getMockServerYunikornConfig(t, ts.URL))

			appsHistory, err := client.GetAppsHistory(context.Background())
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErrMsg, err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, appsHistory)
			}
		})
	}
}

func TestRESTClient_GetContainersHistory(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() *httptest.Server
		expected       []*dao.ContainerHistoryDAOInfo
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name: "200 OK Response",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := []*dao.ContainerHistoryDAOInfo{
						{
							Timestamp:       1595939966153460000,
							TotalContainers: "1",
						},
						{
							Timestamp:       1595940026152892000,
							TotalContainers: "2",
						},
					}
					writeResponse(t, w, response)
				}))
			},
			expected: []*dao.ContainerHistoryDAOInfo{
				{
					Timestamp:       1595939966153460000,
					TotalContainers: "1",
				},
				{
					Timestamp:       1595940026152892000,
					TotalContainers: "2",
				},
			},
			wantErr: false,
		},
		{
			name: "Server Error",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}))
			},
			expected:       nil,
			wantErr:        true,
			expectedErrMsg: "yunicorn api returned non-OK status code: 500",
		},
		{
			name: "Unexpected JSON",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					type unexpected struct {
						Unexpected string `json:"unexpected"`
					}
					writeResponse(t, w, &unexpected{Unexpected: "unexpected"})
				}))
			},
			expected:       nil,
			wantErr:        true,
			expectedErrMsg: "json: cannot unmarshal object into Go value of type []*dao.ContainerHistoryDAOInfo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := tt.setup()
			defer ts.Close()

			client := NewRESTClient(getMockServerYunikornConfig(t, ts.URL))

			containersHistory, err := client.GetContainersHistory(context.Background())
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErrMsg, err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, containersHistory)
			}
		})
	}
}

func getMockServerYunikornConfig(t *testing.T, serverURL string) *config.YunikornConfig {
	parsedURL, err := url.Parse(serverURL)
	require.NoError(t, err)

	portNum, err := strconv.Atoi(parsedURL.Port())
	require.NoError(t, err)

	return &config.YunikornConfig{
		Host: parsedURL.Hostname(),
		Port: portNum,
	}
}

func writeResponse(t *testing.T, w http.ResponseWriter, response any) {
	t.Helper()
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		t.Fatalf("error writing response: %v", err)
	}
}

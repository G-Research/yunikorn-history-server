package ykclient

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

	"github.com/apache/yunikorn-core/pkg/webservice/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetApplication(t *testing.T) {
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
			serverURL, err := url.Parse(ts.URL)
			require.NoError(t, err)

			portNum, err := strconv.Atoi(serverURL.Port())
			require.NoError(t, err)

			ykclient := Client{
				httpProto: serverURL.Scheme,
				ykHost:    serverURL.Hostname(),
				ykPort:    portNum,
				repo:      nil,
				appMap:    map[string]*dao.ApplicationDAOInfo{},
			}

			app, err := ykclient.GetApplication(context.Background(), tt.partName, tt.queueName, tt.appId)
			require.NoError(t, err)
			assert.NotNil(t, app)
			assert.Equal(t, tt.expected.Partition, app.Partition)
			assert.Equal(t, tt.expected.QueueName, app.QueueName)
			assert.Equal(t, tt.expected.ApplicationID, app.ApplicationID)
		})
	}
}

func Test_GetApplications(t *testing.T) {
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
			serverURL, err := url.Parse(ts.URL)
			require.NoError(t, err)

			portNum, err := strconv.Atoi(serverURL.Port())
			require.NoError(t, err)

			ykclient := Client{
				httpProto: serverURL.Scheme,
				ykHost:    serverURL.Hostname(),
				ykPort:    portNum,
				repo:      nil,
				appMap:    map[string]*dao.ApplicationDAOInfo{},
			}

			apps, err := ykclient.GetApplications(context.Background(), tt.partName, tt.queueName)
			require.NoError(t, err)
			assert.NotNil(t, apps)
			for i := range tt.expected {
				assert.Equal(t, *tt.expected[i], *apps[i])
			}
		})
	}
}

func Test_GetPartitions(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *httptest.Server
		expected []*dao.PartitionInfo
		wantErr  bool
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
					err := json.NewEncoder(w).Encode(response)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
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
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Unexpected JSON",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					type unexpected struct {
						Unexpected string `json:"unexpected"`
					}
					err := json.NewEncoder(w).Encode(&unexpected{Unexpected: "unexpected"})
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				}))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := tt.setup()
			defer ts.Close()

			serverURL, err := url.Parse(ts.URL)
			require.NoError(t, err)

			portNum, err := strconv.Atoi(serverURL.Port())
			require.NoError(t, err)

			ykclient := Client{
				httpProto: serverURL.Scheme,
				ykHost:    serverURL.Hostname(),
				ykPort:    portNum,
				repo:      nil,
				appMap:    map[string]*dao.ApplicationDAOInfo{},
			}

			partitions, err := ykclient.GetPartitions(context.Background())
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, partitions)
			}
		})
	}
}

func Test_GetPartitionQueues(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *httptest.Server
		expected []*dao.PartitionQueueDAOInfo
		wantErr  bool
	}{
		{
			name: "200 OK Response",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := []*dao.PartitionQueueDAOInfo{
						{
							QueueName: "queue1",
						},
						{
							QueueName: "queue2",
						},
					}
					err := json.NewEncoder(w).Encode(response)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				}))
			},
			expected: []*dao.PartitionQueueDAOInfo{
				{
					QueueName: "queue1",
				},
				{
					QueueName: "queue2",
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
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Unexpected JSON",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					type unexpected struct {
						Unexpected string `json:"unexpected"`
					}
					err := json.NewEncoder(w).Encode(&unexpected{Unexpected: "unexpected"})
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				}))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := tt.setup()
			defer ts.Close()

			serverURL, err := url.Parse(ts.URL)
			require.NoError(t, err)

			portNum, err := strconv.Atoi(serverURL.Port())
			require.NoError(t, err)

			ykclient := Client{
				httpProto: serverURL.Scheme,
				ykHost:    serverURL.Hostname(),
				ykPort:    portNum,
				repo:      nil,
				appMap:    map[string]*dao.ApplicationDAOInfo{},
			}

			queues, err := ykclient.GetPartitionQueues(context.Background(), "testPartition")
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, queues)
			}
		})
	}
}

func Test_GetPartitionNodes(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *httptest.Server
		expected []*dao.NodeDAOInfo
		wantErr  bool
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
					err := json.NewEncoder(w).Encode(response)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
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
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Unexpected JSON",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					type unexpected struct {
						Unexpected string `json:"unexpected"`
					}
					err := json.NewEncoder(w).Encode(&unexpected{Unexpected: "unexpected"})
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				}))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := tt.setup()
			defer ts.Close()

			serverURL, err := url.Parse(ts.URL)
			require.NoError(t, err)

			portNum, err := strconv.Atoi(serverURL.Port())
			require.NoError(t, err)

			ykclient := Client{
				httpProto: serverURL.Scheme,
				ykHost:    serverURL.Hostname(),
				ykPort:    portNum,
				repo:      nil,
				appMap:    map[string]*dao.ApplicationDAOInfo{},
			}

			nodes, err := ykclient.GetPartitionNodes(context.Background(), "testPartition")
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, nodes)
			}
		})
	}
}

func Test_GetNodeUtil(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *httptest.Server
		expected *[]dao.PartitionNodesUtilDAOInfo
		wantErr  bool
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
					err := json.NewEncoder(w).Encode(response)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				}))
			},
			expected: &[]dao.PartitionNodesUtilDAOInfo{
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
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Unexpected JSON",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					type unexpected struct {
						Unexpected string `json:"unexpected"`
					}
					err := json.NewEncoder(w).Encode(&unexpected{Unexpected: "unexpected"})
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				}))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := tt.setup()
			defer ts.Close()

			serverURL, err := url.Parse(ts.URL)
			require.NoError(t, err)

			portNum, err := strconv.Atoi(serverURL.Port())
			require.NoError(t, err)

			ykclient := Client{
				httpProto: serverURL.Scheme,
				ykHost:    serverURL.Hostname(),
				ykPort:    portNum,
				repo:      nil,
				appMap:    map[string]*dao.ApplicationDAOInfo{},
			}

			nodeUtil, err := ykclient.GetNodeUtil(context.Background())
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, nodeUtil)
			}
		})
	}
}

func Test_GetAppsHistory(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *httptest.Server
		expected []*dao.ApplicationHistoryDAOInfo
		wantErr  bool
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
					err := json.NewEncoder(w).Encode(response)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
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
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Unexpected JSON",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					type unexpected struct {
						Unexpected string `json:"unexpected"`
					}
					err := json.NewEncoder(w).Encode(&unexpected{Unexpected: "unexpected"})
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				}))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := tt.setup()
			defer ts.Close()

			serverURL, err := url.Parse(ts.URL)
			require.NoError(t, err)

			portNum, err := strconv.Atoi(serverURL.Port())
			require.NoError(t, err)

			ykclient := Client{
				httpProto: serverURL.Scheme,
				ykHost:    serverURL.Hostname(),
				ykPort:    portNum,
				repo:      nil,
				appMap:    map[string]*dao.ApplicationDAOInfo{},
			}

			appsHistory, err := ykclient.GetAppsHistory(context.Background())
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, appsHistory)
			}
		})
	}
}

func Test_GetContainersHistory(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *httptest.Server
		expected []*dao.ContainerHistoryDAOInfo
		wantErr  bool
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
					err := json.NewEncoder(w).Encode(response)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
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
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Unexpected JSON",
			setup: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					type unexpected struct {
						Unexpected string `json:"unexpected"`
					}
					err := json.NewEncoder(w).Encode(&unexpected{Unexpected: "unexpected"})
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				}))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := tt.setup()
			defer ts.Close()

			serverURL, err := url.Parse(ts.URL)
			require.NoError(t, err)

			portNum, err := strconv.Atoi(serverURL.Port())
			require.NoError(t, err)

			ykclient := Client{
				httpProto: serverURL.Scheme,
				ykHost:    serverURL.Hostname(),
				ykPort:    portNum,
				repo:      nil,
				appMap:    map[string]*dao.ApplicationDAOInfo{},
			}

			containersHistory, err := ykclient.GetContainersHistory(context.Background())
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, containersHistory)
			}
		})
	}
}

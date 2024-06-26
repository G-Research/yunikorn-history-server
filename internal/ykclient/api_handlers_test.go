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

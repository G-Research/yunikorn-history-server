package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/G-Research/yunikorn-core/pkg/webservice/dao"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/ptr"

	"github.com/G-Research/yunikorn-history-server/cmd/yunikorn-history-server/commands"
	"github.com/G-Research/yunikorn-history-server/internal/health"
	"github.com/G-Research/yunikorn-history-server/internal/yunikorn/model"
	"github.com/G-Research/yunikorn-history-server/test/k8s"
	"github.com/G-Research/yunikorn-history-server/test/util"
)

const (
	testNamespacePrefix = "yunikorn-e2e-"
	serverURL           = "http://localhost:8989"
)

var (
	ctx    context.Context
	cancel context.CancelFunc
)

func TestMain(m *testing.M) {
	// Setup
	ctx, cancel = context.WithCancel(context.Background())
	go runApp(ctx)

	// Run the tests
	code := m.Run()

	// Teardown
	defer cancel()
	os.Exit(code)
}

func TestYunikornApp_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}
	ctx, cancel = context.WithCancel(context.Background())
	t.Cleanup(cancel)

	ns := createTestNamespace(ctx, t, k8s.GetTestK8sClient(t))
	t.Cleanup(func() {
		deleteTestNamespace(context.Background(), t, k8s.GetTestK8sClient(t), ns)
	})

	assert.Eventually(t, func() bool {
		healthy, err := getReadinessStatus(serverURL)
		return healthy && err == nil
	}, 10*time.Second, 500*time.Millisecond)

	// sleep for 2 seconds just in case so all goroutines are ready
	time.Sleep(2 * time.Second)

	k8sClient := k8s.GetTestK8sClient(t)
	appID := "yunikorn-app-test-pod"

	_, err := k8sClient.CoreV1().Pods(ns).Create(ctx, testApp(appID), metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error creating test app: %v", err)
	}
	assert.Eventually(t, func() bool {
		appsResponse, err := getApps(serverURL, ns)
		if err != nil {
			return false
		}
		for _, app := range appsResponse {
			fmt.Println(app.ApplicationID)
			if app.ApplicationID == appID {
				return true
			}
		}
		return false
	}, 400*time.Second, 5*time.Second)
}

func TestYunikornEventStream_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	ctx, cancel = context.WithCancel(context.Background())
	t.Cleanup(cancel)

	ns := createTestNamespace(ctx, t, k8s.GetTestK8sClient(t))
	t.Cleanup(func() {
		deleteTestNamespace(context.Background(), t, k8s.GetTestK8sClient(t), ns)
	})

	assert.Eventually(t, func() bool {
		healthy, err := getReadinessStatus(serverURL)
		return healthy && err == nil
	}, 10*time.Second, 500*time.Millisecond)

	// sleep for 2 seconds just in case so all goroutines are ready
	time.Sleep(2 * time.Second)

	k8sClient := k8s.GetTestK8sClient(t)

	job := testJob()
	_, err := k8sClient.BatchV1().Jobs(ns).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error creating test job: %v", err)
	}

	expectedCounts := model.EventTypeCounts{
		"APP-ADD":          7,
		"APP-REMOVE":       7,
		"APP-SET":          5,
		"NODE-ADD":         3,
		"NODE-REMOVE":      3,
		"USERGROUP-ADD":    3,
		"USERGROUP-REMOVE": 3,
		"QUEUE-ADD":        2,
		"QUEUE-REMOVE":     2,
	}
	assert.Eventually(t, func() bool {
		counts, err := getEventStatistics(serverURL)
		if err != nil {
			return false
		}
		diff := cmp.Diff(expectedCounts, counts)
		return diff == ""
	}, 100*time.Second, 5*time.Second)
}

func TestYunikornQueueCreation_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}
	ns := "yunikorn"
	queueName := util.GenerateRandomAlphanum(t, 8) + "test-queue"
	configMap := testQueueConfigMap(queueName)

	ctx, cancel = context.WithCancel(context.Background())
	t.Cleanup(cancel)

	t.Cleanup(func() {
		k8sClient := k8s.GetTestK8sClient(t)
		err := k8sClient.CoreV1().ConfigMaps(ns).Delete(ctx, configMap.Name, metav1.DeleteOptions{})
		if err != nil {
			t.Fatalf("error deleting configmap: %v", err)
		}
	})

	assert.Eventually(t, func() bool {
		healthy, err := getReadinessStatus(serverURL)
		return healthy && err == nil
	}, 10*time.Second, 500*time.Millisecond)

	// sleep for 2 seconds just in case so all goroutines are ready
	time.Sleep(2 * time.Second)

	k8sClient := k8s.GetTestK8sClient(t)

	// create the configmap which has the queue definition
	_, err := k8sClient.CoreV1().ConfigMaps(ns).Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error creating test queue: %v", err)
	}

	expectedCount := 5
	assert.Eventually(t, func() bool {
		counts, err := getEventStatistics(serverURL)
		if err != nil {
			return false
		}
		actualCount, ok := counts["QUEUE-ADD"]
		return ok && actualCount == expectedCount
	}, 100*time.Second, 5*time.Second)

	assert.Eventually(t, func() bool {
		queuesResponse, err := getQueues(serverURL)
		if err != nil {
			return false
		}
		for _, queue := range queuesResponse {
			if queue.QueueName == "root" {
				for _, childQueue := range queue.Children {
					if childQueue.QueueName == "root."+queueName {
						return true
					}
				}
			}
		}
		return false
	}, 400*time.Second, 5*time.Second)
}

// createTestNamespace creates a test namespace for the e2e test and returns the name of the namespace.
func createTestNamespace(ctx context.Context, t *testing.T, k8sClient kubernetes.Interface) string {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: testNamespacePrefix,
		},
	}
	created, err := k8sClient.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error creating namespace: %v", err)
	}

	return created.Name
}

func deleteTestNamespace(ctx context.Context, t *testing.T, k8sClient kubernetes.Interface, ns string) {
	if err := k8sClient.CoreV1().Namespaces().Delete(ctx, ns, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("error deleting namespace: %v", err)
	}
}

// getReadinessStatus performs a readiness check on the yunikorn history server.
func getReadinessStatus(serverURL string) (bool, error) {
	var status health.ReadinessStatus
	url := fmt.Sprintf("%s/api/v1/health/readiness", serverURL)
	if err := httpGet(url, &status); err != nil {
		return false, err
	}

	return status.Healthy, nil
}

func getEventStatistics(serverURL string) (model.EventTypeCounts, error) {
	var counts model.EventTypeCounts
	url := fmt.Sprintf("%s/api/v1/event-statistics", serverURL)
	if err := httpGet(url, &counts); err != nil {
		return nil, err
	}
	return counts, nil
}

func getQueues(serverURL string) ([]*dao.PartitionQueueDAOInfo, error) {
	var queues []*dao.PartitionQueueDAOInfo
	url := fmt.Sprintf("%s/api/v1/partition/default/queues", serverURL)
	if err := httpGet(url, &queues); err != nil {
		return nil, err
	}
	return queues, nil
}

func getApps(serverURL string, namespace string) ([]*dao.ApplicationDAOInfo, error) {
	var apps []*dao.ApplicationDAOInfo
	url := fmt.Sprintf("%s/api/v1/partition/default/queue/root.%s/applications", serverURL, namespace)
	if err := httpGet(url, &apps); err != nil {
		return nil, err
	}
	return apps, nil
}

// httpGet performs an HTTP GET request and decodes the response into the provided out parameter.
func httpGet(url string, out any) error {
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error getting response: %v", resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

func runApp(ctx context.Context) {
	os.Args = []string{"yunikorn-history-server", "--config", "../../config/yunikorn-history-server/local.yml"}
	if err := commands.New().ExecuteContext(ctx); err != nil {
		panic(err)
	}
}

func testJob() *batchv1.Job {
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "batch-sleep-job-",
		},
		Spec: batchv1.JobSpec{
			Completions: ptr.To[int32](3),
			Parallelism: ptr.To[int32](3),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":           "sleep",
						"applicationId": "batch-sleep-job-1",
						"queue":         "root",
					},
				},
				Spec: corev1.PodSpec{
					SchedulerName: "yunikorn",
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "sleep",
							Image: "alpine:latest",
							Command: []string{
								"sleep",
								"5",
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("50m"),
									corev1.ResourceMemory: resource.MustParse("100Mi"),
								},
							},
						},
					},
				},
			},
		},
	}
}

func testQueueConfigMap(queue string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "yunikorn-configs",
		},
		Data: map[string]string{
			"queues.yaml": fmt.Sprintf(`
partitions:
  - name: default
    queues:
      - name: root
        queues:
          - name: %s
`, queue),
		},
	}
}

func testApp(appID string) *corev1.Pod {
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "yunikorn-nginx-pod",
			Labels: map[string]string{
				"app":           "nginx",
				"applicationId": appID,
			},
		},
		Spec: corev1.PodSpec{
			SchedulerName: "yunikorn",
			Containers: []corev1.Container{
				{
					Name:  "nginx",
					Image: "nginx:latest",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("20M"),
						},
					},
				},
			},
		},
	}
}

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/G-Research/yunikorn-history-server/cmd/yunikorn-history-server/commands"
	"github.com/G-Research/yunikorn-history-server/internal/health"
	"github.com/G-Research/yunikorn-history-server/internal/yunikorn/model"
	"github.com/G-Research/yunikorn-history-server/test/k8s"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/ptr"
	"net/http"
	"os"
	"testing"
	"time"
)

const (
	testNamespacePrefix = "yunikorn-e2e-"
)

func TestYunikornEventStream_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	ns := createTestNamespace(ctx, t, k8s.GetTestK8sClient(t))
	t.Cleanup(func() {
		deleteTestNamespace(context.Background(), t, k8s.GetTestK8sClient(t), ns)
	})

	serverURL := "http://localhost:8989"
	go runApp(ctx)

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
	url := fmt.Sprintf("%s/ws/v1/health/readiness", serverURL)
	if err := httpGet(url, &status); err != nil {
		return false, err
	}

	return status.Healthy, nil
}

func getEventStatistics(serverURL string) (model.EventTypeCounts, error) {
	var counts model.EventTypeCounts
	url := fmt.Sprintf("%s/ws/v1/event-statistics", serverURL)
	if err := httpGet(url, &counts); err != nil {
		return nil, err
	}

	return counts, nil
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

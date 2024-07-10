package k8s

import (
	"errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"testing"
)

func GetTestK8sClient(t *testing.T) kubernetes.Interface {
	config := getInClusterK8sClient(t)
	if config == nil {
		config = getOutClusterK8sClient(t, "")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Fatalf("error creating clientset: %v", err)
	}

	return clientset
}

func getInClusterK8sClient(t *testing.T) *rest.Config {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		if errors.Is(err, rest.ErrNotInCluster) {
			return nil
		}
		t.Fatalf("error creating in-cluster config: %v", err)
	}
	return config
}

// GetOutClusterK8sClient returns a k8s client for out-of-cluster use.
// If kubeconfigFilepath is not provided, it will default to $HOME/.kube/config.
func getOutClusterK8sClient(t *testing.T, kubeconfigFilepath string) *rest.Config {
	var kubeconfig string
	home := homedir.HomeDir()
	switch {
	case kubeconfigFilepath != "":
		kubeconfig = kubeconfigFilepath
	case home != "":
		kubeconfig = filepath.Join(home, ".kube", "config")
	default:
		t.Fatalf("error creating out-cluster config: no kubeconfig specified")
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		t.Fatalf("error creating out-cluster config: %v", err)
	}

	return config
}

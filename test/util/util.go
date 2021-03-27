package util

import (
	"context"
	"os"
	"testing"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// RetryInterval is time for test to retry.
	RetryInterval = time.Second * 5
	// Timeout is time after which tests stops and fails.
	Timeout = time.Minute * 2
)

func CreateKubernetesClient(t *testing.T) *kubernetes.Clientset {
	kubeconfigPath := os.ExpandEnv(os.Getenv("KUBECONFIG"))
	if kubeconfigPath == "" {
		t.Fatalf("env var KUBECONFIG was not set")
	}

	c, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		t.Fatalf("failed building rest client: %v", err)
	}

	cs, err := kubernetes.NewForConfig(c)
	if err != nil {
		t.Fatalf("failed creating new clientset: %v", err)
	}

	return cs
}

func WaitForNamespaceToBeDeleted(t *testing.T, client kubernetes.Interface, name string, retryInterval, timeout time.Duration) {
	if err := wait.PollImmediate(retryInterval, timeout, func() (done bool, err error) {
		_, err = client.CoreV1().Namespaces().Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				t.Logf("namespace %s deleted", name)
				return true, nil
			}

			return false, nil
		}

		return false, nil

	}); err != nil {
		t.Fatalf("waiting for the namespace to be deleted: %v", err)
	}
}

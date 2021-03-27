// +build e2e

package cloner_test

import (
	"context"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/impochi/cloner/pkg/registry"
	"github.com/impochi/cloner/test/util"
)

const daemonsetManifest = `apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: prometheus-daemonset
spec:
  selector:
    matchLabels:
      tier: monitoring
      name: prometheus-exporter
  template:
    metadata:
      labels:
        tier: monitoring
        name: prometheus-exporter
    spec:
      initContainers:
      - name: init-1
        image: busybox:1.33.0
        command: ['sh', '-c', 'echo Init Container executing!']
      containers:
      - name: prometheus
        image: prom/node-exporter
        ports:
        - containerPort: 80
`

//nolint:funlen
func TestDaemonSetImageClone(t *testing.T) {
	client := util.CreateKubernetesClient(t)

	namespace := &corev1.Namespace{}
	if err := yaml.Unmarshal([]byte(NamespaceManifest), namespace); err != nil {
		t.Fatalf("failed unmarshaling manifest: %v", err)
	}

	namespace, err := client.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create Namespace: %v", err)
	}

	daemonset := &appsv1.DaemonSet{}
	if err := yaml.Unmarshal([]byte(daemonsetManifest), daemonset); err != nil {
		t.Fatalf("failed unmarshaling manifest: %v", err)
	}

	initContainerImage := daemonset.Spec.Template.Spec.InitContainers[0].Image
	containerImage := daemonset.Spec.Template.Spec.Containers[0].Image

	dstInitContainerImage, err := registry.GetDestinationImage(initContainerImage)
	if err != nil {
		t.Fatalf("failed to get destination image: %v", err)
	}

	dstContainerImage, err := registry.GetDestinationImage(containerImage)
	if err != nil {
		t.Fatalf("failed to get destination image: %v", err)
	}

	daemonset, err = client.AppsV1().DaemonSets(namespace.Name).Create(context.TODO(), daemonset, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create DaemonSet: %v", err)
	}

	// wait for 1 minute so that the image is cloned and daemonset updaed by the cloner controller.
	// TODO: implement dynamic wait, that checks if the daemonset is ready or not before proceeding.
	time.Sleep(1 * time.Minute)

	daemonset, err = client.AppsV1().DaemonSets(namespace.Name).Get(context.TODO(), daemonset.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to create DaemonSet: %v", err)
	}

	initContainerImage = daemonset.Spec.Template.Spec.InitContainers[0].Image
	containerImage = daemonset.Spec.Template.Spec.Containers[0].Image

	if dstInitContainerImage != initContainerImage {
		t.Fatalf("expected %q, got %q", dstInitContainerImage, initContainerImage)
	}

	if dstContainerImage != containerImage {
		t.Fatalf("expected %q, got %q", dstContainerImage, containerImage)
	}

	t.Cleanup(func() {
		if err := client.AppsV1().DaemonSets(namespace.Name).Delete(
			context.TODO(), daemonset.ObjectMeta.Name, metav1.DeleteOptions{}); err != nil {
			t.Logf("failed to remove DaemonSet: %v", err)
		}

		if err := client.CoreV1().Namespaces().Delete(
			context.TODO(), namespace.ObjectMeta.Name, metav1.DeleteOptions{}); err != nil {
			t.Logf("failed to remove Namespace: %v", err)
		}

		util.WaitForNamespaceToBeDeleted(t, client, namespace.ObjectMeta.Name, time.Second*5, time.Minute*5)
	})
}

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

const NamespaceManifest = `apiVersion: v1
kind: Namespace
metadata:
  name: e2e
`

const deploymentManifest = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      initContainers:
      - name: init-1
        image: busybox:1.33.0
        command: ['sh', '-c', 'echo Init Container executing!']
      containers:
      - name: nginx
        image: nginx:1.14.2
        ports:
        - containerPort: 80
`

func TestDeploymentImageClone(t *testing.T) {

	client := util.CreateKubernetesClient(t)

	namespace := &corev1.Namespace{}
	if err := yaml.Unmarshal([]byte(NamespaceManifest), namespace); err != nil {
		t.Fatalf("failed unmarshaling manifest: %v", err)
	}

	namespace, err := client.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create Namespace: %v", err)
	}

	deployment := &appsv1.Deployment{}
	if err := yaml.Unmarshal([]byte(deploymentManifest), deployment); err != nil {
		t.Fatalf("failed unmarshaling manifest: %v", err)
	}

	initContainerImage := deployment.Spec.Template.Spec.InitContainers[0].Image
	containerImage := deployment.Spec.Template.Spec.Containers[0].Image

	dstInitContainerImage, err := registry.GetDestinationImage(initContainerImage)
	if err != nil {
		t.Fatalf("failed to get destination image: %v", err)
	}

	dstContainerImage, err := registry.GetDestinationImage(containerImage)
	if err != nil {
		t.Fatalf("failed to get destination image: %v", err)
	}

	deployment, err = client.AppsV1().Deployments(namespace.Name).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create Deployment: %v", err)
	}

	// wait for 1 minute so that the image is cloned and deployment updated by the cloner controller.
	// TODO: implement dynamic wait, that checks if the deployment is ready or not before proceeding.
	time.Sleep(1 * time.Minute)

	deployment, err = client.AppsV1().Deployments(namespace.Name).Get(context.TODO(), deployment.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to create Deployment: %v", err)
	}

	initContainerImage = deployment.Spec.Template.Spec.InitContainers[0].Image
	containerImage = deployment.Spec.Template.Spec.Containers[0].Image

	if dstInitContainerImage != initContainerImage {
		t.Fatalf("expected %q, got %q", dstInitContainerImage, initContainerImage)
	}

	if dstContainerImage != containerImage {
		t.Fatalf("expected %q, got %q", dstContainerImage, containerImage)
	}

	t.Cleanup(func() {
		if err := client.AppsV1().Deployments(namespace.Name).Delete(
			context.TODO(), deployment.ObjectMeta.Name, metav1.DeleteOptions{}); err != nil {
			t.Logf("failed to remove Deployment: %v", err)
		}

		if err := client.CoreV1().Namespaces().Delete(
			context.TODO(), namespace.ObjectMeta.Name, metav1.DeleteOptions{}); err != nil {
			t.Logf("failed to remove Namespace: %v", err)
		}

		util.WaitForNamespaceToBeDeleted(t, client, namespace.ObjectMeta.Name, time.Second*5, time.Minute*5)
	})
}

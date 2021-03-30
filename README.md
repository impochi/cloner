# Image Cloner Controller

Image Clone controller is a Kubernetes controller that watches Deployments and DaemonSets and copies/caches the public
images into a separate registry provided as cli flags and updates the said resources to use the backed up image
location.

The result is that your Deployments and DaemonSets do not depend on the public images on which you do not have any
control over.

The controller ignores the Deployments/DaemonSets in the `kube-system` namespace.

## Prerequisites
* Kubernetes cluster
* Secret containing the container registry credentials.

## Installation

1. Update the file [deploy/01-secrets.yaml](https://github.com/impochi/cloner/blob/main/deploy/01-secrets.yaml)
  Provide the credentials for the container registry.

2. Apply the manifests in the `deploy` directory:

  ```bash
  kubectl apply -f deploy/
  ```

## Testing

Sample Deployment and Daemonset manifests are provided in order to test the controller:

```bash
kubectl create namespace test
kubectl apply -n test -f examples/nginx.yaml
kubectl apply -n test -f examples/node-exporter.yaml
```

Once the pods are in running status, Controller will backup the images and restart the Deployment and Daemonset.
Successful back up of the controller can be verified by the `image` field in `containers` and `initContainers`

The repository also provides e2e tests, you can also create a Deployment/DaemonSet and check that the images would have
been copied over to the provided registry and the Deployment/DaemonSet restarted in order to use the new image location.

In order to run the e2e tests, execute the following command:

```bash
## Note: REGISTRY_USERNAME and REGISTRY_PROVIDER must be the same as the ones provided in the secret.
## Note: REGISTRY_PASSWORD is not actually required, it can be any non-empty value. See Issue #17
REGISTRY_PROVIDER=docker.io REGISTRY_PASSWORD=abc REGISTRY_USERNAME=username KUBECONFIG=~/.kube/config go test -mod=vendor -tags=e2e -covermode=atomic -buildmode=exe -v -count=1 ./test/...
```

## Build locally

Clone this repository and build the project, once done create a Docker image and push to your repository.
The above steps would be as follows:

```bash
git clone git@github.com:impochi/cloner.git
cd cloner

make docker-build

# If using docker
docker tag imranpochi/cloner:latest <repo_name>/<image_name>:<tag>
docker push <repo_name>/<image_name>:<tag>
```

## Limitations

* Deployments/DaemonSets created in the controllers namespace are Ignored. See issue [#7](https://github.com/impochi/cloner/issues/7)
* Controller only takes backup of public images and leaves the private images untouched i.e if the Deployment/DaemonSet
  contains an `imagePullSecret` the resource images are not backed up.

## Demo

[![asciicast](https://asciinema.org/a/dir7mBgybU1KKjIgdtKkad105.svg)](https://asciinema.org/a/dir7mBgybU1KKjIgdtKkad105)

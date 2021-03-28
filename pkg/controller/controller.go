// Package controller handles the logic of Image clone controller, watching the resources
// and reconciling objects with the Kubernetes API server.
package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	pkglog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	pkgregistry "github.com/impochi/cloner/pkg/registry"
)

// ClonerReconciler is the controller's reconciler object.
type ClonerReconciler struct {
	Client client.Client
}

// Reconcile reconciles the object that is in question. In this case its either a Deployment or
// DaemonSet.
func (cr *ClonerReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := pkglog.FromContext(ctx)

	// Get the Deployents or DaemonSet from the cache
	deployment := &appsv1.Deployment{}
	daemonset := &appsv1.DaemonSet{}

	kind := "Deployment"

	err := cr.Client.Get(ctx, req.NamespacedName, deployment)
	if errors.IsNotFound(err) {
		log.Info("not a Deployment, checking for Daemonset")

		err = cr.Client.Get(ctx, req.NamespacedName, daemonset)
		if errors.IsNotFound(err) {
			log.Info("not a Daemonset")

			return reconcile.Result{}, nil
		}

		kind = "DaemonSet"
	}

	if err != nil {
		log.Error(err, "could not fetch Deployment or DaemonSet")

		return reconcile.Result{}, err
	}

	log.Info("reconciling Deployment", "deployment name", deployment.Name)

	if kind == "Deployment" && isDeploymentReady(deployment) && len(deployment.Spec.Template.Spec.ImagePullSecrets) == 0 {
		return cr.reconcileDeployment(ctx, deployment)
	}

	if kind == "DaemonSet" && isDaemonSetReady(daemonset) && len(daemonset.Spec.Template.Spec.ImagePullSecrets) == 0 {
		return cr.reconcileDaemonSet(ctx, daemonset)
	}

	return reconcile.Result{}, nil
}

func isDaemonSetReady(ds *appsv1.DaemonSet) bool {
	status := ds.Status
	desired := status.DesiredNumberScheduled

	ready := status.NumberReady
	if desired == ready && desired > 0 {
		return true
	}

	return false
}

func isDeploymentReady(deployments *appsv1.Deployment) bool {
	status := deployments.Status
	desired := status.Replicas
	ready := status.ReadyReplicas

	if desired == ready && desired > 0 {
		return true
	}

	return false
}

func (cr *ClonerReconciler) reconcileDeployment(ctx context.Context,
	deployment *appsv1.Deployment) (reconcile.Result, error) {
	log := pkglog.FromContext(ctx)

	needsUpdate := false

	for index, container := range deployment.Spec.Template.Spec.InitContainers {
		dstImage, err := pkgregistry.GetDestinationImage(container.Image)
		if err != nil {
			log.Error(err, "failed to get destination image")

			return reconcile.Result{}, err
		}

		if container.Image != dstImage {
			if err := pkgregistry.Backup(container.Image, dstImage); err != nil {
				log.Error(err, "failed to push image")

				return reconcile.Result{}, err
			}

			deployment.Spec.Template.Spec.InitContainers[index].Image = dstImage
			needsUpdate = true
		}
	}

	for index, container := range deployment.Spec.Template.Spec.Containers {
		dstImage, err := pkgregistry.GetDestinationImage(container.Image)
		if err != nil {
			log.Error(err, "failed to get destination image")

			return reconcile.Result{}, err
		}

		if container.Image != dstImage {
			if err := pkgregistry.Backup(container.Image, dstImage); err != nil {
				log.Error(err, "failed to push image")

				return reconcile.Result{}, err
			}

			deployment.Spec.Template.Spec.Containers[index].Image = dstImage
			needsUpdate = true
		}
	}
	// Update Deployment
	if needsUpdate {
		if err := cr.Client.Update(ctx, deployment); err != nil {
			log.Error(err, "failed to update Deployment")

			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (cr *ClonerReconciler) reconcileDaemonSet(ctx context.Context,
	daemonset *appsv1.DaemonSet) (reconcile.Result, error) {
	log := pkglog.FromContext(ctx)

	needsUpdate := false

	for index, container := range daemonset.Spec.Template.Spec.InitContainers {
		dstImage, err := pkgregistry.GetDestinationImage(container.Image)
		if err != nil {
			log.Error(err, "failed to get destination image")

			return reconcile.Result{}, err
		}

		if container.Image != dstImage {
			if err := pkgregistry.Backup(container.Image, dstImage); err != nil {
				log.Error(err, "failed to push image")

				return reconcile.Result{}, err
			}

			daemonset.Spec.Template.Spec.InitContainers[index].Image = dstImage
			needsUpdate = true
		}
	}

	for index, container := range daemonset.Spec.Template.Spec.Containers {
		dstImage, err := pkgregistry.GetDestinationImage(container.Image)
		if err != nil {
			log.Error(err, "failed to get destination image")

			return reconcile.Result{}, err
		}

		if container.Image != dstImage {
			if err := pkgregistry.Backup(container.Image, dstImage); err != nil {
				log.Error(err, "failed to push image")

				return reconcile.Result{}, err
			}

			daemonset.Spec.Template.Spec.Containers[index].Image = dstImage
			needsUpdate = true
		}
	}

	// Update Daemonset
	if needsUpdate {
		if err := cr.Client.Update(ctx, daemonset); err != nil {
			log.Error(err, "failed to update DaemonSet")

			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

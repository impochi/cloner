package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	pkgregistry "github.com/impochi/cloner/pkg/registry"
)

type ClonerReconciler struct {
	Client client.Client
}

func (cr *ClonerReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx)

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

		kind = "Daemonset"
	}

	if err != nil {
		log.Error(err, "could not fetch Deployment or DaemonSet")
		return reconcile.Result{}, err
	}

	log.Info("reconciling Deployment", "deployment name", deployment.Name)

	if kind == "Deployment" {
		needsUpdate := false
		for index, container := range deployment.Spec.Template.Spec.InitContainers {
			dstImage, err := pkgregistry.GetDestinationImage(container.Image)
			if err != nil {
				log.Error(err, "failed to get destination image")
			}
			if container.Image != dstImage {
				if err := pkgregistry.Backup(container.Image, dstImage); err != nil {
					log.Error(err, "failed to push image")
				}

				deployment.Spec.Template.Spec.InitContainers[index].Image = dstImage
				needsUpdate = true
			}
		}

		for index, container := range deployment.Spec.Template.Spec.Containers {
			dstImage, err := pkgregistry.GetDestinationImage(container.Image)
			if err != nil {
				log.Error(err, "failed to get destination image")
			}
			if container.Image != dstImage {
				if err := pkgregistry.Backup(container.Image, dstImage); err != nil {
					log.Error(err, "failed to push image")
				}

				deployment.Spec.Template.Spec.Containers[index].Image = dstImage
				needsUpdate = true
			}
		}

		// Update Deployment
		if needsUpdate {
			if err := cr.Client.Update(ctx, deployment); err != nil {
				log.Error(err, "failed to update Deployment")
			}
		}
	}

	if kind == "DaemonSet" {
		needsUpdate := false
		for index, container := range deployment.Spec.Template.Spec.InitContainers {
			dstImage, err := pkgregistry.GetDestinationImage(container.Image)
			if err != nil {
				log.Error(err, "failed to get destination image")
			}
			if container.Image != dstImage {
				if err := pkgregistry.Backup(container.Image, dstImage); err != nil {
					log.Error(err, "failed to push image")
				}

				deployment.Spec.Template.Spec.InitContainers[index].Image = dstImage
				needsUpdate = true
			}
		}

		for index, container := range deployment.Spec.Template.Spec.Containers {
			dstImage, err := pkgregistry.GetDestinationImage(container.Image)
			if err != nil {
				log.Error(err, "failed to get destination image")
			}
			if container.Image != dstImage {
				if err := pkgregistry.Backup(container.Image, dstImage); err != nil {
					log.Error(err, "failed to push image")
				}

				deployment.Spec.Template.Spec.Containers[index].Image = dstImage
				needsUpdate = true
			}
		}

		// Update Deployment
		if needsUpdate {
			if err := cr.Client.Update(ctx, deployment); err != nil {
				log.Error(err, "failed to update Deployment")
			}
		}
	}

	return reconcile.Result{}, nil
}

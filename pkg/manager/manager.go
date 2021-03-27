// Package manager handles the manager lifecycle and initiates the controller.
package manager

import (
	"os"

	appsv1 "k8s.io/api/apps/v1"

	clonercontroller "github.com/impochi/cloner/pkg/controller"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/impochi/cloner/cli/config"
)

// Run starts the manager.
func Run(config *config.Config) { //nolint:funlen
	controllerruntime.SetLogger(config.Logger)
	log := controllerruntime.Log.WithName("manager")

	mgr, err := controllerruntime.NewManager(
		controllerruntime.GetConfigOrDie(),
		controllerruntime.Options{},
	)
	if err != nil {
		log.Error(err, "failed to create manager")
		os.Exit(1)
	}

	// Setup Cloner controller
	log.Info("setting up Cloner controller")

	ctrller, err := controller.New("cloner", mgr,
		controller.Options{
			Reconciler: &clonercontroller.ClonerReconciler{
				Client: mgr.GetClient(),
			},
			Log: log,
		})
	if err != nil {
		log.Error(err, "failed to create controller")
	}

	if err := ctrller.Watch(
		&source.Kind{Type: &appsv1.Deployment{}},
		&handler.EnqueueRequestForObject{},
		predicate.NewPredicateFuncs(func(obj client.Object) bool {
			deployment := obj.(*appsv1.Deployment)
			for _, namespace := range config.IgnoreNamespaces {
				if namespace == deployment.Namespace {
					return false
				}
			}

			return true
		}),
	); err != nil {
		log.Error(err, "failed to watch Deployment")
	}

	if err := ctrller.Watch(
		&source.Kind{Type: &appsv1.DaemonSet{}},
		&handler.EnqueueRequestForObject{},
		predicate.NewPredicateFuncs(func(obj client.Object) bool {
			daemonset := obj.(*appsv1.DaemonSet)
			for _, namespace := range config.IgnoreNamespaces {
				if namespace == daemonset.Namespace {
					return false
				}
			}

			return true
		}),
	); err != nil {
		log.Error(err, "failed to watch DaemonSet")
	}

	// Starting the controller manager
	log.Info("starting the controller manager")

	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "failed to start controller manager")
		os.Exit(1)
	}
}

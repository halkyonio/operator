package framework

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"halkyon.io/operator/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strings"
)

func NewGenericReconciler(resource Resource) *GenericReconciler {
	return &GenericReconciler{resource: resource}
}

type GenericReconciler struct {
	resource Resource
	helper   *K8SHelper
}

func (b *GenericReconciler) watchedSecondaryResourcesTypes() []runtime.Object {
	resources := b.resource.GetDependentResourcesTypes()
	watched := make([]runtime.Object, 0, len(resources))
	for _, dep := range resources {
		if dep.ShouldWatch() {
			watched = append(watched, dep.Prototype())
		}
	}
	return watched
}

func (b *GenericReconciler) Helper() *K8SHelper {
	if b.helper == nil {
		b.helper = GetHelperFor(b.resource.PrimaryResourceType())
	}
	return b.helper
}

func (b *GenericReconciler) logger() logr.Logger {
	return b.Helper().ReqLogger
}

func (b *GenericReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	b.logger().WithValues("namespace", request.Namespace)

	// Fetch the primary resource
	resource, err := b.resource.FetchAndCreateNew(request.Name, request.Namespace)
	typeName := util.GetObjectName(b.resource.PrimaryResourceType())
	if err != nil {
		if errors.IsNotFound(err) {
			// Return and don't create
			if resource.ShouldDelete() {
				b.logger().Info(typeName + " resource is marked for deletion. Running clean-up.")
				err := resource.Delete()
				return reconcile.Result{Requeue: resource.NeedsRequeue()}, err
			}
			return reconcile.Result{}, nil
		}
		// Error reading the object - create the request.
		b.logger().Error(err, "failed to get "+typeName)
		return reconcile.Result{}, err
	}

	initialStatus := resource.GetStatusAsString()
	if resource.GetGeneration() == 1 && len(initialStatus) == 0 {
		resource.SetInitialStatus("Initializing")
	}

	if resource.Init() {
		if e := b.Helper().Client.Update(context.Background(), resource.GetAPIObject()); e != nil {
			b.logger().Error(e, fmt.Sprintf("failed to update '%s' %s", resource.GetName(), typeName))
		}
		return reconcile.Result{}, nil
	}

	if err := resource.CheckValidity(); err != nil {
		b.updateStatusIfNeeded(resource, err)
		return reconcile.Result{}, nil
	}

	b.logger().Info("-> "+typeName, "name", resource.GetName(), "status", initialStatus)

	err = resource.CreateOrUpdate()
	if err != nil {
		err = fmt.Errorf("failed to create or update %s '%s': %s", typeName, resource.GetName(), err.Error())
	}

	// always check status for updates
	b.updateStatusIfNeeded(resource, err)

	requeue := resource.NeedsRequeue()

	// only log exit if status changed to avoid being too verbose
	newStatus := resource.GetStatusAsString()
	if newStatus != initialStatus {
		msg := "<- " + typeName
		if requeue {
			msg += " (requeued)"
		}
		b.logger().Info(msg, "name", resource.GetName(), "status", newStatus)
	}
	return reconcile.Result{Requeue: requeue}, err
}

func (b *GenericReconciler) updateStatusIfNeeded(instance Resource, err error) {
	// update the resource if the status has changed
	updateStatus := false
	if err == nil {
		updateStatus = instance.ComputeStatus()
	} else {
		updateStatus = instance.SetErrorStatus(err)
		b.logger().Error(err, fmt.Sprintf("'%s' %s has an error", instance.GetName(), util.GetObjectName(instance.GetAPIObject())))
	}
	if updateStatus {
		object := instance.GetAPIObject()
		if e := b.Helper().Client.Status().Update(context.Background(), object); e != nil {
			b.logger().Error(e, fmt.Sprintf("failed to update status for '%s' %s", instance.GetName(), util.GetObjectName(object)))
		}
	}
}

func RegisterNewReconciler(resource Resource, mgr manager.Manager) error {
	resourceType := resource.PrimaryResourceType()

	// Create a new controller
	controllerName := controllerNameFor(resourceType)
	reconciler := NewGenericReconciler(resource)
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: reconciler})
	if err != nil {
		return err
	}

	// Register helper
	registerHelper(controllerName, resourceType, mgr)

	// Watch for changes to primary resource
	if err = c.Watch(&source.Kind{Type: resourceType}, &handler.EnqueueRequestForObject{}); err != nil {
		return err
	}

	// Watch for changes of child/secondary resources
	owner := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    resourceType,
	}
	for _, t := range reconciler.watchedSecondaryResourcesTypes() {
		if err = c.Watch(&source.Kind{Type: t}, owner); err != nil {
			return err
		}
	}

	return nil
}

func controllerNameFor(resource runtime.Object) string {
	return strings.ToLower(util.GetObjectName(resource)) + "-controller"
}

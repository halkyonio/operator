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

func NewBaseGenericReconciler(primaryResourceManager PrimaryResourceManager) *BaseGenericReconciler {
	return &BaseGenericReconciler{resourceManager: primaryResourceManager}
}

type BaseGenericReconciler struct {
	resourceManager PrimaryResourceManager
}

func (b *BaseGenericReconciler) watchedSecondaryResourcesTypes() []runtime.Object {
	resources := b.resourceManager.GetDependentResourcesTypes()
	watched := make([]runtime.Object, 0, len(resources))
	for _, dep := range resources {
		if dep.ShouldWatch() {
			watched = append(watched, dep.Prototype())
		}
	}
	return watched
}

func (b *BaseGenericReconciler) Helper() *K8SHelper {
	return b.resourceManager.Helper()
}

func (b *BaseGenericReconciler) logger() logr.Logger {
	return b.Helper().ReqLogger
}

func (b *BaseGenericReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	b.logger().WithValues("namespace", request.Namespace)

	// Fetch the primary resource
	resource, err := b.resourceManager.NewFrom(request.Name, request.Namespace)
	typeName := util.GetObjectName(b.resourceManager.PrimaryResourceType())
	if err != nil {
		if errors.IsNotFound(err) {
			// Return and don't create
			if resource.ShouldDelete() {
				b.logger().Info(typeName + " resource is marked for deletion. Running clean-up.")
				err := b.resourceManager.Delete(resource)
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
		return reconcile.Result{Requeue: true}, err
	}

	b.logger().Info("-> "+typeName, "name", resource.GetName(), "status", initialStatus)

	err = b.resourceManager.CreateOrUpdate(resource)
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

func (b *BaseGenericReconciler) updateStatusIfNeeded(instance Resource, err error) {
	// compute the status and update the resource if the status has changed
	if needsStatusUpdate := instance.ComputeStatus(err, b.Helper()); needsStatusUpdate {
		object := instance.GetAPIObject()
		if e := b.Helper().Client.Status().Update(context.Background(), object); e != nil {
			b.logger().Error(e, fmt.Sprintf("failed to update status for '%s' %s", instance.GetName(), util.GetObjectName(object)))
		}
	}
}

func RegisterNewReconciler(factory PrimaryResourceManager, mgr manager.Manager) error {
	resourceType := factory.PrimaryResourceType()

	// Create a new controller
	controllerName := controllerNameFor(resourceType)
	reconciler := NewBaseGenericReconciler(factory)
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: reconciler})
	if err != nil {
		return err
	}

	// Create helper and set it on the resource manager
	helper := NewHelper(controllerName, mgr)
	factory.SetHelper(helper)

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

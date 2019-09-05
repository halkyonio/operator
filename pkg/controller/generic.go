package controller

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"halkyon.io/operator/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strings"
)

type ReconcilerFactory interface {
	PrimaryResourceType() Resource
	WatchedSecondaryResourceTypes() []runtime.Object
	Delete(object Resource) error
	CreateOrUpdate(object Resource) error
	IsDependentResourceReady(resource Resource) (depOrTypeName string, ready bool)
	Helper() ReconcilerHelper
	GetDependentResourceFor(owner Resource, resourceType runtime.Object) (DependentResource, error)
	AddDependentResource(resource DependentResource)
}

type ReconcilerHelper struct {
	Client    client.Client
	Config    *rest.Config
	Scheme    *runtime.Scheme
	ReqLogger logr.Logger
}

func (rh ReconcilerHelper) Fetch(name, namespace string, into runtime.Object) (runtime.Object, error) {
	if err := rh.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, into); err != nil {
		if errors.IsNotFound(err) {
			return into, err
		}
		return into, fmt.Errorf("couldn't fetch '%s' %s from namespace '%s': %s", name, GetObjectName(into), namespace, err.Error())
	}
	return into, nil
}

func NewBaseGenericReconciler(primaryResourceType Resource, mgr manager.Manager) *BaseGenericReconciler {
	return &BaseGenericReconciler{
		ReconcilerHelper: newHelper(primaryResourceType, mgr),
		dependents:       make(map[string]DependentResource, 7),
		primary:          primaryResourceType,
	}
}

func (b *BaseGenericReconciler) SetReconcilerFactory(factory ReconcilerFactory) {
	b._factory = factory
}

type BaseGenericReconciler struct {
	ReconcilerHelper
	dependents       map[string]DependentResource
	primary          Resource
	_factory         ReconcilerFactory
	onOpenShift      *bool
	openShiftVersion int
}

func (b *BaseGenericReconciler) OpenShiftVersion() int {
	b.IsTargetClusterRunningOpenShift() // make sure things are properly initialized
	return b.openShiftVersion
}

func (b *BaseGenericReconciler) IsTargetClusterRunningOpenShift() bool {
	if b.onOpenShift == nil {
		discoveryClient, err := discovery.NewDiscoveryClientForConfig(b.Config)
		if err != nil {
			panic(err)
		}
		apiList, err := discoveryClient.ServerGroups()
		if err != nil {
			panic(err)
		}
		apiGroups := apiList.Groups
		const openShiftGroupSuffix = ".openshift.io"
		const openShift4GroupName = "config" + openShiftGroupSuffix
		for _, group := range apiGroups {
			if strings.HasSuffix(group.Name, openShiftGroupSuffix) {
				if b.onOpenShift == nil {
					b.onOpenShift = util.NewTrue()
					b.openShiftVersion = 3
				}
				if group.Name == openShift4GroupName {
					b.openShiftVersion = 4
					break
				}
			}
		}

		if b.onOpenShift == nil {
			// we didn't find any api group with the openshift.io suffix, so we're not on OpenShift!
			b.onOpenShift = util.NewFalse()
			b.openShiftVersion = 0
		}
	}

	return *b.onOpenShift
}

func (b *BaseGenericReconciler) computeStatus(current Resource, err error) bool {
	depOrTypeName, ready := b.IsDependentResourceReady(current)
	if !ready {
		errMsg := ""
		if err != nil {
			errMsg = ": " + err.Error()
		}
		msg := fmt.Sprintf("%s is not ready for %s '%s' in namespace '%s'%s",
			depOrTypeName, GetObjectName(current), current.GetName(), current.GetNamespace(), errMsg)
		b.ReqLogger.Info(msg)
		return current.SetInitialStatus(msg)
	}

	return current.SetSuccessStatus(depOrTypeName, "Ready")
}

func (b *BaseGenericReconciler) PrimaryResourceType() Resource {
	return b.primary.Clone()
}

func (b *BaseGenericReconciler) factory() ReconcilerFactory {
	if b._factory == nil {
		panic(fmt.Errorf("factory needs to be set on BaseGenericReconciler before use"))
	}
	return b._factory
}

func (b *BaseGenericReconciler) primaryResourceTypeName() string {
	return GetObjectName(b.primary)
}

func (b *BaseGenericReconciler) WatchedSecondaryResourceTypes() []runtime.Object {
	watched := make([]runtime.Object, 0, len(b.dependents))
	for _, dep := range b.dependents {
		if dep.ShouldWatch() {
			watched = append(watched, dep.Prototype())
		}
	}
	return watched
}

func (b *BaseGenericReconciler) Delete(object Resource) error {
	return b.factory().Delete(object)
}

func (b *BaseGenericReconciler) Fetch(into Resource) (Resource, error) {
	object, e := b.Helper().Fetch(into.GetName(), into.GetNamespace(), into.GetAPIObject())
	if e != nil {
		return into, e
	}
	into.SetAPIObject(object)
	return into, nil
}

func (b *BaseGenericReconciler) CreateOrUpdate(object Resource) error {
	return b.factory().CreateOrUpdate(object)
}

func (b *BaseGenericReconciler) Helper() ReconcilerHelper {
	return b.ReconcilerHelper
}

func getKeyFor(resourceType runtime.Object) (key string) {
	t := reflect.TypeOf(resourceType)
	pkg := t.PkgPath()
	kind := GetObjectName(resourceType)
	key = pkg + "/" + kind
	return
}

func (b *BaseGenericReconciler) AddDependentResource(resource DependentResource) {
	prototype := resource.Prototype()
	key := getKeyFor(prototype)
	b.dependents[key] = resource
}

func (b *BaseGenericReconciler) MustGetDependentResourceFor(owner Resource, resourceType runtime.Object) (resource DependentResource) {
	var e error
	if resource, e = b.GetDependentResourceFor(owner, resourceType); e != nil {
		panic(e)
	}
	return resource
}

func (b *BaseGenericReconciler) GetDependentResourceFor(owner Resource, resourceType runtime.Object) (DependentResource, error) {
	resource, ok := b.dependents[getKeyFor(resourceType)]
	if !ok {
		return nil, fmt.Errorf("couldn't find any dependent resource of kind '%s'", GetObjectName(resourceType))
	}
	return resource.NewInstanceWith(owner), nil
}

func (b *BaseGenericReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	b.ReqLogger.WithValues("namespace", request.Namespace)

	// Fetch the primary resource
	resource := b.PrimaryResourceType()
	resource.SetName(request.Name)
	resource.SetNamespace(request.Namespace)
	typeName := b.primaryResourceTypeName()
	resource, err := b.Fetch(resource)
	if err != nil {
		if errors.IsNotFound(err) {
			// Return and don't create
			b.ReqLogger.Info(typeName + " resource not found.")
			if resource.ShouldDelete() {
				b.ReqLogger.Info(typeName + " resource is marked for deletion. Running clean-up.")
				err := b.Delete(resource)
				return reconcile.Result{Requeue: resource.NeedsRequeue()}, err
			}
			return reconcile.Result{}, nil
		}
		// Error reading the object - create the request.
		b.ReqLogger.Error(err, "failed to get "+typeName)
		return reconcile.Result{}, err
	}

	if resource.GetGeneration() == 1 && len(resource.GetStatusAsString()) == 0 {
		resource.SetInitialStatus("Initializing")
	}

	if !resource.IsValid() {
		return reconcile.Result{Requeue: true}, nil
	}

	b.ReqLogger.Info("==> Reconciling "+typeName,
		"name", resource.GetName(),
		"status", resource.GetStatusAsString(),
		"created", resource.GetCreationTimestamp())

	err = b.CreateOrUpdate(resource)
	if err != nil {
		err = fmt.Errorf("failed to create or update %s '%s': %s", typeName, resource.GetName(), err.Error())
	}

	// always check status for updates
	b.updateStatusIfNeeded(resource, err)

	b.ReqLogger.Info("<== Reconciled "+typeName, "name", resource.GetName())
	return reconcile.Result{Requeue: resource.NeedsRequeue()}, err
}

func (b *BaseGenericReconciler) updateStatusIfNeeded(instance Resource, err error) {
	// compute the status and update the resource if the status has changed
	if needsStatusUpdate := b.computeStatus(instance, err); needsStatusUpdate {
		if e := b.Client.Status().Update(context.Background(), instance.GetAPIObject()); e != nil {
			b.ReqLogger.Error(e, "failed to update status for component "+instance.GetName())
		}
	}
}

func newHelper(resourceType runtime.Object, mgr manager.Manager) ReconcilerHelper {
	helper := ReconcilerHelper{
		Client:    mgr.GetClient(),
		Config:    mgr.GetConfig(),
		Scheme:    mgr.GetScheme(),
		ReqLogger: logf.Log.WithName(controllerNameFor(resourceType)),
	}
	return helper
}

type GenericReconciler interface {
	ReconcilerFactory
	reconcile.Reconciler
}

func RegisterNewReconciler(factory GenericReconciler, mgr manager.Manager) error {
	resourceType := factory.PrimaryResourceType().GetAPIObject()

	// Create a new controller
	c, err := controller.New(controllerNameFor(resourceType), mgr, controller.Options{Reconciler: factory})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource
	if err = c.Watch(&source.Kind{Type: resourceType}, &handler.EnqueueRequestForObject{}); err != nil {
		return err
	}

	// Watch for changes of child/secondary resources
	owner := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    resourceType,
	}

	for _, t := range factory.WatchedSecondaryResourceTypes() {
		if err = c.Watch(&source.Kind{Type: t}, owner); err != nil {
			return err
		}
	}

	return nil
}

func controllerNameFor(resource runtime.Object) string {
	return strings.ToLower(GetObjectName(resource)) + "-controller"
}

func (b *BaseGenericReconciler) CreateIfNeeded(owner Resource, resourceType runtime.Object) error {
	resource, err := b.GetDependentResourceFor(owner, resourceType)
	if err != nil {
		return err
	}

	// if the resource specifies that it shouldn't be created, exit fast
	if !resource.CanBeCreatedOrUpdated() {
		return nil
	}

	kind := GetObjectName(resourceType)
	res, err := resource.Fetch(b.Helper())
	if err != nil {
		if errors.IsNotFound(err) {
			// create the object
			obj, errBuildObject := resource.Build()
			if errBuildObject != nil {
				return errBuildObject
			}

			// set controller reference if the resource should be owned
			if resource.ShouldBeOwned() {
				// in most instances, resourceDefinedOwner == owner but some resources might want to return a different one
				resourceDefinedOwner := resource.Owner()
				if e := controllerutil.SetControllerReference(resourceDefinedOwner.GetAPIObject().(v1.Object), obj.(v1.Object), b.Scheme); e != nil {
					b.ReqLogger.Error(err, "Failed to set owner", "owner", resourceDefinedOwner, "resource", resource.Name())
					return e
				}
			}

			alreadyExists := false
			if err = b.Client.Create(context.TODO(), obj); err != nil {
				// ignore error if it's to state that obj already exists
				alreadyExists = errors.IsAlreadyExists(err)
				if !alreadyExists {
					b.ReqLogger.Error(err, "Failed to create new ", "kind", kind)
					return err
				}
			}
			if !alreadyExists {
				b.ReqLogger.Info("Created successfully", "kind", kind, "name", obj.(v1.Object).GetName())
			}
			return nil
		}
		b.ReqLogger.Error(err, "Failed to get", "kind", kind)
		return err
	} else {
		// if the resource defined an updater, use it to try to update the resource
		updated, err := resource.Update(res)
		if err != nil {
			return err
		}
		if updated {
			if err = b.Client.Update(context.TODO(), res); err != nil {
				b.ReqLogger.Error(err, "Failed to update", "kind", kind)
			}
		}
		return err
	}
}

func (b *BaseGenericReconciler) IsDependentResourceReady(resource Resource) (depOrTypeName string, ready bool) {
	return b.factory().IsDependentResourceReady(resource)
}

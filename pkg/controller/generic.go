package controller

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"github.com/snowdrop/component-operator/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

type DependentResource interface {
	Name() string
	Fetch(helper ReconcilerHelper) (v1.Object, error)
	Build() (runtime.Object, error)
	Update(toUpdate v1.Object) (bool, error)
	NewInstanceWith(owner v1.Object) DependentResource
	Owner() v1.Object
	Prototype() runtime.Object
	ShouldWatch() bool
}

type DependentResourceHelper struct {
	_owner     v1.Object
	_prototype runtime.Object
	_delegate  DependentResource
}

func (res DependentResourceHelper) ShouldWatch() bool {
	return true
}

func NewDependentResource(primaryResourceType runtime.Object, owner v1.Object) *DependentResourceHelper {
	return &DependentResourceHelper{_prototype: primaryResourceType, _owner: owner}
}

func (res *DependentResourceHelper) SetDelegate(delegate DependentResource) {
	res._delegate = delegate
}

func (res DependentResourceHelper) Name() string {
	return res._owner.GetName()
}

func (res DependentResourceHelper) Fetch(helper ReconcilerHelper) (v1.Object, error) {
	delegate := res._delegate
	into := delegate.Prototype()
	if err := helper.Client.Get(context.TODO(), types.NamespacedName{Name: delegate.Name(), Namespace: delegate.Owner().GetNamespace()}, into); err != nil {
		return nil, err
	}
	return into.(v1.Object), nil
}

func (res DependentResourceHelper) Owner() v1.Object {
	return res._owner
}

func (res DependentResourceHelper) Prototype() runtime.Object {
	return res._prototype.DeepCopyObject()
}

type ReconcilerFactory interface {
	PrimaryResourceType() v1alpha2.Resource
	AsResource(object runtime.Object) v1alpha2.Resource
	WatchedSecondaryResourceTypes() []runtime.Object
	Delete(object runtime.Object) (bool, error)
	CreateOrUpdate(object runtime.Object) (bool, error)
	ComputeStatus(current v1alpha2.Resource, err error) (statusChanged bool, needRequeue bool)
	IsDependentResourceReady(resource v1alpha2.Resource) (depOrTypeName string, ready bool)
	Helper() ReconcilerHelper
	GetDependentResourceFor(owner v1.Object, resourceType runtime.Object) (DependentResource, error)
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
		return nil, fmt.Errorf("couldn't fetch '%s' %s from namespace '%s'", name, util.GetObjectName(into), namespace)
	}
	return into, nil
}

func NewBaseGenericReconciler(primaryResourceType v1alpha2.Resource, mgr manager.Manager) *BaseGenericReconciler {
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
	dependents map[string]DependentResource
	primary    runtime.Object
	_factory   ReconcilerFactory
}

func (b *BaseGenericReconciler) ComputeStatus(current v1alpha2.Resource, err error) (bool, bool) {
	depOrTypeName, ready := b.IsDependentResourceReady(current)
	if !ready {
		return b.MakePending(depOrTypeName, current)
	}

	return current.SetSuccessStatus(depOrTypeName, "Ready"), false
}

func (b *BaseGenericReconciler) PrimaryResourceType() v1alpha2.Resource {
	return b.AsResource(b.primary.DeepCopyObject())
}

func (b *BaseGenericReconciler) AsResource(object runtime.Object) v1alpha2.Resource {
	return b.factory().AsResource(object)
}

func (b *BaseGenericReconciler) factory() ReconcilerFactory {
	if b._factory == nil {
		panic(fmt.Errorf("factory needs to be set on BaseGenericReconciler before use"))
	}
	return b._factory
}

func (b *BaseGenericReconciler) primaryResourceTypeName() string {
	return util.GetObjectName(b.primary)
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

func (b *BaseGenericReconciler) Delete(object runtime.Object) (bool, error) {
	return b.factory().Delete(object)
}

func (b *BaseGenericReconciler) Fetch(into v1alpha2.Resource) (v1alpha2.Resource, error) {
	object, e := b.Helper().Fetch(into.GetName(), into.GetNamespace(), into)
	if e != nil {
		return nil, e
	}
	return b.AsResource(object), nil
}

func (b *BaseGenericReconciler) MakePending(dependencyName string, resource v1alpha2.Resource) (changed bool, wantsRequeue bool) {
	msg := fmt.Sprintf("'%s' is not ready for %s '%s' in namespace '%s'",
		dependencyName, util.GetObjectName(resource), resource.GetName(), resource.GetNamespace())
	b.ReqLogger.Info(msg)
	resource.SetInitialStatus(msg)
	return true, true
}

func (b *BaseGenericReconciler) CreateOrUpdate(object runtime.Object) (bool, error) {
	return b.factory().CreateOrUpdate(object)
}

func (b *BaseGenericReconciler) Helper() ReconcilerHelper {
	return b.ReconcilerHelper
}

func getKeyFor(resourceType runtime.Object) (key string) {
	t := reflect.TypeOf(resourceType)
	pkg := t.PkgPath()
	kind := util.GetObjectName(resourceType)
	key = pkg + "/" + kind
	return
}

func (b *BaseGenericReconciler) AddDependentResource(resource DependentResource) {
	prototype := resource.Prototype()
	key := getKeyFor(prototype)
	b.dependents[key] = resource
}

func (b *BaseGenericReconciler) GetDependentResourceFor(owner v1.Object, resourceType runtime.Object) (DependentResource, error) {
	resource, ok := b.dependents[getKeyFor(resourceType)]
	if !ok {
		return nil, fmt.Errorf("couldn't find any dependent resource of kind '%s'", util.GetObjectName(resourceType))
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
	err := b.Client.Get(context.TODO(), request.NamespacedName, resource)
	if err != nil {
		if errors.IsNotFound(err) {
			// Return and don't create
			b.ReqLogger.Info(typeName + " resource not found.")
			if resource.ShouldDelete() {
				b.ReqLogger.Info(typeName + " resource is marked for deletion. Running clean-up.")
				requeue, err := b.Delete(resource)
				return reconcile.Result{Requeue: requeue}, err
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

	changed, err := b.CreateOrUpdate(resource)
	if err != nil {
		err = fmt.Errorf("failed to create or update %s '%s': %s", typeName, resource.GetName(), err.Error())
	}

	b.ReqLogger.Info("<== Reconciled "+typeName, "name", resource.GetName())
	changed = changed || b.updateStatus(resource, err)
	return reconcile.Result{Requeue: changed}, err
}

func (b *BaseGenericReconciler) updateStatus(instance v1alpha2.Resource, err error) bool {
	// fetch latest version to avoid optimistic lock error
	current := instance
	var e error
	current, e = b.Fetch(instance)
	if e != nil {
		b.ReqLogger.Error(e, fmt.Sprintf("failed to fetch latest version of '%s' %s", instance.GetName(), util.GetObjectName(instance)))
	}

	// compute the status and update the resource if the status has changed
	if needsStatusUpdate, needsRequeue := b.ComputeStatus(current, err); needsStatusUpdate {
		e = b.Client.Status().Update(context.TODO(), current)
		if e != nil {
			b.ReqLogger.Error(e, "failed to update status for component "+current.GetName())
		}
		return needsRequeue
	}
	return false
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
	resourceType := factory.PrimaryResourceType()

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
	return strings.ToLower(util.GetObjectName(resource)) + "-controller"
}

func (b *BaseGenericReconciler) CreateIfNeeded(owner v1.Object, resourceType runtime.Object) (bool, error) {
	resource, err := b.GetDependentResourceFor(owner, resourceType)
	if err != nil {
		return false, err
	}

	kind := util.GetObjectName(resourceType)
	res, err := resource.Fetch(b.Helper())
	if err != nil {
		// create the object
		obj, errBuildObject := resource.Build()
		if errBuildObject != nil {
			return false, errBuildObject
		}
		if errors.IsNotFound(err) {
			controllerutil.SetControllerReference(resource.Owner(), obj.(v1.Object), b.Scheme)
			err = b.Client.Create(context.TODO(), obj)
			if err != nil {
				b.ReqLogger.Error(err, "Failed to create new ", "kind", kind)
				return false, err
			}
			b.ReqLogger.Info("Created successfully", "kind", kind)
			return true, nil
		}
		b.ReqLogger.Error(err, "Failed to get", "kind", kind)
		return false, err
	} else {
		// if the resource defined an updater, use it to try to update the resource
		return resource.Update(res)
	}
}

func (b *BaseGenericReconciler) IsDependentResourceReady(resource v1alpha2.Resource) (depOrTypeName string, ready bool) {
	return b.factory().IsDependentResourceReady(resource)
}

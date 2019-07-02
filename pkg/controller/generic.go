package controller

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type ResourceMetadata struct {
	Name         string
	Status       string
	Created      v1.Time
	ShouldDelete bool
}

type DependentResource interface {
	Name() string
	Fetch(helper ReconcilerHelper) (v1.Object, error)
	Build() (runtime.Object, error)
	Update(toUpdate v1.Object) (bool, error)
	NewInstanceWith(owner v1.Object) DependentResource
	Owner() v1.Object
	Prototype() runtime.Object
	AsObject(object runtime.Object) v1.Object
}

type BaseDependentResource struct {
	_owner     v1.Object
	_prototype runtime.Object
}

func NewDependentResource(primaryResourceType runtime.Object) BaseDependentResource {
	return BaseDependentResource{_prototype: primaryResourceType}
}

func (res BaseDependentResource) AsObject(object runtime.Object) v1.Object {
	panic("implement me")
}

func (res BaseDependentResource) Build() (runtime.Object, error) {
	panic("implement me")
}

func (res BaseDependentResource) Update(toUpdate v1.Object) (bool, error) {
	panic("implement me")
}

func (res BaseDependentResource) Name() string {
	return res._owner.GetName()
}

func (res BaseDependentResource) Fetch(helper ReconcilerHelper) (v1.Object, error) {
	into := res.Prototype()
	if err := helper.Client.Get(context.TODO(), types.NamespacedName{Name: res.Name(), Namespace: res.Owner().GetNamespace()}, into); err != nil {
		return nil, err
	}
	return res.AsObject(into), nil
}

func (res BaseDependentResource) NewInstanceWith(owner v1.Object) DependentResource {
	return BaseDependentResource{_owner: owner, _prototype: res._prototype}
}

func (res BaseDependentResource) Owner() v1.Object {
	return res._owner
}

func (res BaseDependentResource) Prototype() runtime.Object {
	return res._prototype.DeepCopyObject()
}

type ReconcilerFactory interface {
	PrimaryResourceType() runtime.Object
	SecondaryResourceTypes() []runtime.Object
	IsPrimaryResourceValid(object runtime.Object) bool
	ResourceMetadata(object runtime.Object) ResourceMetadata
	Delete(object runtime.Object) (bool, error)
	CreateOrUpdate(object runtime.Object) (bool, error)
	SetErrorStatus(object runtime.Object, e error)
	SetSuccessStatus(object runtime.Object)
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

type GenericReconciler struct {
	ReconcilerHelper
	ReconcilerFactory
}

func NewBaseGenericReconciler(primaryResourceType runtime.Object, mgr manager.Manager) *BaseGenericReconciler {
	return &BaseGenericReconciler{
		ReconcilerHelper: NewHelper(primaryResourceType, mgr),
		dependents:       make(map[runtime.Object]DependentResource, 7),
		primary:          primaryResourceType,
	}
}

type BaseGenericReconciler struct {
	ReconcilerHelper
	dependents map[runtime.Object]DependentResource
	primary    runtime.Object
}

func (b *BaseGenericReconciler) PrimaryResourceType() runtime.Object {
	return b.primary
}

func (b *BaseGenericReconciler) SecondaryResourceTypes() []runtime.Object {
	panic("implement me")
}

func (b *BaseGenericReconciler) IsPrimaryResourceValid(object runtime.Object) bool {
	//todo: implement
	return true
}

func (b *BaseGenericReconciler) ResourceMetadata(object runtime.Object) ResourceMetadata {
	panic("implement me")
}

func (b *BaseGenericReconciler) Delete(object runtime.Object) (bool, error) {
	panic("implement me")
}

func (b *BaseGenericReconciler) CreateOrUpdate(object runtime.Object) (bool, error) {
	panic("implement me")
}

func (b *BaseGenericReconciler) SetErrorStatus(object runtime.Object, e error) {
	panic("implement me")
}

func (b *BaseGenericReconciler) SetSuccessStatus(object runtime.Object) {
	panic("implement me")
}

func (b *BaseGenericReconciler) Helper() ReconcilerHelper {
	return b.ReconcilerHelper
}

func (b *BaseGenericReconciler) AddDependentResource(resource DependentResource) {
	b.dependents[resource.Prototype()] = resource
}

func (b *BaseGenericReconciler) GetDependentResourceFor(owner v1.Object, resourceType runtime.Object) (DependentResource, error) {
	resource, ok := b.dependents[resourceType]
	if !ok {
		return nil, fmt.Errorf("couldn't find any dependent resource of kind '%s'", resourceType.GetObjectKind().GroupVersionKind().Kind)
	}
	return resource.NewInstanceWith(owner), nil
}

func (g *GenericReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	g.ReqLogger.WithValues("namespace", request.Namespace)

	// Fetch the primary resource
	resource := g.PrimaryResourceType()
	typeName := resource.GetObjectKind().GroupVersionKind().Kind
	err := g.Client.Get(context.TODO(), request.NamespacedName, resource)
	if err != nil {
		if errors.IsNotFound(err) {
			// Return and don't create
			g.ReqLogger.Info(typeName + " resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		// Error reading the object - create the request.
		g.ReqLogger.Error(err, "failed to get "+typeName)
		return reconcile.Result{}, err
	}

	if !g.IsPrimaryResourceValid(resource) {
		return reconcile.Result{Requeue: true}, nil
	}

	metadata := g.ResourceMetadata(resource)
	g.ReqLogger.Info("==> Reconciling "+typeName,
		"name", metadata.Name,
		"status", metadata.Status,
		"created", metadata.Created)

	if metadata.ShouldDelete {
		requeue, err := g.Delete(resource)
		return reconcile.Result{Requeue: requeue}, err
	}

	changed, err := g.CreateOrUpdate(resource)
	if err != nil {
		g.ReqLogger.Error(err, fmt.Sprintf("failed to create or update %s '%s'", typeName, metadata.Name))
		g.SetErrorStatus(resource, err)
		return reconcile.Result{}, err
	}

	g.ReqLogger.Info("<== Reconciled "+typeName, "name", metadata.Name)
	g.SetSuccessStatus(resource)
	return reconcile.Result{Requeue: changed}, nil
}

func NewGenericReconciler(rf ReconcilerFactory) *GenericReconciler {
	reconciler := &GenericReconciler{
		ReconcilerHelper:  rf.Helper(),
		ReconcilerFactory: rf,
	}

	return reconciler
}

func NewHelper(resourceType runtime.Object, mgr manager.Manager) ReconcilerHelper {
	helper := ReconcilerHelper{
		Client:    mgr.GetClient(),
		Config:    mgr.GetConfig(),
		Scheme:    mgr.GetScheme(),
		ReqLogger: logf.Log.WithName(controllerNameFor(resourceType)),
	}
	return helper
}

func RegisterNewReconciler(factory ReconcilerFactory, mgr manager.Manager) error {
	r := NewGenericReconciler(factory)

	resourceType := factory.PrimaryResourceType()

	// Create a new controller
	c, err := controller.New(controllerNameFor(resourceType), mgr, controller.Options{Reconciler: r})
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

	for _, t := range factory.SecondaryResourceTypes() {
		if err = c.Watch(&source.Kind{Type: t}, owner); err != nil {
			return err
		}
	}

	return nil
}

func controllerNameFor(resource runtime.Object) string {
	return resource.GetObjectKind().GroupVersionKind().Kind + "-controller"
}

func (b *BaseGenericReconciler) CreateIfNeeded(owner v1.Object, resourceType runtime.Object) (bool, error) {
	resource, err := b.GetDependentResourceFor(owner, resourceType)
	kind := resourceType.GetObjectKind().GroupVersionKind().Kind
	if err != nil {
		return false, fmt.Errorf("unknown dependent type %s", kind)
	}

	res, err := resource.Fetch(b.Helper())
	if err != nil {
		// create the object
		obj, errBuildObject := resource.Build()
		if errBuildObject != nil {
			return false, errBuildObject
		}
		if errors.IsNotFound(err) {
			err = b.Client.Create(context.TODO(), obj)
			if err != nil {
				b.ReqLogger.Error(err, "Failed to create new ", "kind", kind)
				return false, err
			}
			b.ReqLogger.Info("Created successfully", "kind", kind)
			return true, controllerutil.SetControllerReference(resource.Owner(), res, b.Scheme)
		}
		b.ReqLogger.Error(err, "Failed to get", "kind", kind)
		return false, err
	} else {
		// if the resource defined an updater, use it to try to update the resource
		return resource.Update(res)
	}
}

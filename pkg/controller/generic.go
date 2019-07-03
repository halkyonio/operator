package controller

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
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

func (rh ReconcilerHelper) Fetch(name, namespace string, into runtime.Object) (runtime.Object, error) {
	if err := rh.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, into); err != nil {
		return nil, fmt.Errorf("couldn't fetch '%s' %s from namespace '%s'", name, util.GetObjectName(into), namespace)
	}
	return into, nil
}

func NewBaseGenericReconciler(primaryResourceType runtime.Object, secondaryResourceTypes []runtime.Object, mgr manager.Manager) *BaseGenericReconciler {
	return &BaseGenericReconciler{
		ReconcilerHelper: newHelper(primaryResourceType, mgr),
		dependents:       make(map[string]DependentResource, 7),
		primary:          primaryResourceType,
		secondary:        secondaryResourceTypes,
	}
}

func (b *BaseGenericReconciler) SetReconcilerFactory(factory ReconcilerFactory) {
	b._factory = factory
}

type BaseGenericReconciler struct {
	ReconcilerHelper
	dependents map[string]DependentResource
	primary    runtime.Object
	secondary  []runtime.Object
	_factory   ReconcilerFactory
}

func (b *BaseGenericReconciler) factory() ReconcilerFactory {
	if b._factory == nil {
		panic(fmt.Errorf("factory needs to be set on BaseGenericReconciler before use"))
	}
	return b._factory
}

func (b *BaseGenericReconciler) PrimaryResourceType() runtime.Object {
	return b.primary.DeepCopyObject()
}

func (b *BaseGenericReconciler) primaryResourceTypeName() string {
	return util.GetObjectName(b.primary)
}

func (b *BaseGenericReconciler) SecondaryResourceTypes() []runtime.Object {
	return b.secondary
}

func (b *BaseGenericReconciler) IsPrimaryResourceValid(object runtime.Object) bool {
	return b.factory().IsPrimaryResourceValid(object)
}

func (b *BaseGenericReconciler) ResourceMetadata(object runtime.Object) ResourceMetadata {
	return b.factory().ResourceMetadata(object)
}

func (b *BaseGenericReconciler) Delete(object runtime.Object) (bool, error) {
	return b.factory().Delete(object)
}

func (b *BaseGenericReconciler) Fetch(name, namespace string) (runtime.Object, error) {
	into := b.PrimaryResourceType()
	return b.Helper().Fetch(name, namespace, into)
}

func (b *BaseGenericReconciler) CreateOrUpdate(object runtime.Object) (bool, error) {
	return b.factory().CreateOrUpdate(object)
}

func (b *BaseGenericReconciler) SetErrorStatus(object runtime.Object, e error) {
	b.factory().SetErrorStatus(object, e)
}

func (b *BaseGenericReconciler) SetSuccessStatus(object runtime.Object) {
	b.factory().SetSuccessStatus(object)
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
	typeName := b.primaryResourceTypeName()
	err := b.Client.Get(context.TODO(), request.NamespacedName, resource)
	if err != nil {
		if errors.IsNotFound(err) {
			// Return and don't create
			b.ReqLogger.Info(typeName + " resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		// Error reading the object - create the request.
		b.ReqLogger.Error(err, "failed to get "+typeName)
		return reconcile.Result{}, err
	}

	if !b.IsPrimaryResourceValid(resource) {
		return reconcile.Result{Requeue: true}, nil
	}

	metadata := b.ResourceMetadata(resource)
	b.ReqLogger.Info("==> Reconciling "+typeName,
		"name", metadata.Name,
		"status", metadata.Status,
		"created", metadata.Created)

	if metadata.ShouldDelete {
		requeue, err := b.Delete(resource)
		return reconcile.Result{Requeue: requeue}, err
	}

	changed, err := b.CreateOrUpdate(resource)
	if err != nil {
		b.ReqLogger.Error(err, fmt.Sprintf("failed to create or update %s '%s'", typeName, metadata.Name))
		b.SetErrorStatus(resource, err)
		return reconcile.Result{}, err
	}

	b.ReqLogger.Info("<== Reconciled "+typeName, "name", metadata.Name)
	b.SetSuccessStatus(resource)
	return reconcile.Result{Requeue: changed}, nil
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

	for _, t := range factory.SecondaryResourceTypes() {
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

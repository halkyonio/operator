package controller

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
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

type ReconcilerFactory interface {
	PrimaryResourceName() string
	PrimaryResourceType() runtime.Object
	SecondaryResourceTypes() []runtime.Object
	IsPrimaryResourceValid(object runtime.Object) bool
	ResourceMetadata(object runtime.Object) ResourceMetadata
	Delete(object runtime.Object) error
	CreateOrUpdate(object runtime.Object) (bool, error)
	SetErrorStatus(object runtime.Object, e error)
	SetSuccessStatus(object runtime.Object)
	Helper() ReconcilerHelper
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

func (g *GenericReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	g.ReqLogger.WithValues("namespace", request.Namespace)

	// Fetch the primary resource
	resource := g.PrimaryResourceType()
	typeName := g.PrimaryResourceName()
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
		err := g.Delete(resource)
		return reconcile.Result{}, err
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

func NewHelper(rf ReconcilerFactory, mgr manager.Manager) ReconcilerHelper {
	helper := ReconcilerHelper{
		Client:    mgr.GetClient(),
		Config:    mgr.GetConfig(),
		Scheme:    mgr.GetScheme(),
		ReqLogger: logf.Log.WithName(controllerNameFor(rf)),
	}
	return helper
}

func RegisterNewReconciler(factory ReconcilerFactory, mgr manager.Manager) error {
	r := NewGenericReconciler(factory)

	// Create a new controller
	c, err := controller.New(controllerNameFor(factory), mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource
	if err = c.Watch(&source.Kind{Type: factory.PrimaryResourceType()}, &handler.EnqueueRequestForObject{}); err != nil {
		return err
	}

	// Watch for changes of child/secondary resources
	owner := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    factory.PrimaryResourceType(),
	}

	for _, t := range factory.SecondaryResourceTypes() {
		if err = c.Watch(&source.Kind{Type: t}, owner); err != nil {
			return err
		}
	}

	return nil
}

func controllerNameFor(factory ReconcilerFactory) string {
	return factory.PrimaryResourceName() + "-controller"
}

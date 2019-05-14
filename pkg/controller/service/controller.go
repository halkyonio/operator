package service

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/snowdrop/component-operator/pkg/util"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strconv"
)

const (
	controllerName = "service-controller"
)

var log = logf.Log.WithName("service.controller")

// New creates a new Component Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func New(mgr manager.Manager) error {
	return Add(mgr, NewReconciler(mgr))
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func Add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Service
	err = c.Watch(&source.Kind{Type: &v1alpha2.Service{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// newReconciler returns a new reconcile.Reconciler
func NewReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileService{
		client:    mgr.GetClient(),
		config:    mgr.GetConfig(),
		scheme:    mgr.GetScheme(),
		reqLogger: log,
	}
}

type ReconcileService struct {
	client    client.Client
	config    *rest.Config
	scheme    *runtime.Scheme
	reqLogger logr.Logger
}

//Update the factory object and requeue
func (r *ReconcileService) update(obj runtime.Object) (reconcile.Result, error) {
	err := r.client.Update(context.TODO(), obj)
	if err != nil {
		r.reqLogger.Error(err, "Failed to update spec")
		return reconcile.Result{}, err
	}
	r.reqLogger.Info("Spec updated - return and create")
	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.reqLogger = log.WithValues("Namespace", request.Namespace)

	// Fetch the Service created, deleted or updated
	service := &v1alpha2.Service{}
	err := r.client.Get(context.TODO(), request.NamespacedName, service)
	if err != nil {
		return r.fetch(err)
	}

	r.reqLogger.Info("-----------------------")
	r.reqLogger.Info("Reconciling Service")
	r.reqLogger.Info("Status of the Service", "Status phase", service.Status.Phase)
	r.reqLogger.Info("Creation time          ", "Creation time", service.ObjectMeta.CreationTimestamp)
	r.reqLogger.Info("Resource version       ", "Resource version", service.ObjectMeta.ResourceVersion)
	r.reqLogger.Info("Generation version     ", "Generation version", strconv.FormatInt(service.ObjectMeta.Generation, 10))
	// r.reqLogger.Info("Deletion time          ","Deletion time", Service.ObjectMeta.DeletionTimestamp)

	_, err = util.IsOpenshift(r.config)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Add the Status Service when we process the first time the Service CR
	if service.Status.Phase == "" && service.Generation == 1 {
		// Update Status to value "Installing"
/*		err = r.updateStatusInstance(v1alpha2.PhaseServiceing, Service)
		if err != nil {
			r.reqLogger.Info("Status update failed !")
			return reconcile.Result{}, err
		}*/
	}

	// See finalizer doc for more info : https://book.kubebuilder.io/beyond_basics/using_finalizers.html
	// If DeletionTimeStamp is not equal zero, then the resource has been marked for deletion
/*	if !component.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted
		if ContainsString(component.ObjectMeta.Finalizers, svcFinalizerName) {
			// Component has been deleted like also its dependencies
			operation = deletionOperation

			// our finalizer is present, so lets handle our external dependency
			// Check if the component is a Service and then delete the ServiceInstance, ServiceBinding
			// TODO: Move this code under the ServiceController !!
			if component.Spec.Services != nil {
				removeServiceInstanceStep := servicecatalog.RemoveServiceInstanceStep()
				r.reqLogger.Info("## Invoking'service catalog', action '%s' on %s", "delete", component.Name)
				//log.Infof("## Invoking'service catalog', action '%s' on %s", "delete", component.Name)
				if err := removeServiceInstanceStep.Handle(component, r.config, &r.client, request.Namespace, r.scheme); err != nil {
					r.reqLogger.Error(err, "Removing Service Instance, binding failed")
				}
			}

			// remove our finalizer from the list and update it.
			component.ObjectMeta.Finalizers = RemoveString(component.ObjectMeta.Finalizers, svcFinalizerName)
			if err := r.client.Update(context.Background(), component); err != nil {
				return reconcile.Result{Requeue: true}, nil
			}
		}
		r.reqLogger.Info("***** Reconciled Component %s, namespace %s", request.Name, request.Namespace)
		r.reqLogger.Info("***** Operation performed : %s", operation)
		return reconcile.Result{}, nil
	}*/

    // Check if the ServiceBinding AND ServiceInstance exist
    if _, err := r.fetchServiceBindings(service); err != nil {
    	return reconcile.Result{}, err
	}

	// Process the Service if the status is not ....

	r.reqLogger.Info(fmt.Sprintf("Reconciled : %s", service.Name))
	return reconcile.Result{}, nil
}


// TODO
func (r *ReconcileService) DeleteService(service v1alpha2.Service, config rest.Config, c client.Client, namespace string, scheme *runtime.Scheme) error {
	/*// Let's retrieve the ServiceBindings to delete them first
	list, err := listServiceBindings(&{}, c)
	if err != nil {
		return err
	}
	// Delete ServiceBinding(s) linked to the ServiceInstance
	for _, sb := range list.Items {
		if sb.Name == s.Name {
			err := c.Delete(context.TODO(), &sb)
			if err != nil {
				return err
			}
			log.Infof("### Deleted serviceBinding '%s' for the service '%s'", sb.Name, s.Name)
		}
	}

	// Retrieve ServiceInstances
	serviceInstanceList := new(servicecatalog.ServiceInstanceList)
	serviceInstanceList.TypeMeta = metav1.TypeMeta{
		Kind:       "ServiceInstance",
		APIVersion: "servicecatalog.k8s.io/v1beta1",
	}
	listOps := &client.ListOptions{
		Namespace: service.ObjectMeta.Namespace,
	}
	err = c.List(context.TODO(), listOps, serviceInstanceList)
	if err != nil {
		return err
	}

	// Delete ServiceInstance(s)
	for _, si := range serviceInstanceList.Items {
		err := c.Delete(context.TODO(), &si)
		if err != nil {
			return err
		}
		log.Infof("### Deleted serviceInstance '%s' for the service '%s'", si.Name, s.Name)
	}*/
	return nil
}

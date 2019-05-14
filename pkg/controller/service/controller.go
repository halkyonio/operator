package service

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"k8s.io/apimachinery/pkg/api/errors"
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
	controllerName   = "service-controller"
	SERVICEINSTANCE  = "ServiceInstance"
	SERVICEBINDING   = "ServiceBinding"
	svcFinalizerName = "service.component.k8s.io"
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

	/** Watch for changes of child/secondary resources **/
	//ServiceInstance
/*	if err := watchServiceInstance(c); err != nil {
		return err
	}

	//ServiceBinding
	if err := watchServiceBinding(c); err != nil {
		return err
	}*/

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

//buildFactory will return the resource according to the kind defined
func (r *ReconcileService) buildFactory(instance *v1alpha2.Service, kind string) (runtime.Object, error) {
	r.reqLogger.Info("Check "+kind, "into the namespace", instance.Namespace)
	switch kind {
	case SERVICEBINDING:
		return r.buildServiceBinding(instance), nil
	case SERVICEINSTANCE:
		return r.buildServiceInstance(instance)
	default:
		msg := "Failed to recognize type of object" + kind + " into the Namespace " + instance.Namespace
		panic(msg)
	}
}

//Create the factory object and requeue
func (r *ReconcileService) create(instance *v1alpha2.Service, kind string, err error) (reconcile.Result, error) {
	obj, errBuildObject := r.buildFactory(instance, kind)
	if errBuildObject != nil {
		return reconcile.Result{}, errBuildObject
	}
	if errors.IsNotFound(err) {
		r.reqLogger.Info("Creating a new ", "kind", kind, "Namespace", instance.Namespace)
		err = r.client.Create(context.TODO(), obj)
		if err != nil {
			r.reqLogger.Error(err, "Failed to create new ", "kind", kind, "Namespace", instance.Namespace)
			return reconcile.Result{}, err
		}
		r.reqLogger.Info("Created successfully - return and create", "kind", kind, "Namespace", instance.Namespace)
		return reconcile.Result{Requeue: true}, nil
	}
	r.reqLogger.Error(err, "Failed to get", "kind", kind, "Namespace", instance.Namespace)
	return reconcile.Result{}, err

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

	// Service has been marked for deletion or deleted
	if !service.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted
		if r.ContainsString(service.ObjectMeta.Finalizers, svcFinalizerName) {
			// our finalizer is present, so lets handle our external dependency
			if service.Spec.Name != "" {
				// TODO Call action to remove
				// r.reqLogger.Info("## Invoking'service catalog', action '%s' on %s", "delete", component.Name)
				if err := r.DeleteService(service); err != nil {
					r.reqLogger.Error(err, "Removing Service Instance & ServiceBinding failed")
				}
			}

			// remove our finalizer from the list and update it.
			service.ObjectMeta.Finalizers = r.RemoveString(service.ObjectMeta.Finalizers, svcFinalizerName)
			if err := r.client.Update(context.Background(), service); err != nil {
				return reconcile.Result{Requeue: true}, nil
			}
		}
		r.reqLogger.Info("Reconciled Component %s, namespace %s", request.Name, request.Namespace)
		return reconcile.Result{}, nil
	}

	// Add the Status Service when we process the first time the Service CR
	if service.Generation == 1 && service.Status.Phase != "" {
		if err := r.updateServiceStatus(service, v1alpha2.PhaseServiceCreation); err != nil {
			r.reqLogger.Info("Status update failed !")
			return reconcile.Result{Requeue: true}, err
		}
	}

	// Check if the ServiceInstance exists
	if _, err := r.fetchServiceInstance(service); err != nil {
		return r.create(service, SERVICEBINDING, err)
	}

	// Check if the ServiceBinding exists
	if _, err := r.fetchServiceBinding(service); err != nil {
		return r.create(service, SERVICEBINDING, err)
	}

	// Update Service object to add a k8s ObjectMeta finalizer
	if !r.ContainsString(service.ObjectMeta.Finalizers, svcFinalizerName) {
		service.ObjectMeta.Finalizers = append(service.ObjectMeta.Finalizers, svcFinalizerName)
		r.update(service)
	}

	// Update Status
	serviceBindingStatus, err := r.updateServiceBindingStatus(service)
	if err != nil {
		return reconcile.Result{}, err
	}

	serviceInstanceStatus, err := r.updateServiceInstanceStatus(service)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Update Status of the Service
	//Update status for App
	if err:= r.updateStatus(serviceBindingStatus, serviceInstanceStatus, service); err != nil {
		return reconcile.Result{}, err
	}

	r.reqLogger.Info(fmt.Sprintf("Reconciled : %s", service.Name))
	return reconcile.Result{}, nil
}

func (r *ReconcileService) DeleteService(service *v1alpha2.Service) error {
	// Let's retrieve the ServiceBinding to delete it first
	serviceBinding, err := r.fetchServiceBinding(service)
	if err != nil {
		return err
	}
	// Delete ServiceBinding linked to the ServiceInstance
	if serviceBinding.Name == service.Name {
		err := r.client.Delete(context.TODO(), serviceBinding)
		if err != nil {
			return err
		}
		r.reqLogger.Info(fmt.Sprintf("Deleted serviceBinding '%s' for the service '%s'", serviceBinding.Name, service.Name))
	}

	serviceInstance, err := r.fetchServiceInstance(service)
	if err != nil {
		return err
	}
	// Delete the ServiceInstance
	err = r.client.Delete(context.TODO(), serviceInstance)
	if err != nil {
		return err
	}
	r.reqLogger.Info(fmt.Sprintf("Deleted serviceInstance '%s' for the service '%s'", serviceInstance.Name, service.Name))

	return nil
}

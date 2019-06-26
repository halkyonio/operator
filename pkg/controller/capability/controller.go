package capability

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
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
	controllerName   = "service-controller"
	SECRET           = "Secret"
	PG_DATABASE         = "Postgres"
	svcFinalizerName = "capability.devexp.runtime.redhat.com"
)

var log = logf.Log.WithName("capability.controller")

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

	// Watch for changes to primary resource Capability
	err = c.Watch(&source.Kind{Type: &v1alpha2.Capability{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	/** Watch for changes of child/secondary resources **/
	//KubeDB
	/*	if err := watchServiceInstance(c); err != nil {
		return err
	}*/

	return nil
}

// newReconciler returns a new reconcile.Reconciler
func NewReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCapability{
		client:    mgr.GetClient(),
		config:    mgr.GetConfig(),
		scheme:    mgr.GetScheme(),
		reqLogger: log,
	}
}

type ReconcileCapability struct {
	client    client.Client
	config    *rest.Config
	scheme    *runtime.Scheme
	reqLogger logr.Logger
}

//Update the factory object and requeue
func (r *ReconcileCapability) update(obj runtime.Object) (reconcile.Result, error) {
	err := r.client.Update(context.TODO(), obj)
	if err != nil {
		r.reqLogger.Error(err, "Failed to update spec")
		return reconcile.Result{}, err
	}
	r.reqLogger.Info("Spec updated successfully")
	return reconcile.Result{}, nil
}

//buildFactory will return the resource according to the kind defined
func (r *ReconcileCapability) buildFactory(instance *v1alpha2.Capability, kind string) (runtime.Object, error) {
	r.reqLogger.Info("Check "+kind, "into the namespace", instance.Namespace)
	switch kind {
	case SECRET:
		return r.buildSecret(instance)
	case PG_DATABASE:
		return r.buildKubeDbPostgres(instance)
	default:
		msg := "Failed to recognize type of object" + kind + " into the Namespace " + instance.Namespace
		panic(msg)
	}
}

//Create the factory object and don't requeue
func (r *ReconcileCapability) create(instance *v1alpha2.Capability, kind string) error {
	obj, err := r.buildFactory(instance, kind)
	if err != nil {
		return err
	}
	r.reqLogger.Info("Creating a new ", "kind", kind, "Namespace", instance.Namespace)
	err = r.client.Create(context.TODO(), obj)
	if err != nil {
		r.reqLogger.Error(err, "Failed to create new ", "kind", kind, "Namespace", instance.Namespace)
		return err
	}
	r.reqLogger.Info("Created successfully", "kind", kind, "Namespace", instance.Namespace)
	return nil
}

func (r *ReconcileCapability) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.reqLogger = log.WithValues("Namespace", request.Namespace)

	// Fetch the Capability created, deleted or updated
	service := &v1alpha2.Capability{}
	err := r.client.Get(context.TODO(), request.NamespacedName, service)
	if err != nil {
		return r.fetch(err)
	}

	r.reqLogger.Info("-----------------------")
	r.reqLogger.Info("Reconciling Capability")
	r.reqLogger.Info("Status of the Capability", "Status phase", service.Status.Phase)
	r.reqLogger.Info("Creation time          ", "Creation time", service.ObjectMeta.CreationTimestamp)
	r.reqLogger.Info("Resource version       ", "Resource version", service.ObjectMeta.ResourceVersion)
	r.reqLogger.Info("Generation version     ", "Generation version", strconv.FormatInt(service.ObjectMeta.Generation, 10))
	// r.reqLogger.Info("Deletion time          ","Deletion time", Capability.ObjectMeta.DeletionTimestamp)


	// Add the Status Capability Creation when we process the first time the Capability CR
	// as we will start to create different resources
	if service.Generation == 1 && service.Status.Phase == "" {
		if err := r.updateServiceStatus(service, v1alpha2.CapabilityPending, request); err != nil {
			r.reqLogger.Info("Status update failed !")
			return reconcile.Result{}, err
		}
		r.reqLogger.Info(fmt.Sprintf("Status is now : %s", v1alpha2.CapabilityPending))
	}

	// Check if the ServiceInstance exists
	if _, err := r.fetchSecret(service); err != nil {
		if err = r.create(service, SECRET); err != nil {
			return reconcile.Result{}, err
		}
	}


	// Update Status
/*	serviceInstanceStatus, err := r.updateServiceInstanceStatus(service, request)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Update Status of the Capability
	if err := r.updateStatus(serviceInstanceStatus, service, request); err != nil {
		return reconcile.Result{}, err
	}*/

	r.reqLogger.Info(fmt.Sprintf("Reconciled : %s", service.Name))
	return reconcile.Result{}, nil
}

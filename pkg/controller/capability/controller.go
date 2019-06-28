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
	"strings"
)

const (
	controllerName        = "service-controller"
	SECRET                = "Secret"
	PG_DATABASE           = "Postgres"
	PG_VAR_DATABASE_NAME  = "POSTGRES_DB"
)

var (
	log              = logf.Log.WithName("capability.controller")
)

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
	// KubeDB Postgres
	if err := watchPostgresDB(c); err != nil {
		return err
	}

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
		return r.buildKubeDBPostgres(instance)
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
	capability := &v1alpha2.Capability{}
	err := r.client.Get(context.TODO(), request.NamespacedName, capability)
	if err != nil {
		return r.fetch(err)
	}

	r.reqLogger.Info("-----------------------")
	r.reqLogger.Info("Reconciling Capability")
	r.reqLogger.Info("Status of the Capability", "Status phase", capability.Status.Phase)
	r.reqLogger.Info("Creation time          ", "Creation time", capability.ObjectMeta.CreationTimestamp)
	r.reqLogger.Info("Resource version       ", "Resource version", capability.ObjectMeta.ResourceVersion)
	r.reqLogger.Info("Generation version     ", "Generation version", strconv.FormatInt(capability.ObjectMeta.Generation, 10))


	if strings.ToLower(string(v1alpha2.DatabaseCategory)) == string(capability.Spec.Category) {
		installFn := r.installDB

		// Define the initial status to pending to indicates to the user that we have started to process
		// the CRD and we are creating/installing the kube resources
		r.setInitialStatus(capability, v1alpha2.CapabilityPending)

		// Install the 2nd resources and check if the status of the watched resources has changed
		result, e := r.installAndUpdateStatus(capability, request, installFn)
		r.reqLogger.Info("<== Reconciled Capability", "name", capability.Name)
		return result, e
	} else {
		r.reqLogger.Info(fmt.Sprintf("<== Reconciled but Capability not supported : %s", capability.Spec.Category))
		return reconcile.Result{}, nil
	}
}

type installFnType func(c *v1alpha2.Capability) (bool, error)

func (r *ReconcileCapability) installAndUpdateStatus(c *v1alpha2.Capability, request reconcile.Request, install installFnType) (reconcile.Result, error) {
	changed, err := install(c)
	if err != nil {
		r.reqLogger.Error(err, fmt.Sprintf("failed to install %s", c.Spec.Kind))
		r.setErrorStatus(c, err)
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: changed}, r.updateStatus(c, v1alpha2.CapabilityReady)
}

// Add the Status Capability Creation when we process the first time the Capability CR
// as we will start to create different resources
func (r *ReconcileCapability) setInitialStatus(c *v1alpha2.Capability, phase v1alpha2.CapabilityPhase) error {
	if c.Generation == 1 && c.Status.Phase == "" {
		if err := r.updateStatus(c, phase); err != nil {
			r.reqLogger.Info("Status update failed !")
			return err
		}
	}
	return nil
}

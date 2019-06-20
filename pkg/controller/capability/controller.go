package capability

import (
	"context"
	"fmt"
	servicecatalogv1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	controller2 "github.com/snowdrop/component-operator/pkg/controller"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	controllerName          = "service-controller"
	SECRET                  = "Secret"
	// KubeDB Postgres const
	KUBEDB_PG_DATABASE      = "Postgres"
	KUBEDB_PG_DATABASE_NAME = "POSTGRES_DB"
	KUBEDB_PG_USER          = "POSTGRES_USER"
	KUBEDB_PG_PASSWORD      = "POSTGRES_PASSWORD"
	// Capability const
	DB_CONFIG_NAME          = "DB_CONFIG_NAME"
	DB_HOST                 = "DB_HOST"
	DB_PORT                 = "DB_PORT"
	DB_NAME                 = "DB_NAME"
	DB_USER                 = "DB_USER"
	DB_PASSWORD             = "DB_PASSWORD"
)

func NewCapabilityReconciler(mgr manager.Manager) *ReconcileCapability {
	r := &ReconcileCapability{}
	r.ReconcilerHelper = controller2.NewHelper(r, mgr)
	return r
}

type ReconcileCapability struct {
	controller2.ReconcilerHelper
}

func (r *ReconcileCapability) PrimaryResourceName() string {
	return "capability"
}

func (r *ReconcileCapability) PrimaryResourceType() runtime.Object {
	return new(v1alpha2.Capability)
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

func (r *ReconcileCapability) SecondaryResourceTypes() []runtime.Object {
	return []runtime.Object{
		&servicecatalogv1.ServiceInstance{},
		&servicecatalogv1.ServiceBinding{},
	}
}

func (ReconcileCapability) asCapability(object runtime.Object) *v1alpha2.Capability {
	return object.(*v1alpha2.Capability)
}

func (r *ReconcileCapability) IsPrimaryResourceValid(object runtime.Object) bool {
	// todo: implement
	return true
}

func (r *ReconcileCapability) ResourceMetadata(object runtime.Object) controller2.ResourceMetadata {
	capability := r.asCapability(object)
	return controller2.ResourceMetadata{
		Name:         capability.Name,
		Status:       capability.Status.Phase.String(),
		Created:      capability.ObjectMeta.CreationTimestamp,
		ShouldDelete: !capability.ObjectMeta.DeletionTimestamp.IsZero(),
	}
}

func (r *ReconcileCapability) Delete(object runtime.Object) (bool, error) {
	// todo: implement
	service := r.asCapability(object)
	// The object is being deleted
	if r.ContainsString(service.ObjectMeta.Finalizers, svcFinalizerName) {
		// our finalizer is present, so let's handle our external dependency
		if service.Name != "" {
			// TODO Call action to remove
			// r.ReqLogger.Info("## Invoking'service catalog', action '%s' on %s", "delete", component.Name)
			if err := r.DeleteService(service); err != nil {
				r.ReqLogger.Error(err, "Removing Capability Instance & ServiceBinding failed")
			}
		}

		// remove our finalizer from the list and update it.
		service.ObjectMeta.Finalizers = r.RemoveString(service.ObjectMeta.Finalizers, svcFinalizerName)
		if err := r.Client.Update(context.Background(), service); err != nil {
			return true, err
		}
	}
	r.ReqLogger.Info(fmt.Sprintf("Deleted capability %s, namespace %s", service.Name, service.Namespace))
	return false, nil
}

func (r *ReconcileCapability) CreateOrUpdate(object runtime.Object) (bool, error) {
	service := r.asCapability(object)
	request := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: service.Namespace,
			Name:      service.Name,
		},
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
	if service.Generation == 1 && service.Status.Phase == "" {
		if err := r.updateServiceStatus(service, v1alpha2.CapabilityPending, request); err != nil {
			r.ReqLogger.Info("Status update failed !")
			return false, err
		}
		r.ReqLogger.Info(fmt.Sprintf("Status is now : %s", v1alpha2.CapabilityPending))
	}

	// Check if the ServiceInstance exists
	if _, err := r.fetchServiceInstance(service); err != nil {
		if err = r.create(service, SERVICEINSTANCE); err != nil {
			return false, err
		}
	}

	// Check if the ServiceBinding exists
	if _, err := r.fetchServiceBinding(service); err != nil {
		if err = r.create(service, SERVICEBINDING); err != nil {
			return false, err
		}
	}

	// Update Capability object to add a k8s ObjectMeta finalizer
	if !r.ContainsString(service.ObjectMeta.Finalizers, svcFinalizerName) {
		// Get a more recent version of the CR
		service, err := r.fetchCapability(request)
		if err != nil {
			return false, err
		}
		service.ObjectMeta.Finalizers = append(service.ObjectMeta.Finalizers, svcFinalizerName)
		_, err = r.update(service)
		if err != nil {
			return false, err
		}
	}

	return false, nil
}

func (r *ReconcileCapability) SetErrorStatus(object runtime.Object, e error) {
	panic("implement me")
}

func (r *ReconcileCapability) SetSuccessStatus(object runtime.Object) {
	service := r.asCapability(object)
	request := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: service.Namespace,
			Name:      service.Name,
		},
	}

	serviceBindingStatus, err := r.updateServiceBindingStatus(service, request)
	if err != nil {
		panic(err)
	}

	serviceInstanceStatus, err := r.updateServiceInstanceStatus(service, request)
	if err != nil {
		panic(err)
	}

	// Update Status of the Capability
	if err := r.updateStatus(serviceBindingStatus, serviceInstanceStatus, service, request); err != nil {
		panic(err)
	}
}

func (r *ReconcileCapability) Helper() controller2.ReconcilerHelper {
	return r.ReconcilerHelper
}

//Update the factory object and requeue
func (r *ReconcileCapability) update(obj runtime.Object) (reconcile.Result, error) {
	err := r.Client.Update(context.TODO(), obj)
	if err != nil {
		r.ReqLogger.Error(err, "Failed to update spec")
		return reconcile.Result{}, err
	}
	r.ReqLogger.Info("Spec updated successfully")
	return reconcile.Result{}, nil
}

//buildFactory will return the resource according to the kind defined
func (r *ReconcileCapability) buildFactory(instance *v1alpha2.Capability, kind string) (runtime.Object, error) {
	r.ReqLogger.Info("Check "+kind, "into the namespace", instance.Namespace)
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

//Create the factory object and don't requeue
func (r *ReconcileCapability) create(instance *v1alpha2.Capability, kind string) error {
	obj, err := r.buildFactory(instance, kind)
	if err != nil {
		return err
	}
	r.ReqLogger.Info("Creating a new ", "kind", kind, "Namespace", instance.Namespace)
	err = r.Client.Create(context.TODO(), obj)
	if err != nil {
		r.ReqLogger.Error(err, "Failed to create new ", "kind", kind, "Namespace", instance.Namespace)
		return err
	}
	r.ReqLogger.Info("Created successfully", "kind", kind, "Namespace", instance.Namespace)
	return nil
}

func (r *ReconcileCapability) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	return controller2.NewGenericReconciler(r).Reconcile(request)
}

func (r *ReconcileCapability) DeleteService(service *v1alpha2.Capability) error {
	// Let's retrieve the ServiceBinding to delete it first
	serviceBinding, err := r.fetchServiceBinding(service)
	if err != nil {
		return err
	}
	// Delete ServiceBinding linked to the ServiceInstance
	if serviceBinding.Name == service.Name {
		err := r.Client.Delete(context.TODO(), serviceBinding)
		if err != nil {
			return err
		}
		r.ReqLogger.Info(fmt.Sprintf("Deleted serviceBinding '%s' for the service '%s'", serviceBinding.Name, service.Name))
	}

	serviceInstance, err := r.fetchServiceInstance(service)
	if err != nil {
		return err
	}
	// Delete the ServiceInstance
	err = r.Client.Delete(context.TODO(), serviceInstance)
	if err != nil {
		return err
	}
	r.ReqLogger.Info(fmt.Sprintf("Deleted serviceInstance '%s' for the service '%s'", serviceInstance.Name, service.Name))

	return nil
}

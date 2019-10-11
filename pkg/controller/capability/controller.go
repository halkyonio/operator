package capability

import (
	"fmt"
	capability "halkyon.io/api/capability/v1beta1"
	controller2 "halkyon.io/operator/pkg/controller"
	"halkyon.io/operator/pkg/controller/framework"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	SECRET = "Secret"
	// KubeDB Postgres const
	KUBEDB_PG_DATABASE      = "Postgres"
	KUBEDB_PG_DATABASE_NAME = "POSTGRES_DB"
	KUBEDB_PG_USER          = "POSTGRES_USER"
	KUBEDB_PG_PASSWORD      = "POSTGRES_PASSWORD"
	// Capability const
	DB_CONFIG_NAME = "DB_CONFIG_NAME"
	DB_HOST        = "DB_HOST"
	DB_PORT        = "DB_PORT"
	DB_NAME        = "DB_NAME"
	DB_USER        = "DB_USER"
	DB_PASSWORD    = "DB_PASSWORD"
)

func NewCapabilityReconciler(mgr manager.Manager) *ReconcileCapability {
	baseReconciler := framework.NewBaseGenericReconciler(controller2.NewCapability(), mgr)
	r := &ReconcileCapability{
		BaseGenericReconciler: baseReconciler,
	}
	baseReconciler.SetReconcilerFactory(r)

	r.AddDependentResource(newSecret())
	r.AddDependentResource(newPostgres())
	r.AddDependentResource(controller2.NewRole())
	r.AddDependentResource(controller2.NewRoleBinding())
	return r
}

type ReconcileCapability struct {
	*framework.BaseGenericReconciler
}

func asCapability(object runtime.Object) *controller2.Capability {
	return object.(*controller2.Capability)
}

func (r *ReconcileCapability) Delete(object framework.Resource) error {
	return nil
}

func (r *ReconcileCapability) CreateOrUpdate(object framework.Resource) (e error) {
	c := asCapability(object)
	if capability.DatabaseCategory.Equals(c.Spec.Category) {
		// Install the 2nd resources and check if the status of the watched resources has changed
		e = r.installDB(c)
	} else {
		e = fmt.Errorf("unsupported '%s' capability category", c.Spec.Category)
	}
	return e
}

func (r *ReconcileCapability) SetPrimaryResourceStatus(primary framework.Resource, statuses []framework.DependentResourceStatus) bool {
	return primary.SetSuccessStatus(statuses, "Ready")
}

package capability

import (
	"halkyon.io/api/capability/v1beta1"
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
	r := &ReconcileCapability{}
	r.K8SHelper = framework.NewHelper(r.PrimaryResourceType(), mgr)
	return r
}

type ReconcileCapability struct {
	*framework.K8SHelper
}

func (r *ReconcileCapability) GetDependentResourcesTypes() []framework.DependentResource {
	return []framework.DependentResource{
		newSecret(),
		newPostgres(),
		controller2.NewRole(),
		controller2.NewRoleBinding(),
	}
}

func (r *ReconcileCapability) PrimaryResourceType() runtime.Object {
	return &v1beta1.Capability{}
}

func (r *ReconcileCapability) NewFrom(name string, namespace string) (framework.Resource, error) {
	c := controller2.NewCapability()
	_, err := r.K8SHelper.Fetch(name, namespace, c.Capability)
	resourcesTypes := r.GetDependentResourcesTypes()
	for _, rType := range resourcesTypes {
		c.AddDependentResource(rType.NewInstanceWith(c))
	}
	return c, err
}

func asCapability(object runtime.Object) *controller2.Capability {
	return object.(*controller2.Capability)
}

func (r *ReconcileCapability) Delete(object framework.Resource) error {
	return nil
}

func (r *ReconcileCapability) CreateOrUpdate(object framework.Resource) (e error) {
	c := asCapability(object)
	return r.installDB(c)
}

func (r *ReconcileCapability) SetPrimaryResourceStatus(primary framework.Resource, statuses []framework.DependentResourceStatus) bool {
	return primary.SetSuccessStatus(statuses, "Ready")
}

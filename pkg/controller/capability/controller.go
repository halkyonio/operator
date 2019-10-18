package capability

import (
	"halkyon.io/api/capability/v1beta1"
	"halkyon.io/operator/pkg/controller/framework"
	"k8s.io/apimachinery/pkg/runtime"
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

func NewCapabilityManager() *CapabilityManager {
	return &CapabilityManager{}
}

type CapabilityManager struct {
	dependentTypes []framework.DependentResource
}

func (r *CapabilityManager) Helper() *framework.K8SHelper {
	return framework.GetHelperFor(r.PrimaryResourceType())
}

func (r *CapabilityManager) GetDependentResourcesTypes() []framework.DependentResource {
	if len(r.dependentTypes) == 0 {
		r.dependentTypes = []framework.DependentResource{
			newSecret(),
			newPostgres(),
			newRole(nil),
			newRoleBinding(nil),
		}
	}
	return r.dependentTypes
}

func (r *CapabilityManager) PrimaryResourceType() runtime.Object {
	return &v1beta1.Capability{}
}

func (r *CapabilityManager) NewFrom(name string, namespace string) (framework.Resource, error) {
	return NewCapability().FetchAndInit(name, namespace, r)
}

func asCapability(object runtime.Object) *Capability {
	return object.(*Capability)
}

func (r *CapabilityManager) Delete(object framework.Resource) error {
	return nil
}

func (r *CapabilityManager) CreateOrUpdate(object framework.Resource) (e error) {
	c := asCapability(object)
	return r.installDB(c)
}

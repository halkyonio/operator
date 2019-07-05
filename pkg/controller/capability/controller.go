package capability

import (
	"fmt"
	kubedbv1 "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	controller2 "github.com/snowdrop/component-operator/pkg/controller"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"strings"
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
	baseReconciler := controller2.NewBaseGenericReconciler(&v1alpha2.Capability{}, mgr)
	r := &ReconcileCapability{
		BaseGenericReconciler: baseReconciler,
	}
	baseReconciler.SetReconcilerFactory(r)

	r.AddDependentResource(newSecret())
	r.AddDependentResource(newPostgres())

	return r
}

type ReconcileCapability struct {
	*controller2.BaseGenericReconciler
}

func asCapability(object runtime.Object) *v1alpha2.Capability {
	return object.(*v1alpha2.Capability)
}

func (r *ReconcileCapability) IsDependentResourceReady(resource v1alpha2.Resource) (depOrTypeName string, ready bool) {
	capability := asCapability(resource)
	db, err := r.fetchKubeDBPostgres(capability)
	if err != nil || !r.isDBReady(db) {
		return "postgreSQL db", false
	}
	return db.Name, true
}

func (r *ReconcileCapability) Delete(name, namespace string) (bool, error) {
	panic("implement me")
}

func (r *ReconcileCapability) CreateOrUpdate(object runtime.Object) (bool, error) {
	capability := asCapability(object)
	if strings.ToLower(string(v1alpha2.DatabaseCategory)) == string(capability.Spec.Category) {
		// Install the 2nd resources and check if the status of the watched resources has changed
		return r.installDB(capability)
	} else {
		return false, fmt.Errorf("unsupported '%s' capability category", capability.Spec.Category)
	}
}

func (r *ReconcileCapability) isDBReady(p *kubedbv1.Postgres) bool {
	return p.Status.Phase == kubedbv1.DatabasePhaseRunning
}

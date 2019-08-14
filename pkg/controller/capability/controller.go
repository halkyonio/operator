package capability

import (
	"fmt"
	kubedbv1 "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"halkyon.io/operator/pkg/apis/halkyon/v1beta1"
	controller2 "halkyon.io/operator/pkg/controller"
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
	baseReconciler := controller2.NewBaseGenericReconciler(&v1beta1.Capability{}, mgr)
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
	*controller2.BaseGenericReconciler
}

func asCapability(object runtime.Object) *v1beta1.Capability {
	return object.(*v1beta1.Capability)
}

func (r *ReconcileCapability) IsDependentResourceReady(resource v1beta1.Resource) (depOrTypeName string, ready bool) {
	db, err := r.MustGetDependentResourceFor(resource, &kubedbv1.Postgres{}).Fetch(r.Helper())
	if err != nil || !r.isDBReady(db.(*kubedbv1.Postgres)) {
		return "postgreSQL db", false
	}
	return db.(*kubedbv1.Postgres).Name, true
}

func (r *ReconcileCapability) Delete(object v1beta1.Resource) (bool, error) {
	panic("implement me")
}

func (r *ReconcileCapability) CreateOrUpdate(object v1beta1.Resource) (e error) {
	capability := asCapability(object)
	if strings.ToLower(string(v1beta1.DatabaseCategory)) == string(capability.Spec.Category) {
		// Install the 2nd resources and check if the status of the watched resources has changed
		e = r.installDB(capability)
	} else {
		e = fmt.Errorf("unsupported '%s' capability category", capability.Spec.Category)
	}
	return e
}

func (r *ReconcileCapability) isDBReady(p *kubedbv1.Postgres) bool {
	return p.Status.Phase == kubedbv1.DatabasePhaseRunning
}

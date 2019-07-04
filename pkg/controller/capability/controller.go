package capability

import (
	"fmt"
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

func (r *ReconcileCapability) IsPrimaryResourceValid(object runtime.Object) bool {
	// todo: implement
	return true
}

func (r *ReconcileCapability) ResourceMetadata(object runtime.Object) controller2.ResourceMetadata {
	capability := asCapability(object)
	return controller2.ResourceMetadata{
		Name:         capability.Name,
		Status:       capability.Status.Phase.String(),
		Created:      capability.CreationTimestamp,
		ShouldDelete: !capability.DeletionTimestamp.IsZero(),
	}
}

func (r *ReconcileCapability) Delete(object runtime.Object) (bool, error) {
	panic("implement me")
}

func (r *ReconcileCapability) CreateOrUpdate(object runtime.Object) (bool, error) {
	capability := asCapability(object)
	if strings.ToLower(string(v1alpha2.DatabaseCategory)) == string(capability.Spec.Category) {
		err := r.setInitialStatus(capability, v1alpha2.CapabilityPending)
		if err != nil {
			return false, err
		}

		// Install the 2nd resources and check if the status of the watched resources has changed
		return r.installDB(capability)
	} else {
		return false, fmt.Errorf("unsupported '%s' capability category", capability.Spec.Category)
	}
}

func (r *ReconcileCapability) SetErrorStatus(object runtime.Object, e error) {
	r.setErrorStatus(asCapability(object), e)
}

func (r *ReconcileCapability) SetSuccessStatus(object runtime.Object) {
	c := asCapability(object)
	if c.Status.Phase != v1alpha2.CapabilityReady {
		err := r.updateStatus(c, v1alpha2.CapabilityReady)
		if err != nil {
			panic(err)
		}
	}
}

// Add the Status Capability Creation when we process the first time the Capability CR
// as we will start to create different resources
func (r *ReconcileCapability) setInitialStatus(c *v1alpha2.Capability, phase v1alpha2.CapabilityPhase) error {
	if c.Generation == 1 && c.Status.Phase == "" {
		if err := r.updateStatus(c, phase); err != nil {
			r.ReqLogger.Info("Status update failed !")
			return err
		}
	}
	return nil
}

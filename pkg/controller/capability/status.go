package capability

import (
	"context"
	"fmt"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	kubedbv1 "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

//updateStatus returns error when status regards the all required resources could not be updated
func (r *ReconcileCapability) updateStatus(p *kubedbv1.Postgres ,instance *v1alpha2.Capability, request reconcile.Request) error {
	if r.isServiceInstanceReady(p.Status) {
		r.reqLogger.Info(fmt.Sprintf("Updating Status of the Capability to %v", v1alpha2.CapabilityReady))

		if v1alpha2.CapabilityReady != instance.Status.Phase {
			// Get a more recent version of the CR
			service, err := r.fetchCapability(request)
			if err != nil {
				r.reqLogger.Error(err, "Failed to get the Capability")
				return err
			}

			service.Status.Phase = v1alpha2.CapabilityReady

			err = r.client.Status().Update(context.TODO(), service)
			if err != nil {
				r.reqLogger.Error(err, "Failed to update Status for the Capability App")
				return err
			}
		}
		return nil
	} else {
		r.reqLogger.Info("Capability is pending. So, we won't update the status of the Capability to Ready", "Namespace", instance.Namespace, "Name", instance.Name)
		return nil
	}
}

//updateStatus
func (r *ReconcileCapability) updateCapabilityStatus(instance *v1alpha2.Capability, phase v1alpha2.CapabilityPhase, request reconcile.Request) error {
	if !reflect.DeepEqual(phase, instance.Status.Phase) {
		// Get a more recent version of the CR
		service, err := r.fetchCapability(request)
		if err != nil {
			return err
		}

		service.Status.Phase = phase

		err = r.client.Status().Update(context.TODO(), service)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update Status of the Capability")
			return err
		}
	}
	r.reqLogger.Info(fmt.Sprintf("Updating Status of the Capability to %v", phase))
	return nil
}

func (r *ReconcileCapability) updateKubeDBStatus(c *v1alpha2.Capability, request reconcile.Request) (*kubedbv1.Postgres, error) {
	for {
		_, err := r.fetchPostgres(c)
		if err != nil {
			r.reqLogger.Info("Failed to get KubeDB Postgres. We retry till ...", "Namespace", c.Namespace, "Name", c.Name)
			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}

	r.reqLogger.Info("Updating KubeDB Postgres Status for the Capability")
	postgresDB, err := r.fetchPostgres(c)
	if err != nil {
		r.reqLogger.Error(err, "Failed to get KubeDB Postgres for Status", "Namespace", c.Namespace, "Name", c.Name)
		return postgresDB, err
	}
	if !reflect.DeepEqual(postgresDB.Name, c.Status.DatabaseName) || !reflect.DeepEqual(postgresDB.Status, c.Status.DatabaseStatus) {
		// Get a more recent version of the CR
		service, err := r.fetchCapability(request)
		if err != nil {
			r.reqLogger.Error(err, "Failed to get the Capability")
			return postgresDB, err
		}

		service.Status.DatabaseName = postgresDB.Name
		service.Status.DatabaseStatus = postgresDB.Status.Reason

		err = r.client.Status().Update(context.TODO(), service)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update postgresDB Name and Status for the Capability")
			return postgresDB, err
		}
		r.reqLogger.Info("postgresDB Status updated for the Capability")
	}
	return postgresDB, nil
}

func (r *ReconcileCapability) isServiceInstanceReady(p kubedbv1.PostgresStatus) bool {
		if p.Phase == kubedbv1.DatabasePhaseRunning {
			return true
		} else {
			return false
		}
}

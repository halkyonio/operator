package capability

import (
	"context"
	"fmt"
	kubedbv1 "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
)

func (r *ReconcileCapability) setErrorStatus(instance *v1alpha2.Capability, err error) {
	instance.Status.Phase = v1alpha2.CapabilityFailed
	r.updateStatusWithMessage(instance, err.Error(), true)
}

func (r *ReconcileCapability) updateStatusWithMessage(instance *v1alpha2.Capability, msg string, fetch bool) {
	// fetch latest version to avoid optimistic lock error
	current := instance
	var err error
	if fetch {
		current, err = r.fetchLatestVersion(instance)
		if err != nil {
			r.ReqLogger.Error(err, "failed to fetch latest version of capability "+instance.Name)
		}
	}

	r.ReqLogger.Info("updating capability status",
		"phase", instance.Status.Phase, "podName", instance.Status.PodName, "message", msg)
	current.Status.PodName = instance.Status.PodName
	current.Status.Phase = instance.Status.Phase
	current.Status.Message = msg

	err = r.Client.Status().Update(context.TODO(), current)
	if err != nil {
		r.ReqLogger.Error(err, "failed to update status for capability "+current.Name)
	}
}

func (r *ReconcileCapability) fetchLatestVersion(instance *v1alpha2.Capability) (*v1alpha2.Capability, error) {
	component, err := r.Fetch(instance.Name, instance.Namespace)
	if err != nil {
		r.ReqLogger.Error(err, "failed to get the Capability")
		return nil, err
	}
	return asCapability(component), nil
}

//updateStatus returns error when status regards the all required resources could not be updated

func (r *ReconcileCapability) updateStatus(instance *v1alpha2.Capability, phase v1alpha2.CapabilityPhase) error {
	// Get a more recent version of the CR
	capability, err := r.fetchLatestVersion(instance)
	if err != nil {
		return err
	}

	db, err := r.fetchKubeDBPostgres(capability)
	if err != nil || !r.isDBReady(db) {
		r.makePending("pod", capability)
		return nil
	}

	if db.Name != instance.Status.PodName || phase != instance.Status.Phase {
		capability.Status.PodName = db.Name
		capability.Status.Phase = phase
		r.updateStatusWithMessage(capability, "", false)
	}

	return nil
}

func (r *ReconcileCapability) makePending(dependencyName string, c *v1alpha2.Capability) {
	msg := fmt.Sprintf(dependencyName+" is not ready for component '%s' in namespace '%s'", c.Name, c.Namespace)
	r.ReqLogger.Info(msg)
	c.Status.Phase = v1alpha2.CapabilityPending
	r.updateStatusWithMessage(c, msg, false)
}

func (r *ReconcileCapability) isDBReady(p *kubedbv1.Postgres) bool {
	return p.Status.Phase == kubedbv1.DatabasePhaseRunning
}

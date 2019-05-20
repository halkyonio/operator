package link

import (
	"context"
	"fmt"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileLink) updateStatusInstance(status v1alpha2.LinkPhase, instance *v1alpha2.Link, request reconcile.Request) error {
	if status != instance.Status.Phase {
		r.reqLogger.Info("Updating App Status for the Link")
		// Get a more recent version of the CR
		link, err := r.fetchLink(request)
		if err != nil {
			return err
		}

		link.Status.Phase = status
		err = r.client.Status().Update(context.TODO(), link)
		//err := r.client.Update(context.TODO(),instance)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update Status for the Link")
			return err
		}
		r.reqLogger.Info(fmt.Sprintf("Status updated : %s", instance.Status.Phase))
	}
	return nil
}

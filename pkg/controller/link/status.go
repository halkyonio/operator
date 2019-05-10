package link

import (
	"context"
	"fmt"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"reflect"
)

func (r *ReconcileLink) updateStatusInstance(status v1alpha2.Phase, instance *v1alpha2.Link) error {
	r.reqLogger.Info("Updating App Status for the Link")

	if !reflect.DeepEqual(status, instance.Status.Phase) {
		instance.Status.Phase = status
		//err := r.client.Status().Update(context.TODO(), instance)
		err := r.client.Update(context.TODO(),instance)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update Status for the Component")
			return err
		}
	}

	r.reqLogger.Info(fmt.Sprintf("Status updated : %s", instance.Status.Phase))
	return nil
}

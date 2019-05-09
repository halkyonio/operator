package component

import (
	"context"
	"fmt"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"reflect"
	routev1 "github.com/openshift/api/route/v1"
)

//updateStatus returns error when status regards the all required resources could not be updated
func (r *ReconcileComponent) updateStatus(configMapStatus *corev1.ConfigMap, deploymentStatus *v1beta1.Deployment, serviceStatus *corev1.Service, routeStatus *routev1.Route, instance *v1alpha2.Component) error {
	r.reqLogger.Info("Updating App Status for the MobileSecurityService")
	if len(configMapStatus.UID) < 1 && len(deploymentStatus.UID) < 1 && len(serviceStatus.UID) < 1 && len(routeStatus.Name) < 1 {
		err := fmt.Errorf("Failed to get OK Status for Component")
		r.reqLogger.Error(err, "One of the resources are not created", "Component.Namespace", instance.Namespace, "Component.Name", instance.Name)
		return err
	}
	//status:= "OK"
	status := v1alpha2.PhaseDeploying
	if !reflect.DeepEqual(status, instance.Status.Phase) {
		instance.Status.Phase = status
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update Status for the MobileSecurityService App")
			return err
		}
	}
	return nil
}

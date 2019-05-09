package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	imagev1 "github.com/openshift/api/image/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//buildImageStream returns the service resource
func (r *ReconcileComponent) buildImageStream(c *v1alpha2.Component) *imagev1.ImageStream {
	ls := r.getAppLabels(c.Name)
	ser := &imagev1.ImageStream{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "",
			Namespace: c.Namespace,
			Labels:    ls,
		},
		Spec: imagev1.ImageStreamSpec{
		},
	}
	// Set Component instance as the owner and controller
	controllerutil.SetControllerReference(c, ser, r.scheme)
	return ser
}

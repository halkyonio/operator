package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"k8s.io/apimachinery/pkg/api/resource"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//buildPVC returns the PVC resource
func (r *ReconcileComponent) buildPVC(c *v1alpha2.Component) *corev1.PersistentVolumeClaim {
	ls := r.getAppLabels(c.Name)
	pvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Spec.Storage.Name,
			Namespace: c.Namespace,
			Labels:    ls,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				getAccessMode(c.Spec.Storage.Mode),
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(c.Spec.Storage.Capacity),
				},
			},
		},
	}
	// Set Component instance as the owner and controller
	controllerutil.SetControllerReference(c, pvc, r.scheme)
	return pvc
}

func getAccessMode(storage string) corev1.PersistentVolumeAccessMode {
	switch storage {
	case "ReadWriteOnce":
		return corev1.ReadWriteOnce
	case "ReadWriteMany":
		return corev1.ReadWriteMany
	case "ReadOnlyMany":
		return corev1.ReadOnlyMany
	default:
		return corev1.ReadWriteOnce
	}
}
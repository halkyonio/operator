package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type pvc struct {
	BaseDependentResource
}

func newPVC(manager manager.Manager) pvc {
	return pvc{BaseDependentResource: NewDependentResource(&corev1.PersistentVolumeClaim{}, manager)}
}

func (p pvc) Name(object runtime.Object) string {
	c := asComponent(object)
	specified := c.Spec.Storage.Name
	if len(specified) > 0 {
		return specified
	}
	return "m2-data-" + c.Name
}

func (p pvc) Build(object runtime.Object) (runtime.Object, error) {
	c := asComponent(object)
	ls := getAppLabels(c.Name)
	name := p.Name(c)
	pvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.Namespace,
			Labels:    ls,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				getAccessMode(c),
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: getCapacity(c),
				},
			},
		},
	}

	// Specify the default Storage data - value
	c.Spec.Storage.Name = name

	// Set Component instance as the owner and controller
	return pvc, controllerutil.SetControllerReference(c, pvc, p.Scheme)
}

//buildPVC returns the PVC resource
func (r *ReconcileComponent) buildPVC(res dependentResource, c *v1alpha2.Component) (runtime.Object, error) {
	ls := getAppLabels(c.Name)
	name := res.name(c)
	pvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.Namespace,
			Labels:    ls,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				getAccessMode(c),
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: getCapacity(c),
				},
			},
		},
	}

	// Specify the default Storage data - value
	c.Spec.Storage.Name = name

	// Set Component instance as the owner and controller
	return pvc, controllerutil.SetControllerReference(c, pvc, r.Scheme)
}

func getCapacity(c *v1alpha2.Component) resource.Quantity {
	specified := c.Spec.Storage.Capacity
	if len(specified) == 0 {
		specified = "1Gi"
		c.Spec.Storage.Capacity = specified
	}
	return resource.MustParse(specified)
}

func getAccessMode(c *v1alpha2.Component) corev1.PersistentVolumeAccessMode {
	storage := c.Spec.Storage.Mode
	mode := corev1.ReadWriteOnce
	switch storage {
	case "ReadWriteMany":
		mode = corev1.ReadWriteMany
	case "ReadOnlyMany":
		mode = corev1.ReadOnlyMany
	}
	c.Spec.Storage.Mode = string(mode)
	return mode
}

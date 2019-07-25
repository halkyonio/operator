package component

import (
	"github.com/snowdrop/component-api/pkg/apis/component/v1alpha2"
	"github.com/snowdrop/component-operator/pkg/controller"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type pvc struct {
	base
}

func (res pvc) NewInstanceWith(owner v1alpha2.Resource) controller.DependentResource {
	return newOwnedPvc(owner)
}

func newOwnedPvc(owner v1alpha2.Resource) pvc {
	dependent := newBaseDependent(&corev1.PersistentVolumeClaim{}, owner)
	p := pvc{base: dependent}
	dependent.SetDelegate(p)
	return p
}

func newPvc() pvc {
	return newOwnedPvc(nil)
}

func (res pvc) Build() (runtime.Object, error) {
	c := res.ownerAsComponent()
	ls := getAppLabels(c.Name)
	name := res.Name()
	pvc := &corev1.PersistentVolumeClaim{
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
	return pvc, nil
}

func (res pvc) Name() string {
	c := res.ownerAsComponent()
	specified := c.Spec.Storage.Name
	if len(specified) > 0 {
		return specified
	}
	return "m2-data-" + c.Name
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

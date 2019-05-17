package component

import (
	imagev1 "github.com/openshift/api/image/v1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
)

//buildImageStream returns the service resource
func (r *ReconcileComponent) buildImageStream(c *v1alpha2.Component, imageName string) *imagev1.ImageStream {
	ls := r.getAppLabels(c.Name)
	image, err := r.getImageInfo(c.Spec)
	if err != nil {
		panic(err)
	}
	is := &imagev1.ImageStream{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Capability",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      imageName,
			Namespace: c.Namespace,
			Labels:    ls,
		},
		Spec: imagev1.ImageStreamSpec{
			LookupPolicy: imagev1.ImageLookupPolicy{false},
			Tags: []imagev1.TagReference{
				{
					Annotations: image.defaultEnv,
					From:        &corev1.ObjectReference{Kind: "DockerImage", Name: image.registryRef},
					Name:        latestVersionTag,
					Reference:   true,
				},
			},
		},
	}
	// Set Component instance as the owner and controller
	controllerutil.SetControllerReference(c, is, r.scheme)
	return is
}

func (r *ReconcileComponent) getRuntimeImageName(c *v1alpha2.Component) string {
	return strings.Join([]string{"dev-runtime", strings.ToLower(c.Spec.Runtime)}, "-")
}

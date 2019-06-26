package capability

import (
	kubedbv1 "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/apis/apps"
	"k8s.io/kubernetes/pkg/apis/core"
	ofst "kmodules.xyz/offshoot-api/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

/*
spec:
  podTemplate:
    spec:
      env:
      - name: POSTGRES_DB
        value: my-database
*/

//buildSecret returns the secret resource
func (r *ReconcileCapability) buildKubeDbPostgres(c *v1alpha2.Capability) (*kubedbv1.Postgres, error) {
	ls := r.GetAppLabels(c.Name)
	postgres := &kubedbv1.Postgres{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
			Namespace: c.Namespace,
			Labels:    ls,
		},
		Spec: kubedbv1.PostgresSpec{
			// TODO : to be specify as parameter
			Version:  "10.2-v2",
			Replicas: replicaNumber(1),
			UpdateStrategy: apps.StatefulSetUpdateStrategy{
				Type: apps.RollingUpdateStatefulSetStrategyType,
			},
			DatabaseSecret: &core.SecretVolumeSource{
				SecretName: c.Name,
			},
			StorageType:       kubedbv1.StorageTypeEphemeral,
			TerminationPolicy: kubedbv1.TerminationPolicyDelete,
			PodTemplate: ofst.PodTemplateSpec{
			},
		},
	}

	// Set Component instance as the owner and controller
	controllerutil.SetControllerReference(c, postgres, r.scheme)
	return postgres, nil
}

func replicaNumber(num int) *int32 {
	q := int32(num)
	return &q
}

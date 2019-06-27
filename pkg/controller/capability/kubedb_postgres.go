package capability

import (
	kubedbv1 "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)
//buildSecret returns the secret resource
func (r *ReconcileCapability) buildKubeDBPostgres(c *v1alpha2.Capability) (*kubedbv1.Postgres, error) {
	ls := r.GetAppLabels(c.Name)

	//postgresRes := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
/*	postgres := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
		},
	}*/

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
			PodTemplate: ofst.PodTemplateSpec {
				Spec:ofst.PodSpec{
					Env: []core.EnvVar{
						{Name:PG_DATABASE, Value: "my-database"},
					},
				},
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

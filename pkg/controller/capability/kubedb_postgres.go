package capability

import (
	"github.com/appscode/go/encoding/json/types"
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
	paramsMap := r.ParametersAsMap(c.Spec.Parameters)

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
			Version:  SetDefaultDatabaseVersionIfEmpty(c.Spec.Version),
			Replicas: replicaNumber(1),
			UpdateStrategy: apps.StatefulSetUpdateStrategy{
				Type: apps.RollingUpdateStatefulSetStrategyType,
			},
			DatabaseSecret: &core.SecretVolumeSource{
				SecretName: r.SetDefaultSecretNameIfEmpty(c.Spec.SecretName),
			},
			StorageType:       kubedbv1.StorageTypeEphemeral,
			TerminationPolicy: kubedbv1.TerminationPolicyDelete,
			PodTemplate: ofst.PodTemplateSpec{
				Spec: ofst.PodSpec{
					Env: []core.EnvVar{
						{Name: PG_DATABASE, Value: paramsMap["DB_NAME"]},
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

func SetDefaultDatabaseVersionIfEmpty(version string) types.StrYo {
	if version == "10.6-v2" {
		return types.StrYo("10.6")
	} else {
		// Map DB Version with the KubeDB Version
		switch version {
		case "9":
			return types.StrYo("9.6-v4")
		case "10":
			return types.StrYo("10.6-v2")
		case "11":
			return types.StrYo("11.2")
		default:
			return types.StrYo("10.6-v2")
		}
	}
}

/*
	// https://github.com/kubernetes/client-go/tree/master/examples/dynamic-create-update-delete-deployment
	// Approach to create dynamically tyhe object without type imported
	postgresRes := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	postgres := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
		},
	}
	// Create Postgres DB
		fmt.Println("Creating Postgres DB ...")
		result, err := client.Resource(postgresRes).Namespace(namespace).Create(postgres, metav1.CreateOptions{})
		if err != nil {
			panic(err)
		}
*/

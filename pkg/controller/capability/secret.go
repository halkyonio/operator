package capability

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//buildSecret returns the secret resource
func (r *ReconcileCapability) buildSecret(c *v1alpha2.Capability) (*v1.Secret, error) {
	ls := r.GetAppLabels(c.Name)
	paramsMap := r.ParametersAsMap(c.Spec.Parameters)
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.SetDefaultSecretNameIfEmpty(c.Name, paramsMap[DB_CONFIG_NAME]),
			Namespace: c.Namespace,
			Labels:    ls,
		},
		Data: map[string][]byte{
			KUBEDB_PG_USER:     []byte(paramsMap[DB_USER]),
			KUBEDB_PG_PASSWORD: []byte(paramsMap[DB_PASSWORD]),
			KUBEDB_PG_DATABASE_NAME: []byte(paramsMap[DB_NAME]),
			DB_HOST: []byte(r.SetDefaultDatabaseHost(c.Name,paramsMap[DB_HOST]),
			DB_PORT: []byte(r.SetDefaultDatabasePort(paramsMap[DB_PORT])),
		},
	}

	// Set Capability instance as the owner and controller
	controllerutil.SetControllerReference(c, secret, r.scheme)
	return secret, nil
}
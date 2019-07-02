package capability

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"github.com/snowdrop/component-operator/pkg/controller"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
)

type secret struct {
	controller.BaseDependentResource
}

func newSecret() secret {
	return secret{BaseDependentResource: controller.NewDependentResource(&v1.Secret{})}
}

func (res secret) ownerAsCapability() *v1alpha2.Capability {
	return res.Owner().(*v1alpha2.Capability)
}

//buildSecret returns the secret resource
func (res secret) Build() (runtime.Object, error) {
	c := res.ownerAsCapability()
	ls := getAppLabels(c.Name)
	paramsMap := parametersAsMap(c.Spec.Parameters)
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
			KUBEDB_PG_PASSWORD:      []byte(paramsMap[DB_PASSWORD]),
			KUBEDB_PG_DATABASE_NAME: []byte(r.SetDefaultDatabaseName(paramsMap[DB_NAME])),
			// TODO : To be reviewed according to the discussion started with issue #75
			// as we will create another secret when a link will be issued
			DB_HOST:     []byte(r.SetDefaultDatabaseHost(c.Name,paramsMap[DB_HOST])),
			DB_PORT:     []byte(r.SetDefaultDatabasePort(paramsMap[DB_PORT])),
			DB_NAME:     []byte(r.SetDefaultDatabaseName(paramsMap[DB_NAME])),
			DB_USER:     []byte((paramsMap[DB_USER])),
			DB_PASSWORD: []byte(paramsMap[DB_PASSWORD]),
		},
	}

	return secret, nil
}

func (res secret) Name() string {
	return secretNamer(res.ownerAsCapability())
}

func secretNamer(c *v1alpha2.Capability) string {
	name := c.Spec.SecretName
	if len(name) == 0 {
		return strings.ToLower(string(v1alpha2.PostgresKind)) + "-auth"
	} else {
		return name
	}
}

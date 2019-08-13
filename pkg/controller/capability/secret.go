package capability

import (
	"github.com/halkyonio/operator/pkg/apis/halkyon/v1beta1"
	"github.com/halkyonio/operator/pkg/controller"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type secret struct {
	*controller.DependentResourceHelper
}

func (res secret) Update(toUpdate runtime.Object) (bool, error) {
	return false, nil
}

func (res secret) NewInstanceWith(owner v1beta1.Resource) controller.DependentResource {
	return newOwnedSecret(owner)
}

func newSecret() secret {
	return newOwnedSecret(nil)
}

func newOwnedSecret(owner v1beta1.Resource) secret {
	resource := controller.NewDependentResource(&v1.Secret{}, owner)
	s := secret{DependentResourceHelper: resource}
	resource.SetDelegate(s)
	return s
}

func (res secret) ownerAsCapability() *v1beta1.Capability {
	return res.Owner().(*v1beta1.Capability)
}

//buildSecret returns the secret resource
func (res secret) Build() (runtime.Object, error) {
	c := res.ownerAsCapability()
	ls := getAppLabels(c.Name)
	paramsMap := parametersAsMap(c.Spec.Parameters)
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.Name(),
			Namespace: c.Namespace,
			Labels:    ls,
		},
		Data: map[string][]byte{
			KUBEDB_PG_USER:          []byte(paramsMap[DB_USER]),
			KUBEDB_PG_PASSWORD:      []byte(paramsMap[DB_PASSWORD]),
			KUBEDB_PG_DATABASE_NAME: []byte(SetDefaultDatabaseName(paramsMap[DB_NAME])),
			// TODO : To be reviewed according to the discussion started with issue #75
			// as we will create another secret when a link will be issued
			DB_HOST:     []byte(SetDefaultDatabaseHost(c.Name, paramsMap[DB_HOST])),
			DB_PORT:     []byte(SetDefaultDatabasePort(paramsMap[DB_PORT])),
			DB_NAME:     []byte(SetDefaultDatabaseName(paramsMap[DB_NAME])),
			DB_USER:     []byte((paramsMap[DB_USER])),
			DB_PASSWORD: []byte(paramsMap[DB_PASSWORD]),
		},
	}

	return secret, nil
}

func (res secret) Name() string {
	c := res.ownerAsCapability()
	paramsMap := parametersAsMap(c.Spec.Parameters)
	return SetDefaultSecretNameIfEmpty(c.Name, paramsMap[DB_CONFIG_NAME])
}

func (res secret) ShouldWatch() bool {
	return false
}

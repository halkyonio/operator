package openshift

import (
	"context"
	"github.com/openshift/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetDeploymentConfig(namespace string, name string, c client.Client) (*v1.DeploymentConfig, error) {
	dc := &v1.DeploymentConfig{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, dc); err != nil {
		return nil, err
	}
	return dc, nil
}

package capability

import (
	kubedbv1 "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

//Watch for changes to secondary resources and create the owner Capability
//Watch KubeDB PostgresDB resources created in the project/namespace
func watchPostgresDB(c controller.Controller) error {
	err := c.Watch(&source.Kind{Type: &kubedbv1.Postgres{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha2.Capability{},
	})
	return err
}


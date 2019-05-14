package service

import (
	servicecatalogv1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

//Watch for changes to secondary resources and create the owner Service
//Watch ServiceInstance resources created in the project/namespace
func watchServiceInstance(c controller.Controller) error {
	err := c.Watch(&source.Kind{Type: &servicecatalogv1.ServiceInstance{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha2.Service{},
	})
	return err
}

//Watch for changes to secondary resources and create the owner Service
//Watch ServiceBinding resources created in the project/namespace
func watchServiceBinding(c controller.Controller) error {
	err := c.Watch(&source.Kind{Type: &servicecatalogv1.ServiceBinding{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha2.Service{},
	})
	return err
}


package framework

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"halkyon.io/operator/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ReconcilerFactory interface {
	PrimaryResourceType() Resource
	WatchedSecondaryResourceTypes() []runtime.Object
	Delete(object Resource) error
	CreateOrUpdate(object Resource) error
	Helper() ReconcilerHelper
	GetDependentResourceFor(owner Resource, resourceType runtime.Object) (DependentResource, error)
	AddDependentResource(resource DependentResource)
	SetPrimaryResourceStatus(primary Resource, statuses []DependentResourceStatus) (needsUpdate bool)
}

type ReconcilerHelper struct {
	Client    client.Client
	Config    *rest.Config
	Scheme    *runtime.Scheme
	ReqLogger logr.Logger
}

func (rh ReconcilerHelper) Fetch(name, namespace string, into runtime.Object) (runtime.Object, error) {
	if err := rh.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, into); err != nil {
		if errors.IsNotFound(err) {
			return into, err
		}
		return into, fmt.Errorf("couldn't fetch '%s' %s from namespace '%s': %s", name, util.GetObjectName(into), namespace, err.Error())
	}
	return into, nil
}

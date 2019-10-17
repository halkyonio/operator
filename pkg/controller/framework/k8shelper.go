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
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var helpers = make(map[string]*K8SHelper, 7)

type K8SHelper struct {
	Client    client.Client
	Config    *rest.Config
	Scheme    *runtime.Scheme
	ReqLogger logr.Logger
}

func (rh K8SHelper) Fetch(name, namespace string, into runtime.Object) (runtime.Object, error) {
	if err := rh.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, into); err != nil {
		if errors.IsNotFound(err) {
			return into, err
		}
		return into, fmt.Errorf("couldn't fetch '%s' %s from namespace '%s': %s", name, util.GetObjectName(into), namespace, err.Error())
	}
	return into, nil
}

func NewHelper(nameForLogger string, resourceType runtime.Object, mgr manager.Manager) *K8SHelper {
	config := mgr.GetConfig()
	helper := &K8SHelper{
		Client:    mgr.GetClient(),
		Config:    config,
		Scheme:    mgr.GetScheme(),
		ReqLogger: log.Log.WithName(nameForLogger),
	}
	helpers[util.GetObjectName(resourceType)] = helper
	checkIfOpenShift(config)
	return helper
}

func GetHelperFor(resourceType runtime.Object) *K8SHelper {
	return helpers[util.GetObjectName(resourceType)]
}

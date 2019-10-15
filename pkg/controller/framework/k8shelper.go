package framework

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"halkyon.io/operator/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strings"
)

type K8SHelper struct {
	Client           client.Client
	Config           *rest.Config
	Scheme           *runtime.Scheme
	ReqLogger        logr.Logger
	onOpenShift      *bool
	openShiftVersion int
}

func (rh *K8SHelper) OpenShiftVersion() int {
	rh.IsTargetClusterRunningOpenShift() // make sure things are properly initialized
	return rh.openShiftVersion
}

func (rh *K8SHelper) IsTargetClusterRunningOpenShift() bool {
	if rh.onOpenShift == nil {
		discoveryClient, err := discovery.NewDiscoveryClientForConfig(rh.Config)
		if err != nil {
			panic(err)
		}
		apiList, err := discoveryClient.ServerGroups()
		if err != nil {
			panic(err)
		}
		apiGroups := apiList.Groups
		const openShiftGroupSuffix = ".openshift.io"
		const openShift4GroupName = "config" + openShiftGroupSuffix
		for _, group := range apiGroups {
			if strings.HasSuffix(group.Name, openShiftGroupSuffix) {
				if rh.onOpenShift == nil {
					rh.onOpenShift = util.NewTrue()
					rh.openShiftVersion = 3
				}
				if group.Name == openShift4GroupName {
					rh.openShiftVersion = 4
					break
				}
			}
		}

		if rh.onOpenShift == nil {
			// we didn't find any api group with the openshift.io suffix, so we're not on OpenShift!
			rh.onOpenShift = util.NewFalse()
			rh.openShiftVersion = 0
		}
	}

	return *rh.onOpenShift
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

func NewHelper(nameForLogger string, mgr manager.Manager) *K8SHelper {
	helper := &K8SHelper{
		Client:    mgr.GetClient(),
		Config:    mgr.GetConfig(),
		Scheme:    mgr.GetScheme(),
		ReqLogger: log.Log.WithName(nameForLogger),
	}
	return helper
}

package framework

import (
	"halkyon.io/operator/pkg/util"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"strings"
)

var (
	onOpenShift      *bool
	openShiftVersion int
)

func OpenShiftVersion() int {
	return openShiftVersion
}

func IsTargetClusterRunningOpenShift() bool {
	return *onOpenShift
}

func checkIfOpenShift(config *rest.Config) {
	if onOpenShift == nil {
		discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
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
				if onOpenShift == nil {
					onOpenShift = util.NewTrue()
					openShiftVersion = 3
				}
				if group.Name == openShift4GroupName {
					openShiftVersion = 4
					break
				}
			}
		}

		if onOpenShift == nil {
			// we didn't find any api group with the openshift.io suffix, so we're not on OpenShift!
			onOpenShift = util.NewFalse()
			openShiftVersion = 0
		}
	}
}

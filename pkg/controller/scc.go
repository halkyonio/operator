package controller

import (
	"fmt"
	securityv1 "github.com/openshift/api/security/v1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"github.com/snowdrop/component-operator/pkg/util"
	"k8s.io/apimachinery/pkg/runtime"
)

type serviceAccountNamer func(owner v1alpha2.Resource) string
type scc struct {
	*DependentResourceHelper
	serviceAccountNamer serviceAccountNamer
	reconciler          *BaseGenericReconciler
}

func (res scc) NewInstanceWith(owner v1alpha2.Resource) DependentResource {
	return newOwnedScc(owner, res.serviceAccountNamer)
}

func NewScc(serviceAccountNamer serviceAccountNamer) scc {
	return newOwnedScc(nil, serviceAccountNamer)
}

func newOwnedScc(owner v1alpha2.Resource, serviceAccountNamer serviceAccountNamer) scc {
	dependent := NewDependentResource(&securityv1.SecurityContextConstraints{}, owner)
	scc := scc{
		DependentResourceHelper: dependent,
		serviceAccountNamer:     serviceAccountNamer,
	}
	dependent.SetDelegate(scc)
	return scc
}

func (res scc) Name() string {
	return "privileged"
}

func (res scc) Build() (runtime.Object, error) {
	panic("scc.Build should never be called: check your code!")
}

func (res scc) Update(toUpdate runtime.Object) (bool, error) {
	toUpdateSCC := toUpdate.(*securityv1.SecurityContextConstraints)
	owner := res.Owner()
	sccUser := fmt.Sprintf("system:serviceaccount:%s:%s", owner.GetNamespace(), res.serviceAccountNamer(owner))
	if util.Index(toUpdateSCC.Users, sccUser) >= 0 {
		toUpdateSCC.Users = append(toUpdateSCC.Users, sccUser)
		return true, nil
	}
	return false, nil
}

func (res scc) ShouldWatch() bool {
	return false
}

func (res scc) CanBeCreatedOrUpdated() bool {
	return res.reconciler.IsTargetClusterRunningOpenShift()
}

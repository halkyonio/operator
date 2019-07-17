package component

import (
	"fmt"
	securityv1 "github.com/openshift/api/security/v1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"github.com/snowdrop/component-operator/pkg/controller"
	"github.com/snowdrop/component-operator/pkg/util"
	"k8s.io/apimachinery/pkg/runtime"
)

type scc struct {
	base
}

func (res scc) NewInstanceWith(owner v1alpha2.Resource) controller.DependentResource {
	return newOwnedScc(owner)
}

func newScc() scc {
	return newOwnedScc(nil)
}

func newOwnedScc(owner v1alpha2.Resource) scc {
	dependent := newBaseDependent(&securityv1.SecurityContextConstraints{}, owner)
	scc := scc{
		base: dependent,
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
	sccUser := fmt.Sprintf("system:serviceaccount:%s:%s", res.Owner().GetNamespace(), serviceAccountName)
	if util.Index(toUpdateSCC.Users, sccUser) >= 0 {
		toUpdateSCC.Users = append(toUpdateSCC.Users, sccUser)
		return true, nil
	}
	return false, nil
}

func (res scc) ShouldWatch() bool {
	return false
}

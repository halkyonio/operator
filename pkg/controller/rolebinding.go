package controller

import (
	"halkyon.io/operator-framework"
	authorizv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type RoleBinding struct {
	*framework.DependentResourceHelper
	namer               func() string
	associatedRoleNamer func() string
	serviceAccountNamer func() string
}

func (res RoleBinding) Update(toUpdate runtime.Object) (bool, error) {
	// add appropriate subject for owner
	rb := toUpdate.(*authorizv1.RoleBinding)
	owner := res.Owner()

	// check if the binding contains the current owner as subject
	namespace := owner.GetNamespace()
	name := res.serviceAccountNamer()
	found := false
	for _, subject := range rb.Subjects {
		if subject.Name == name && subject.Namespace == namespace {
			found = true
			break
		}
	}

	if !found {
		rb.Subjects = append(rb.Subjects, authorizv1.Subject{
			Kind:      "ServiceAccount",
			Namespace: namespace,
			Name:      name,
		})
	}

	return !found, nil
}

func (res RoleBinding) NewInstanceWith(owner framework.Resource) framework.DependentResource {
	return NewOwnedRoleBinding(owner, res.namer, res.associatedRoleNamer, res.serviceAccountNamer)
}

func NewOwnedRoleBinding(owner framework.Resource, namer, associatedRoleNamer, serviceAccountNamer func() string) RoleBinding {
	dependent := framework.NewDependentResource(&authorizv1.RoleBinding{}, owner)
	rolebinding := RoleBinding{
		DependentResourceHelper: dependent,
		namer:                   namer,
		associatedRoleNamer:     associatedRoleNamer,
		serviceAccountNamer:     serviceAccountNamer,
	}
	dependent.SetDelegate(rolebinding)
	return rolebinding
}

func (res RoleBinding) Name() string {
	return res.namer()
}

func (res RoleBinding) Build() (runtime.Object, error) {
	c := res.Owner()
	namespace := c.GetNamespace()
	ser := &authorizv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.Name(),
			Namespace: namespace,
		},
		RoleRef: authorizv1.RoleRef{
			Kind: "Role",
			Name: res.associatedRoleNamer(),
		},
		Subjects: []authorizv1.Subject{
			{Kind: "ServiceAccount", Name: res.serviceAccountNamer(), Namespace: namespace},
		},
	}
	return ser, nil
}

func (res RoleBinding) ShouldWatch() bool {
	return false
}

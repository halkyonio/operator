package v1alpha1

import (
	"fmt"
	"github.com/onsi/gomega"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

type Scheme struct {
	Types map[string]reflect.Type
}

func NewScheme() *Scheme {
	return &Scheme{
		Types: map[string]reflect.Type{},
	}
}

type Foo struct {
	Name   string
	Ref    *corev1.ObjectReference
	Object interface{}
}

func TestComponentObject(t *testing.T) {
	s := NewScheme()
	component := Foo{
		Name: "Hello",
	}

	env, tp := s.CreateEnvVar()
	component.Object = env
	s.Types[tp.String()] = tp

	//envVar := &corev1.EnvVar{}
	for idx, _ := range s.Types {
		if s.Types[idx].String() == "**v1.EnvVar" {
			envVar := component.Object.(*corev1.EnvVar)
			g := gomega.NewGomegaWithT(t)
			g.Expect(envVar.Name).To(gomega.Equal("Foo"))
			g.Expect(envVar.Value).To(gomega.Equal("Bar"))
		}
	}

}

func (s *Scheme) CreateEnvVar() (interface{}, reflect.Type) {
	env := &corev1.EnvVar{
		Name:  "Foo",
		Value: "Bar",
	}
	return env, reflect.TypeOf(&env)
}

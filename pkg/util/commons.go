package util

import (
	"fmt"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
)

func GetImageReference(imageName string, version ...string) string {
	runtimeVersion := "latest"
	if len(version) == 1 && len(version[0]) > 0 {
		runtimeVersion = version[0]
	}
	return fmt.Sprintf("%s:%s", imageName, runtimeVersion)
}

func GetObjectName(object runtime.Object) string {
	t := reflect.TypeOf(object)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

func Index(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

func DefaultDependentResourceNameFor(owner v1alpha2.Resource) string {
	return owner.GetName()
}

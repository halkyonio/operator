package util

import (
	"fmt"
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

func Index(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

func NewTrue() *bool {
	b := true
	return &b
}

func NewFalse() *bool {
	b := false
	return &b
}

func GetObjectName(object runtime.Object) string {
	t := reflect.TypeOf(object)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

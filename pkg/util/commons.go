package util

import (
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
)

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

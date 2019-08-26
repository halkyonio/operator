package controller

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Resource interface {
	v1.Object
	runtime.Object
	NeedsRequeue() bool
	SetNeedsRequeue(requeue bool)
	GetStatusAsString() string
	ShouldDelete() bool
	SetErrorStatus(err error) bool
	SetSuccessStatus(dependentName, msg string) bool
	SetInitialStatus(msg string) bool
	IsValid() bool
}

package controller

import (
	halkyon "halkyon.io/api/link/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Link struct {
	*halkyon.Link
	requeue bool
}

func (in *Link) SetAPIObject(object runtime.Object) {
	in.Link = object.(*halkyon.Link)
}

func (in *Link) GetAPIObject() runtime.Object {
	return in.Link
}

func (in *Link) Clone() Resource {
	link := NewLink(in.Link)
	link.requeue = in.requeue
	return link
}

func NewLink(link ...*halkyon.Link) *Link {
	l := &halkyon.Link{}
	if link != nil {
		l = link[0]
	}
	return &Link{
		Link:    l,
		requeue: false,
	}
}

func (in *Link) SetNeedsRequeue(requeue bool) {
	in.requeue = in.requeue || requeue
}

func (in *Link) NeedsRequeue() bool {
	return in.requeue
}

func (in *Link) SetInitialStatus(msg string) bool {
	if halkyon.LinkPending != in.Status.Phase || msg != in.Status.Message {
		in.Status.Phase = halkyon.LinkPending
		in.Status.Message = msg
		in.SetNeedsRequeue(true)
		return true
	}
	return false
}

func (in *Link) IsValid() bool {
	return true // todo: implement me
}

func (in *Link) SetErrorStatus(err error) bool {
	errMsg := err.Error()
	if halkyon.LinkFailed != in.Status.Phase || errMsg != in.Status.Message {
		in.Status.Phase = halkyon.LinkFailed
		in.Status.Message = errMsg
		in.SetNeedsRequeue(true)
		return true
	}
	return false
}

func (in *Link) SetSuccessStatus(statuses []DependentResourceStatus, msg string) bool {
	if halkyon.LinkReady != in.Status.Phase || msg != in.Status.Message {
		in.Status.Phase = halkyon.LinkReady
		in.Status.Message = msg
		in.requeue = true
		return true
	}
	return false
}

func (in *Link) GetStatusAsString() string {
	return in.Status.Phase.String()
}

func (in *Link) ShouldDelete() bool {
	return true
}

package controller

import (
	halkyon "halkyon.io/api/capability/v1beta1"
)

type Capability struct {
	*halkyon.Capability
	requeue bool
}

func (in *Capability) SetNeedsRequeue(requeue bool) {
	in.requeue = in.requeue || requeue
}

func (in *Capability) NeedsRequeue() bool {
	return in.requeue
}

func (in *Capability) SetInitialStatus(msg string) bool {
	if halkyon.CapabilityPending != in.Status.Phase || in.Status.Message != msg {
		in.Status.Phase = halkyon.CapabilityPending
		in.Status.Message = msg
		in.SetNeedsRequeue(true)
		return true
	}
	return false
}

func (in *Capability) IsValid() bool {
	return true // todo: implement me
}

func (in *Capability) SetErrorStatus(err error) bool {
	errMsg := err.Error()
	if halkyon.CapabilityFailed != in.Status.Phase || errMsg != in.Status.Message {
		in.Status.Phase = halkyon.CapabilityFailed
		in.Status.Message = errMsg
		in.SetNeedsRequeue(false)
		return true
	}
	return false
}

func (in *Capability) SetSuccessStatus(dependentName, msg string) bool {
	if dependentName != in.Status.PodName || halkyon.CapabilityReady != in.Status.Phase || msg != in.Status.Message {
		in.Status.Phase = halkyon.CapabilityReady
		in.Status.PodName = dependentName
		in.Status.Message = msg
		in.requeue = false
		return true
	}
	return false
}

func (in *Capability) GetStatusAsString() string {
	return in.Status.Phase.String()
}

func (in *Capability) ShouldDelete() bool {
	return true
}

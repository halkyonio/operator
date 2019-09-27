package controller

import "testing"

func TestCapabilitySetSuccessStatus(t *testing.T) {
	c := NewCapability()
	c.Status.PodName = "initial"

	const s = "new pod name"
	const msg = "foo"
	changed := c.SetSuccessStatus([]DependentResourceStatus{NewReadyDependentResourceStatus(s, "PodName")}, msg)
	if !changed || c.Status.PodName != s || c.Status.Message != msg {
		t.Fail()
	}
}

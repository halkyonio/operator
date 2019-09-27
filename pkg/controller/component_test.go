package controller

import "testing"

func TestComponentSetSuccessStatus(t *testing.T) {
	c := NewComponent()
	c.Status.PodName = "initial"

	const s = "new pod name"
	const msg = "foo"
	changed := c.SetSuccessStatus([]DependentResourceStatus{NewReadyDependentResourceStatus(s, "PodName")}, msg)
	if !changed || c.Status.PodName != s || c.Status.Message != msg {
		t.Fail()
	}
}

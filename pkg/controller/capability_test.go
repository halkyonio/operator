package controller

import (
	"strings"
	"testing"
)

func TestCapabilitySetSuccessStatus(t *testing.T) {
	c := NewCapability()
	c.Status.PodName = "initial"

	const s = "new pod name"
	const msg = "foo"
	const fieldName = "PodName"
	changed := c.SetSuccessStatus([]DependentResourceStatus{NewReadyDependentResourceStatus(s, fieldName)}, msg)
	if !changed {
		t.Errorf("expected updates from SetSuccessStatus, got none")
	}
	if c.Status.PodName != s {
		t.Errorf("expected pod name to be changed to '%s', got '%s'", s, c.Status.PodName)
	}
	if !strings.HasPrefix(c.Status.Message, msg) && !strings.Contains(c.Status.Message, fieldName) {
		t.Errorf("expected message to start with '%s' and contain '%s', got \"%s\"", msg, fieldName, c.Status.Message)
	}
}

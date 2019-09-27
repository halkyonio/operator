package controller

import (
	"halkyon.io/api/component/v1beta1"
	"testing"
)

func TestSetNamedStringField(t *testing.T) {
	var tests = []struct {
		testName string
		object   interface{}
		field    string
		value    string
		want     bool
		error    bool
	}{
		{"correct", &v1beta1.ComponentStatus{PodName: "foo"}, "PodName", "bar", true, false},
		{"unchanged value", &v1beta1.ComponentStatus{PodName: "foo"}, "PodName", "foo", false, false},
		{"need to pass pointer to set value", v1beta1.ComponentStatus{PodName: "foo"}, "PodName", "bar", false, true},
		{"inexistent field name", &v1beta1.ComponentStatus{}, "inexistent", "bar", false, true},
		{"empty field name == noop", &v1beta1.ComponentStatus{}, "", "bar", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			changed, err := SetNamedStringField(tt.object, tt.field, tt.value)
			if err != nil && !tt.error {
				t.Errorf("got error '%v' when none was expected", err)
			}
			if changed != tt.want {
				t.Errorf("expected changed status to be %t, got %t", tt.want, changed)
			}
		})
	}
}

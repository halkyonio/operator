package v1alpha1

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

// Spec
type Spec struct {
  App        string      `json:"app"`
  InfoItems  []InfoItem  `json:"infoitems"`
}

// InfoItemType is a string that describes the value of InfoItem
type InfoItemType string

// InfoItemSource represents a source for the value of an InfoItem.
type InfoItemSource struct {}

// InfoItem is a human readable key,value pair containing important information about how to access the Application.
type InfoItem struct {
	// Name is a human readable title for this piece of information.
	Name string `json:"name,omitempty"`

	// Type of the value for this InfoItem.
	Type InfoItemType `json:"type,omitempty"`

	// Value is human readable content.
	Value string `json:"value,omitempty"`

	// ValueFrom defines a reference to derive the value from another source.
	ValueFrom *InfoItemSource `json:"valueFrom,omitempty"`

	Object interface{}
}

var data = `
app: my-app
infoitems:
- Name: Spring Boot v1
  Type: EnvVar
  Object: 
    - key: JAVA_HOME
      value: /usr/local/bin/java
    - key: JAVA_DIR
      value: /usr/local/my-app
`

func TestUnMarshalling(t *testing.T) {
	spec := Spec{}
	err := yaml.Unmarshal([]byte(data), &spec)
	if err != nil {
		fmt.Print(err.Error())
	}

	for _, info := range spec.InfoItems {
		spew.Dump(CreateEnvVar(info.Type, info.Object))
	}
}

func CreateEnvVar(Type InfoItemType, obj interface{}) *[]corev1.EnvVar {
	envVars := []corev1.EnvVar{}

	r := obj.([]interface{})
	for _, i := range r {
		m := i.(map[string]interface{})
		n := m["key"].(string)
		v := m["value"].(string)

		if Type == "EnvVar" {
			env := corev1.EnvVar{
				Name:  n,
				Value: v,
			}
			envVars = append(envVars, env)
			return &envVars
		}
	}
	return nil
}

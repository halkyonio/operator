package capability

import (
	"fmt"
	halkyon "halkyon.io/api/capability/v1beta1"
	"halkyon.io/operator-framework"
	"halkyon.io/plugins/capability"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"path/filepath"
	"plugin"
)

const (
	SECRET = "Secret"
	// KubeDB Postgres const
	KUBEDB_PG_DATABASE      = "Postgres"
	KUBEDB_PG_DATABASE_NAME = "POSTGRES_DB"
	KUBEDB_PG_USER          = "POSTGRES_USER"
	KUBEDB_PG_PASSWORD      = "POSTGRES_PASSWORD"
	// Capability const
	DB_CONFIG_NAME = "DB_CONFIG_NAME"
	DB_HOST        = "DB_HOST"
	DB_PORT        = "DB_PORT"
	DB_NAME        = "DB_NAME"
	DB_USER        = "DB_USER"
	DB_PASSWORD    = "DB_PASSWORD"
)

type typeRegistry map[halkyon.CapabilityType]bool
type categoryRegistry map[halkyon.CapabilityCategory]typeRegistry

var (
	plugins             []capability.Plugin
	supportedCategories categoryRegistry
)

// blank assignment to check that Capabilit implements Resource
var _ framework.Resource = &Capability{}

type Capability struct {
	*halkyon.Capability
	*framework.BaseResource
}

func (in *Capability) Delete() error {
	return nil
}

func (in *Capability) CreateOrUpdate() error {
	return in.CreateOrUpdateDependents()
}

func (in *Capability) FetchAndCreateNew(name, namespace string) (framework.Resource, error) {
	return in.BaseResource.FetchAndInitNewResource(name, namespace, NewCapability())
}

func (in *Capability) ComputeStatus() (needsUpdate bool) {
	statuses, notReadyWantsUpdate := in.BaseResource.ComputeStatus(in)
	return notReadyWantsUpdate || in.SetSuccessStatus(statuses, "Ready")
}

func (in *Capability) Init() bool {
	return false
}

func (in *Capability) GetAPIObject() runtime.Object {
	return in.Capability
}

func NewCapability() *Capability {
	dependents := framework.NewHasDependents(&halkyon.Capability{})
	c := &Capability{
		Capability:   &halkyon.Capability{},
		BaseResource: dependents,
	}
	for _, p := range plugins {
		dependents.AddDependentResource(p.GetDependentResources()...)
	}
	return c
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

func (in *Capability) CheckValidity() error {
	category := in.Spec.Category
	types := supportedCategories[category]
	if len(types) == 0 {
		return fmt.Errorf("unsupported '%s' capability category", category)
	}
	t := in.Spec.Type
	if !types[t] {
		return fmt.Errorf("unsupported '%s' type for '%s'", t, category)
	}
	return nil
}

func (in *Capability) SetErrorStatus(err error) bool {
	if err != nil {
		errMsg := err.Error()
		if halkyon.CapabilityFailed != in.Status.Phase || errMsg != in.Status.Message {
			in.Status.Phase = halkyon.CapabilityFailed
			in.Status.Message = errMsg
			in.SetNeedsRequeue(false)
			return true
		}
	}
	return false
}

func (in *Capability) DependentStatusFieldName() string {
	_ = in.Status.PodName // to make sure we update the value below if that field changes as returned value must match field name
	return "PodName"
}

func (in *Capability) SetSuccessStatus(statuses []framework.DependentResourceStatus, msg string) bool {
	changed, updatedMsg := framework.HasChangedFromStatusUpdate(&in.Status, statuses, msg)
	if changed || halkyon.CapabilityReady != in.Status.Phase {
		in.Status.Phase = halkyon.CapabilityReady
		in.Status.Message = updatedMsg
		in.SetNeedsRequeue(false)
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

func init() {
	plugins = make([]capability.Plugin, 0, 7)
	supportedCategories = make(categoryRegistry, 7)
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	pluginsDir := filepath.Join(currentDir, "plugins")
	goPlugins, err := ioutil.ReadDir(pluginsDir)
	if err != nil {
		panic(err)
	}
	for _, p := range goPlugins {
		pluginPath := filepath.Join(pluginsDir, p.Name())
		if goPlugin, err := plugin.Open(pluginPath); err == nil {
			if maybePlugin, err := goPlugin.Lookup("Plugin"); err == nil {
				if plug, ok := maybePlugin.(capability.Plugin); ok {
					plugins = append(plugins, plug)
					category := plug.GetCategory()
					types, ok := supportedCategories[category]
					if !ok {
						types = make(typeRegistry, 3)
						supportedCategories[category] = types
					}
					types[plug.GetType()] = true
				}
			} else {
				panic("Couldn't load Plugin var from plugin " + pluginPath)
			}
		} else {
			panic(err)
		}
	}
}

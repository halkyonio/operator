package capability

import (
	"fmt"
	halkyon "halkyon.io/api/capability/v1beta1"
	"halkyon.io/operator/pkg/controller/framework"
	"k8s.io/apimachinery/pkg/runtime"
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

type Capability struct {
	*halkyon.Capability
	*framework.HasDependents
	dependentTypes []framework.DependentResource
}

func (in *Capability) PrimaryResourceType() runtime.Object {
	return &halkyon.Capability{}
}

func (in *Capability) Delete() error {
	return nil
}

func (in *Capability) CreateOrUpdate() error {
	helper := framework.GetHelperFor(in.PrimaryResourceType())
	return in.CreateOrUpdateDependents(helper)
}

func (in *Capability) GetDependentResourcesTypes() []framework.DependentResource {
	if len(in.dependentTypes) == 0 {
		in.dependentTypes = []framework.DependentResource{
			newSecret(),
			newPostgres(),
			newRole(nil),
			newRoleBinding(nil),
		}
	}
	return in.dependentTypes
}

func (in *Capability) FetchAndInit(name, namespace string) (framework.Resource, error) {
	return in.HasDependents.FetchAndInitNewResource(name, namespace, in)
}

func (in *Capability) ComputeStatus(err error, helper *framework.K8SHelper) (needsUpdate bool) {
	statuses, update := in.HasDependents.ComputeStatus(in, err, helper)
	return in.SetSuccessStatus(statuses, "Ready") || update
}

func (in *Capability) Init() bool {
	return false
}

func (in *Capability) GetAPIObject() runtime.Object {
	return in.Capability
}

func NewCapability(capability ...*halkyon.Capability) *Capability {
	c := &halkyon.Capability{}
	if capability != nil {
		c = capability[0]
	}
	return &Capability{
		Capability:    c,
		HasDependents: framework.NewHasDependents(),
	}
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
	if !halkyon.DatabaseCategory.Equals(in.Spec.Category) {
		return fmt.Errorf("unsupported '%s' capability category", in.Spec.Category)
	}
	if !halkyon.PostgresType.Equals(in.Spec.Type) {
		return fmt.Errorf("unsupported '%s' database type", in.Spec.Type)
	}
	return nil
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

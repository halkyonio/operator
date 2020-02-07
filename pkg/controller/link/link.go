package link

import (
	"context"
	"fmt"
	"halkyon.io/api/component/v1beta1"
	halkyon "halkyon.io/api/link/v1beta1"
	halkyon2 "halkyon.io/api/v1beta1"
	"halkyon.io/operator-framework"
	"halkyon.io/operator/pkg"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

type Link struct {
	*halkyon.Link
	*framework.BaseResource
}

func (in *Link) InitDependents() []framework.DependentResource {
	res := []framework.DependentResource{newComponent(in.Link)}
	in.BaseResource.AddDependentResource(res...)
	return res
}

// blank assignment to check that Link implements Resource
var _ framework.Resource = &Link{}

func (in *Link) Delete() error {
	return nil
}

func (in *Link) CreateOrUpdate() error {
	// get the associated component
	fetch, err := in.FetchUpdatedDependent(framework.TypePredicateFor(v1beta1.SchemeGroupVersion.WithKind(v1beta1.Kind)))
	if err != nil {
		return fmt.Errorf("couldn't find associated component named '%s'", in.Spec.ComponentName)
	}
	c := fetch.(*v1beta1.Component)

	// Enrich the Deployment object using the information passed within the Link Spec (e.g Env Vars, EnvFrom, ...) if needed
	return updateComponentWithLinkInfo(in, c)
}

func (in *Link) NewEmpty() framework.Resource {
	return NewLink()
}

func (in *Link) InitDependentResources() ([]framework.DependentResource, error) {
	return in.BaseResource.AddDependentResource(newComponent(in.Link)), nil
}

func (in *Link) ComputeStatus() (needsUpdate bool) {
	statuses, notReadyWantsUpdate := in.BaseResource.ComputeStatus(in)
	return notReadyWantsUpdate || in.SetSuccessStatus(statuses, "Ready")
}

func (in *Link) Init() bool {
	return false
}

func (in *Link) GetAsHalkyonResource() halkyon2.HalkyonResource {
	return in.Link
}

func NewLink() *Link {
	dependents := framework.NewHasDependents(&halkyon.Link{})
	l := &Link{
		Link:         &halkyon.Link{},
		BaseResource: dependents,
	}
	return l
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

func (in *Link) CheckValidity() error {
	return nil // todo: implement me
}

func (in *Link) SetErrorStatus(err error) bool {
	if err != nil {
		errMsg := err.Error()
		if halkyon.LinkFailed != in.Status.Phase || errMsg != in.Status.Message {
			in.Status.Phase = halkyon.LinkFailed
			in.Status.Message = errMsg
			in.SetNeedsRequeue(false)
			return true
		}
	}
	return false
}

func (in *Link) SetSuccessStatus(statuses []framework.DependentResourceStatus, msg string) bool {
	if halkyon.LinkReady != in.Status.Phase || msg != in.Status.Message {
		in.Status.Phase = halkyon.LinkReady
		in.Status.Message = msg
		in.SetNeedsRequeue(false)
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

func updateComponentWithLinkInfo(l *Link, component *v1beta1.Component) error {
	var isModified = false
	linkType := l.Spec.Type
	deployment := &appsv1.Deployment{}
	if err := framework.Helper.Client.Get(context.TODO(), types.NamespacedName{Name: pkg.DeploymentName(component), Namespace: l.Namespace}, deployment); err != nil {
		return fmt.Errorf("couldn't retrieve deployment for component '%s'", component.Name)
	}
	containers := deployment.Spec.Template.Spec.Containers
	switch linkType {
	case halkyon.SecretLinkType:
		secretName := l.Spec.Ref

		// Check if EnvFrom already exists
		// If this is the case, exit without error
		for i := 0; i < len(containers); i++ {
			var isEnvFromExist = false
			for _, env := range containers[i].EnvFrom {
				if env.SecretRef.Name == secretName {
					// EnvFrom already exists for the Secret Ref
					isEnvFromExist = true
				}
			}
			if !isEnvFromExist {
				// Add the Secret as EnvVar to the container
				containers[i].EnvFrom = append(containers[i].EnvFrom, addSecretAsEnvFromSource(secretName))
				isModified = true
			}
		}

	case halkyon.EnvLinkType:
		// Check if Env already exists
		// If this is the case, exit without error
		for i := 0; i < len(containers); i++ {
			var isEnvExist = false
			for _, specEnv := range l.Spec.Envs {
				for _, env := range containers[i].Env {
					if specEnv.Name == env.Name && specEnv.Value == env.Value {
						// EnvFrom already exists for the Secret Ref
						isEnvExist = true
					}
				}
				if !isEnvExist {
					// Add the Secret as EnvVar to the container
					containers[i].Env = append(containers[i].Env, addKeyValueAsEnvVar(specEnv.Name, specEnv.Value))
					isModified = true
				}
			}
		}
	}

	if isModified {
		deployment.Spec.Template.Spec.Containers = containers
		if err := framework.Helper.Client.Update(context.TODO(), deployment); err != nil {
			// As it could be possible that we can't update the Deployment as it has been modified by another
			// process, then we will requeue
			l.SetNeedsRequeue(true)
			return err
		}
		compStatus := &component.Status
		compStatus.Phase = v1beta1.ComponentLinking
		compStatus.SetLinkingStatus(l.Name, v1beta1.Started, compStatus.PodName)
		compStatus.PodName = ""
		compStatus.Message = fmt.Sprintf("Initiating link %s", l.Name)
		err := framework.Helper.Client.Status().Update(context.TODO(), component)
		if err != nil {
			return err
		}
	}

	return nil
}

func addSecretAsEnvFromSource(secretName string) v1.EnvFromSource {
	return v1.EnvFromSource{
		SecretRef: &v1.SecretEnvSource{
			LocalObjectReference: v1.LocalObjectReference{Name: secretName},
		},
	}
}

func addKeyValueAsEnvVar(key, value string) v1.EnvVar {
	return v1.EnvVar{
		Name:  key,
		Value: value,
	}
}

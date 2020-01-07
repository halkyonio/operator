package link

import (
	"context"
	"fmt"
	"halkyon.io/api/component/v1beta1"
	halkyon "halkyon.io/api/link/v1beta1"
	halkyon2 "halkyon.io/api/v1beta1"
	"halkyon.io/operator-framework"
	"halkyon.io/operator-framework/util"
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
	found, err := fetchDeployment(in.Link)
	if err != nil {
		in.SetNeedsRequeue(true)
		return err
	}
	// Enrich the Deployment object using the information passed within the Link Spec (e.g Env Vars, EnvFrom, ...) if needed
	if containers, isModified := updateContainersWithLinkInfo(in.Link, found.Spec.Template.Spec.Containers); isModified {
		found.Spec.Template.Spec.Containers = containers

		if err = updateDeploymentWithLink(found, in); err != nil {
			// As it could be possible that we can't update the Deployment as it has been modified by another
			// process, then we will requeue
			in.SetNeedsRequeue(true)
			return err
		}

		// if the deployment has been updated, we need to update the component's status
		fetch, err := in.FetchUpdatedDependent(util.GetObjectName(&v1beta1.Component{}))
		if err != nil {
			return fmt.Errorf("cannot retrieve associated component")
		}
		c := fetch.(*v1beta1.Component)
		compStatus := &c.Status
		compStatus.Phase = v1beta1.ComponentLinking
		compStatus.SetLinkingStatus(in.Name, v1beta1.Started, compStatus.PodName)
		compStatus.PodName = ""
		compStatus.Message = fmt.Sprintf("Initiating link %s", in.Name)
		err = framework.Helper.Client.Status().Update(context.TODO(), c)
		if err != nil {
			return err
		}
	}

	return nil
}

func (in *Link) FetchAndCreateNew(name, namespace string, callback framework.WatchCallback) (framework.Resource, error) {
	return framework.FetchAndInitNewResource(name, namespace, NewLink(), callback, func(toInit halkyon2.HalkyonResource) ([]framework.DependentResource, error) {
		return []framework.DependentResource{newComponent(toInit.(*halkyon.Link))}, nil
	})
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

func fetchDeployment(link *halkyon.Link) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	name := link.Spec.ComponentName
	if err := framework.Helper.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: link.Namespace}, deployment); err == nil {
		return deployment, nil
	} else if err := framework.Helper.Client.Get(context.TODO(), types.NamespacedName{Name: name + "-build", Namespace: link.Namespace}, deployment); err == nil {
		return deployment, nil
	} else {
		framework.LoggerFor(link).Info("Deployment doesn't exist", "Name", name)
		return deployment, err
	}
}

func updateContainersWithLinkInfo(l *halkyon.Link, containers []v1.Container) ([]v1.Container, bool) {
	var isModified = false
	linkType := l.Spec.Type
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

	return containers, isModified
}

func updateDeploymentWithLink(d *appsv1.Deployment, link *Link) error {
	// Update the Deployment of the component
	logger := framework.LoggerFor(link.GetAsHalkyonResource())
	if err := framework.Helper.Client.Update(context.TODO(), d); err != nil {
		logger.Info("Failed to update deployment.")
		return err
	}
	logger.Info("Deployment updated.")

	name := link.Spec.ComponentName
	link.SetSuccessStatus([]framework.DependentResourceStatus{}, fmt.Sprintf("linked to '%s' component", name))
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

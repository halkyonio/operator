package link

import (
	"context"
	"fmt"
	"halkyon.io/api/component/v1beta1"
	halkyon "halkyon.io/api/link/v1beta1"
	"halkyon.io/operator/pkg/controller/framework"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type Link struct {
	*halkyon.Link
	*framework.Requeueable
	*framework.HasDependents
	dependentTypes []framework.DependentResource
}

func (in *Link) PrimaryResourceType() runtime.Object {
	return &halkyon.Link{}
}

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
		helper := framework.GetHelperFor(in.PrimaryResourceType())
		fetch, err := in.FetchUpdatedDependent(&v1beta1.Component{}, helper)
		if err != nil {
			return fmt.Errorf("cannot retrieve associated component")
		}
		c := fetch.(*v1beta1.Component)
		compStatus := &c.Status
		compStatus.Phase = v1beta1.ComponentLinking
		compStatus.SetLinkingStatus(in.Name, v1beta1.Started, compStatus.PodName)
		compStatus.PodName = ""
		compStatus.Message = fmt.Sprintf("Initiating link %s", in.Name)
		err = helper.Client.Status().Update(context.TODO(), c)
		if err != nil {
			return err
		}
	}

	return nil
}

func (in *Link) GetDependentResourcesTypes() []framework.DependentResource {
	if len(in.dependentTypes) == 0 {
		in.dependentTypes = []framework.DependentResource{newComponent()}
	}
	return in.dependentTypes
}

func (in *Link) FetchAndInit(name, namespace string) (framework.Resource, error) {
	return in.HasDependents.FetchAndInitNewResource(name, namespace, in)
}

func (in *Link) ComputeStatus(err error, helper *framework.K8SHelper) (needsUpdate bool) {
	statuses, update := in.HasDependents.ComputeStatus(in, err, helper)
	return in.SetSuccessStatus(statuses, "Ready") || update
}

func (in *Link) Init() bool {
	return false
}

func (in *Link) GetAPIObject() runtime.Object {
	return in.Link
}

func NewLink(link ...*halkyon.Link) *Link {
	l := &halkyon.Link{}
	if link != nil {
		l = link[0]
	}
	return &Link{
		Link:          l,
		Requeueable:   new(framework.Requeueable),
		HasDependents: framework.NewHasDependents(),
	}
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
	errMsg := err.Error()
	if halkyon.LinkFailed != in.Status.Phase || errMsg != in.Status.Message {
		in.Status.Phase = halkyon.LinkFailed
		in.Status.Message = errMsg
		in.SetNeedsRequeue(false)
		return true
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
	helper := framework.GetHelperFor(link)
	deployment := &appsv1.Deployment{}
	name := link.Spec.ComponentName
	if err := helper.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: link.Namespace}, deployment); err == nil {
		return deployment, nil
	} else if err := helper.Client.Get(context.TODO(), types.NamespacedName{Name: name + "-build", Namespace: link.Namespace}, deployment); err == nil {
		return deployment, nil
	} else {
		helper.ReqLogger.Info("Deployment doesn't exist", "Name", name)
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

func update(d *appsv1.Deployment) error {
	helper := framework.GetHelperFor(&halkyon.Link{})
	err := helper.Client.Update(context.TODO(), d)
	if err != nil {
		return err
	}

	helper.ReqLogger.Info("Deployment updated.")
	return nil
}

func updateDeploymentWithLink(d *appsv1.Deployment, link *Link) error {
	// Update the Deployment of the component
	helper := framework.GetHelperFor(&halkyon.Link{})
	if err := update(d); err != nil {
		helper.ReqLogger.Info("Failed to update deployment.")
		return err
	}

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

package link

import (
	"context"
	"fmt"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	controller2 "github.com/snowdrop/component-operator/pkg/controller"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func NewLinkReconciler(mgr manager.Manager) *ReconcileLink {
	baseReconciler := controller2.NewBaseGenericReconciler(&v1alpha2.Link{}, mgr)
	r := &ReconcileLink{BaseGenericReconciler: baseReconciler}
	baseReconciler.SetReconcilerFactory(r)
	return r
}

type ReconcileLink struct {
	*controller2.BaseGenericReconciler
}

func (ReconcileLink) asLink(object runtime.Object) *v1alpha2.Link {
	return object.(*v1alpha2.Link)
}

func (r *ReconcileLink) IsDependentResourceReady(resource v1alpha2.Resource) (depOrTypeName string, ready bool) {
	link := r.asLink(resource)
	component := &v1alpha2.Component{}
	component.Name = link.Spec.ComponentName
	component.Namespace = link.Namespace
	_, err := r.Fetch(component)
	if err != nil || (v1alpha2.ComponentReady != component.Status.Phase && v1alpha2.ComponentRunning != component.Status.Phase) {
		return "component", false
	}
	return component.Name, true
}

func (r *ReconcileLink) CreateOrUpdate(object v1alpha2.Resource) (changed bool, err error) {
	link := r.asLink(object)
	if link.Status.Phase != v1alpha2.LinkReady {
		found, err := r.fetchDeployment(link)
		if err != nil {
			link.SetNeedsRequeue(true)
			return false, err
		}
		// Enrich the Deployment object using the information passed within the Link Spec (e.g Env Vars, EnvFrom, ...)
		if containers, isModified := r.updateContainersWithLinkInfo(link, found.Spec.Template.Spec.Containers); isModified {
			found.Spec.Template.Spec.Containers = containers
			if err = r.updateDeploymentWithLink(found, link); err != nil {
				// As it could be possible that we can't update the Deployment as it has been modified by another
				// process, then we will requeue
				link.SetNeedsRequeue(true)
			}
			return isModified, err
		}
	}
	return false, nil
}

func (r *ReconcileLink) updateContainersWithLinkInfo(link *v1alpha2.Link, containers []v1.Container) ([]v1.Container, bool) {
	var isModified = false
	kind := link.Spec.Kind
	switch kind {
	case v1alpha2.SecretLinkKind:
		secretName := link.Spec.Ref

		// Check if EnvFrom already exists
		// If this is the case, exit without error
		for i := 0; i < len(containers); i++ {
			var isEnvFromExist = false
			for _, env := range containers[i].EnvFrom {
				if env.String() == secretName {
					// EnvFrom already exists for the Secret Ref
					isEnvFromExist = true
				}
			}
			if !isEnvFromExist {
				// Add the Secret as EnvVar to the container
				containers[i].EnvFrom = append(containers[i].EnvFrom, r.addSecretAsEnvFromSource(secretName))
				isModified = true
			}
		}

	case v1alpha2.EnvLinkKind:
		// Check if Env already exists
		// If this is the case, exit without error
		for i := 0; i < len(containers); i++ {
			var isEnvExist = false
			for _, specEnv := range link.Spec.Envs {
				for _, env := range containers[i].Env {
					if specEnv.Name == env.Name && specEnv.Value == env.Value {
						// EnvFrom already exists for the Secret Ref
						isEnvExist = true
					}
				}
				if !isEnvExist {
					// Add the Secret as EnvVar to the container
					containers[i].Env = append(containers[i].Env, r.addKeyValueAsEnvVar(specEnv.Name, specEnv.Value))
					isModified = true
				}
			}
		}
	}

	return containers, isModified
}

func (r *ReconcileLink) update(d *appsv1.Deployment) error {
	err := r.Client.Update(context.TODO(), d)
	if err != nil {
		return err
	}

	r.ReqLogger.Info("Deployment updated.")
	return nil
}

//fetchDeployment returns the deployment resource created for this instance
func (r *ReconcileLink) fetchDeployment(link *v1alpha2.Link) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	name := link.Spec.ComponentName
	if err := r.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: link.Namespace}, deployment); err != nil {
		r.ReqLogger.Info("Deployment doesn't exist", "Name", name)
		return deployment, err
	} else {
		return deployment, nil
	}
}

func (r *ReconcileLink) updateDeploymentWithLink(d *appsv1.Deployment, link *v1alpha2.Link) error {
	// Update the Deployment of the component
	if err := r.update(d); err != nil {
		r.ReqLogger.Info("Failed to update deployment.")
		return err
	}

	name := link.Spec.ComponentName
	link.SetSuccessStatus(name, fmt.Sprintf("linked to '%s' component", name))
	return nil
}

package component

import (
	"context"
	"fmt"
	"halkyon.io/api/component/v1beta1"
	"halkyon.io/operator-framework"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type pod struct {
	base
}

var _ framework.DependentResource = &pod{}

func newPod(owner *v1beta1.Component, ownerStatusField string) pod {
	config := framework.NewConfig(corev1.SchemeGroupVersion.WithKind("Pod"))
	config.CheckedForReadiness = v1beta1.DevDeploymentMode == owner.Spec.DeploymentMode
	config.OwnerStatusField = ownerStatusField
	config.CreatedOrUpdated = false
	return pod{base: newConfiguredBaseDependent(owner, config)}
}

func (res pod) Build(empty bool) (runtime.Object, error) {
	if empty {
		return &corev1.Pod{}, nil
	}
	// we don't want to be building anything: the pod is under the deployment's control
	return nil, nil
}

func (res pod) NameFrom(underlying runtime.Object) string {
	return underlying.(*corev1.Pod).Name
}

func (res pod) IsReady(underlying runtime.Object) (bool, string) {
	p := underlying.(*corev1.Pod)
	for _, c := range p.Status.Conditions {
		if c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue {
			return true, ""
		}
	}
	msg := ""
	if len(p.Status.Message) > 0 {
		msg = ": " + p.Status.Message
	}
	return false, fmt.Sprintf("%s is not ready%s", p.Name, msg)
}

func (res pod) Fetch() (runtime.Object, error) {
	pods := &corev1.PodList{}
	lo := &client.ListOptions{}
	component := res.ownerAsComponent()
	lo.InNamespace(component.Namespace)
	lo.MatchingLabels(map[string]string{"app": component.Name})
	if err := framework.Helper.Client.List(context.TODO(), lo, pods); err != nil {
		return nil, err
	} else {
		// We assume that there is only one Pod containing the label app=component name AND we return it
		if len(pods.Items) > 0 {
			return &pods.Items[0], nil
		} else {
			err := fmt.Errorf("failed to get pod created for the component")
			return nil, err
		}
	}
}

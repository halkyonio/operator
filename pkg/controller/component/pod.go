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

func newPod(owner framework.Resource) pod {
	dependent := newBaseDependent(&corev1.Pod{}, owner)
	i := pod{base: dependent}
	dependent.SetDelegate(i)
	return i
}

func (res pod) Build() (runtime.Object, error) {
	// we don't want to be building anything: the pod is under the deployment's control
	return nil, nil
}

func (res pod) NameFrom(underlying runtime.Object) string {
	return underlying.(*corev1.Pod).Name
}

func (res pod) CanBeCreatedOrUpdated() bool {
	return false
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

func (res pod) ShouldBeCheckedForReadiness() bool {
	return v1beta1.DevDeploymentMode == res.ownerAsComponent().Spec.DeploymentMode
}

func (res pod) OwnerStatusField() string {
	return res.ownerAsComponent().DependentStatusFieldName()
}

func (res pod) Fetch(helper *framework.K8SHelper) (runtime.Object, error) {
	pods := &corev1.PodList{}
	lo := &client.ListOptions{}
	component := res.ownerAsComponent()
	lo.InNamespace(component.Namespace)
	lo.MatchingLabels(map[string]string{"app": component.Name})
	if err := helper.Client.List(context.TODO(), lo, pods); err != nil {
		helper.ReqLogger.Info("Pod(s) don't exist")
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

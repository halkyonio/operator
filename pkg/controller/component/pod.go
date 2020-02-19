package component

import (
	"context"
	"fmt"
	"halkyon.io/api/component/v1beta1"
	beta1 "halkyon.io/api/v1beta1"
	"halkyon.io/operator-framework"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type pod struct {
	base
}

var _ framework.DependentResource = &pod{}
var podGVK = corev1.SchemeGroupVersion.WithKind("Pod")

func newPod(owner *v1beta1.Component) pod {
	config := framework.NewConfig(podGVK)
	config.CheckedForReadiness = v1beta1.DevDeploymentMode == owner.Spec.DeploymentMode
	config.Created = false
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

func (res pod) GetCondition(underlying runtime.Object, err error) *beta1.DependentCondition {
	return framework.DefaultCustomizedGetConditionFor(res, err, underlying, func(underlying runtime.Object, cond *beta1.DependentCondition) {
		p := underlying.(*corev1.Pod)
		msg := ""
		ready := true
		for _, c := range p.Status.Conditions {
			if c.Type == corev1.PodReady {
				if c.Status != corev1.ConditionTrue {
					ready = false
					if "ContainersNotReady" == c.Reason {
						// extract list of not ready containers
						openBracket := strings.IndexRune(c.Message, '[')
						var notReadyContainers []string
						if openBracket > 1 {
							containerList := c.Message[openBracket+1 : strings.IndexRune(c.Message, ']')]
							notReadyContainers = strings.Split(containerList, ",")
						}
						msgArr := make([]string, 0, len(notReadyContainers))
						for _, c := range notReadyContainers {
							for _, status := range p.Status.ContainerStatuses {
								waiting := status.State.Waiting
								if status.Name == c && waiting != nil {
									msgArr = append(msgArr, fmt.Sprintf("%s: %s => %s", c, waiting.Reason, waiting.Message))
								}
							}
						}
						msg = strings.Join(msgArr, " & ")
					} else {
						msg = c.Message
					}
					msg = fmt.Sprintf("%s pod is not ready: %s => %s", p.Name, c.Reason, msg)
				} else {
					msg = fmt.Sprintf("%s is ready", p.Name)
					cond.SetAttribute("PodName", p.Name)
				}
			}
		}
		if len(p.Status.Message) > 0 {
			msg = p.Status.Message + ": " + msg
		}
		cond.Message = msg
		cond.Type = beta1.DependentPending
		if ready {
			cond.Type = beta1.DependentReady
		}

	})
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

package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	v1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileComponent) buildTaskRunS2iBuildahPush(c *v1alpha2.Component) (*v1alpha1.TaskRun, error) {
	taskRun := &v1alpha1.TaskRun{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tekton.dev/v1alpha1",
			Kind:       "TaskRun",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: c.Namespace,
			Name:      taskS2iBuildahPusName,
		},
		Spec: v1alpha1.TaskRunSpec{
			ServiceAccount: serviceAccountName,
			TaskRef: &v1alpha1.TaskRef{
				Name: taskS2iBuildahPusName,
			},
			Inputs: v1alpha1.TaskRunInputs{
				Params: []v1alpha1.Param{
					{Name: "baseImage", Value: "registry.access.redhat.com/redhat-openjdk-18/openjdk18-openshift"},
					{Name: "verifyTLS", Value: "false"},
					{Name: "contextFolder", Value: "/workspace"},
				},
				Resources: []v1alpha1.TaskResourceBinding{
					{
						Name: "workspace-git",
						ResourceSpec: &v1alpha1.PipelineResourceSpec{
							Type: "git",
							Params: []v1alpha1.Param{
								{
									Name: "revision",
									Value: "2.1.3-2",
								},
								{
									Name: "url",
									Value: "https://github.com/snowdrop/rest-http-example",
								},
							},
						},
					},
				},
			},
			Outputs: v1alpha1.TaskRunOutputs{
				Resources: []v1alpha1.TaskResourceBinding{
					{
						Name: "image",
						ResourceSpec: &v1alpha1.PipelineResourceSpec{
							Type: "image",
							Params: []v1alpha1.Param{
								{
									Name: "url",
									// OCP, OKD
									Value: "docker-registry.default.svc.cluster.local:5000/demo/spring-boot-example",
									// Kubernetes
									// Value: "kube-registry.kube-system.svc:5000/demo/spring-boot-example",
								},
							},
						},
					},
				},
			},
		},
	}

	// Set Component instance as the owner and controller
	controllerutil.SetControllerReference(c, taskRun, r.scheme)
	return taskRun, nil
}

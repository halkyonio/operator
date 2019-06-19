package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	v1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileComponent) buildTaskRunS2iBuildahPush(res dependentResource, c *v1alpha2.Component) (runtime.Object, error) {
	ls := r.getBuildLabels(c.Name)
	taskRun := &v1alpha1.TaskRun{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tekton.dev/v1alpha1",
			Kind:       "TaskRun",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: c.Namespace,
			Name:      res.name(c),
			Labels:    ls,
		},
		Spec: v1alpha1.TaskRunSpec{
			ServiceAccount: serviceAccountName,
			TaskRef: &v1alpha1.TaskRef{
				Name: taskS2iBuildahPushName,
			},
			Inputs: v1alpha1.TaskRunInputs{
				Params: []v1alpha1.Param{
					// See description of the parameters within the Tasks
					// We only override parameters here. Defaults are defined within the Tasks
					//{Name: "baseImage", Value: "registry.access.redhat.com/redhat-openjdk-18/openjdk18-openshift"},
					{Name: "moduleDirName", Value: c.Spec.BuildConfig.ModuleDirName},
				},
				Resources: []v1alpha1.TaskResourceBinding{
					{
						Name: "git",
						ResourceSpec: &v1alpha1.PipelineResourceSpec{
							Type: "git",
							Params: []v1alpha1.Param{
								{
									Name:  "revision",
									Value: c.Spec.BuildConfig.Ref,
								},
								{
									Name:  "url",
									Value: c.Spec.BuildConfig.URL,
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
									Value: r.dockerImageURL(c),
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

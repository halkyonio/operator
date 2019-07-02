package component

import (
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type taskRun struct {
	base
	reconciler *ReconcileComponent // todo: remove
}

func newTaskRun(reconciler *ReconcileComponent) taskRun {
	return taskRun{
		base:       newBaseDependent(&v1alpha1.TaskRun{}),
		reconciler: reconciler,
	}
}

func (res taskRun) Build() (runtime.Object, error) {
	c := res.ownerAsComponent()
	ls := getBuildLabels(c.Name)
	taskRun := &v1alpha1.TaskRun{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tekton.dev/v1alpha1",
			Kind:       "TaskRun",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: c.Namespace,
			Name:      res.Name(),
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
									Value: res.reconciler.dockerImageURL(c),
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

	return taskRun, nil
}

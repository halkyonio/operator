package component

import (
	"fmt"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"halkyon.io/api/component/v1beta1"
	"halkyon.io/operator-framework"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/pkg/apis"
)

type taskRun struct {
	base
}

var _ framework.DependentResource = &taskRun{}

func newTaskRun(owner *v1beta1.Component, ownerStatusField string) taskRun {
	config := framework.NewConfig(v1alpha1.SchemeGroupVersion.WithKind("TaskRun"), owner.GetNamespace())
	config.CheckedForReadiness = v1beta1.BuildDeploymentMode == owner.Spec.DeploymentMode
	config.CreatedOrUpdated = config.CheckedForReadiness
	config.OwnerStatusField = ownerStatusField
	return taskRun{base: newConfiguredBaseDependent(owner, config)}
}

func (res taskRun) Build(empty bool) (runtime.Object, error) {
	taskRun := &v1alpha1.TaskRun{}
	if !empty {
		c := res.ownerAsComponent()
		ls := getBuildLabels(c.Name)
		taskRun.ObjectMeta = metav1.ObjectMeta{
			Namespace: c.Namespace,
			Name:      res.Name(),
			Labels:    ls,
		}
		taskRun.Spec = v1alpha1.TaskRunSpec{
			ServiceAccountName: ServiceAccountName(c),
			TaskRef: &v1alpha1.TaskRef{
				Name: TaskName(c),
			},
			Inputs: v1alpha1.TaskRunInputs{
				Params: []v1alpha1.Param{
					// See description of the parameters within the Tasks
					// We only override parameters here. Defaults are defined within the Tasks
					{Name: "baseImage", Value: v1alpha1.ArrayOrString{Type: v1alpha1.ParamTypeString, StringVal: baseImage(c)}},
					{Name: "moduleDirName", Value: v1alpha1.ArrayOrString{Type: v1alpha1.ParamTypeString, StringVal: moduleDirName(c)}},
					{Name: "contextPath", Value: v1alpha1.ArrayOrString{Type: v1alpha1.ParamTypeString, StringVal: contextPath(c)}},
				},
				Resources: []v1alpha1.TaskResourceBinding{{
					PipelineResourceBinding: v1alpha1.PipelineResourceBinding{
						Name: "git",
						ResourceSpec: &v1alpha1.PipelineResourceSpec{
							Type: "git",
							Params: []v1alpha1.ResourceParam{
								{
									Name:  "revision",
									Value: gitRevision(c),
								},
								{
									Name:  "url",
									Value: c.Spec.BuildConfig.URL},
							},
						},
					},
				}},
			},
			Outputs: v1alpha1.TaskRunOutputs{
				Resources: []v1alpha1.TaskResourceBinding{{
					PipelineResourceBinding: v1alpha1.PipelineResourceBinding{
						Name: "image",
						ResourceSpec: &v1alpha1.PipelineResourceSpec{
							Type: "image",
							Params: []v1alpha1.ResourceParam{
								{
									Name: "url",
									// Calculate the path of the image using docker registry URL for ocp or k8s cluster
									Value: dockerImageURL(c),
								},
							},
						},
					},
				}},
			},
		}
	}

	return taskRun, nil
}

func (res taskRun) IsReady(underlying runtime.Object) (ready bool, message string) {
	tr := underlying.(*v1alpha1.TaskRun)
	succeeded := tr.Status.GetCondition(apis.ConditionSucceeded)
	if succeeded != nil {
		if succeeded.IsTrue() {
			return true, succeeded.Message
		} else {
			return false, fmt.Sprintf("%s didn't succeed: %s", tr.Name, succeeded.Message)
		}
	} else {
		return false, fmt.Sprintf("%s is not ready", tr.Name)
	}
}

func (res taskRun) NameFrom(underlying runtime.Object) string {
	return underlying.(*v1alpha1.TaskRun).Name
}

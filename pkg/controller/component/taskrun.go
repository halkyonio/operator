package component

import (
	"fmt"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"halkyon.io/api/component/v1beta1"
	beta1 "halkyon.io/api/v1beta1"
	"halkyon.io/operator-framework"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/pkg/apis"
)

type taskRun struct {
	base
}

var _ framework.DependentResource = &taskRun{}

func newTaskRun(owner *v1beta1.Component) taskRun {
	config := framework.NewConfig(v1alpha1.SchemeGroupVersion.WithKind("TaskRun"))
	config.CheckedForReadiness = v1beta1.BuildDeploymentMode == owner.Spec.DeploymentMode
	config.Created = config.CheckedForReadiness
	config.Updated = config.CheckedForReadiness
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

func (res taskRun) GetCondition(underlying runtime.Object, err error) *beta1.DependentCondition {
	return framework.DefaultCustomizedGetConditionFor(res, err, underlying, func(underlying runtime.Object, cond *beta1.DependentCondition) {
		tr := underlying.(*v1alpha1.TaskRun)
		succeeded := tr.Status.GetCondition(apis.ConditionSucceeded)
		if succeeded != nil {
			cond.Message = succeeded.Message
			cond.Reason = succeeded.Reason
			if succeeded.IsTrue() {
				cond.Type = beta1.DependentReady
				return
			}
			if succeeded.IsFalse() {
				cond.Type = beta1.DependentFailed
				return
			}
		}
		cond.Type = beta1.DependentPending
		cond.Message = fmt.Sprintf("%s is not ready", tr.Name)
	})
}

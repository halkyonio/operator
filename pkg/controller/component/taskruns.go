package component

import (
	"github.com/knative/pkg/apis"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"halkyon.io/api/component/v1beta1"
	"halkyon.io/operator/pkg/controller"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type taskRun struct {
	base
	reconciler *ReconcileComponent // todo: remove
}

func (res taskRun) NewInstanceWith(owner controller.Resource) controller.DependentResource {
	return newOwnedTaskRun(res.reconciler, owner)
}

func newTaskRun(reconciler *ReconcileComponent) taskRun {
	return newOwnedTaskRun(reconciler, nil)
}

func newOwnedTaskRun(reconciler *ReconcileComponent, owner controller.Resource) taskRun {
	dependent := newBaseDependent(&v1alpha1.TaskRun{}, owner)
	t := taskRun{
		base:       dependent,
		reconciler: reconciler,
	}
	dependent.SetDelegate(t)
	return t
}

func (res taskRun) Build() (runtime.Object, error) {
	c := res.ownerAsComponent()
	ls := getBuildLabels(c.Name)
	taskRun := &v1alpha1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: c.Namespace,
			Name:      res.Name(),
			Labels:    ls,
		},
		Spec: v1alpha1.TaskRunSpec{
			ServiceAccount: controller.ServiceAccountName(c),
			TaskRef: &v1alpha1.TaskRef{
				Name: controller.TaskName(c),
			},
			Inputs: v1alpha1.TaskRunInputs{
				Params: []v1alpha1.Param{
					// See description of the parameters within the Tasks
					// We only override parameters here. Defaults are defined within the Tasks
					{Name: "baseImage", Value: res.reconciler.baseImage(c)},
					{Name: "moduleDirName", Value: res.reconciler.moduleDirName(c)},
					{Name: "contextPath", Value: res.reconciler.contextPath(c)},
				},
				Resources: []v1alpha1.TaskResourceBinding{
					{
						Name: "git",
						ResourceSpec: &v1alpha1.PipelineResourceSpec{
							Type: "git",
							Params: []v1alpha1.Param{
								{
									Name:  "revision",
									Value: res.reconciler.gitRevision(c),
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
									// Calculate the path of the image using docker registry URL for ocp or k8s cluster
									Value: res.reconciler.dockerImageURL(c),
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

func (res taskRun) OwnerStatusField() string {
	return res.ownerAsComponent().DependentStatusFieldName()
}

func (res taskRun) ShouldBeCheckedForReadiness() bool {
	return v1beta1.BuildDeploymentMode == res.ownerAsComponent().Spec.DeploymentMode
}

func (res taskRun) IsReady(underlying runtime.Object) bool {
	tr := underlying.(*v1alpha1.TaskRun)
	for _, c := range tr.Status.Conditions {
		if c.Type == apis.ConditionSucceeded && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func (res taskRun) NameFrom(underlying runtime.Object) string {
	return underlying.(*v1alpha1.TaskRun).Name
}

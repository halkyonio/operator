package component

import (
	"fmt"
	component "halkyon.io/api/component/v1beta1"
	"halkyon.io/operator/pkg/controller"
	"halkyon.io/operator/pkg/controller/framework"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type deployment struct {
	base
	reconciler *ComponentManager // todo: remove
}

func (res deployment) NewInstanceWith(owner framework.Resource) framework.DependentResource {
	return newOwnedDeployment(res.reconciler, owner)
}

func newDeployment(reconciler *ComponentManager) deployment {
	return newOwnedDeployment(reconciler, nil)
}

func newOwnedDeployment(reconciler *ComponentManager, owner framework.Resource) deployment {
	dependent := newBaseDependent(&appsv1.Deployment{}, owner)
	d := deployment{
		base:       dependent,
		reconciler: reconciler,
	}
	dependent.SetDelegate(d)
	return d
}

func (res deployment) Build() (runtime.Object, error) {
	c := res.ownerAsComponent()
	if component.BuildDeploymentMode == c.Spec.DeploymentMode {
		return res.installBuild()
	}
	return res.installDev()
}

func (res deployment) Name() string {
	return controller.DeploymentName(res.ownerAsComponent())
}

/*func (res deployment) Handler() handler.EventHandler {
	return handler.Funcs{
		CreateFunc: func(e event.CreateEvent, q workqueue.RateLimitingInterface) {
			q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
				Namespace: e.Meta.GetNamespace(),
				Name:      e.Meta.GetLabels()["app"],
			}})
		},
	}
}
*/
func (res deployment) Predicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(createEvent event.CreateEvent) bool {
			fmt.Println(fmt.Sprintf("create %v", createEvent))
			return true
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			fmt.Println(fmt.Sprintf("delete %v", deleteEvent))
			return true
		},
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			fmt.Println(fmt.Sprintf("update %v", updateEvent))
			return true
		},
		GenericFunc: func(genericEvent event.GenericEvent) bool {
			fmt.Println(fmt.Sprintf("generic %v", genericEvent))
			return true
		},
	}
}

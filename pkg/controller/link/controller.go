package link

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strconv"
)

const (
	controllerName = "link-controller"
)

var log = logf.Log.WithName("link.controller")

// New creates a new Component Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func New(mgr manager.Manager) error {
	return Add(mgr, NewReconciler(mgr))
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func Add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Link
	err = c.Watch(&source.Kind{Type: &v1alpha2.Link{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// newReconciler returns a new reconcile.Reconciler
func NewReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileLink{
		client:    mgr.GetClient(),
		config:    mgr.GetConfig(),
		scheme:    mgr.GetScheme(),
		reqLogger: log,
	}
}

type ReconcileLink struct {
	client    client.Client
	config    *rest.Config
	scheme    *runtime.Scheme
	reqLogger logr.Logger
}

//Update the factory object and requeue
func (r *ReconcileLink) update(d *appsv1.Deployment) error {
	err := r.client.Update(context.TODO(), d)
	if err != nil {
		return err
	}

	r.reqLogger.Info("Deployment updated.")
	return nil
}

func (r *ReconcileLink) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.reqLogger = log.WithValues("Namespace", request.Namespace)

	// Fetch the Link created, deleted or updated
	link := &v1alpha2.Link{}
	err := r.client.Get(context.TODO(), request.NamespacedName, link)
	if err != nil {
		return r.fetch(err)
	}

	r.reqLogger.Info("-----------------------")
	r.reqLogger.Info("Reconciling Link")
	r.reqLogger.Info("Status of the Link", "Status phase", link.Status.Phase)
	r.reqLogger.Info("Creation time          ", "Creation time", link.ObjectMeta.CreationTimestamp)
	r.reqLogger.Info("Resource version       ", "Resource version", link.ObjectMeta.ResourceVersion)
	r.reqLogger.Info("Generation version     ", "Generation version", strconv.FormatInt(link.ObjectMeta.Generation, 10))
	// r.reqLogger.Info("Deletion time          ","Deletion time", Link.ObjectMeta.DeletionTimestamp)

	// Add the Status Link Creation when we process the first time the Link CR
	// as we will start to create/update resources
	if link.Generation == 1 && link.Status.Phase == "" {
		// Update Status to value "Linking" as we will try to update the Deployment
		if err := r.updateStatusInstance(v1alpha2.LinkPending, link, request); err != nil {
			r.reqLogger.Info("Status update failed !")
			return reconcile.Result{}, err
		}
	}

	// Process the Link if the status is not Linked
	if link.Status.Phase != v1alpha2.LinkReady {
		found, err := r.fetchDeployment(request.Namespace, link.Spec.ComponentName)
		if err != nil {
			r.reqLogger.Info("Component not found")
			// TODO Update status of the link to report the error
			return reconcile.Result{Requeue: true}, nil
		}

		// Enrich the Deployment object using the information passed within the Link Spec (e.g Env Vars, EnvFrom, ...)
		if containers, isModified := r.updateContainersWithLinkInfo(link, found.Spec.Template.Spec.Containers, request); isModified {
			found.Spec.Template.Spec.Containers = containers
			if err := r.updateDeploymentWithLink(found, link, request); err != nil {
				// As it could be possible that we can't update the Deployment as it has been modified by another
				// process, then we will requeue
				return reconcile.Result{Requeue: true}, err
			}
		}
	}

	r.reqLogger.Info(fmt.Sprintf("Reconciled : %s", link.Name))
	return reconcile.Result{}, nil
}

func (r *ReconcileLink) updateContainersWithLinkInfo(link *v1alpha2.Link, containers []v1.Container, request reconcile.Request) ([]v1.Container, bool) {
	var isModified = false
	kind := link.Spec.Kind
	switch kind {
	case v1alpha2.SecretLinkKind:
		secretName := link.Spec.Ref

		// Check if EnvFrom already exists
		// If this is the case, exit without error
		for i := 0; i < len(containers); i++ {
			var isEnvFromExist = false
			for _, env := range containers[i].EnvFrom {
				if env.String() == secretName {
					// EnvFrom already exists for the Secret Ref
					isEnvFromExist = true
				}
			}
			if !isEnvFromExist {
				// Add the Secret as EnvVar to the container
				containers[i].EnvFrom = append(containers[i].EnvFrom, r.addSecretAsEnvFromSource(secretName))
				isModified = true
			}
		}

	case v1alpha2.EnvLinkKind:
		// Check if Env already exists
		// If this is the case, exit without error
		for i := 0; i < len(containers); i++ {
			var isEnvExist = false
			for _, specEnv := range link.Spec.Envs {
				for _, env := range containers[i].Env {
					if specEnv.Name == env.Name && specEnv.Value == env.Value {
						// EnvFrom already exists for the Secret Ref
						isEnvExist = true
					}
				}
				if !isEnvExist {
					// Add the Secret as EnvVar to the container
					containers[i].Env = append(containers[i].Env, r.addKeyValueAsEnvVar(specEnv.Name, specEnv.Value))
					isModified = true
				}
			}
		}
	}

	return containers, isModified
}

func (r *ReconcileLink) updateDeploymentWithLink(d *appsv1.Deployment, link *v1alpha2.Link, request reconcile.Request) error {
	// Update the Deployment of the component
	if err := r.update(d); err != nil {
		r.reqLogger.Info( "Failed to update deployment.")
		return err
	}

	// Update Status to value "Linked"
	if err := r.updateStatusInstance(v1alpha2.LinkReady, link, request); err != nil {
		r.reqLogger.Info("Failed to update link Status !")
		return err
	}

	r.reqLogger.Info("Added link to the component", "Name", link.Spec.ComponentName)
	return nil
}

package link

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"github.com/snowdrop/component-operator/pkg/util"
	"github.com/snowdrop/component-operator/pkg/util/kubernetes"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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
func (r *ReconcileLink) update(obj runtime.Object) (reconcile.Result, error) {
	err := r.client.Update(context.TODO(), obj)
	if err != nil {
		r.reqLogger.Error(err, "Failed to update spec")
		return reconcile.Result{}, err
	}
	r.reqLogger.Info("Spec updated - return and create")
	return reconcile.Result{Requeue: true}, nil
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
	r.reqLogger.Info("Reconciling Link  ")
	r.reqLogger.Info("Status of the Link", "Status phase", link.Status.Phase)
	r.reqLogger.Info("Creation time          ", "Creation time", link.ObjectMeta.CreationTimestamp)
	r.reqLogger.Info("Resource version       ", "Resource version", link.ObjectMeta.ResourceVersion)
	r.reqLogger.Info("Generation version     ", "Generation version", strconv.FormatInt(link.ObjectMeta.Generation, 10))
	// r.reqLogger.Info("Deletion time          ","Deletion time", Link.ObjectMeta.DeletionTimestamp)

	isOpenShift, err := util.IsOpenshift(r.config)
	if err != nil {
		return reconcile.Result{}, err
	}

	if (isOpenShift) {
		// Search about the DeploymentConfig to be updated using the "Component Name"
		found, err := r.fetchDeploymentConfig(request.Namespace, link.Spec.ComponentName)
		if err != nil {
			r.reqLogger.Info("Component not found")
			// TODO Update status of the link to report the error
			return reconcile.Result{}, nil
		}

		// Enrich the DeploymentConfig of the Component using the information passed within the Link Spec
		var logMessage = ""
		kind := link.Spec.Kind
		switch kind {
		case "Secret":
			secretName := link.Spec.Ref
			// Add the Secret as EnvVar to the container
			found.Spec.Template.Spec.Containers[0].EnvFrom = r.addSecretAsEnvFromSource(secretName)
			logMessage = "Added the deploymentConfig's EnvFrom reference of the secret " + secretName
		case "Env":
			// TODO Iterate through Env vars
			key := link.Spec.Envs[0].Name
			val := link.Spec.Envs[0].Value
			found.Spec.Template.Spec.Containers[0].Env = append(found.Spec.Template.Spec.Containers[0].Env, r.addKeyValueAsEnvVar(key, val))
			logMessage = "Added the deploymentConfig's EnvVar : " + key + ", " + val
		}

		// Update the DeploymentConfig of the component
		r.update(found)
		r.reqLogger.Info(logMessage)

		// Rollout the DC
		err = r.rolloutDeploymentConfig(link.Spec.ComponentName, link.Namespace)
		if err != nil {
			r.reqLogger.Info("Deployment Config rollout failed !")
			return reconcile.Result{}, err
		}

		r.reqLogger.Info("### Added %s link's CRD component", link.Spec.ComponentName)
		r.reqLogger.Info("### Rollout the DeploymentConfig of the '%s' component", link.Spec.ComponentName)
	} else {
        /*
        // K8s platform. We will fetch a deployment
		d, err := kubernetes.GetDeployment(namespace, componentName, c)
		if err != nil {
			return false, err
		}
		logMessage := ""
		kind := link.Spec.Kind
		switch kind {
		case "Secret":
			secretName := link.Spec.Ref
			// Add the Secret as EnvVar to the container
			d.Spec.Template.Spec.Containers[0].EnvFrom = addSecretAsEnvFromSource(secretName)
			logMessage = "### Added the deploymentConfig's EnvFrom reference of the secret " + secretName
		case "Env":
			// TODO Iterate through Env vars
			key := link.Spec.Envs[0].Name
			val := link.Spec.Envs[0].Value
			d.Spec.Template.Spec.Containers[0].Env = append(d.Spec.Template.Spec.Containers[0].Env, addKeyValueAsEnvVar(key, val))
			logMessage = "### Added the deploymentConfig's EnvVar : " + key + ", " + val
		}

		// Update the Deployment
		err = c.Update(context.TODO(), d)
		if err != nil && k8serrors.IsConflict(err) {
			// Retry function on conflict
			return false, nil
		}
		if err != nil {
			return false, err
		}
		log.Info(logMessage)

		r.reqLogger.Info("### Added link", "Target Component",link.Spec.TargetComponentName)
		r.reqLogger.Info("### Rollout Deployment of the '%s' component", component.Name)
		return true, nil
        */
	}

	r.reqLogger.Info(fmt.Sprintf("Reconciled : %s", link.Name))
	return reconcile.Result{}, nil
}

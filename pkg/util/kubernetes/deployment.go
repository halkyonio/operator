package kubernetes

import (
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
	"time"

	appv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

// WaitForDeployment checks to see if a given deployment has a certain number of available replicas after a specified amount of time
// If the deployment does not have the required number of replicas after 5 * retries seconds, the function returns an error
// This can be used in multiple ways, like verifying that a required resource is ready before trying to use it, or to test
// failure handling, like simulated in SimulatePodFail.
func WaitForDeployment(t *testing.T, kubeclient kubernetes.Interface, namespace, name string, replicas int, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		deployment, err := kubeclient.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{IncludeUninitialized: true})
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of %s deployment\n", name)
				return false, nil
			}
			return false, err
		}

		if int(deployment.Status.AvailableReplicas) == replicas {
			return true, nil
		}
		t.Logf("Waiting for full availability of %s deployment (%d/%d)\n", name, deployment.Status.AvailableReplicas, replicas)
		return false, nil
	})
	if err != nil {
		return err
	}
	t.Logf("Deployment available (%d/%d)\n", replicas, replicas)
	return nil
}

func GetDeployment(namespace string, name string, c client.Client) (*appv1.Deployment, error) {
	d := &appv1.Deployment{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, d); err != nil {
		return nil, err
	}
	return d, nil
}


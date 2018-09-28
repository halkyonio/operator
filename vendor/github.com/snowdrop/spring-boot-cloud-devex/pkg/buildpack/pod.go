package buildpack

import (
	"encoding/json"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"time"

	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/snowdrop/spring-boot-cloud-devex/pkg/buildpack/types"
)

// WaitAndGetPod block and waits until pod matching selector is in in Running state
func WaitAndGetPod(c *kubernetes.Clientset, application types.Application) (*corev1.Pod, error) {

	selector := podSelector(application)
	log.Debugf("Waiting for %s pod", selector)

	w, err := c.CoreV1().Pods(application.Namespace).Watch(selector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to watch pod")
	}
	defer w.Stop()

	const timeoutInSeconds = 30
	duration := timeoutInSeconds * time.Second
	select {
	case val := <-w.ResultChan():
		log.Debugf("Received event of type %s", val.Type)
		if pod, ok := val.Object.(*corev1.Pod); ok {
			return pod, nil
		} else {
			return nil, errors.Errorf("Unable to convert event object to Pod")
		}
	case <-time.After(duration):
		bytes, e := json.Marshal(selector)
		if e != nil {
			return nil, errors.Errorf("Couldn't marshall pod selector to JSON: %s", e)
		}
		return nil, errors.Errorf("Waited %s but couldn't find pod matching '%s' selector", duration, string(bytes))
	}

	bytes, e := json.Marshal(selector)
	if e != nil {
		return nil, errors.Errorf("Couldn't marshall pod selector to JSON in unknown error code-path. JSON error is: %s", e)
	}
	return nil, errors.Errorf("Unknown error while waiting for pod matching '%s' selector", string(bytes))
}

func podSelector(application types.Application) metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: "app=" + application.Name,
		FieldSelector: "status.phase=Running",
	}
}

package kubernetes

import (
	log "github.com/sirupsen/logrus"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	util "github.com/snowdrop/component-operator/pkg/util/template"
	"golang.org/x/net/context"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
	"text/template"
)

func CreateResource(tmpl template.Template, component *v1alpha1.Component, c client.Client, scheme *runtime.Scheme) error {
	res, err := newResourceFromTemplate(tmpl, component, scheme)
	if err != nil {
		return err
	}

	for _, r := range res {
		if obj, ok := r.(metav1.Object); ok {
			obj.SetLabels(PopulateK8sLabels(component, "Backend"))
		}
		err = c.Create(context.TODO(), r)
		log.Infof("##### Error returned : #####",err)
		if err != nil {
			return err
		}
	}

	return nil
}

func newResourceFromTemplate(tmpl template.Template, component *v1alpha1.Component, scheme *runtime.Scheme) ([]runtime.Object, error) {
	var result = []runtime.Object{}

	var b = util.Parse(tmpl, component)
	r, err := PopulateKubernetesObjectFromYaml(b.String())
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(r.GetKind(), "List") {
		l, err := r.ToList()
		if err != nil {
			return nil, err
		}
		for _, item := range l.Items {
			obj, err := RuntimeObjectFromUnstructured(&item)
			if err != nil {
				return nil, err
			}
			ro, ok := obj.(v1.Object)
			ro.SetNamespace(component.Namespace)
			if !ok {
				return nil, err
			}
			controllerutil.SetControllerReference(component, ro, scheme)
			result = append(result, obj)
		}
	} else {
		obj, err := RuntimeObjectFromUnstructured(r)
		if err != nil {
			return nil, err
		}

		ro, ok := obj.(v1.Object)
		ro.SetNamespace(component.Namespace)
		if !ok {
			return nil, err
		}
		controllerutil.SetControllerReference(component, ro, scheme)
		result = append(result, obj)
	}
	return result, nil
}


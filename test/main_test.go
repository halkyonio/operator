// +build !test

package test

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/types"

	"github.com/halkyonio/operator/pkg/apis/halkyon/v1beta1"
	f "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestMain(m *testing.M) {
	f.MainEntry(m)
}

func TestTypeMetaComponent(t *testing.T) {
	component := &v1beta1.Component{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Component",
			APIVersion: "halkyon.io/v1beta1",
		},
	}
	err := f.AddToFrameworkScheme(v1beta1.AddToScheme, component)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
}

func componentTest(t *testing.T, f *f.Framework, ctx *f.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}

	// Create Component CRD
	component := &v1beta1.Component{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Component",
			APIVersion: "halkyon.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-component",
			Namespace: namespace,
		},
		Spec: v1beta1.ComponentSpec{
			Runtime: "Spring Boot",
		},
	}

	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(context.TODO(), component, nil)
	if err != nil {
		return err
	}

	err = f.Client.Get(context.TODO(), types.NamespacedName{Name: "example-component", Namespace: namespace}, component)
	if err != nil {
		return err
	}

	return nil
}

func TestComponentCRD(t *testing.T) {
	ctx := f.NewTestCtx(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&f.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}

	// get global framework variables
	f := f.Global

	// wait for halkyon-operator to be ready
	err = WaitForDeployment(t, f.KubeClient, namespace, "halkyon-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	// Let's create a Component
	if err = componentTest(t, f, ctx); err != nil {
		t.Fatal(err)
	}
}

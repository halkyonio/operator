package component

import (
	"testing"

	"github.com/onsi/gomega"
	compv1alpha "github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	"golang.org/x/net/context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var instance = compv1alpha.Component{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "foo",
		Namespace: "default"},
	Spec: compv1alpha.ComponentSpec{
		Runtime:        "spring-boot",
		DeploymentMode: "innerloop",
	},
}

func TestComponentInstance(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c = mgr.GetClient()

	recFn, _ := SetupTestReconcile(NewReconciler(mgr))
	g.Expect(create(mgr, recFn)).NotTo(gomega.HaveOccurred())

	stopMgr, mgrStopped := StartTestManager(mgr, g)

	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	// Create the Component object and expect the Reconcile and Deployment to be created
	err = c.Create(context.TODO(), &instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), &instance)
}

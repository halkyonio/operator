package e2e

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/errors"

	"testing"
)

func TestComponent(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Component Operator Test Suite")
}

var _ = Describe("ComponentE2E", func() {

	tmpDir, _ := ioutil.TempDir("", "component-operator")
	namespace := "my-spring-boot"
	compA := "my-spring-boot"
	fmt.Print("Temp dir : ", tmpDir, "\n")

	Describe("Component name", func() {
		It("should create a maven project", func() {
			pomfile := tmpDir + "/pom.xml"
			generateSpringBootJavaProject(tmpDir, "crud", "my-spring-boot")
			Expect(pomfile).To(BeAnExistingFile())
		})

		It("should build a maven project", func() {
			err := runMavenBuild(tmpDir)
			Expect(err).To(BeNil())
		})

		It("should install the component", func() {
			cli := crudClient()
			err := cli.Create(context.TODO(), springBootComponent(compA, namespace))
			if err != nil && !errors.IsAlreadyExists(err) {
				logrus.Errorf("failed to create component : %v", err)
			}
		})

		It("should have a component installed", func() {
			// checking component deployed
			componentName := runCmd("oc get cp " + compA + " -o go-template='{{.metadata.name}}'")
			Expect(componentName).To(ContainSubstring("component-a"))
		})

		It("should have deploymentconfig, service, imagestream", func() {
			// checking component deployed
			componentName := runCmd("oc get cp " + compA + " -o go-template='{{.metadata.name}}'")
			Expect(componentName).To(ContainSubstring("component-a"))
		})

	})
})

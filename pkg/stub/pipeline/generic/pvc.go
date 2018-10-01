package generic

import (
	"github.com/snowdrop/spring-boot-operator/pkg/apis/springboot/v1alpha1"
	"github.com/snowdrop/spring-boot-operator/pkg/stub/pipeline"
)

// NewPVCStep creates a step that handles the creation of the DeploymentConfig
func NewPVCStep() pipeline.Step {
	return &pvcStep{}
}

type pvcStep struct {
}

func (pvcStep) Name() string {
	return "pvc"
}

func (pvcStep) CanHandle(springboot *v1alpha1.SpringBoot) bool {
	return true
}

func (pvcStep) Handle(springboot *v1alpha1.SpringBoot) error {
	return nil
}

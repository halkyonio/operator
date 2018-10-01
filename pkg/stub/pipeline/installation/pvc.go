package installation

import "github.com/snowdrop/spring-boot-operator/pkg/apis/springboot/v1alpha1"

// NewPVCStep creates a step that handles the creation of the DeploymentConfig
func NewPVCStep() Step {
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

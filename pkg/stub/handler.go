package stub

import (
	"context"
	"github.com/snowdrop/spring-boot-operator/pkg/apis/springboot/v1alpha1"
	"github.com/snowdrop/spring-boot-operator/pkg/stub/pipeline/installation"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewHandler() sdk.Handler {
	return &Handler{
		installationSteps: []installation.Step{
			installation.NewDeployStep(),
			installation.NewServiceStep(),
		},
	}
}

type Handler struct {
	installationSteps []installation.Step
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.SpringBoot:
		for _, a := range h.installationSteps {
			if a.CanHandle(o) {
				logrus.Debug("Invoking action ", a.Name(), " on Spring Boot ", o.Name)
				if err := a.Handle(o); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

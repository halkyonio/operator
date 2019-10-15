package capability

import (
	"halkyon.io/operator/pkg/controller"
)

func (r *ReconcileCapability) installDB(c *controller.Capability) (e error) {
	return c.CreateOrUpdate(r.K8SHelper)
}

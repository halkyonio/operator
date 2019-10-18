package capability

func (r *CapabilityManager) installDB(c *Capability) (e error) {
	return c.CreateOrUpdate(r.Helper())
}

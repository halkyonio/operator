package component

import (
	halkyon "halkyon.io/api/component/v1beta1"
	framework "halkyon.io/operator-framework"
)

func PVCName(c *halkyon.Component) string {
	specified := c.Spec.Storage.Name
	if len(specified) > 0 {
		return specified
	}
	return "m2-data-" + c.Name // todo: use better default name?
}

func ServiceAccountName(owner framework.SerializableResource) string {
	return "build-bot"
}

func TaskName(owner framework.SerializableResource) string {
	return "s2i-buildah-push"
}

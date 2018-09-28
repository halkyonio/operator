package buildpack

import "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	pvcName = "m2-data"
)

var zero = int64(0)
var deleteOptions = &v1.DeleteOptions{GracePeriodSeconds: &zero}

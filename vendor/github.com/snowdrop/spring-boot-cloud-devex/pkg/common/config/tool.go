package config

import (
	"github.com/snowdrop/spring-boot-cloud-devex/pkg/buildpack/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	SupervisordBin = "/var/lib/supervisord/bin/supervisord"
	SupervisordCtl = "ctl"
	RunCmdName     = "run-java"
	CompileCmdName = "compile-java"
)

type Tool struct {
	Application types.Application
	KubeConfig  Kube
	RestConfig  *rest.Config
	Clientset   *kubernetes.Clientset
}

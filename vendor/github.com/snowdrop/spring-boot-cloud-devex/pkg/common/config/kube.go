package config

import (
	log "github.com/sirupsen/logrus"
	"os/user"
)

const (
	KUBECONFILE = "/.kube/config"
)

type Kube struct {
	MasterURL string
	Config    string
}

func NewKube() *Kube {
	return &Kube{}
}

func HomeKubePath() string {
	usr, err := user.Current()
	if err != nil {
		log.Debugf("Can't get current user:\n%v", err)
	}
	return usr.HomeDir + KUBECONFILE
}

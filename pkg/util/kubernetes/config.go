/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubernetes

import (
	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
	"os/user"
)

const (
	KUBECONFILE = "/.kube/config"
)

// InitKubeClient initialize the k8s client
func InitKubeClient(kubeconfig string) error {
	if kubeconfig == "" {
		kubeconfig = getK8Config("")
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}

	k8sclient.GetKubeConfig() = config
	return nil
}

func getK8Config(kubeconfig string) string {
	log.Info("Get K8s config file")
	if kubeconfig == "" {
		return homeKubePath()
	} else {
		return kubeconfig
	}
	log.Debug("Kubeconfig : ", kubeconfig)
}

func homeKubePath() string {
	usr, err := user.Current()
	if err != nil {
		log.Debugf("Can't get current user:\n%v", err)
	}
	return usr.HomeDir + KUBECONFILE
}

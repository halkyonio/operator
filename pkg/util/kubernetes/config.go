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
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
	restclient "k8s.io/client-go/rest"
	"os/user"
	"strings"
)

const (
	KUBECONFILE = ".kube/config"
)

type Kube struct {
	MasterURL string
	Config    string
}

// InitKubeClient initialize the k8s client
func InitKubeClient(kubeconfig string) error {
	if kubeconfig == "" {
		kubeconfig = getK8Config("")
	}

	// use the current context in kubeconfig
	_, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}

	return nil
}

func GetK8RestConfig() *restclient.Config {
   return createKubeRestconfig()
}

// Create Kube Rest's Config Client
func createKubeRestconfig() *restclient.Config {
	kube := Kube{
		Config: HomeKubePath(),
		MasterURL: "192.168.99.50:8443",
	}
	kubeRestClient, err := clientcmd.BuildConfigFromFlags(kube.MasterURL, kube.Config)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
	}
	return kubeRestClient
}

func getK8Config(kubeconfig string) string {
	log.Debug("Get K8s config file")
	if kubeconfig == "" {
		return HomeKubePath()
	} else {
		return kubeconfig
	}
	log.Debug("Kubeconfig : ", kubeconfig)
	return ""
}

func HomeKubePath() string {
	usr, err := user.Current()
	if err != nil {
		log.Debugf("Can't get current user:\n%v", err)
	}
	return strings.Join([]string{usr.HomeDir, KUBECONFILE}, "/")
}

// GetClientCurrentNamespace --
func GetClientCurrentNamespace(kubeconfig string) (string, error) {
	if kubeconfig == "" {
		kubeconfig = getK8Config(kubeconfig)
	}
	if kubeconfig == "" {
		return "default", nil
	}

	data, err := ioutil.ReadFile(kubeconfig)
	if err != nil {
		return "", err
	}
	config := clientcmdapi.NewConfig()
	if len(data) == 0 {
		return "", errors.New("kubernetes config file is empty")
	}

	decoded, _, err := clientcmdlatest.Codec.Decode(data, &schema.GroupVersionKind{Version: clientcmdlatest.Version, Kind: "Config"}, config)
	if err != nil {
		return "", err
	}

	clientcmdconfig := decoded.(*clientcmdapi.Config)

	cc := clientcmd.NewDefaultClientConfig(*clientcmdconfig, &clientcmd.ConfigOverrides{})
	ns, _, err := cc.Namespace()
	return ns, err
}

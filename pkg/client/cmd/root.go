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

package cmd

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/snowdrop/spring-boot-operator/pkg/util/kubernetes"
	"github.com/spf13/cobra"
	"os"
)

var (
	namespace string
	appName   string
)

// RootCmdOptions --
type RootCmdOptions struct {
	Context         context.Context
	KubeConfig      string
	Namespace       string
	ApplicationName string
}

// NewSpringBootCommand --
func NewSpringBootCommand(ctx context.Context) (*cobra.Command, error) {
	options := RootCmdOptions{
		Context: ctx,
	}

	var cmd = cobra.Command{
		Use:   "sb",
		Short: "sb client",
		Long:  `spring boot client handling cloud deployment on kubernetes`,

		Example: `        # To deploy a Spring Boot application on Kubernetes
        sb init -n namespace # to initialize/create the environment within the cloud machine
        sb push --mode binary # to push the uber jar file
        sb exec start`,
	}

	cmd.PersistentFlags().StringP("kubeconfig", "k", "", "Path to a kubeconfig ($HOME/.kube/config). Only required if out-of-cluster.")
	cmd.PersistentFlags().StringP("masterurl", "", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	// Global flag(s)
	cmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Namespace/project (defaults to current project)")
	cmd.PersistentFlags().StringVarP(&appName, "application", "a", "", "Application name (defaults to current directory name)")

	// Parse the flags before setting the defaults
	cmd.ParseFlags(os.Args)

	// Set ENV VAR to let SDK to configure the k8s client running outside of the cluster
	os.Setenv("KUBERNETES_CONFIG", kubernetes.HomeKubePath())

	if options.Namespace == "" {
		current, err := kubernetes.GetClientCurrentNamespace(options.KubeConfig)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get current namespace")
		}
		cmd.Flag("namespace").Value.Set(current)
	}

	cmd.AddCommand(newCmdVersion())
	cmd.AddCommand(newCmdInstall(&options))

	return &cmd, nil
}

// checkError prints the cause of the given error and exits the code with an
// exit code of 1.
// If the context is provided, then that is printed, if not, then the cause is
// detected using errors.Cause(err)
func checkError(err error, context string, a ...interface{}) {
	if err != nil {
		log.Debugf("Error:\n%v", err)
		if context == "" {
			fmt.Println(errors.Cause(err))
		} else {
			fmt.Printf(fmt.Sprintf("%s\n", context), a...)
		}
		os.Exit(1)
	}
}

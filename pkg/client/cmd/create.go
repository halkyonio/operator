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
	"fmt"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	log "github.com/sirupsen/logrus"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
)
func newCmdInstall(rootCmdOptions *RootCmdOptions) *cobra.Command {
	options := installCmdOptions{
		RootCmdOptions: rootCmdOptions,
	}
	cmd := cobra.Command{
		Use:     "create [flags]",
		Short:   "Create a development's pod for the component",
		Long:    `Create a development's pod for the component.`,
		Example: ` sb create -n [namespace]`,
		Args:    options.validateArgs,
		RunE:    options.create,
	}

	cmd.Annotations = map[string]string{"command": "init"}
	cmd.ParseFlags(os.Args)

	return &cmd
}

type installCmdOptions struct {
	*RootCmdOptions
}

func (*installCmdOptions) validateArgs(cmd *cobra.Command, args []string) error {
	//cobra.RangeArgs(0, 1)
	return nil
}

func (o *installCmdOptions) create(cmd *cobra.Command, args []string) error {
	log.Info("Start command called")
	name := ""
	if len(args) == 0 {
		name = "sb-a"
	} else {
		name = args[0]
	}
	component, err := o.createComponent(name)
	if err != nil {
		return err
	}
	fmt.Print("Component : ", component)
	return nil
}

func (o *installCmdOptions) createComponent(name string) (*v1alpha1.Component, error) {
	component := v1alpha1.Component{
		TypeMeta: v1.TypeMeta{
			Kind:       v1alpha1.ComponentKind,
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.ComponentSpec{
			DeploymentMode: "innerloop",
		},
	}

	existed := false
	err := sdk.Create(&component)
	if err != nil && k8serrors.IsAlreadyExists(err) {
		existed = true
		clone := component.DeepCopy()
		err = sdk.Get(clone)
		if err != nil {
			return nil, err
		}
		component.ResourceVersion = clone.ResourceVersion
		err = sdk.Update(&component)
	}

	if err != nil {
		return nil, err
	}

	if !existed {
		fmt.Printf("component \"%s\" created\n", name)
	} else {
		fmt.Printf("component \"%s\" updated\n", name)
	}

	return &component, nil
}

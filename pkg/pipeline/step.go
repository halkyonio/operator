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

package pipeline

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
)

// Action --
type Step interface {

	// a user friendly name for the action
	Name() string

	// returns true if the action can handle the integration
	CanHandle(component *v1alpha1.Component) bool

	// executes the handling function
	Handle(component *v1alpha1.Component, config *rest.Config, client *client.Client, namespace string, scheme *runtime.Scheme) error
}

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

package innerloop

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
)

var (
	defaultEnvVar = make(map[string]string)
	envs          = []v1alpha1.Env{}
)

func init() {
	defaultEnvVar["JAVA_APP_DIR"] = "/deployment"
	defaultEnvVar["JAVA_DEBUG"] = "\"false\""
	defaultEnvVar["JAVA_DEBUG_PORT"] = "\"5005\""
	defaultEnvVar["JAVA_APP_JAR"] = "app.jar"
	// defaultEnvVar["CMDS"] = "run-java:/usr/local/s2i/run;run-node:/usr/libexec/s2i;compile-java:/usr/local/s2i/assemble;build:/deployments/buildapp"
}

func populateEnvVar(component *v1alpha1.Component) {
	envs := component.Spec.Envs
	tmpEnvVar := make(map[string]string)

	// Convert Slice to Map
	for i := 0; i < len(envs); i += 1 {
		tmpEnvVar[envs[i].Name] = envs[i].Value
	}

	// Check if Component EnvVar contains the defaults
	for k, _ := range defaultEnvVar {
		if tmpEnvVar[k] == "" {
			tmpEnvVar[k] = defaultEnvVar[k]
		}
	}

	// Convert Map to Slice
	newEnvVars := []v1alpha1.Env{}
	for k, v := range tmpEnvVar {
		newEnvVars = append(newEnvVars, v1alpha1.Env{Name: k, Value: v})
	}

	// Store result
	component.Spec.Envs = newEnvVars
}

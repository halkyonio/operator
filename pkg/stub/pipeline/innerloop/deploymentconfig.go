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
	appsv1 "github.com/openshift/api/apps/v1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func InnerLoopDeploymentconfig(component *v1alpha1.Component, commands string) *appsv1.DeploymentConfig {
	if commands == "" {
		commands = "run-java:/usr/local/s2i/run;compile-java:/usr/local/s2i/assemble;build:/deployments/buildapp"
	}
	return &appsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: component.Name,
			Labels: map[string]string{
				"app":        component.Name,
			},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Replicas: 1,
			Selector: map[string]string{
				"app":              component.Name,
				"deploymentconfig": component.Name,
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.DeploymentStrategyTypeRolling,
			},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: component.Name,
					Labels: map[string]string{
						"app":              component.Name,
						"deploymentconfig": component.Name,
					},
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{*supervisordInitContainer(component.Spec.SupervisordName, commands)},
					Containers: []corev1.Container{
						{
							Image: "dev-s2i:latest",
							Name:  component.Name,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: component.Spec.Port,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env: populateEnvVar(component),
							/*							Resources: corev1.ResourceRequirements{
														Limits: corev1.ResourceList{
															corev1.ResourceCPU: resource.MustParse(appConfig.Cpu),
															corev1.ResourceMemory: resource.MustParse(appConfig.Memory),
														},
													},*/
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "shared-data",
									MountPath: "/var/lib/supervisord",
								},
								{
									Name:      "m2-data",
									MountPath: "/tmp/artifacts",
								},
							},
							Command: []string{
								"/var/lib/supervisord/bin/supervisord",
							},
							Args: []string{
								"-c",
								"/var/lib/supervisord/conf/supervisor.conf",
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "shared-data",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "m2-data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "m2-data",
								},
							},
						},
					},
				},
			},
			Triggers: []appsv1.DeploymentTriggerPolicy{
				{
					Type: "ImageChange",
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							component.Spec.SupervisordName,
						},
						From: corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: component.Spec.SupervisordName + ":latest",
						},
					},
				},
				{
					Type: "ImageChange",
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							component.Name,
						},
						From: corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "dev-s2i:latest",
						},
					},
				},
			},
		},
	}
}

func populateEnvVar(component v1alpha1.Component) []corev1.EnvVar {
	envs := []corev1.EnvVar{}

	// Add default values
	envs = append(envs,
		corev1.EnvVar{Name: "JAVA_APP_DIR", Value: "/deployments"},
		corev1.EnvVar{Name: "JAVA_DEBUG", Value: "true"},
		corev1.EnvVar{Name: "JAVA_DEBUG_PORT", Value: "5005"})

	// enrich with User's env var from MANIFEST
	for _, e := range component.Spec.Env {
		envs = append(envs, corev1.EnvVar{Name: e.Name, Value: e.Value})
	}

	if ! contains(envs,"JAVA_APP_JAR") {
		envs = append(envs, corev1.EnvVar{Name: "JAVA_APP_JAR", Value: "app.jar"})
	}

	return envs
}

func contains(envs []corev1.EnvVar, key string) bool {
	for _, env := range envs {
		if env.Name == key {
			return true
		}
	}
	return false
}

func supervisordInitContainer(name string, commands string) *corev1.Container {
	return &corev1.Container{
		Name:  name,
		Image: name + ":latest",
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "shared-data",
				MountPath: "/var/lib/supervisord",
			},
		},
		// TODO : The following list should be calculated based on the labels of the S2I image
		Env: []corev1.EnvVar{
			{
				Name:  "CMDS",
				Value: commands,
			},
		},
	}
}


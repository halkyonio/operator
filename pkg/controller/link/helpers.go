package link

import (
	corev1 "k8s.io/api/core/v1"
)

func (r *LinkManager) addSecretAsEnvFromSource(secretName string) corev1.EnvFromSource {
	return corev1.EnvFromSource{
		SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
		},
	}
}

func (r *LinkManager) addKeyValueAsEnvVar(key, value string) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  key,
		Value: value,
	}
}

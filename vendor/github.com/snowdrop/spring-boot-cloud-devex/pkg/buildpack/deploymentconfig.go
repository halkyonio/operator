package buildpack

import (
	log "github.com/sirupsen/logrus"

	restclient "k8s.io/client-go/rest"

	appsv1 "github.com/openshift/api/apps/v1"
	appsocpv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/snowdrop/spring-boot-cloud-devex/pkg/buildpack/types"
	"github.com/snowdrop/spring-boot-cloud-devex/pkg/common/oc"
	"k8s.io/client-go/kubernetes"
)

const OdoLabelName = "io.openshift.odo"
const OdoLabelValue = "inject-supervisord"

func CreatePVC(clientset *kubernetes.Clientset, application types.Application, size string) {
	if !oc.Exists("pvc", pvcName) {
		quantity, err := resource.ParseQuantity(size)
		if err != nil {
			log.Fatal(err.Error())
		}
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: pvcName,
				Labels: map[string]string{
					"app": application.Name,
				},
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: quantity,
					},
				},
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
			},
		}
		_, errPVC := clientset.CoreV1().PersistentVolumeClaims(application.Namespace).Create(pvc)
		if errPVC != nil {
			log.Fatal(errPVC.Error())
		}
	} else {
		log.Infof("'%s' PVC already exists, skipping", pvcName)
	}
}

func DeletePVC(clientset *kubernetes.Clientset, application types.Application) {
	if oc.Exists("pvc", pvcName) {
		errPVC := clientset.CoreV1().PersistentVolumeClaims(application.Namespace).Delete(pvcName, deleteOptions)
		if errPVC != nil {
			log.Fatal(errPVC.Error())
		}
	}
}

func CreateOrRetrieveDeploymentConfig(config *restclient.Config, application types.Application, commands string) *appsv1.DeploymentConfig {
	deploymentConfigV1client := getAppsClient(config)

	deploymentConfigs := deploymentConfigV1client.DeploymentConfigs(application.Namespace)

	var dc *appsv1.DeploymentConfig
	var errCreate error
	if oc.Exists("dc", application.Name) {
		dc, errCreate = deploymentConfigs.Get(application.Name, metav1.GetOptions{})
		log.Infof("'%s' DeploymentConfig already exists, skipping", application.Name)
	} else {
		dc, errCreate = deploymentConfigs.Create(javaDeploymentConfig(application, commands))
	}
	if errCreate != nil {
		log.Fatalf("DeploymentConfig not created: %s", errCreate.Error())
	}
	return dc
}

func getAppsClient(config *restclient.Config) *appsocpv1.AppsV1Client {
	deploymentConfigV1client, err := appsocpv1.NewForConfig(config)
	if err != nil {
		log.Fatalf("Can't get DeploymentConfig Clientset: %s", err.Error())
	}
	return deploymentConfigV1client
}

func DeleteDeploymentConfig(config *restclient.Config, application types.Application) {
	if oc.Exists("dc", application.Name) {
		errPVC := getAppsClient(config).DeploymentConfigs(application.Namespace).Delete(application.Name, deleteOptions)
		if errPVC != nil {
			log.Fatal(errPVC.Error())
		}
	}
}

func javaDeploymentConfig(application types.Application, commands string) *appsv1.DeploymentConfig {
	if commands == "" {
		commands = "run-java:/usr/local/s2i/run;compile-java:/usr/local/s2i/assemble;build:/deployments/buildapp"
	}
	return &appsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: application.Name,
			Labels: map[string]string{
				"app":        application.Name,
				OdoLabelName: OdoLabelValue,
			},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Replicas: 1,
			Selector: map[string]string{
				"app":              application.Name,
				"deploymentconfig": application.Name,
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.DeploymentStrategyTypeRolling,
			},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: application.Name,
					Labels: map[string]string{
						"app":              application.Name,
						"deploymentconfig": application.Name,
					},
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{*supervisordInitContainer(application.SupervisordName, commands)},
					Containers: []corev1.Container{
						{
							Image: "dev-s2i:latest",
							Name:  application.Name,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: application.Port,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env: populateEnvVar(application),
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
							application.SupervisordName,
						},
						From: corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: application.SupervisordName + ":latest",
						},
					},
				},
				{
					Type: "ImageChange",
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							application.Name,
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

func populateEnvVar(application types.Application) []corev1.EnvVar {
	envs := []corev1.EnvVar{}

	// Add default values
	envs = append(envs,
		corev1.EnvVar{Name: "JAVA_APP_DIR", Value: "/deployments"},
		corev1.EnvVar{Name: "JAVA_DEBUG", Value: "true"},
		corev1.EnvVar{Name: "JAVA_DEBUG_PORT", Value: "5005"})

	// enrich with User's env var from MANIFEST
	for _, e := range application.Env {
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

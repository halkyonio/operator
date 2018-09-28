package buildpack

import (
	buildclientsetv1 "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	restclient "k8s.io/client-go/rest"

	buildv1 "github.com/openshift/api/build/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/snowdrop/spring-boot-cloud-devex/pkg/buildpack/types"
	"log"
)

func CreateBuild(config *restclient.Config, appConfig types.Application) {
	buildClient, err := buildclientsetv1.NewForConfig(config)
	if err != nil {
	}

	//_, errbuild := buildClient.Builds(appConfig.Namespace).Create(devBuild(appConfig.Name))
	_, errbuild := buildClient.BuildConfigs(appConfig.Namespace).Create(devBuildConfig("dev-s2i", appConfig.Name))
	if errbuild != nil {
		log.Fatalf("Unable to create Build: %s", errbuild.Error())
	}
}

// Using a Build resource doesn't work as we get as response when build is executed : Error from server (NotFound): buildconfigs.build.openshift.io "spring-boot-http" not found
func devBuild(name string) *buildv1.Build {
	return &buildv1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: buildv1.BuildSpec{
			CommonSpec: buildv1.CommonSpec{
				Source: buildv1.BuildSource{
					Type: buildv1.BuildSourceBinary,
				},
				/*
					Strategy: buildv1.BuildStrategy{
						Type: buildv1.DockerBuildStrategyType,
					},
					Output:buildv1.BuildOutput{
						To: &corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: appConfig.Name + "2" + ":latest",
						},
					},
				*/
				Strategy: buildv1.BuildStrategy{
					SourceStrategy: &buildv1.SourceBuildStrategy{
						From: corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: name + ":latest",
						},
					},
				},
				Output: buildv1.BuildOutput{
					To: &corev1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: name + "2" + ":latest",
					},
				},
			},
		},
	}
}

func devBuildConfig(fromName string, toName string) *buildv1.BuildConfig {
	return &buildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: toName,
		},
		Spec: buildv1.BuildConfigSpec{
			CommonSpec: buildv1.CommonSpec{
				Output: buildv1.BuildOutput{
					To: &corev1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: toName + ":latest",
					},
				},
				Source: buildv1.BuildSource{
					Type: buildv1.BuildSourceBinary,
				},
				Strategy: buildv1.BuildStrategy{
					SourceStrategy: &buildv1.SourceBuildStrategy{
						From: corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: fromName + ":latest",
						},
					},
				},
			},
		},
	}
}

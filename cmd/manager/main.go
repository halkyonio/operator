package main

import (
	"flag"
	"log"
	"runtime"

	"github.com/snowdrop/component-operator/pkg/apis"
	"github.com/snowdrop/component-operator/pkg/controller"
	k8sutil "github.com/snowdrop/component-operator/pkg/util/kubernetes"

	servicecatalog "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	deploymentconfig "github.com/openshift/api/apps/v1"
	image "github.com/openshift/api/image/v1"
	route "github.com/openshift/api/route/v1"

	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

var (
	Version   = "unset"
	GitCommit = "HEAD"
)

func printVersion() {
	log.Printf("Go Version: %s", runtime.Version())
	log.Printf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	log.Printf("component-operator version: %v", Version)
	log.Printf("component-operator git commit: %v", GitCommit)
}

func main() {
	printVersion()
	flag.Parse()

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		log.Fatalf("failed to get watch namespace: %v", err)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{Namespace: namespace})
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Fatal(err)
	}

	// 3rd party Resources (ServiceCatalog, DeploymentConfig, Route, Image)
	if err := servicecatalog.AddToScheme(mgr.GetScheme()); err != nil {
		log.Fatal(err)
	}
	if err := deploymentconfig.AddToScheme(mgr.GetScheme()); err != nil {
		log.Fatal(err)
	}
	if err := image.AddToScheme(mgr.GetScheme()); err != nil {
		log.Fatal(err)
	}
	if err := route.AddToScheme(mgr.GetScheme()); err != nil {
		log.Fatal(err)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Fatal(err)
	}

	log.Print("Starting the Cmd.")

	// Start the Cmd
	log.Fatal(mgr.Start(signals.SetupSignalHandler()))
}

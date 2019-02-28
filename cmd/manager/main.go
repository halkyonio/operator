package main

import (
	"flag"
	"github.com/snowdrop/component-operator/pkg/apis"
	"github.com/snowdrop/component-operator/pkg/controller/component"
	k8sutil "github.com/snowdrop/component-operator/pkg/util/kubernetes"
	"log"
	"runtime"

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

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{Namespace: namespace})
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Registering Components")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Fatal(err)
	}

	// Create component controller and add it to the manager
	if err := component.New(mgr); err != nil {
		log.Fatal(err)
	}

	// Start the manager
	log.Print("Start the manager")
	log.Fatal(mgr.Start(signals.SetupSignalHandler()))
}

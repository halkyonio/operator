package main

import (
	"flag"
	"fmt"
	servicecatalogv1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	build "github.com/openshift/api/build/v1"
	image "github.com/openshift/api/image/v1"
	route "github.com/openshift/api/route/v1"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"github.com/snowdrop/component-operator/pkg/controller"
	util "github.com/snowdrop/component-operator/pkg/util"
	"github.com/spf13/pflag"
	"os"
	"runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

var (
	Version = "Unset"
	GitCommit = "HEAD"
)

var log = logf.Log.WithName("cmd")

func printVersion() {
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("Version of operator-sdk: %v", sdkVersion.Version))
	log.Info(fmt.Sprintf("component-operator version: %v", Version))
	log.Info(fmt.Sprintf("component-operator git commit: %v", GitCommit))
}

func main() {
	// Add the zap logger flag set to the CLI. The flag set must
	// be added before calling pflag.Parse().
	pflag.CommandLine.AddFlagSet(zap.FlagSet())

	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.Parse()

	// Use a zap logr.Logger implementation. If none of the zap
	// flags are configured (or if the zap flag set is not being
	// used), this defaults to a production zap logger.
	//
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	logf.SetLogger(zap.Logger())

	printVersion()

	namespace, err := util.GetWatchNamespace()
	if err != nil {
		log.Error(err,"failed to get watch namespace")
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{Namespace: namespace})
	if err != nil {
		log.Error(err,"")
	}

	// Setup Scheme for all resources
	log.Info("Registering Components")
	if err := v1alpha2.Install(mgr.GetScheme()); err != nil {
		log.Error(err,"")
	}

	log.Info("Registering 3rd party resources")
	registerAdditionalResources(mgr)

	// Create component controller and add it to the manager
	if err := controller.AddToManager(mgr); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	log.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}
}

func registerAdditionalResources(m manager.Manager) {
	if err := servicecatalogv1.AddToScheme(m.GetScheme()); err != nil {
		log.Error(err,"")
	}
	if err := route.Install(m.GetScheme()); err != nil {
		log.Error(err,"")
	}
	if err := build.Install(m.GetScheme()); err != nil {
		log.Error(err,"")
	}
	if err := image.Install(m.GetScheme()); err != nil {
		log.Error(err,"")
	}
}
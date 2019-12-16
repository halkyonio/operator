package main

import (
	"flag"
	"fmt"
	authorizv1 "github.com/openshift/api/authorization/v1"
	image "github.com/openshift/api/image/v1"
	route "github.com/openshift/api/route/v1"
	securityv1 "github.com/openshift/api/security/v1"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/spf13/pflag"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	halkyon "halkyon.io/api"
	"halkyon.io/operator-framework"
	"halkyon.io/operator/pkg/controller/capability"
	"halkyon.io/operator/pkg/controller/component"
	"halkyon.io/operator/pkg/controller/link"
	capability2 "halkyon.io/plugins/capability"
	"io/ioutil"
	"os"
	"path/filepath"
	plugin2 "plugin"
	"runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

const (
	// WatchNamespaceEnvVar holds the name of the env variable containing the name of the namespace to watch for components
	// If left empty, the operator will watch all namespaces
	WatchNamespaceEnvVar = "WATCH_NAMESPACE"
)

var (
	Version   = "Unset"
	GitCommit = "HEAD"
	log       = logf.Log.WithName("cmd")
)

func printVersion() {
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("Version of operator-sdk: %v", sdkVersion.Version))
	log.Info(fmt.Sprintf("halkyon-operator version: %v", Version))
	log.Info(fmt.Sprintf("halkyon-operator git commit: %v", GitCommit))
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

	// check if we want to watch a single namespace
	namespace, found := os.LookupEnv(WatchNamespaceEnvVar)
	syncPeriod := 30 * time.Second
	options := manager.Options{SyncPeriod: &syncPeriod}
	if found {
		options.Namespace = namespace
		log.Info("watching namespace " + namespace)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(config.GetConfigOrDie(), options)
	if err != nil {
		log.Error(err, "")
	}

	// Setup Scheme for all resources
	log.Info("Registering Halkyon resources")
	if err := halkyon.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
	}

	log.Info("Registering 3rd party resources")
	registerAdditionalResources(mgr)

	// load plugins
	capability2.SupportedCategories = make(capability2.CategoryRegistry, 7)
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	pluginsDir := filepath.Join(currentDir, "plugins")
	goPlugins, err := ioutil.ReadDir(pluginsDir)
	capability2.Plugins = make([]capability2.Plugin, 0, len(goPlugins))
	if err != nil {
		panic(err)
	}
	for _, p := range goPlugins {
		pluginPath := filepath.Join(pluginsDir, p.Name())
		if plugin, err := capability2.NewPlugin(pluginPath); err == nil {
			capability2.Plugins = append(capability2.Plugins, plugin)
			category := plugin.GetCategory()
			types, ok := capability2.SupportedCategories[category]
			if !ok {
				types = make(capability2.TypeRegistry, 3)
				capability2.SupportedCategories[category] = types
			}
			types[plugin.GetType()] = true

			goPluginPath := pluginPath + "-registration"
			if goPlugin, err := plugin2.Open(goPluginPath); err == nil {
				initializer, err := goPlugin.Lookup("Initializer")
				if err != nil {
					panic(err)
				}
				initializer.(capability2.SchemeInitializer).Init(mgr.GetScheme())
			} else {
				panic(err)
			}
			defer plugin.Kill()
		} else {
			panic(err)
		}
	}

	// Create component controller and add it to the manager
	if err := framework.RegisterNewReconciler(component.NewComponent(), mgr); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	if err := framework.RegisterNewReconciler(capability.NewCapability(), mgr); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	if err := framework.RegisterNewReconciler(link.NewLink(), mgr); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}
}

func registerAdditionalResources(m manager.Manager) {
	scheme := m.GetScheme()
	if err := authorizv1.Install(scheme); err != nil {
		log.Error(err, "")
	}
	if err := securityv1.Install(scheme); err != nil {
		log.Error(err, "")
	}
	/*if err := kubedbv1.AddToScheme(scheme); err != nil {
		log.Error(err, "")
	}*/
	if err := route.Install(scheme); err != nil {
		log.Error(err, "")
	}
	if err := tektonv1.AddToScheme(scheme); err != nil {
		log.Error(err, "")
	}
	if err := image.Install(scheme); err != nil {
		log.Error(err, "")
	}
}

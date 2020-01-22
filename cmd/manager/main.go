package main

import (
	"flag"
	"fmt"
	"github.com/hashicorp/go-getter"
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
	capability2 "halkyon.io/operator-framework/plugins/capability"
	"halkyon.io/operator/pkg/controller/capability"
	"halkyon.io/operator/pkg/controller/component"
	"halkyon.io/operator/pkg/controller/link"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

const (
	// WatchNamespaceEnvVar holds the name of the env variable containing the name of the namespace to watch for components
	// If left empty, the operator will watch all namespaces
	WatchNamespaceEnvVar = "WATCH_NAMESPACE"
	// HalkyonPluginsEnvVar holds the name of the env variable defining which plugins need to be downloaded as a comma-separated
	// list of values following the <github org>/<github project>@<version> format e.g. halkyonio/postgresql-capability@v1.0.0-beta.3
	HalkyonPluginsEnvVar = "HALKYON_PLUGINS"
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
	config := config.GetConfigOrDie()
	mgr, err := manager.New(config, options)
	if err != nil {
		log.Error(err, "")
	}

	// check if we run on OpenShift early so that things are initialized for DependentResources which might depend on it
	framework.InitHelper(mgr)

	// Setup Scheme for all resources
	log.Info("Registering Halkyon resources")
	if err := halkyon.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
	}

	log.Info("Registering 3rd party resources")
	registerAdditionalResources(mgr)

	// load plugins based on specified list
	log.Info("Loading plugins")
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	pluginsDir := filepath.Join(currentDir, "plugins")
	if pluginList, found := os.LookupEnv(HalkyonPluginsEnvVar); found {
		pluginDefs := strings.Split(pluginList, ",")
		for _, pluginDef := range pluginDefs {
			// only download the plugin if we haven't already done so before
			markerFileName := strings.ReplaceAll("."+pluginDef, "/", "___")
			markerFileName = filepath.Join(pluginsDir, markerFileName)
			if _, err := os.Stat(markerFileName); err == nil {
				log.Info(pluginDef + ": already downloaded")
			} else {
				pluginParts := strings.Split(pluginDef, "@")
				url := "https://github.com/" + pluginParts[0] + "/releases/download/" + pluginParts[1] + "/halkyon_plugin_" + runtime.GOOS + ".tar.gz"
				log.Info(pluginDef + ": downloading from " + url)
				err := getter.GetAny(pluginsDir, url)
				if err != nil {
					log.Error(err, "couldn't download plugin at "+url)
				}
				// create marker file to avoid re-downloading the plugin at next re-start
				marker, err := os.Create(markerFileName)
				if err != nil {
					panic(err)
				}
				_ = marker.Close()
			}
		}
	}
	// initialize plugins
	goPlugins, err := ioutil.ReadDir(pluginsDir)
	pluginCount := 0
	typeCount := 0
	if err != nil {
		log.Error(err, "cannot read plugins directory")
	} else {
		for _, p := range goPlugins {
			// ignore marker files
			if !strings.HasPrefix(p.Name(), ".") {
				pluginPath := filepath.Join(pluginsDir, p.Name())
				if runtime.GOOS == "windows" {
					pluginPath += ".exe"
				}
				if plugin, err := capability2.NewPlugin(pluginPath, log); err == nil {
					pluginCount++
					typeCount += len(plugin.GetTypes())
					defer plugin.Kill()
				} else {
					log.Error(err, "ignoring "+pluginPath+" plugin which couldn't be loaded")
				}
			}
		}
		log.Info(fmt.Sprintf("Loaded %d plugin(s) for a total of %d capabilities", pluginCount, typeCount))
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

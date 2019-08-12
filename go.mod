module github.com/halkyonio/operator

go 1.12

require (
	github.com/appscode/go v0.0.0-20190424183524-60025f1135c9
	github.com/go-logr/logr v0.1.0
	github.com/knative/pkg v0.0.0-20190409220258-28cfa161499b
	github.com/kubedb/apimachinery v0.0.0-20190506191700-871d6b5d30ee
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/openshift/api v0.0.0-20190322043348-8741ff068a47
	github.com/operator-framework/operator-sdk v0.8.1
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.1
	github.com/snowdrop/component-operator v0.0.3 // indirect
	github.com/spf13/pflag v1.0.3
	github.com/tektoncd/pipeline v0.3.1
	golang.org/x/net v0.0.0-20190503192946-f4e77d36d62c
	k8s.io/api v0.0.0-20190503110853-61630f889b3c
	k8s.io/apimachinery v0.0.0-20190508063446-a3da69d3723c
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/code-generator v0.0.0-20190419212335-ff26e7842f9d // needed for code generation
	k8s.io/kubernetes v1.14.5
	kmodules.xyz/custom-resources v0.0.0-20190508103408-464e8324c3ec // indirect
	kmodules.xyz/monitoring-agent-api v0.0.0-20190513065523-186af167f817 // indirect
	kmodules.xyz/objectstore-api v0.0.0-20190516233206-ea3ba546e348 // indirect
	kmodules.xyz/offshoot-api v0.0.0-20190527060812-295f97bb8061
	sigs.k8s.io/controller-runtime v0.1.9
)

replace (
	github.com/graymeta/stow => github.com/appscode/stow v0.0.0-20190506085026-ca5baa008ea3
	k8s.io/api => k8s.io/api v0.0.0-20181126151915-b503174bad59
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery => github.com/kmodules/apimachinery v0.0.0-20190423074438-a3cbe62563e6
	k8s.io/apiserver => github.com/kmodules/apiserver v0.0.0-20190423074744-1ff296932385
	k8s.io/client-go => k8s.io/client-go v0.0.0-20181126152608-d082d5923d3c
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190311093542-50b561225d70
)

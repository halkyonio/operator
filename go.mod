module github.com/snowdrop/component-operator

go 1.12

require (
	github.com/Azure/azure-sdk-for-go v30.0.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.1.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.2.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.1.0 // indirect
	github.com/Masterminds/semver v1.4.2 // indirect
	github.com/appscode/go v0.0.0-20190424183524-60025f1135c9
	github.com/aws/aws-sdk-go v1.19.41 // indirect
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/gobuffalo/envy v1.7.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/google/go-containerregistry v0.0.0-20190531175139-2687bd5ba651 // indirect
	github.com/gophercloud/gophercloud v0.0.0-20190509032623-7892efa714f1 // indirect
	github.com/iancoleman/strcase v0.0.0-20190422225806-e506e3ef7365 // indirect
	github.com/knative/pkg v0.0.0-20190409220258-28cfa161499b
	github.com/kubedb/apimachinery v0.0.0-20190506191700-871d6b5d30ee
	github.com/markbates/inflect v1.0.4 // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/openshift/api v0.0.0-20190322043348-8741ff068a47
	github.com/operator-framework/operator-sdk v0.7.0
	github.com/pkg/errors v0.8.1
	github.com/rogpeppe/go-internal v1.3.0 // indirect
	github.com/sirupsen/logrus v1.4.1
	github.com/spf13/pflag v1.0.3
	github.com/tektoncd/pipeline v0.3.1
	golang.org/x/lint v0.0.0-20190409202823-959b441ac422 // indirect
	golang.org/x/net v0.0.0-20190503192946-f4e77d36d62c
	golang.org/x/tools v0.0.0-20190506145303-2d16b83fe98c // indirect
	google.golang.org/api v0.5.0 // indirect
	google.golang.org/genproto v0.0.0-20190508193815-b515fa19cec8 // indirect
	k8s.io/api v0.0.0-20190503110853-61630f889b3c
	k8s.io/apimachinery v0.0.0-20190508063446-a3da69d3723c
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/code-generator v0.0.0-20190419212335-ff26e7842f9d // needed for code generation
	k8s.io/helm v2.13.1+incompatible // indirect
	k8s.io/kubernetes v1.14.2
	kmodules.xyz/client-go v0.0.0-20190524133821-9c8a87771aea
	kmodules.xyz/custom-resources v0.0.0-20190508103408-464e8324c3ec // indirect
	kmodules.xyz/monitoring-agent-api v0.0.0-20190513065523-186af167f817 // indirect
	kmodules.xyz/objectstore-api v0.0.0-20190516233206-ea3ba546e348 // indirect
	kmodules.xyz/offshoot-api v0.0.0-20190527060812-295f97bb8061
	sigs.k8s.io/controller-runtime v0.1.9
	sigs.k8s.io/controller-tools v0.1.10 // indirect
	sigs.k8s.io/testing_frameworks v0.1.1 // indirect
)

replace (
	github.com/graymeta/stow => github.com/appscode/stow v0.0.0-20190506085026-ca5baa008ea3
	k8s.io/api => k8s.io/api v0.0.0-20181126151915-b503174bad59
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery => github.com/kmodules/apimachinery v0.0.0-20190508045248-a52a97a7a2bf
	k8s.io/apiserver => github.com/kmodules/apiserver v0.0.0-20190508082252-8397d761d4b5
	k8s.io/client-go => k8s.io/client-go v0.0.0-20181126152608-d082d5923d3c
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190311093542-50b561225d70
)

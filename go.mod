module halkyon.io/operator

go 1.12

require (
	contrib.go.opencensus.io/exporter/prometheus v0.1.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.12.8 // indirect
	github.com/Azure/go-autorest/autorest v0.9.3 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.8.1 // indirect
	github.com/aws/aws-sdk-go v1.26.2 // indirect
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/gobuffalo/envy v1.7.0 // indirect
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9 // indirect
	github.com/gophercloud/gophercloud v0.0.0-20190509032623-7892efa714f1 // indirect
	github.com/hashicorp/go-getter v1.4.1
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/markbates/inflect v1.0.4 // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
	github.com/mitchellh/gox v1.0.1 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/openshift/api v0.0.0-20190322043348-8741ff068a47
	github.com/openzipkin/zipkin-go v0.2.2 // indirect
	github.com/operator-framework/operator-sdk v0.8.1
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.2.1 // indirect
	github.com/prometheus/client_model v0.0.0-20191202183732-d1d2010b5bee // indirect
	github.com/sergi/go-diff v1.0.0 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/tektoncd/pipeline v0.9.1
	go.opencensus.io v0.22.2 // indirect
	go.uber.org/zap v1.13.0 // indirect
	golang.org/x/net v0.0.0-20191209160850-c0dbc17a3553
	google.golang.org/api v0.15.0 // indirect
	halkyon.io/api v1.0.0-rc.1
	halkyon.io/operator-framework v1.0.0-beta.2
	halkyon.io/plugins v1.0.0-beta.3
	k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apiextensions-apiserver v0.0.0-00010101000000-000000000000 // indirect
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v11.0.1-0.20190805182715-88a2adca7e76+incompatible
	knative.dev/pkg v0.0.0-20200108000451-298f22bea61f
	sigs.k8s.io/controller-runtime v0.3.0
	sigs.k8s.io/controller-tools v0.1.10 // indirect
	sigs.k8s.io/testing_frameworks v0.1.2 // indirect
)

replace (
	github.com/go-check/check => github.com/go-check/check v0.0.0-20180628173108-788fd7840127
	github.com/graymeta/stow => github.com/appscode/stow v0.0.0-20190506085026-ca5baa008ea3
	k8s.io/api => k8s.io/api v0.0.0-20181126151915-b503174bad59
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery => github.com/kmodules/apimachinery v0.0.0-20190423074438-a3cbe62563e6
	k8s.io/apiserver => github.com/kmodules/apiserver v0.0.0-20190423074744-1ff296932385
	k8s.io/client-go => k8s.io/client-go v0.0.0-20181126152608-d082d5923d3c
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190311093542-50b561225d70
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.1.9
)

module halkyon.io/operator

go 1.12

require (
	github.com/Azure/azure-sdk-for-go v30.0.0+incompatible // indirect
	github.com/aws/aws-sdk-go v1.19.41 // indirect
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/gobuffalo/envy v1.7.0 // indirect
	github.com/google/go-containerregistry v0.0.0-20190531175139-2687bd5ba651 // indirect
	github.com/gophercloud/gophercloud v0.0.0-20190509032623-7892efa714f1 // indirect
	github.com/knative/pkg v0.0.0-20190409220258-28cfa161499b
	github.com/kubedb/apimachinery v0.0.0-20190506191700-871d6b5d30ee
	github.com/markbates/inflect v1.0.4 // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/openshift/api v0.0.0-20190322043348-8741ff068a47
	github.com/operator-framework/operator-sdk v0.8.1
	github.com/pkg/errors v0.8.1
	github.com/rogpeppe/go-internal v1.3.0 // indirect
	github.com/sirupsen/logrus v1.4.1
	github.com/spf13/pflag v1.0.3
	github.com/tektoncd/pipeline v0.9.1
	golang.org/x/net v0.0.0-20190724013045-ca1201d0de80
	golang.org/x/tools v0.0.0-20190506145303-2d16b83fe98c
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543 // indirect
	halkyon.io/api v1.0.0-beta.7
	halkyon.io/operator-framework v0.0.0-20191108175501-3d0a053bc383
	halkyon.io/plugins v0.0.0-20191108175809-0426164e2120
	k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/client-go v11.0.1-0.20190805182715-88a2adca7e76+incompatible
	k8s.io/kubernetes v1.14.5 // indirect
	kmodules.xyz/custom-resources v0.0.0-20190508103408-464e8324c3ec // indirect
	kmodules.xyz/monitoring-agent-api v0.0.0-20190513065523-186af167f817 // indirect
	kmodules.xyz/objectstore-api v0.0.0-20190516233206-ea3ba546e348 // indirect
	kmodules.xyz/offshoot-api v0.0.0-20190527060812-295f97bb8061
	knative.dev/pkg v0.0.0-20191211150249-bebd5557feae
	sigs.k8s.io/controller-runtime v0.1.9
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

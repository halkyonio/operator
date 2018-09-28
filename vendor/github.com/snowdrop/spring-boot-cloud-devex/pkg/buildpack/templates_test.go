package buildpack_test

import (
	"bytes"
	"fmt"
	"github.com/snowdrop/spring-boot-cloud-devex/pkg/buildpack/types"
	"path"
	"runtime"
	"testing"
	"text/template"
)

func TestServiceTemplate(t *testing.T) {

	builderpath := "tmpl/java/"

	const service = `apiVersion: v1
kind: Service
metadata:
  name: service-test
  labels:
    app: service-test
    name: service-test
spec:
  ports:
  - port: 8080
    protocol: TCP
    targetPort: 8080
  selector:
    app: service-test
    deploymentconfig: service-test`

	application := types.Application{
		Name: "service-test",
		Port: 8080,
	}

	// Get package full path
	serviceFile := packageDirectory() + "/" + builderpath + "/service"
	templ, _ := template.New("service").ParseFiles(serviceFile)

	var b bytes.Buffer
	templ.Execute(&b, application)
	r := b.String()

	fmt.Println(service)
	fmt.Println("================")
	fmt.Println(r)

	if service != r {
		t.Errorf("Result was incorrect, got: "+
			"%s"+
			", want: "+
			"%s", r, service)
	}
}

func packageDirectory() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	return path.Dir(filename)
}

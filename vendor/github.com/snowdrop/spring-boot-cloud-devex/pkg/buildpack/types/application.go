package types

import "github.com/pkg/errors"

type Application struct {
	Name            string
	Version         string
	Namespace       string
	Replica         int
	Cpu             string `default:"100m"`
	Memory          string `default:"250Mi"`
	Port            int32  `default:"8080"`
	Image           Image
	SupervisordName string
	Env             []Env
	Services        []Service
}

func (app *Application) GetService(name string) (Service, error) {
	for _, service := range app.Services {
		if service.Name == name {
			return service, nil
		}
	}

	return Service{}, errors.Errorf("Couldn't find Service named %s", name)
}

type Service struct {
	Class      string
	Name       string
	Plan       string `default:"dev"`
	ExternalId string
	Parameters []Parameter
}

func (service *Service) GetParameter(name string) (Parameter, error) {
	for _, parameter := range service.Parameters {
		if parameter.Name == name {
			return parameter, nil
		}
	}

	return Parameter{}, errors.Errorf("Couldn't find Parameter named %s", name)
}

func (service *Service) ParametersAsMap() map[string]string {
	result := make(map[string]string, len(service.Parameters))
	for _, parameter := range service.Parameters {
		result[parameter.Name] = parameter.Value
	}

	return result
}

type Parameter struct {
	Name  string
	Value string
}

type Env struct {
	Name  string
	Value string
}

type Image struct {
	Name           string
	AnnotationCmds bool
	Repo           string
	Tag            string
	DockerImage    bool
}

func NewApplication() Application {
	return Application{
		Version:         "1.0",
		Cpu:             "100m",
		Memory:          "250Mi",
		Replica:         1,
		Port:            8080,
		SupervisordName: "copy-supervisord",
	}
}

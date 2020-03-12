package component

import (
	"context"
	"fmt"
	v1beta12 "halkyon.io/api/capability/v1beta1"
	"halkyon.io/api/component/v1beta1"
	beta1 "halkyon.io/api/v1beta1"
	framework "halkyon.io/operator-framework"
	"halkyon.io/operator/pkg"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type requiredCapability struct {
	base
	capabilityConfig v1beta1.RequiredCapabilityConfig
}

var _ framework.DependentResource = &requiredCapability{}
var capabilityGVK = v1beta12.SchemeGroupVersion.WithKind(v1beta12.Kind)

func newRequiredCapability(owner *v1beta1.Component, capConfig v1beta1.RequiredCapabilityConfig) requiredCapability {
	config := framework.NewConfig(capabilityGVK)
	config.CheckedForReadiness = true
	config.Created = false
	config.Updated = true
	config.TypeName = "Required Capability"
	c := requiredCapability{base: newConfiguredBaseDependent(owner, config), capabilityConfig: capConfig}
	c.NameFn = c.Name
	return c
}

func (res requiredCapability) Build(empty bool) (runtime.Object, error) {
	if empty {
		return &v1beta12.Capability{}, nil
	}
	// we don't want to be building anything: the capability is under halkyon's control, it's not up to the component to create it
	return nil, nil
}

func (res requiredCapability) Update(toUpdate runtime.Object) (bool, runtime.Object, error) {
	c := toUpdate.(*v1beta12.Capability)
	return res.updateWithParametersIfNeeded(c, false)
}

func (res requiredCapability) updateWithParametersIfNeeded(c *v1beta12.Capability, doUpdate bool) (bool, *v1beta12.Capability, error) {
	updated := false

	// examine all given parameters that starts by `halkyon` as these denote parameters to pass to the underlying plugin
	// as opposed to parameters used to match a capability
	wanted := res.capabilityConfig.Spec.Parameters
	for _, parameter := range wanted {
		name := parameter.Name
		if strings.HasPrefix(name, "halkyon.") {
			value := parameter.Value
			// try to see if that parameter was already set for that capability
			found := false
			for i, pair := range c.Spec.Parameters {
				if pair.Name == name {
					if pair.Value != value {
						updated = true
						c.Spec.Parameters[i] = beta1.NameValuePair{Name: name, Value: value}
					}
					found = true
					break
				}
			}
			// if we didn't find the parameter, add it
			if !found {
				updated = true
				c.Spec.Parameters = append(c.Spec.Parameters, beta1.NameValuePair{Name: name, Value: value})
			}
		}
	}
	if doUpdate && updated {
		err := framework.Helper.Client.Update(context.Background(), c)
		if err != nil {
			return updated, nil, &contractError{msg: fmt.Sprintf("couldn't update capability '%s': %s", c.Name, err.Error())}
		}
	}

	return updated, c, nil
}

func (res requiredCapability) Name() string {
	return res.capabilityConfig.Name
}

func (res requiredCapability) GetCondition(underlying runtime.Object, err error) *beta1.DependentCondition {
	return framework.DefaultCustomizedGetConditionFor(res, err, underlying, func(underlying runtime.Object, cond *beta1.DependentCondition) {
		c := underlying.(*v1beta12.Capability)
		if c.Status.Reason != beta1.ReasonReady {
			cond.Type = beta1.DependentPending
		}
		cond.Message = c.Status.Message
	})
}

type contractError struct {
	msg string
}

func (e contractError) Error() string {
	return e.msg
}

func (res requiredCapability) Fetch() (runtime.Object, error) {
	config := res.capabilityConfig
	spec := config.Spec
	selector := selectorFor(spec)

	var result *v1beta12.Capability
	component := res.ownerAsComponent()

	// if the component defines a bound value, try to retrieve it and check that it conforms to requirements
	if len(config.BoundTo) > 0 {
		result = &v1beta12.Capability{}
		_, err := framework.Helper.Fetch(config.BoundTo, component.Namespace, result)
		if err != nil {
			return nil, &contractError{msg: err.Error()}
		}

		// if the referenced capability matches, return it
		foundSpec := result.Spec
		if foundSpec.Matches(spec) {
			_, result, err = res.updateWithParametersIfNeeded(result, true)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
		return nil, &contractError{msg: fmt.Sprintf("specified '%s' bound to capability doesn't match %v requirements, was: %v", config.BoundTo, selector, selectorFor(foundSpec))}
	}

	// retrieve names of matching capabilities along with last (and hopefully, only) matching one
	names, result, err := capabilitiesNameMatching(spec)
	if err != nil {
		return nil, &contractError{msg: fmt.Sprintf("couldn't find matching capabilities: %s", err.Error())}
	}

	// otherwise, check if we can auto-bind to an available capability
	if config.AutoBindable {
		if len(names) > 1 {
			return nil, &contractError{msg: fmt.Sprintf("cannot autobind because several capabilities match %v: '%s', use explicit binding instead", selector, strings.Join(names, ", "))}
		}
		if result != nil {
			// set the boundTo attribute on the required capability
			requires := component.Spec.Capabilities.Requires
			for i, require := range requires {
				if require.Name == config.Name {
					requires[i].BoundTo = result.Name
					break
				}
			}
			_, result, err = res.updateWithParametersIfNeeded(result, true)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
	}

	msg := ""
	switch len(names) {
	case 0:
		msg = fmt.Sprintf("no capability matching '%v' was found", selector)
	case 1:
		msg = fmt.Sprintf("no capability bound, found one matching candidate: '%s'", result.Name)
	default:
		msg = fmt.Sprintf("no capability bound, several matching candidates were found: '%s'", strings.Join(names, ", "))
	}

	return nil, &contractError{msg: msg}
}

func selectorFor(spec v1beta12.CapabilitySpec) fields.Selector {
	selector := fields.AndSelectors(fields.OneTermEqualSelector("spec.category", spec.Category.String()), fields.OneTermEqualSelector("spec.type", spec.Type.String()))
	if len(spec.Version) > 0 {
		selector = fields.AndSelectors(selector, fields.OneTermEqualSelector("spec.version", spec.Version))
	}
	return selector
}

func capabilitiesNameMatching(spec v1beta12.CapabilitySpec) (names []string, lastMatching *v1beta12.Capability, err error) {
	matching := &v1beta12.CapabilityList{}
	err = framework.Helper.Client.List(context.TODO(), &client.ListOptions{ /*FieldSelector: selector*/ }, matching)
	if err != nil {
		return nil, nil, err
	}
	capabilityNb := len(matching.Items)
	names = make([]string, 0, capabilityNb)
	for _, capability := range matching.Items {
		if capability.Spec.Matches(spec) {
			names = append(names, capability.Name)
			lastMatching = &capability
		}
	}
	return names, lastMatching, nil
}

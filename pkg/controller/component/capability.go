package component

import (
	"context"
	"fmt"
	v1beta12 "halkyon.io/api/capability/v1beta1"
	"halkyon.io/api/component/v1beta1"
	framework "halkyon.io/operator-framework"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type capability struct {
	base
	capabilityConfig v1beta1.CapabilityConfig
}

var _ framework.DependentResource = &capability{}

func newCapability(owner *v1beta1.Component, capConfig v1beta1.CapabilityConfig) capability {
	config := framework.NewConfig(v1beta12.SchemeGroupVersion.WithKind(v1beta12.Kind))
	config.CheckedForReadiness = true
	config.CreatedOrUpdated = false
	return capability{base: newConfiguredBaseDependent(owner, config), capabilityConfig: capConfig}
}

func (res capability) Build(empty bool) (runtime.Object, error) {
	if empty {
		return &v1beta12.Capability{}, nil
	}
	// we don't want to be building anything: the capability is under halkyon's control, it's not up to the component to create it
	return nil, nil
}

func (res capability) NameFrom(underlying runtime.Object) string {
	return underlying.(*v1beta12.Capability).Name
}

func (res capability) IsReady(underlying runtime.Object) (bool, string) {
	c := underlying.(*v1beta12.Capability)
	return c.Status.Phase == v1beta12.CapabilityReady, c.Status.Message
}

func (res capability) Fetch() (runtime.Object, error) {
	config := res.capabilityConfig
	spec := config.Spec
	selector := fields.AndSelectors(fields.OneTermEqualSelector("spec.category", spec.Category.String()), fields.OneTermEqualSelector("spec.type", spec.Type.String()))
	if len(spec.Version) > 0 {
		selector = fields.AndSelectors(selector, fields.OneTermEqualSelector("spec.version", spec.Version))
	}

	var result *v1beta12.Capability
	component := res.ownerAsComponent()

	// if the component defines a bound value, try to retrieve it and check that it conforms to requirements
	if len(config.BoundTo) > 0 {
		result = &v1beta12.Capability{}
		fetch, err := framework.Helper.Fetch(config.BoundTo, component.Namespace, result)

		// if the referenced capability matches, return it and change the link status to started if we're not already linked
		if matches(fetch.(*v1beta12.Capability).Spec, spec) {
			if err := updateLinkingStatus(component, config, true); err != nil {
				return nil, err
			}
			return fetch, nil
		}
		return nil, err
	}

	// retrieve names of matching capabilities along with last (and hopefully, only) matching one
	names, result, err := capabilitiesNameMatching(spec)
	if err != nil {
		return nil, err
	}

	// otherwise, check if we can auto-bind to an available capability
	if config.AutoBindable {
		if len(names) > 1 {
			return nil, fmt.Errorf("cannot autobind because several capabilities match %v: '%s', use explicit binding instead", selector, strings.Join(names, ", "))
		}
		if result != nil {
			if err := updateLinkingStatus(component, config, false); err != nil {
				return nil, err
			}
			// todo: replace by status update instead of modifying spec
			requires := component.Spec.Capabilities.Requires
			for i, cc := range requires {
				if config.Name == cc.Name {
					requires[i] = v1beta1.CapabilityConfig{
						Name:         config.Name,
						BoundTo:      result.Name,
						AutoBindable: config.AutoBindable,
						Spec:         config.Spec,
					}
				}
			}
			err := framework.Helper.Client.Update(context.Background(), component)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
	}

	switch len(names) {
	case 0:
		err = fmt.Errorf("no capability matching '%v' was found", selector)
	case 1:
		err = fmt.Errorf("no capability bound, found one matching candidate: '%s'", result.Name)
	default:
		err = fmt.Errorf("no capability bound, several matching candidates were found: '%s'", strings.Join(names, ", "))
	}

	return nil, err
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
		if matches(spec, capability.Spec) {
			names = append(names, capability.Name)
			lastMatching = &capability
		}
	}
	return names, lastMatching, nil
}

func updateLinkingStatus(component *v1beta1.Component, config v1beta1.CapabilityConfig, boundFound bool) error {
	status := &component.Status
	var linkingStatus v1beta1.LinkingStatus
	updateStatus := false
	switch status.GetLinkStatus(config.Name) {
	case v1beta1.Started:
		if boundFound {
			linkingStatus = v1beta1.Linked
			updateStatus = true
		}
	default:
		linkingStatus = v1beta1.Started
		updateStatus = true
	}

	if updateStatus {
		status.SetLinkingStatus(config.Name, linkingStatus, status.PodName)
		return framework.Helper.Client.Status().Update(context.Background(), component)
	}
	return nil
}

func matches(requested, actual v1beta12.CapabilitySpec) bool {
	// first check that category and type match
	if requested.Category.Equals(actual.Category) && requested.Type.Equals(actual.Type) {
		// if we're asking for a specific version then we need to provide a capability with that version
		// todo: implement range matching on version?
		return len(requested.Version) == 0 || requested.Version == actual.Version
	}
	return false
}

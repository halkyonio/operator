/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubernetes

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/yaml"
	cgoscheme "k8s.io/client-go/kubernetes/scheme"

	deploymentconfig "github.com/openshift/api/apps/v1"
	image "github.com/openshift/api/image/v1"
	route "github.com/openshift/api/route/v1"
    servicecatalog "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
)

var (
	scheme      = runtime.NewScheme()
	codecs      = serializer.NewCodecFactory(scheme)
	decoderFunc = decoder
)

func init() {
	// Add the standard kubernetes [GVK:Types] type registry
	// e.g (v1,Pods):&v1.Pod{}
	metav1.AddToGroupVersion(scheme, schema.GroupVersion{Version: "v1"})
	cgoscheme.AddToScheme(scheme)

	//add openshift types
	deploymentconfig.AddToScheme(scheme)
	image.AddToScheme(scheme)
	route.AddToScheme(scheme)

	//add kubernetes types
	servicecatalog.AddToScheme(scheme)
}

// UtilDecoderFunc retrieve the correct decoder from a GroupVersion
// and the schemes codec factory.
type UtilDecoderFunc func(schema.GroupVersion, serializer.CodecFactory) runtime.Decoder

// SetDecoderFunc sets a non default decoder function
// This is used as a work around to add support for unstructured objects
func SetDecoderFunc(u UtilDecoderFunc) {
	decoderFunc = u
}

func decoder(gv schema.GroupVersion, codecs serializer.CodecFactory) runtime.Decoder {
	codec := codecs.UniversalDecoder(gv)
	return codec
}

type addToSchemeFunc func(*runtime.Scheme) error

// AddToSDKScheme allows CRDs to register their types with the sdk scheme
func AddToSDKScheme(addToScheme addToSchemeFunc) {
	addToScheme(scheme)
}

func PopulateKubernetesObjectFromYaml(data string) (*unstructured.Unstructured, error) {
	yml := []byte(data)
	json, err := yaml.ToJSON(yml)
	if err != nil {
		return nil, err
	}
	u := unstructured.Unstructured{}
	err = u.UnmarshalJSON(json)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// RuntimeObjectFromUnstructured converts an unstructured to a runtime object
func RuntimeObjectFromUnstructured(u *unstructured.Unstructured) (runtime.Object, error) {
	gvk := u.GroupVersionKind()
	decoder := decoderFunc(gvk.GroupVersion(), codecs)

	b, err := u.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("error running MarshalJSON on unstructured object: %v", err)
	}
	ro, _, err := decoder.Decode(b, &gvk, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decode json data with gvk(%v): %v", gvk.String(), err)
	}
	return ro, nil
}

// UnstructuredFromRuntimeObject converts a runtime object to an unstructured
func UnstructuredFromRuntimeObject(ro runtime.Object) (*unstructured.Unstructured, error) {
	b, err := json.Marshal(ro)
	if err != nil {
		return nil, fmt.Errorf("error running MarshalJSON on runtime object: %v", err)
	}
	var u unstructured.Unstructured
	if err := json.Unmarshal(b, &u.Object); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json into unstructured object: %v", err)
	}
	return &u, nil
}

// UnstructuredIntoRuntimeObject unmarshalls an unstructured into a given runtime object
// TODO: https://github.com/operator-framework/operator-sdk/issues/127
func UnstructuredIntoRuntimeObject(u *unstructured.Unstructured, into runtime.Object) error {
	gvk := u.GroupVersionKind()
	decoder := decoderFunc(gvk.GroupVersion(), codecs)

	b, err := u.MarshalJSON()
	if err != nil {
		return err
	}
	_, _, err = decoder.Decode(b, &gvk, into)
	if err != nil {
		return fmt.Errorf("failed to decode json data with gvk(%v): %v", gvk.String(), err)
	}
	return nil
}

/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha2

import (
	v1alpha2 "github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// LinkLister helps list Links.
type LinkLister interface {
	// List lists all Links in the indexer.
	List(selector labels.Selector) (ret []*v1alpha2.Link, err error)
	// Links returns an object that can list and get Links.
	Links(namespace string) LinkNamespaceLister
	LinkListerExpansion
}

// linkLister implements the LinkLister interface.
type linkLister struct {
	indexer cache.Indexer
}

// NewLinkLister returns a new LinkLister.
func NewLinkLister(indexer cache.Indexer) LinkLister {
	return &linkLister{indexer: indexer}
}

// List lists all Links in the indexer.
func (s *linkLister) List(selector labels.Selector) (ret []*v1alpha2.Link, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha2.Link))
	})
	return ret, err
}

// Links returns an object that can list and get Links.
func (s *linkLister) Links(namespace string) LinkNamespaceLister {
	return linkNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// LinkNamespaceLister helps list and get Links.
type LinkNamespaceLister interface {
	// List lists all Links in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha2.Link, err error)
	// Get retrieves the Link from the indexer for a given namespace and name.
	Get(name string) (*v1alpha2.Link, error)
	LinkNamespaceListerExpansion
}

// linkNamespaceLister implements the LinkNamespaceLister
// interface.
type linkNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Links in the indexer for a given namespace.
func (s linkNamespaceLister) List(selector labels.Selector) (ret []*v1alpha2.Link, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha2.Link))
	})
	return ret, err
}

// Get retrieves the Link from the indexer for a given namespace and name.
func (s linkNamespaceLister) Get(name string) (*v1alpha2.Link, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha2.Resource("link"), name)
	}
	return obj.(*v1alpha2.Link), nil
}

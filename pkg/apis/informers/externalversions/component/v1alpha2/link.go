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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha2

import (
	time "time"

	versioned "github.com/snowdrop/component-operator/pkg/apis/clientset/versioned"
	componentv1alpha2 "github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	internalinterfaces "github.com/snowdrop/component-operator/pkg/apis/informers/externalversions/internalinterfaces"
	v1alpha2 "github.com/snowdrop/component-operator/pkg/apis/listers/component/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// LinkInformer provides access to a shared informer and lister for
// Links.
type LinkInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha2.LinkLister
}

type linkInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewLinkInformer constructs a new informer for Link type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewLinkInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredLinkInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredLinkInformer constructs a new informer for Link type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredLinkInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.DevexpV1alpha2().Links(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.DevexpV1alpha2().Links(namespace).Watch(options)
			},
		},
		&componentv1alpha2.Link{},
		resyncPeriod,
		indexers,
	)
}

func (f *linkInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredLinkInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *linkInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&componentv1alpha2.Link{}, f.defaultInformer)
}

func (f *linkInformer) Lister() v1alpha2.LinkLister {
	return v1alpha2.NewLinkLister(f.Informer().GetIndexer())
}

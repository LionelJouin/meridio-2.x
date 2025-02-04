/*
Copyright (c) 2024 OpenInfra Foundation Europe

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

package v1alpha1

import (
	"context"
	time "time"

	apisv1alpha1 "github.com/lioneljouin/meridio-experiment/apis/v1alpha1"
	versioned "github.com/lioneljouin/meridio-experiment/pkg/client/clientset/versioned"
	internalinterfaces "github.com/lioneljouin/meridio-experiment/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/lioneljouin/meridio-experiment/pkg/client/listers/apis/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// GatewayRouterInformer provides access to a shared informer and lister for
// GatewayRouters.
type GatewayRouterInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.GatewayRouterLister
}

type gatewayRouterInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewGatewayRouterInformer constructs a new informer for GatewayRouter type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewGatewayRouterInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredGatewayRouterInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredGatewayRouterInformer constructs a new informer for GatewayRouter type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredGatewayRouterInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.MeridioV1alpha1().GatewayRouters(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.MeridioV1alpha1().GatewayRouters(namespace).Watch(context.TODO(), options)
			},
		},
		&apisv1alpha1.GatewayRouter{},
		resyncPeriod,
		indexers,
	)
}

func (f *gatewayRouterInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredGatewayRouterInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *gatewayRouterInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&apisv1alpha1.GatewayRouter{}, f.defaultInformer)
}

func (f *gatewayRouterInformer) Lister() v1alpha1.GatewayRouterLister {
	return v1alpha1.NewGatewayRouterLister(f.Informer().GetIndexer())
}

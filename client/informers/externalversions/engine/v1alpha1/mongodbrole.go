/*
Copyright 2019 The Kube Vault Authors.

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
	time "time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
	enginev1alpha1 "kubevault.dev/operator/apis/engine/v1alpha1"
	versioned "kubevault.dev/operator/client/clientset/versioned"
	internalinterfaces "kubevault.dev/operator/client/informers/externalversions/internalinterfaces"
	v1alpha1 "kubevault.dev/operator/client/listers/engine/v1alpha1"
)

// MongoDBRoleInformer provides access to a shared informer and lister for
// MongoDBRoles.
type MongoDBRoleInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.MongoDBRoleLister
}

type mongoDBRoleInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewMongoDBRoleInformer constructs a new informer for MongoDBRole type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewMongoDBRoleInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredMongoDBRoleInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredMongoDBRoleInformer constructs a new informer for MongoDBRole type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredMongoDBRoleInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.EngineV1alpha1().MongoDBRoles(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.EngineV1alpha1().MongoDBRoles(namespace).Watch(options)
			},
		},
		&enginev1alpha1.MongoDBRole{},
		resyncPeriod,
		indexers,
	)
}

func (f *mongoDBRoleInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredMongoDBRoleInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *mongoDBRoleInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&enginev1alpha1.MongoDBRole{}, f.defaultInformer)
}

func (f *mongoDBRoleInformer) Lister() v1alpha1.MongoDBRoleLister {
	return v1alpha1.NewMongoDBRoleLister(f.Informer().GetIndexer())
}

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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	v1alpha1 "kubevault.dev/operator/apis/engine/v1alpha1"
)

// SecretEngineLister helps list SecretEngines.
type SecretEngineLister interface {
	// List lists all SecretEngines in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.SecretEngine, err error)
	// SecretEngines returns an object that can list and get SecretEngines.
	SecretEngines(namespace string) SecretEngineNamespaceLister
	SecretEngineListerExpansion
}

// secretEngineLister implements the SecretEngineLister interface.
type secretEngineLister struct {
	indexer cache.Indexer
}

// NewSecretEngineLister returns a new SecretEngineLister.
func NewSecretEngineLister(indexer cache.Indexer) SecretEngineLister {
	return &secretEngineLister{indexer: indexer}
}

// List lists all SecretEngines in the indexer.
func (s *secretEngineLister) List(selector labels.Selector) (ret []*v1alpha1.SecretEngine, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.SecretEngine))
	})
	return ret, err
}

// SecretEngines returns an object that can list and get SecretEngines.
func (s *secretEngineLister) SecretEngines(namespace string) SecretEngineNamespaceLister {
	return secretEngineNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// SecretEngineNamespaceLister helps list and get SecretEngines.
type SecretEngineNamespaceLister interface {
	// List lists all SecretEngines in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.SecretEngine, err error)
	// Get retrieves the SecretEngine from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.SecretEngine, error)
	SecretEngineNamespaceListerExpansion
}

// secretEngineNamespaceLister implements the SecretEngineNamespaceLister
// interface.
type secretEngineNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all SecretEngines in the indexer for a given namespace.
func (s secretEngineNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.SecretEngine, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.SecretEngine))
	})
	return ret, err
}

// Get retrieves the SecretEngine from the indexer for a given namespace and name.
func (s secretEngineNamespaceLister) Get(name string) (*v1alpha1.SecretEngine, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("secretengine"), name)
	}
	return obj.(*v1alpha1.SecretEngine), nil
}

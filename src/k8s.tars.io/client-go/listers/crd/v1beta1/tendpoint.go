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

package v1beta1

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	v1beta1 "k8s.tars.io/crd/v1beta1"
)

// TEndpointLister helps list TEndpoints.
type TEndpointLister interface {
	// List lists all TEndpoints in the indexer.
	List(selector labels.Selector) (ret []*v1beta1.TEndpoint, err error)
	// TEndpoints returns an object that can list and get TEndpoints.
	TEndpoints(namespace string) TEndpointNamespaceLister
	TEndpointListerExpansion
}

// tEndpointLister implements the TEndpointLister interface.
type tEndpointLister struct {
	indexer cache.Indexer
}

// NewTEndpointLister returns a new TEndpointLister.
func NewTEndpointLister(indexer cache.Indexer) TEndpointLister {
	return &tEndpointLister{indexer: indexer}
}

// List lists all TEndpoints in the indexer.
func (s *tEndpointLister) List(selector labels.Selector) (ret []*v1beta1.TEndpoint, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.TEndpoint))
	})
	return ret, err
}

// TEndpoints returns an object that can list and get TEndpoints.
func (s *tEndpointLister) TEndpoints(namespace string) TEndpointNamespaceLister {
	return tEndpointNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// TEndpointNamespaceLister helps list and get TEndpoints.
type TEndpointNamespaceLister interface {
	// List lists all TEndpoints in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1beta1.TEndpoint, err error)
	// Get retrieves the TEndpoint from the indexer for a given namespace and name.
	Get(name string) (*v1beta1.TEndpoint, error)
	TEndpointNamespaceListerExpansion
}

// tEndpointNamespaceLister implements the TEndpointNamespaceLister
// interface.
type tEndpointNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all TEndpoints in the indexer for a given namespace.
func (s tEndpointNamespaceLister) List(selector labels.Selector) (ret []*v1beta1.TEndpoint, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.TEndpoint))
	})
	return ret, err
}

// Get retrieves the TEndpoint from the indexer for a given namespace and name.
func (s tEndpointNamespaceLister) Get(name string) (*v1beta1.TEndpoint, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1beta1.Resource("tendpoint"), name)
	}
	return obj.(*v1beta1.TEndpoint), nil
}

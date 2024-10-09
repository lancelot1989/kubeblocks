/*
Copyright (C) 2022-2024 ApeCloud Co., Ltd

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

package v1

import (
	internalinterfaces "github.com/apecloud/kubeblocks/pkg/client/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// Clusters returns a ClusterInformer.
	Clusters() ClusterInformer
	// ClusterDefinitions returns a ClusterDefinitionInformer.
	ClusterDefinitions() ClusterDefinitionInformer
	// Components returns a ComponentInformer.
	Components() ComponentInformer
	// ComponentDefinitions returns a ComponentDefinitionInformer.
	ComponentDefinitions() ComponentDefinitionInformer
	// ComponentVersions returns a ComponentVersionInformer.
	ComponentVersions() ComponentVersionInformer
	// ServiceDescriptors returns a ServiceDescriptorInformer.
	ServiceDescriptors() ServiceDescriptorInformer
	// ShardingDefinitions returns a ShardingDefinitionInformer.
	ShardingDefinitions() ShardingDefinitionInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// Clusters returns a ClusterInformer.
func (v *version) Clusters() ClusterInformer {
	return &clusterInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// ClusterDefinitions returns a ClusterDefinitionInformer.
func (v *version) ClusterDefinitions() ClusterDefinitionInformer {
	return &clusterDefinitionInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// Components returns a ComponentInformer.
func (v *version) Components() ComponentInformer {
	return &componentInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// ComponentDefinitions returns a ComponentDefinitionInformer.
func (v *version) ComponentDefinitions() ComponentDefinitionInformer {
	return &componentDefinitionInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// ComponentVersions returns a ComponentVersionInformer.
func (v *version) ComponentVersions() ComponentVersionInformer {
	return &componentVersionInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// ServiceDescriptors returns a ServiceDescriptorInformer.
func (v *version) ServiceDescriptors() ServiceDescriptorInformer {
	return &serviceDescriptorInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// ShardingDefinitions returns a ShardingDefinitionInformer.
func (v *version) ShardingDefinitions() ShardingDefinitionInformer {
	return &shardingDefinitionInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

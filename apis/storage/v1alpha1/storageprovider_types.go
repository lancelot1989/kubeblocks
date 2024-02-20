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

package v1alpha1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StorageProviderSpec defines the desired state of `StorageProvider`.
type StorageProviderSpec struct {
	// Specifies the name of the CSI driver used to access remote storage.
	// This field can be empty, it indicates that the storage is not accessible via CSI.
	//
	// +optional
	CSIDriverName string `json:"csiDriverName,omitempty"`

	// A Go template that used to render and generate `k8s.io/api/core/v1.Secret`
	// resources for a specific CSI driver.
	// For example, `accessKey` and `secretKey` needed by CSI-S3 are stored in this
	// `Secret` resource.
	//
	// +optional
	CSIDriverSecretTemplate string `json:"csiDriverSecretTemplate,omitempty"`

	// A Go template utilized to render and generate `kubernetes.storage.k8s.io.v1.StorageClass`
	// resources. The `StorageClass' created by this template is aimed at using the CSI driver.
	//
	// +optional
	StorageClassTemplate string `json:"storageClassTemplate,omitempty"`

	// A Go template that renders and generates `k8s.io/api/core/v1.PersistentVolumeClaim`
	// resources. This PVC can reference the `StorageClass` created from `storageClassTemplate`,
	// allowing Pods to access remote storage by mounting the PVC.
	//
	// +optional
	PersistentVolumeClaimTemplate string `json:"persistentVolumeClaimTemplate,omitempty"`

	// A Go template used to render and generate `k8s.io/api/core/v1.Secret`.
	// This `Secret` involves the configuration details required by the `datasafed` tool
	// to access remote storage. For example, the `Secret` should contain `endpoint`,
	// `bucket`, 'region', 'accessKey', 'secretKey', or something else for S3 storage.
	// This field can be empty, it means this kind of storage is not accessible via
	// the `datasafed` tool.
	//
	// +optional
	DatasafedConfigTemplate string `json:"datasafedConfigTemplate,omitempty"`

	// Describes the parameters required for storage.
	// The parameters defined here can be referenced in the above templates,
	// and `kbcli` uses this definition for dynamic command-line parameter parsing.
	//
	// +optional
	ParametersSchema *ParametersSchema `json:"parametersSchema,omitempty"`
}

// ParametersSchema describes the parameters needed for a certain storage.
type ParametersSchema struct {
	// Defines the parameters in OpenAPI V3.
	//
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	// +k8s:conversion-gen=false
	// +optional
	OpenAPIV3Schema *apiextensionsv1.JSONSchemaProps `json:"openAPIV3Schema,omitempty"`

	// Defines which parameters are credential fields, which need to be handled specifically.
	// For instance, these should be stored in a `Secret` instead of a `ConfigMap`.
	//
	// +optional
	CredentialFields []string `json:"credentialFields,omitempty"`
}

// StorageProviderStatus defines the observed state of `StorageProvider`.
type StorageProviderStatus struct {
	// The phase of the `StorageProvider`. Valid phases are `NotReady` and `Ready`.
	//
	// +kubebuilder:validation:Enum={NotReady,Ready}
	Phase StorageProviderPhase `json:"phase,omitempty"`

	// Describes the current state of the `StorageProvider`.
	//
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories={kubeblocks},scope=Cluster
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="CSIDRIVER",type="string",JSONPath=".spec.csiDriverName"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

// StorageProvider comprises specifications that provide guidance on accessing remote storage.
// Currently the supported access methods are via a dedicated CSI driver or the `datasafed` tool.
// In case of CSI driver, the specification expounds on provisioning PVCs for that driver.
// As for the `datasafed` tool, the specification provides insights on generating the necessary
// configuration file.
type StorageProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StorageProviderSpec   `json:"spec,omitempty"`
	Status StorageProviderStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// StorageProviderList contains a list of `StorageProvider`.
type StorageProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StorageProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StorageProvider{}, &StorageProviderList{})
}

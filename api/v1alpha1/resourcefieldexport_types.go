/*
Copyright 2023.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ResourceRef struct {
	// APIVersion is the group version of the resource
	// +kubebuilder:validation:Pattern=^([a-zA-Z0-9.-]+[a-zA-Z0-9-]\/[a-zA-Z0-9]+|[a-zA-Z0-9]+)$
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}

// DestinationType is a ConfigMap or a Secret
// +kubebuilder:validation:Enum=ConfigMap;Secret
type DestinationType string

const (
	ConfigMap DestinationType = "ConfigMap"
	Secret    DestinationType = "Secret"
)

// DestinationRef is where the fields should be written.
type DestinationRef struct {
	Type DestinationType `json:"type"`
	Name string          `json:"name"`
}

type Output struct {
	Key  string `json:"key"`
	Path string `json:"path"`
}

type RequiredFields struct {
	// +kubebuilder:validation:Optional
	StatusConditions []StatusCondition `json:"statusConditions"`
}

type StatusCondition struct {
	Type   string `json:"type"`
	Status string `json:"status"`
}

// ResourceFieldExportSpec defines the desired state of ResourceFieldExport
type ResourceFieldExportSpec struct {
	From ResourceRef    `json:"from"`
	To   DestinationRef `json:"to"`

	// +kubebuilder:validation:Optional
	RequiredFields *RequiredFields `json:"requiredFields"`
	Outputs        []Output        `json:"outputs"`
}

type ConditionType string

type Condition struct {
	// Type is the type of the Condition
	Type ConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// The last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`
	// A human-readable message indicating details about the transition.
	// +optional
	Message *string `json:"message,omitempty"`
}

// ResourceFieldExportStatus defines the observed state of ResourceFieldExport
type ResourceFieldExportStatus struct {
	Conditions []Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ResourceFieldExport is the Schema for the resourcefieldexports API
type ResourceFieldExport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceFieldExportSpec   `json:"spec,omitempty"`
	Status ResourceFieldExportStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ResourceFieldExportList contains a list of ResourceFieldExport
type ResourceFieldExportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceFieldExport `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ResourceFieldExport{}, &ResourceFieldExportList{})
}

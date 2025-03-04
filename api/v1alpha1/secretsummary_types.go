/*
Copyright 2025.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecretSummaryStatus defines the observed state of SecretSummary
type SecretSummaryStatus struct {
	// The type of Kubernetes secret
	Type string `json:"type"`
	// List of keys present in the Secret's data field
	Keys []string `json:"keys"`
	// The number of keys present in the Secret's data field
	KeyCount int `json:"keyCount"`
	// When this object was most recently updated
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=".status.type",description="Type of the secret"
// +kubebuilder:printcolumn:name="Keys",type=integer,JSONPath=".status.keyCount",description="Number of keys"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=".metadata.creationTimestamp",description="Age of the resource"

// SecretSummary allows you to view the metadata of all Kubernetes secrets in a namespace without having access to the data field.
// This CR should only ever be managed by its controller and never manually.
type SecretSummary struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status SecretSummaryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecretSummaryList contains a list of SecretSummary
type SecretSummaryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretSummary `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecretSummary{}, &SecretSummaryList{})
}

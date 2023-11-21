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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AutoscaleConfigSpec defines the desired state of AutoscaleConfig
type AutoscaleConfigSpec struct {
	ThresholdValue resource.Quantity `json:"thresholdValue,omitempty"`
}

// AutoscaleConfigStatus defines the observed state of AutoscaleConfig
type AutoscaleConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AutoscaleConfig is the Schema for the autoscaleconfigs API
type AutoscaleConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AutoscaleConfigSpec   `json:"spec,omitempty"`
	Status AutoscaleConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AutoscaleConfigList contains a list of AutoscaleConfig
type AutoscaleConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AutoscaleConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AutoscaleConfig{}, &AutoscaleConfigList{})
}

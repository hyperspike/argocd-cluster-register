/*
Copyright 2022 Dan Molik.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GeneratorSpec defines the desired state of Generator
type GeneratorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Filters is simple key value map to target Cluster-API Clusters, by label matching
	Filters map[string]string `json:"filters,omitempty"`

	// AppProjectName is the name of the project to add the found clusters to, if no Appproject is given, then Clusters
	// will not be added to a project automatically.
	AppProjectName string `json:"appProjectName,omitempty"`

	// AppProjectNamespace is the namespace to target for the appproject addition. Defaults to; argocd.
	// +kubebuilder:default:="argocd"
	AppProjectNamespace string `json:"appProjectNamespace,omitempty"`
}

// GeneratorStatus defines the observed state of Generator
type GeneratorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Generator is the Schema for the generators API
type Generator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GeneratorSpec   `json:"spec,omitempty"`
	Status GeneratorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GeneratorList contains a list of Generator
type GeneratorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Generator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Generator{}, &GeneratorList{})
}

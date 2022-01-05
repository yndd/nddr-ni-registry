/*
Copyright 2021 NDD.

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
	"reflect"

	nddv1 "github.com/yndd/ndd-runtime/apis/common/v1"
	nddov1 "github.com/yndd/nddo-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Registry struct
type RegistryRegistry struct {
	// +kubebuilder:validation:Enum=`disable`;`enable`
	// +kubebuilder:default:="enable"
	AdminState *string `json:"admin-state,omitempty"`
	// +kubebuilder:validation:Enum=`hash`
	// +kubebuilder:default:="hash"
	AllocationStrategy *string `json:"allocation-strategy,omitempty"`
	// kubebuilder:validation:Minimum=1
	// kubebuilder:validation:Maximum=10000
	Size *uint32 `json:"size"`
	// kubebuilder:validation:MinLength=1
	// kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[A-Za-z0-9 !@#$^&()|+=`~.,'/_:;?-]*"
	Description *string `json:"description,omitempty"`
}

// A RegistrySpec defines the desired state of a Registry.
type RegistrySpec struct {
	nddov1.OdaInfo `json:",inline"`
	Registry       *RegistryRegistry `json:"registry,omitempty"`
}

// A RegistryStatus represents the observed state of a Registry.
type RegistryStatus struct {
	nddv1.ConditionedStatus `json:",inline"`
	nddov1.OdaInfo          `json:",inline"`
	RegistryName            *string               `json:"registry-name,omitempty"`
	Registry                *NddrRegistryRegistry `json:"registry,omitempty"`
}

// +kubebuilder:object:root=true

// Registry is the Schema for the Registry API
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SYNC",type="string",JSONPath=".status.conditions[?(@.kind=='Synced')].status"
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.conditions[?(@.kind=='Ready')].status"
// +kubebuilder:printcolumn:name="ORG",type="string",JSONPath=".status.oda[?(@.key=='organization')].value"
// +kubebuilder:printcolumn:name="DEP",type="string",JSONPath=".status.oda[?(@.key=='deployment')].value"
// +kubebuilder:printcolumn:name="AZ",type="string",JSONPath=".status.oda[?(@.key=='availability-zone')].value"
// +kubebuilder:printcolumn:name="REGISTRY",type="string",JSONPath=".status.registry-name"
// +kubebuilder:printcolumn:name="ALLOCATED",type="string",JSONPath=".status.registry.state.allocated",description="allocated network-instances"
// +kubebuilder:printcolumn:name="AVAILABLE",type="string",JSONPath=".status.registry.state.available",description="available network-instances"
// +kubebuilder:printcolumn:name="TOTAL",type="string",JSONPath=".status.registry.state.total",description="total network-instances"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
type Registry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RegistrySpec   `json:"spec,omitempty"`
	Status RegistryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RegistryList contains a list of Registrys
type RegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Registry `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Registry{}, &RegistryList{})
}

// Registry type metadata.
var (
	RegistryKindKind         = reflect.TypeOf(Registry{}).Name()
	RegistryGroupKind        = schema.GroupKind{Group: Group, Kind: RegistryKindKind}.String()
	RegistryKindAPIVersion   = RegistryKindKind + "." + GroupVersion.String()
	RegistryGroupVersionKind = GroupVersion.WithKind(RegistryKindKind)
)

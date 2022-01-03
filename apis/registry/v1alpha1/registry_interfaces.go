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
	"github.com/yndd/ndd-runtime/pkg/utils"
	"github.com/yndd/nddo-runtime/pkg/odr"
	"github.com/yndd/nddo-runtime/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ RgList = &RegistryList{}

// +k8s:deepcopy-gen=false
type RgList interface {
	client.ObjectList

	GetRegistries() []Rg
}

func (x *RegistryList) GetRegistries() []Rg {
	xs := make([]Rg, len(x.Items))
	for i, r := range x.Items {
		r := r // Pin range variable so we can take its address.
		xs[i] = &r
	}
	return xs
}

var _ Rg = &Registry{}

// +k8s:deepcopy-gen=false
type Rg interface {
	resource.Object
	resource.Conditioned

	GetOrganizationName() string
	GetRegistryName() string
	GetAllocationStrategy() string
	GetSize() uint32
	GetAllocations() uint32
	GetAllocatedNis() []*string
	InitializeResource() error
	SetStatus(uint32, []*string)
	SetOrganizationName(s string)
	SetRegistryName(s string)
}

// GetCondition of this Network Node.
func (x *Registry) GetCondition(ct nddv1.ConditionKind) nddv1.Condition {
	return x.Status.GetCondition(ct)
}

// SetConditions of the Network Node.
func (x *Registry) SetConditions(c ...nddv1.Condition) {
	x.Status.SetConditions(c...)
}

func (x *Registry) GetOrganizationName() string {
	odr, err := odr.GetOdrRegistryInfo(x.GetName())
	if err != nil {
		return ""
	}
	return odr.OrganizationName
}

func (x *Registry) GetRegistryName() string {
	odr, err := odr.GetOdrRegistryInfo(x.GetName())
	if err != nil {
		return ""
	}
	return odr.RegistryName
}

func (n *Registry) GetAllocationStrategy() string {
	if reflect.ValueOf(n.Spec.Registry.AllocationStrategy).IsZero() {
		return ""
	}
	return *n.Spec.Registry.AllocationStrategy
}

func (n *Registry) GetSize() uint32 {
	if reflect.ValueOf(n.Spec.Registry.Size).IsZero() {
		return 0
	}
	return *n.Spec.Registry.Size
}

func (n *Registry) GetAllocations() uint32 {
	if n.Status.Registry != nil && n.Status.Registry.State != nil {
		return *n.Status.Registry.State.Allocated
	}
	return 0
}

func (n *Registry) GetAllocatedNis() []*string {
	return n.Status.Registry.State.Used
}

func (n *Registry) InitializeResource() error {

	// check if the pool was already initialized
	if n.Status.Registry != nil && n.Status.Registry.State != nil {
		// pool was already initialiazed
		return nil
	}
	size := int(*n.Spec.Registry.Size)

	n.Status.Registry = &NddrRegistryRegistry{
		Size:        n.Spec.Registry.Size,
		AdminState:  n.Spec.Registry.AdminState,
		Description: n.Spec.Registry.Description,
		State: &NddrRegistryRegistryState{
			Total:     utils.Uint32Ptr(uint32(size)),
			Allocated: utils.Uint32Ptr(0),
			Available: utils.Uint32Ptr(uint32(size)),
			Used:      make([]*string, 0),
		},
	}
	return nil

}

func (n *Registry) SetStatus(allocated uint32, used []*string) {
	n.Status.Registry.State.Allocated = utils.Uint32Ptr(allocated)
	n.Status.Registry.State.Available = utils.Uint32Ptr(*n.Spec.Registry.Size - allocated)

	n.Status.Registry.State.Used = used
}

func (x *Registry) SetOrganizationName(s string) {
	x.Status.OrganizationName = &s
}

func (x *Registry) SetRegistryName(s string) {
	x.Status.RegistryName = &s
}

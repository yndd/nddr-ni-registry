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

	GetCondition(ct nddv1.ConditionKind) nddv1.Condition
	SetConditions(c ...nddv1.Condition)
	GetOrganization() string
	GetDeployment() string
	GetAvailabilityZone() string
	GetRegistryName() string
	GetAllocationStrategy() string
	GetSize() uint32
	GetAllocations() uint32
	GetAllocatedNis() []*string
	InitializeResource() error
	SetStatus(uint32, []*string)
	SetOrganization(string)
	SetDeployment(string)
	SetAvailabilityZone(s string)
	SetRegistryName(string)
}

// GetCondition of this Network Node.
func (x *Registry) GetCondition(ct nddv1.ConditionKind) nddv1.Condition {
	return x.Status.GetCondition(ct)
}

// SetConditions of the Network Node.
func (x *Registry) SetConditions(c ...nddv1.Condition) {
	x.Status.SetConditions(c...)
}

func (x *Registry) GetOrganization() string {
	return x.Spec.GetOrganization()
}

func (x *Registry) GetDeployment() string {
	return x.Spec.GetDeployment()
}

func (x *Registry) GetAvailabilityZone() string {
	return x.Spec.GetAvailabilityZone()
}

func (x *Registry) GetRegistryName() string {
	return x.GetName()
}

func (x *Registry) GetAllocationStrategy() string {
	if reflect.ValueOf(x.Spec.Registry.AllocationStrategy).IsZero() {
		return ""
	}
	return *x.Spec.Registry.AllocationStrategy
}

func (x *Registry) GetSize() uint32 {
	if reflect.ValueOf(x.Spec.Registry.Size).IsZero() {
		return 0
	}
	return *x.Spec.Registry.Size
}

func (x *Registry) GetAllocations() uint32 {
	if x.Status.Registry != nil && x.Status.Registry.State != nil {
		return *x.Status.Registry.State.Allocated
	}
	return 0
}

func (x *Registry) GetAllocatedNis() []*string {
	return x.Status.Registry.State.Used
}

func (x *Registry) InitializeResource() error {

	// check if the pool was already initialized
	if x.Status.Registry != nil && x.Status.Registry.State != nil {
		// pool was already initialiazed
		return nil
	}
	size := int(*x.Spec.Registry.Size)

	x.Status.Registry = &NddrRegistryRegistry{
		Size:        x.Spec.Registry.Size,
		AdminState:  x.Spec.Registry.AdminState,
		Description: x.Spec.Registry.Description,
		State: &NddrRegistryRegistryState{
			Total:     utils.Uint32Ptr(uint32(size)),
			Allocated: utils.Uint32Ptr(0),
			Available: utils.Uint32Ptr(uint32(size)),
			Used:      make([]*string, 0),
		},
	}
	return nil

}

func (x *Registry) SetStatus(allocated uint32, used []*string) {
	x.Status.Registry.State.Allocated = utils.Uint32Ptr(allocated)
	x.Status.Registry.State.Available = utils.Uint32Ptr(*x.Spec.Registry.Size - allocated)

	x.Status.Registry.State.Used = used
}

func (x *Registry) SetOrganization(s string) {
	x.Status.SetOrganization(s)
}

func (x *Registry) SetDeployment(s string) {
	x.Status.SetDeployment(s)
}

func (x *Registry) SetAvailabilityZone(s string) {
	x.Status.SetAvailabilityZone(s)
}

func (x *Registry) SetRegistryName(s string) {
	x.Status.RegistryName = &s
}

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
	"fmt"
	"reflect"
	"strings"

	nddv1 "github.com/yndd/ndd-runtime/apis/common/v1"
	"github.com/yndd/nddo-runtime/pkg/odr"
	"github.com/yndd/nddo-runtime/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ RrList = &RegisterList{}

// +k8s:deepcopy-gen=false
type RrList interface {
	client.ObjectList

	GetRegisters() []Rr
}

func (x *RegisterList) GetRegisters() []Rr {
	registers := make([]Rr, len(x.Items))
	for i, r := range x.Items {
		r := r // Pin range variable so we can take its address.
		registers[i] = &r
	}
	return registers
}

var _ Rr = &Register{}

// +k8s:deepcopy-gen=false
type Rr interface {
	resource.Object
	resource.Conditioned

	GetCondition(ct nddv1.ConditionKind) nddv1.Condition
	SetConditions(c ...nddv1.Condition)
	GetOrganizationName() string
	GetDeploymentName() string
	GetRegistryName() string
	GetSourceTag() map[string]string
	GetSelector() map[string]string
	SetNi(uint32)
	HasNi() (uint32, bool)
	SetOrganizationName(s string)
	SetDeploymentName(s string)
	SetRegistryName(s string)
}

// GetCondition of this Network Node.
func (x *Register) GetCondition(ct nddv1.ConditionKind) nddv1.Condition {
	return x.Status.GetCondition(ct)
}

// SetConditions of the Network Node.
func (x *Register) SetConditions(c ...nddv1.Condition) {
	x.Status.SetConditions(c...)
}

func (x *Register) GetOrganizationName() string {
	return odr.GetOrganizationName(x.GetNamespace())
}

func (x *Register) GetDeploymentName() string {
	return odr.GetDeploymentName(x.GetNamespace())
}

func (x *Register) GetRegistryName() string {
	split := strings.Split(x.GetName(), ".")
	if len(split) > 1 {
		return split[0]
	}
	return ""
}

func (n *Register) GetSourceTag() map[string]string {
	s := make(map[string]string)
	if reflect.ValueOf(n.Spec.Register.SourceTag).IsZero() {
		return s
	}
	for _, tag := range n.Spec.Register.SourceTag {
		s[*tag.Key] = *tag.Value
	}
	return s
}

func (n *Register) GetSelector() map[string]string {
	s := make(map[string]string)
	if reflect.ValueOf(n.Spec.Register.Selector).IsZero() {
		return s
	}
	for _, tag := range n.Spec.Register.Selector {
		s[*tag.Key] = *tag.Value
	}
	return s
}

func (n *Register) SetNi(idx uint32) {
	n.Status = RegisterStatus{
		Register: &NddrNiRegister{
			State: &NddrRegisterState{
				Index: &idx,
			},
		},
	}
}

func (n *Register) HasNi() (uint32, bool) {
	fmt.Printf("HasNi: %#v\n", n.Status.Register)
	if n.Status.Register != nil && n.Status.Register.State != nil && n.Status.Register.State.Index != nil {
		return *n.Status.Register.State.Index, true
	}
	return 99999999, false

}

func (x *Register) SetOrganizationName(s string) {
	x.Status.OrganizationName = &s
}

func (x *Register) SetDeploymentName(s string) {
	x.Status.DeploymentName = &s
}

func (x *Register) SetRegistryName(s string) {
	x.Status.RegistryName = &s
}

/*
Copyright 2022.

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

type ClusterPhase string

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MemcachedSpec defines the desired state of Memcached
type MemcachedSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Memcached. Edit memcached_types.go to remove/update
	// Foo string `json:"foo,omitempty"`
	// Size int32 `json:"size"`
	Experiments []ExperimentSpec `json:"experiments"`
}

type ExperimentSpec struct {
	// Scope is the area of the experiments, currently support node, pod and container
	Scope string `json:"scope"`
	// Target is the experiment target, such as cpu, network
	Target string `json:"target"`
	// Action is the experiment scenario of the target, such as delay, load
	Action string `json:"action"`
	// Desc is the experiment description
	Desc string `json:"desc,omitempty"`
	// Matchers is the experiment rules
	Matchers []FlagSpec `json:"matchers,omitempty"`
}

type FlagSpec struct {
	// Name is the name of flag
	Name string `json:"name"`
	// TODO: Temporarily defined as an array for all flags
	// Value is the value of flag
	Value []string `json:"value"`
}

// MemcachedStatus defines the observed state of Memcached
type MemcachedStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Nodes []string `json:"nodes"`

	// Phase indicates the state of the experiment
	//   Initial -> Running -> Updating -> Destroying -> Destroyed
	Phase ClusterPhase `json:"phase,omitempty"`

	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	ExpStatuses []ExperimentStatus `json:"expStatuses"`
}

type ExperimentStatus struct {
	// experiment scope for cache
	Scope  string `json:"scope"`
	Target string `json:"target"`
	Action string `json:"action"`
	// Success is used to judge the experiment result
	Success bool `json:"success"`
	// State is used to describe the experiment result
	State string `json:"state"`
	Error string `json:"error,omitempty"`
	// ResStatuses is the details of the experiment
	ResStatuses []ResourceStatus `json:"resStatuses,omitempty"`
}

type ResourceStatus struct {
	// experiment uid in chaosblade
	Id string `json:"id,omitempty"`
	// experiment state
	State string `json:"state"`
	// experiment code
	Code int32 `json:"code,omitempty"`
	// experiment error
	Error string `json:"error,omitempty"`
	// success
	Success bool `json:"success"`

	// Kind
	Kind string `json:"kind"`
	// Resource identifier, rules as following:
	// container: Namespace/NodeName/PodName/ContainerName
	// pod： Namespace/NodeName/PodName
	Identifier string `json:"identifier,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Memcached is the Schema for the memcacheds API
type Memcached struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MemcachedSpec   `json:"spec,omitempty"`
	Status MemcachedStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MemcachedList contains a list of Memcached
type MemcachedList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Memcached `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Memcached{}, &MemcachedList{})
}

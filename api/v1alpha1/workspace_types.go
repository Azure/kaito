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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	PresetLlama2AModel ModelName = "llama-2-7b"
	PresetLlama2BModel ModelName = "llama-2-13b"
	PresetLlama2CModel ModelName = "llama-2-70b"
	PresetLlama2AChat            = PresetLlama2AModel + "-chat"
	PresetLlama2BChat            = PresetLlama2BModel + "-chat"
	PresetLlama2CChat            = PresetLlama2CModel + "-chat"

	ModelAccessModePublic  ModelAccessMode = "public"
	ModelAccessModePrivate ModelAccessMode = "private"
)

// ResourceSpec desicribes the resource requirement of running the workload.
// If the number of nodes in the cluster that meet the InstanceType and
// LabelSelector requirements is small than the Count, controller
// will provision new nodes before deploying the workload.
// The final list of nodes used to run the workload is presented in workspace Status.
type ResourceSpec struct {
	// Count is the required number of GPU nodes.
	// +optional
	// +kubebuilder:default:=1
	Count *int `json:"count,omitempty"`

	// InstanceType specifies the GPU node SKU.
	// This field defaults to "Standard_NC12s_v3" if not specified.
	// +optional
	// +kubebuilder:default:="Standard_NC12s_v3"
	InstanceType string `json:"instanceType,omitempty"`

	// LabelSelector specifies the required labels for the GPU nodes.
	LabelSelector *metav1.LabelSelector `json:"labelSelector"`

	// PreferredNodes is an optional node list specified by the user.
	// If a node in the list does not have the required labels or
	// the required instanceType, it will be ignored.
	// +optional
	PreferredNodes []string `json:"preferredNodes,omitempty"`
}

type ModelName string

// +kubebuilder:validation:Enum=public;private
type ModelAccessMode string

type PresetMeta struct {
	// Name of the supported models with preset configurations.
	Name ModelName `json:"name"`
	// AccessMode specifies whether the containerized model image is accessible via public registry
	// or private registry. This field defaults to "public" if not specified.
	// If this field is "private", user needs to provide the private image information in PresetOptions.
	// +bebuilder:default:="public"
	// +optional
	AccessMode ModelAccessMode `json:"accessMode,omitempty"`
}

type PresetOptions struct {
	// Image is the name of the containerized model image.
	// +optional
	Image string `json:"image,omitempty"`
	// ImagePullSecrets is a list of secret names in the same namespace used for pulling the image.
	// +optional
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`
}

// PresetSpec provides the information for rendering preset configurations to run the model inference service.
type PresetSpec struct {
	PresetMeta `json:",inline"`
	// +optional
	PresetOptions `json:"presetOptions,omitempty"`
}

type InferenceSpec struct {
	// Preset describles the model that will be deployed with preset configurations.
	// +optional
	Preset PresetSpec `json:"preset,omitempty"`
	// Template specifies the Pod template used to run the inference service. Users can specify custom Pod settings
	// if the preset configurations cannot meet the requirements. Note that if Preset is specified, Template should not
	// be specified and vice versa.
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	// +optional
	Template v1.PodTemplateSpec `json:"template,omitempty"`
}

// WorkspaceStatus defines the observed state of Workspace
type WorkspaceStatus struct {
	// WorkerNodes is the list of nodes chosen to run the workload based on the workspace resource requirement.
	// +optional
	WorkerNodes []string `json:"workerNodes,omitempty"`

	// Conditions report the current conditions of the workspace.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// Workspace is the Schema for the workspaces API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=workspaces,scope=Namespaced,categories=workspace,shortName={wk,wks}
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Instance",type="string",JSONPath=".resource.instanceType",description=""
// +kubebuilder:printcolumn:name="ResourceReady",type="string",JSONPath=".status.condition[?(@.type==\"ResourceReady\")].status",description=""
// +kubebuilder:printcolumn:name="InferenceReady",type="string",JSONPath=".status.condition[?(@.type==\"InferenceReady\")].status",description=""
// +kubebuilder:printcolumn:name="WorkspaceReady",type="string",JSONPath=".status.condition[?(@.type==\"WorkspaceReady\")].status",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
type Workspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Resource  ResourceSpec    `json:"resource,omitempty"`
	Inference InferenceSpec   `json:"inference,omitempty"`
	Status    WorkspaceStatus `json:"status,omitempty"`
}

// WorkspaceList contains a list of Workspace
// +kubebuilder:object:root=true
type WorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workspace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Workspace{}, &WorkspaceList{})
}

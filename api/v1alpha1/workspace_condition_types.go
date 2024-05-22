// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package v1alpha1

// ConditionType is a valid value for Condition.Type.
type ConditionType string

const (
	// WorkspaceConditionTypeMachineStatus is the state when checking machine status.
	WorkspaceConditionTypeMachineStatus = ConditionType("MachineReady")

	// WorkspaceConditionTypeNodeClaimStatus is the state when checking nodeClaim status.
	WorkspaceConditionTypeNodeClaimStatus = ConditionType("NodeClaimReady")

	// WorkspaceConditionTypeResourceStatus is the state when Resource has been created.
	WorkspaceConditionTypeResourceStatus = ConditionType("ResourceReady")

	// WorkspaceConditionTypeInferenceStatus is the state when Inference has been created.
	WorkspaceConditionTypeInferenceStatus = ConditionType("InferenceReady")

	//WorkspaceConditionTypeDeleting is the Workspace state when starts to get deleted.
	WorkspaceConditionTypeDeleting = ConditionType("WorkspaceDeleting")

	//WorkspaceConditionTypeReady is the Workspace state that summarize all operations' state.
	WorkspaceConditionTypeReady ConditionType = ConditionType("WorkspaceReady")
)

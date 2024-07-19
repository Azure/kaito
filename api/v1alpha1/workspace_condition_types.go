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

	// WorkspaceConditionTypeInferenceStatus is the state when Inference service has been ready.
	WorkspaceConditionTypeInferenceStatus = ConditionType("InferenceReady")

	// WorkspaceConditionTypeTuningJobStatus is the state when the tuning job starts normally.
	WorkspaceConditionTypeTuningJobStatus ConditionType = ConditionType("JobStarted")

	//WorkspaceConditionTypeDeleting is the Workspace state when starts to get deleted.
	WorkspaceConditionTypeDeleting = ConditionType("WorkspaceDeleting")

	//WorkspaceConditionTypeSucceeded is the Workspace state that summarizes all operations' states.
	//For inference, the "True" condition means the inference service is ready to serve requests.
	//For fine tuning, the "True" condition means the tuning job completes successfully.
	WorkspaceConditionTypeSucceeded ConditionType = ConditionType("WorkspaceSucceeded")
)

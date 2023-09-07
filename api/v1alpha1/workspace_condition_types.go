package v1alpha1

// ConditionType is a valid value for Condition.Type.
type ConditionType string

const (
	// WorkspaceConditionTypeMachineStatus is the state when checking machine status.
	WorkspaceConditionTypeMachineStatus = ConditionType("MachineReady")

	// WorkspaceConditionTypeMachineProvisioned is the state when Machine has been created.
	WorkspaceConditionTypeMachineProvisioned = ConditionType("MachineProvisioned")

	// WorkspaceConditionTypeMachineDeleted is the state when Machine has been deleted.
	WorkspaceConditionTypeMachineDeleted = ConditionType("MachineDeleted")

	// WorkspaceConditionTypeResourceStatus is the state when Resource has been created.
	WorkspaceConditionTypeResourceStatus = ConditionType("ResourceStatus")

	// WorkspaceConditionTypeResourceDeleted is the state when Resource has been deleted.
	WorkspaceConditionTypeResourceDeleted = ConditionType("ResourceDeleted")

	// WorkspaceConditionTypeInferenceStatus is the state when Inference has been created.
	WorkspaceConditionTypeInferenceStatus = ConditionType("InferenceStatus")

	// WorkspaceConditionTypeInferenceDeleted is the state when Inference has been deleted.
	WorkspaceConditionTypeInferenceDeleted = ConditionType("InferenceDeleted")

	//WorkspaceConditionTypeDeleting is the Workspace state when starts to get deleted.
	WorkspaceConditionTypeDeleting = ConditionType("WorkspaceDeleting")

	//WorkspaceConditionTypeReady is the Workspace state that summarize all operations' state.
	WorkspaceConditionTypeReady ConditionType = ConditionType("WorkspaceReady")
)

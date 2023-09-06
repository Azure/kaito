package v1alpha1

// ConditionType is a valid value for Condition.Type.
type ConditionType string

const (
	//WorkspaceConditionTypeDeleting is the Workspace state when starts to get deleted.
	WorkspaceConditionTypeDeleting = ConditionType("WorkspaceDeleting")

	//WorkspaceConditionTypeReady is the Workspace state that summarize all operations' state.
	WorkspaceConditionTypeReady ConditionType = ConditionType("WorkspaceReady")

	// WorkspaceConditionTypeResourceProvisioned is the state when Resources have been created.
	WorkspaceConditionTypeResourceProvisioned = ConditionType("ResourceProvisioned")

	// WorkspaceConditionTypeMachineStatus is the state when checking machine status.
	WorkspaceConditionTypeMachineStatus = ConditionType("MachineReady")

	// WorkspaceConditionTypeResourceDeleted is the state when Resources have been deleted.
	WorkspaceConditionTypeResourceDeleted = ConditionType("ResourceDeleted")

	// WorkspaceConditionTypeMachineProvisioned is the state when Machine has been created.
	WorkspaceConditionTypeMachineProvisioned = ConditionType("MachineProvisioned")

	// WorkspaceConditionTypeMachineDeleted is the state when Machine have been deleted.
	WorkspaceConditionTypeMachineDeleted = ConditionType("MachineDeleted")
)

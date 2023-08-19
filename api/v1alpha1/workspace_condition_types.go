package v1alpha1

// ConditionType is a valid value for Condition.Type.
type ConditionType string

const (
	// WorkspaceConditionTypePending is the first state after being created.
	WorkspaceConditionTypePending = ConditionType("WorkspacePending")

	//WorkspaceConditionTypeDeleting is the Workspace state when starts to get deleted.
	WorkspaceConditionTypeDeleting = ConditionType("WorkspaceDeleting")

	//WorkspaceConditionTypeReady is the Workspace state that summarize all operations' state.
	WorkspaceConditionTypeReady ConditionType = ConditionType("WorkspaceReady")

	// WorkspaceConditionTypeFailed is the Workspace state when Operation failed after reconciliation.
	WorkspaceConditionTypeFailed = ConditionType("WorkspaceFailed")

	// WorkspaceConditionTypeResourceProvisioning is the state when the Workspace starts provisioning Resources.
	WorkspaceConditionTypeResourceProvisioning = ConditionType("ResourceProvisioning")

	// WorkspaceConditionTypeResourceProvisioned is the state when Resources have been created.
	WorkspaceConditionTypeResourceProvisioned = ConditionType("ResourceProvisioned")

	// WorkspaceConditionTypeConfiguring is the state when Workspace starts Configuring Resources.
	WorkspaceConditionTypeConfiguring = ConditionType("ResourceConfiguring")

	// WorkspaceConditionTypeConfigured is the state when Resources have been Configured.
	WorkspaceConditionTypeConfigured = ConditionType("ResourceConfigured")

	// WorkspaceConditionTypeResourceDeleting is the state when the Workspace starts deleting Resources.
	WorkspaceConditionTypeResourceDeleting = ConditionType("ResourceDeleting")

	// WorkspaceConditionTypeResourceDeleted is the state when Resources have been deleted.
	WorkspaceConditionTypeResourceDeleted = ConditionType("ResourceDeleted")

	// WorkspaceConditionTypeWorkloadProvisioning is the state when the Workspace starts provisioning Workload.
	WorkspaceConditionTypeWorkloadProvisioning = ConditionType("WorkloadProvisioning")

	// WorkspaceConditionTypeWorkloadProvisioned is the state when Workload has been created.
	WorkspaceConditionTypeWorkloadProvisioned = ConditionType("WorkloadProvisioned")

	// WorkspaceConditionTypeWorkloadDeleting is the state when the Workspace starts deleting Resources.
	WorkspaceConditionTypeWorkloadDeleting = ConditionType("WorkloadDeleting")

	// WorkspaceConditionTypeWorkloadDeleted is the state when Workload have been deleted.
	WorkspaceConditionTypeWorkloadDeleted = ConditionType("WorkloadDeleted")
)

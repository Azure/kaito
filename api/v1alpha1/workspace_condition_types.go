package v1alpha1

import v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

// ConditionType is a valid value for Condition.Type.
type ConditionType string

// Condition contains details for one aspect of the current state of this API Resource.
// ---
// This struct is intended for direct use as an array at the field path .status.conditions.  For example,
//
//	type FooStatus struct{
//	    // Represents the observations of a foo's current state.
//	    // Known .status.conditions.type are: "Available", "Progressing", and "Degraded"
//	    // +patchMergeKey=type
//	    // +patchStrategy=merge
//	    // +listType=map
//	    // +listMapKey=type
//	    Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
//
//	    // other fields
//	}
type Condition struct {
	// type of condition in CamelCase or in foo.example.com/CamelCase.
	// ---
	// Many .condition.type values are consistent across resources like Available, but because arbitrary conditions can be
	// useful (see .node.status.conditions), the ability to deconflict is important.
	// The regex it matches is (dns1123SubdomainFmt/)?(qualifiedNameFmt)
	// +required
	// +kubebuilder:validation:Enum=True;
	Type ConditionType `json:"type"`
	// status of the condition, one of True, False, Unknown.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=True;False;Unknown
	Status v1.ConditionStatus `json:"status"`
	// observedGeneration represents the .metadata.generation that the condition was set based upon.
	// For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
	// with respect to the current state of the instance.
	// +optional
	// +kubebuilder:validation:Minimum=0
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// lastTransitionTime is the last time the condition transitioned from one status to another.
	// This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=date-time
	LastTransitionTime v1.Time `json:"lastTransitionTime"`
	// reason contains a programmatic identifier indicating the reason for the condition's last transition.
	// Producers of specific condition types may define expected values and meanings for this field,
	// and whether the values are considered a guaranteed API.
	// The value should be a CamelCase string.
	// This field may not be empty.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=`^[A-Za-z]([A-Za-z0-9_,:]*[A-Za-z0-9_])?$`
	Reason string `json:"reason"`
	// message is a human-readable message indicating details about the transition.
	// This may be an empty string.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=32768
	Message string `json:"message"`
}

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package v1alpha1

const (

	// Non-prefixed labels/annotations are reserved for end-use.

	// KAITOPrefix Kubernetes Data Mining prefix.
	KAITOPrefix = "kaito.sh/"

	// AnnotationEnableLB determines whether kaito creates LoadBalancer type service for testing.
	AnnotationEnableLB = KAITOPrefix + "enablelb"

	// LabelWorkspaceName is the label for workspace name.
	LabelWorkspaceName = KAITOPrefix + "workspace"

	// LabelRAGEngineName is the label for ragengine name.
	LabelRAGEngineName = KAITOPrefix + "ragengine"

	// LabelWorkspaceName is the label for workspace namespace.
	LabelWorkspaceNamespace = KAITOPrefix + "workspacenamespace"

	// LabelRAGEngineNamespace is the label for ragengine namespace.
	LabelRAGEngineNamespace = KAITOPrefix + "ragenginenamespace"

	// WorkspaceRevisionAnnotation is the Annotations for revision number
	WorkspaceRevisionAnnotation = "workspace.kaito.io/revision"
)

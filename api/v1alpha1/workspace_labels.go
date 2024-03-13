// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package v1alpha1

const (

	// Non-prefixed labels/annotations are reserved for end-use.

	// KAITOPrefix Kubernetes Data Mining prefix.
	KAITOPrefix = "kaito.sh/"

	// AnnotationEnableLB determines whether kaito creates LoadBalancer type service for testing.
	AnnotationEnableLB = KAITOPrefix + "enablelb"

	// AnnotationEnableSampleFrontEnd determines whether kaito creates a sample front end using OSS Chainlit
	AnnotationEnableSampleFrontEnd = KAITOPrefix + "enable-sample-frontend"

	// LabelWorkspaceName is the label for workspace name.
	LabelWorkspaceName = KAITOPrefix + "workspace"

	// LabelWorkspaceName is the label for workspace namespace.
	LabelWorkspaceNamespace = KAITOPrefix + "workspacenamespace"
)

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package consts

const (
	// WorkspaceFinalizer is used to make sure that workspace controller handles garbage collection.
	WorkspaceFinalizer            = "workspace.finalizer.kaito.sh"
	DefaultReleaseNamespaceEnvVar = "RELEASE_NAMESPACE"
	FeatureFlagKarpenter          = "Karpenter"

	//	RequiredKubernetesVersionForNodeClaim is the latest major version of Kubernetes that Kaito supports.
	RequiredKubernetesVersionForNodeClaim = "29"
)

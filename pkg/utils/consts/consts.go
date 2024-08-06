// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package consts

const (
	// WorkspaceFinalizer is used to make sure that workspace controller handles garbage collection.
	WorkspaceFinalizer            = "workspace.finalizer.kaito.sh"
	DefaultReleaseNamespaceEnvVar = "RELEASE_NAMESPACE"
	FeatureFlagKarpenter          = "Karpenter"
	AzureCloudName                = "azure"
	AWSCloudName                  = "aws"
	GPUString                     = "gpu"
	SKUString                     = "sku"
	MaxRevisionHistoryLimit       = 10
	GiBToBytes                    = 1024 * 1024 * 1024 // Conversion factor from GiB to bytes
	NvidiaGPU                     = "nvidia.com/gpu"
)

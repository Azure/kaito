// Copyright (c) KAITO authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package consts

// Node provisioning failure reasons on the Workspace NodeClaimReady and ResourceReady conditions that need the user to
// act, such as raising a quota or fixing the node configuration. The message starts with the reason reported by the
// node provisioner, so these stay stable across provisioner versions.
const (
	// ReasonQuotaExceeded reports an exhausted subscription vCPU quota (VM SKU family, regional or spot).
	ReasonQuotaExceeded = "QuotaExceeded"
	// ReasonCapacityUnavailable reports that the cloud cannot allocate the VM size, or does not offer it.
	ReasonCapacityUnavailable = "CapacityUnavailable"
	// ReasonNodeImageInvalid reports a node image that does not exist or is not a valid image reference.
	ReasonNodeImageInvalid = "NodeImageInvalid"
	// ReasonNodeBootstrapFailed reports a node that failed its bootstrap script or never registered with the cluster.
	ReasonNodeBootstrapFailed = "NodeBootstrapFailed"
	// ReasonNodeConfigurationInvalid reports a permission or subnet configuration that prevents creating the node.
	ReasonNodeConfigurationInvalid = "NodeConfigurationInvalid"
	// ReasonNodeClassInvalid reports a NodeClass that is missing or Ready=False, which blocks every launch. It is read
	// on every reconcile, so it is reported only while it lasts.
	ReasonNodeClassInvalid = "NodeClassInvalid"
	// ReasonNodePoolInvalid reports a NodePool rejected by Karpenter's validation, which blocks every launch. It is read
	// on every reconcile, so it is reported only while it lasts.
	ReasonNodePoolInvalid = "NodePoolInvalid"
)

// Inference failure reasons on the Workspace InferenceReady condition that need the user to act, such as changing the
// instance type, the model image or the model configuration.
const (
	// ReasonContainerOOMKilled reports an inference container killed for exceeding its memory.
	ReasonContainerOOMKilled = "ContainerOOMKilled"
	// ReasonImagePullError reports an inference image that cannot be pulled.
	ReasonImagePullError = "ImagePullError"
	// ReasonContainerStartError reports an inference container that cannot be created or started, e.g. a missing Secret.
	ReasonContainerStartError = "ContainerStartError"
	// ReasonContainerCrashLoopBackOff reports an inference container that keeps crashing.
	ReasonContainerCrashLoopBackOff = "ContainerCrashLoopBackOff"
	// ReasonSASTokenFetchFailed reports a model streaming init container that cannot obtain a SAS token.
	ReasonSASTokenFetchFailed = "SASTokenFetchFailed"
	// ReasonModelBundleMismatch reports a streamed model bundle that does not match the deployment's configuration.
	ReasonModelBundleMismatch = "ModelBundleMismatch"
	// ReasonModelStreamingStorageClassNotFound reports a missing model streaming StorageClass.
	ReasonModelStreamingStorageClassNotFound = "ModelStreamingStorageClassNotFound"
	// ReasonModelStreamingInvalidStorageClass reports a model streaming StorageClass that cannot be used.
	ReasonModelStreamingInvalidStorageClass = "ModelStreamingInvalidStorageClass"
	// ReasonModelStreamingServiceAccountInvalid reports an invalid model streaming ServiceAccount.
	ReasonModelStreamingServiceAccountInvalid = "ModelStreamingServiceAccountInvalid"
	// ReasonModelStreamingInvalidAnnotations reports invalid model streaming annotations.
	ReasonModelStreamingInvalidAnnotations = "ModelStreamingInvalidAnnotations"
)

// ReasonDownloadOOMKilled on the Workspace ModelMirrorReady condition reports a model download killed for exceeding
// its memory.
const ReasonDownloadOOMKilled = "DownloadOOMKilled"

// IsTerminatedReason reports whether a not-ready Workspace or InferenceSet reason cannot recover without the user
// acting: the controllers keep retrying, but the same failure recurs until the user changes the configuration or the
// cloud subscription. An InferenceSet with a replica failing for one of these reasons reports it on its
// InferenceSetReady condition, so callers can pass the reason of an InferenceSetReady=False condition whose
// observedGeneration matches the InferenceSet's generation.
func IsTerminatedReason(reason string) bool {
	switch reason {
	case ReasonQuotaExceeded,
		ReasonCapacityUnavailable,
		ReasonNodeImageInvalid,
		ReasonNodeBootstrapFailed,
		ReasonNodeConfigurationInvalid,
		ReasonNodeClassInvalid,
		ReasonNodePoolInvalid,
		ReasonContainerOOMKilled,
		ReasonImagePullError,
		ReasonContainerStartError,
		ReasonContainerCrashLoopBackOff,
		ReasonSASTokenFetchFailed,
		ReasonModelBundleMismatch,
		ReasonModelStreamingStorageClassNotFound,
		ReasonModelStreamingInvalidStorageClass,
		ReasonModelStreamingServiceAccountInvalid,
		ReasonModelStreamingInvalidAnnotations,
		ReasonDownloadOOMKilled:
		return true
	}
	return false
}

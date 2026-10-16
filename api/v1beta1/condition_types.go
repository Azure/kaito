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

package v1beta1

import (
	"k8s.io/apimachinery/pkg/util/sets"

	mmconsts "github.com/kaito-project/kaito/pkg/modelmirror/consts"
)

// ConditionType is a valid value for Condition.Type.
type ConditionType string

const (
	// ConditionTypeNodeClaimStatus is the state when checking nodeClaim status.
	// With the Karpenter node provisioner, while False, NodeClaimNotReady means provisioning is in progress and any
	// other reason means the latest attempt failed.
	ConditionTypeNodeClaimStatus = ConditionType("NodeClaimReady")

	// ConditionTypeNodeStatus is the state when checking node status.
	ConditionTypeNodeStatus = ConditionType("NodesReady")

	// ConditionTypeResourceStatus is the state when Resource has been created.
	ConditionTypeResourceStatus = ConditionType("ResourceReady")

	// WorkspaceConditionTypeInferenceStatus is the state when Inference service has been ready.
	WorkspaceConditionTypeInferenceStatus = ConditionType("InferenceReady")

	// RAGEneineConditionTypeServiceStatus is the state when service has been ready.
	RAGEneineConditionTypeServiceStatus = ConditionType("ServiceReady")

	// RAGConditionTypeServiceStatus is the state when RAG Engine service has been ready.
	RAGConditionTypeServiceStatus = ConditionType("RAGEngineServiceReady")

	// WorkspaceConditionTypeTuningJobStatus is the state when the tuning job starts normally.
	WorkspaceConditionTypeTuningJobStatus ConditionType = ConditionType("JobStarted")

	//RAGEngineConditionTypeDeleting is the RAGEngine state when starts to get deleted.
	RAGEngineConditionTypeDeleting = ConditionType("RAGEngineDeleting")

	//WorkspaceConditionTypeDeleting is the Workspace state when starts to get deleted.
	WorkspaceConditionTypeDeleting = ConditionType("WorkspaceDeleting")

	//InferenceSetConditionTypeDeleting is the InferenceSet state when starts to get deleted.
	InferenceSetConditionTypeDeleting = ConditionType("InferenceSetDeleting")

	// InferenceSetConditionTypeReady is the InferenceSet state when starts to get ready.
	InferenceSetConditionTypeReady = ConditionType("InferenceSetReady")

	// InferenceSetConditionTypeBenchmarkCompleted is set when benchmark is enabled (default).
	InferenceSetConditionTypeBenchmarkCompleted = ConditionType("BenchmarkCompleted")

	//WorkspaceConditionTypeSucceeded is the Workspace state that summarizes all operations' states.
	//For inference, the "True" condition means the inference service is ready to serve requests.
	//For fine tuning, the "True" condition means the tuning job completes successfully.
	WorkspaceConditionTypeSucceeded ConditionType = ConditionType("WorkspaceSucceeded")

	RAGEngineConditionTypeSucceeded ConditionType = ConditionType("RAGEngineSucceeded")

	// ConditionTypeScalingDownStatus is the state when scaling down nodeClaim.
	ConditionTypeScalingDownStatus = ConditionType("ScalingDownCompleted")

	// WorkspaceConditionTypeBenchmarkCompleted is the state after the post-load benchmark has run.
	// True means the benchmark completed and results are stored in status.benchmarkResult.
	// Set by default; omitted when kaito.sh/disable-benchmark is "true".
	WorkspaceConditionTypeBenchmarkCompleted = ConditionType("BenchmarkCompleted")

	// WorkspaceConditionTypeModelMirrorReady indicates the ModelMirror download is complete and model is ready for streaming.
	WorkspaceConditionTypeModelMirrorReady = ConditionType("ModelMirrorReady")

	// WorkspaceConditionTypeModelConfigReady reports whether a bring-your-own
	// model's configuration resolved successfully and still matches what the
	// deployment was sized and configured from. It also carries artifact-level
	// failures the controller can observe, so a bring-your-own deployment has a
	// single readiness condition for its model rather than two.
	WorkspaceConditionTypeModelConfigReady = ConditionType("ModelConfigReady")
)

// Node provisioning failure reasons on the NodeClaimReady and ResourceReady conditions that need the user to act,
// such as raising a quota or fixing the node configuration. The message starts with the reason reported by the node
// provisioner, so these stay stable across provisioner versions.
const (
	// ReasonQuotaExceeded reports an exhausted subscription vCPU quota (VM SKU family, regional or spot).
	ReasonQuotaExceeded = "QuotaExceeded"
	// ReasonCapacityUnavailable reports that the cloud cannot allocate the VM size, or does not offer it.
	ReasonCapacityUnavailable = "CapacityUnavailable"
	// ReasonNodeImageInvalid reports a node image that does not exist or is not a valid image reference.
	ReasonNodeImageInvalid = "NodeImageInvalid"
	// ReasonNodeBootstrapFailed reports a node that failed its bootstrap script or never registered with the cluster.
	ReasonNodeBootstrapFailed = "NodeBootstrapFailed"
	// ReasonNodeConfigurationInvalid reports an invalid NodeClass, NodePool, permission or subnet configuration.
	ReasonNodeConfigurationInvalid = "NodeConfigurationInvalid"
)

// Inference failure reasons on the InferenceReady condition that need the user to act, such as changing the instance
// type, the model image or the model configuration.
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

// ReasonDownloadOOMKilled on the ModelMirrorReady condition reports a model download killed for exceeding its memory.
const ReasonDownloadOOMKilled = mmconsts.ReasonDownloadOOMKilled

// TerminalReasons are the reasons of a not-ready Workspace that cannot recover without the user acting: the
// controllers keep retrying, but the same failure recurs until the user changes the configuration or the cloud
// subscription. An InferenceSet with a replica failing for one of these reasons reports it on InferenceSetReady.
var TerminalReasons = sets.New(
	ReasonQuotaExceeded,
	ReasonCapacityUnavailable,
	ReasonNodeImageInvalid,
	ReasonNodeBootstrapFailed,
	ReasonNodeConfigurationInvalid,
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
	ReasonDownloadOOMKilled,
)

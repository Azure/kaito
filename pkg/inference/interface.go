// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package inference

type Model interface {
	GetInferenceParameters() *PresetInferenceParam
	SupportDistributedInference() bool //If true, the model workload will be a StatefulSet, using the torch elastic runtime framework.
}

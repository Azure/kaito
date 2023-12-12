// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package inference

type Model interface {
	GetInferenceParameters() *PresetInferenceParam
	NeedStatefulSet() bool
	NeedHeadlessService() bool
}

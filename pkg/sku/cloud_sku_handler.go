// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sku

type CloudSKUHandler interface {
	GetSupportedSKUs() []string
	GetGPUConfigs() map[string]GPUConfig
}

type GPUConfig struct {
	SKU      string
	GPUCount int
	GPUMem   int
	GPUModel string
}

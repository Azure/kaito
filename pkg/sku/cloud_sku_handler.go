// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sku

type CloudSKUHandler interface {
	GetSupportedSKUs() []string
	GetGPUConfigMap() map[string]GPUConfig
}

type GPUConfig struct {
	SKU      string
	GPUCount float64
	GPUMem   int
	GPUModel string
}

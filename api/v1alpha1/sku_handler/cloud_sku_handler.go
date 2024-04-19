// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package skuhandler

type CloudSKUHandler interface {
	GetSupportedSKUs() string
	IsSupportedSKU(sku string) (GPUConfig, bool)
}

type GPUConfig struct {
	SKU      string
	GPUCount float64
	GPUMem   int
	GPUModel string
}

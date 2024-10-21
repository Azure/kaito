// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sku

var _ CloudSKUHandler = &AksArcSKUHandler{}

type AksArcSKUHandler struct {
	supportedSKUs map[string]GPUConfig
}

func NewAksArcSKUHandler() *AksArcSKUHandler {
	return &AksArcSKUHandler{
		// Supported GPU SKU for AksArc: https://learn.microsoft.com/en-us/azure/aks/hybrid/deploy-gpu-node-pool
		supportedSKUs: map[string]GPUConfig{
			"MOCVirtualMachine": {SKU: "MOCVirtualMachine", GPUCount: 1, GPUMem: 12, GPUModel: "NVIDIA K80"},
		},
	}
}

func (a *AksArcSKUHandler) GetSupportedSKUs() []string {
	return GetMapKeys(a.supportedSKUs)
}

func (a *AksArcSKUHandler) GetGPUConfigs() map[string]GPUConfig {
	return a.supportedSKUs
}

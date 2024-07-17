// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sku

import (
	"github.com/azure/kaito/pkg/utils/consts"
)

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

func GetCloudSKUHandler(cloud string) CloudSKUHandler {
	switch cloud {
	case consts.AzureCloudName:
		return NewAzureSKUHandler()
	case consts.AWSCloudName:
		return NewAwsSKUHandler()
	default:
		return nil
	}
}

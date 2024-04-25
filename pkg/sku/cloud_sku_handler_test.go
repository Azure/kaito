package sku

import (
	"testing"
)

func TestAzureSKUHandler(t *testing.T) {
	handler := NewAzureSKUHandler()

	// Test GetSupportedSKUs
	skus := handler.GetSupportedSKUs()
	if len(skus) == 0 {
		t.Errorf("GetSupportedSKUs returned an empty array")
	}

	// Test GetGPUConfigs with a SKU that is supported
	sku := "Standard_NC6s_v3"
	configMap := handler.GetGPUConfigs()
	config, exists := configMap[sku]
	if !exists {
		t.Errorf("Supported SKU missing from GPUConfigs")
	}
	if config.SKU != sku {
		t.Errorf("Incorrect config returned for a supported SKU")
	}

	// Test GetGPUConfigs with a SKU that is not supported
	sku = "Unsupported_SKU"
	config, exists = configMap[sku]
	if exists {
		t.Errorf("Unsupported SKU found in GPUConfigs")
	}
}

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
	config, supported := configMap[sku]
	if !supported {
		t.Errorf("IsSupportedSKU returned false for a supported SKU")
	}
	if config.SKU != sku {
		t.Errorf("IsSupportedSKU returned incorrect config for a supported SKU")
	}

	// Test IsSupportedSKU with a SKU that is not supported
	sku = "Unsupported_SKU"
	config, supported = configMap[sku]
	if supported {
		t.Errorf("IsSupportedSKU returned true for an unsupported SKU")
	}
}

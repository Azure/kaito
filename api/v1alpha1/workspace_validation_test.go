// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package v1alpha1

import (
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"
)

// Mock data for tests
// var (
// 	supportedGPUConfigs = map[string]GPUConfig{
// 		"gpu-sku-1": {GPUCount: 4, GPUMem: 8},
// 		// ... add other mock SKUs
// 	}

// 	presetRequirementsMap = map[string]PresetRequirement{
// 		"preset-1": {MinGPUCount: 2, MinMemoryPerGPU: 4, MinTotalMemory: 16},
// 		// ... add other mock presets
// 	}

// 	// Define a mock of isValidPreset function based on the presets in the map
// 	isValidPreset = func(preset string) bool {
// 		_, exists := presetRequirementsMap[preset]
// 		return exists
// 	}
// )

func pointerToInt(i int) *int {
	return &i
}

// TestResourceSpecValidateCreate tests the validateCreate function.
func TestResourceSpecValidateCreate(t *testing.T) {
	// Arrange your test cases with different ResourceSpec and InferenceSpec configurations
	tests := []struct {
		name          string
		resourceSpec  *ResourceSpec
		inferenceSpec *InferenceSpec
		errContent    string // Content expect error to include, if any
		expectErrs    bool
	}{
		{
			name: "Insufficient total GPU memory",
			resourceSpec: &ResourceSpec{
				InstanceType: "standard_nc6",
				Count:        pointerToInt(1),
			},
			inferenceSpec: &InferenceSpec{
				Preset: &PresetSpec{
					PresetMeta: PresetMeta{
						Name: ModelName("falcon-7b"),
					},
				},
			},
			errContent: "Insufficient total GPU memory",
			expectErrs: true,
		},

		{
			name: "Insufficient number of GPUs",
			resourceSpec: &ResourceSpec{
				InstanceType: "standard_nc24ads_a100_v4",
				Count:        pointerToInt(1),
			},
			inferenceSpec: &InferenceSpec{
				Preset: &PresetSpec{
					PresetMeta: PresetMeta{
						Name: ModelName("llama-2-13b-chat"),
					},
				},
			},
			errContent: "Insufficient number of GPUs",
			expectErrs: true,
		},

		{
			name: "Invalid SKU",
			resourceSpec: &ResourceSpec{
				InstanceType: "standard_invalid_sku",
				Count:        pointerToInt(1),
			},
			inferenceSpec: &InferenceSpec{
				Preset: &PresetSpec{
					PresetMeta: PresetMeta{
						Name: ModelName("llama-2-70b"),
					},
				},
			},
			errContent: "Unsupported instance",
			expectErrs: true,
		},

		{
			name: "Invalid Preset",
			resourceSpec: &ResourceSpec{
				InstanceType: "standard_nv12s_v3",
				Count:        pointerToInt(1),
			},
			inferenceSpec: &InferenceSpec{
				Preset: &PresetSpec{
					PresetMeta: PresetMeta{
						Name: ModelName("invalid-preset"),
					},
				},
			},
			errContent: "Unsupported preset",
			expectErrs: true,
		},

		{
			name: "Invalid SKU",
			resourceSpec: &ResourceSpec{
				InstanceType: "standard_invalid_sku",
				Count:        pointerToInt(1),
			},
			inferenceSpec: &InferenceSpec{
				Preset: &PresetSpec{
					PresetMeta: PresetMeta{
						Name: ModelName("llama-2-70b"),
					},
				},
			},
			errContent: "Unsupported instance",
			expectErrs: true,
		},

		{
			name: "N-Prefix SKU",
			resourceSpec: &ResourceSpec{
				InstanceType: "standard_nsku",
				Count:        pointerToInt(1),
			},
			inferenceSpec: &InferenceSpec{
				Preset: &PresetSpec{
					PresetMeta: PresetMeta{
						Name: ModelName("llama-2-7b"),
					},
				},
			},
			errContent: "",
			expectErrs: false,
		},

		{
			name: "D-Prefix SKU",
			resourceSpec: &ResourceSpec{
				InstanceType: "standard_dsku",
				Count:        pointerToInt(1),
			},
			inferenceSpec: &InferenceSpec{
				Preset: &PresetSpec{
					PresetMeta: PresetMeta{
						Name: ModelName("llama-2-7b"),
					},
				},
			},
			errContent: "",
			expectErrs: false,
		},
	}

	// Run the tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			errs := tc.resourceSpec.validateCreate(*tc.inferenceSpec)
			hasErrs := errs != nil
			if hasErrs != tc.expectErrs {
				t.Errorf("validateCreate() errors = %v, expectErrs %v", errs, tc.expectErrs)
			}

			// If there is an error and errContent is not empty, check that the error contains the expected content.
			if hasErrs && tc.errContent != "" {
				errMsg := errs.Error()
				if !strings.Contains(errMsg, tc.errContent) {
					t.Errorf("validateCreate() error message = %v, expected to contain = %v", errMsg, tc.errContent)
				}
			}
		})
	}
}

func TestInferenceSpecValidateCreate(t *testing.T) {
	tests := []struct {
		name          string
		inferenceSpec *InferenceSpec
		errContent    string // Content expected error to include, if any
		expectErrs    bool
	}{
		{
			name: "Invalid Preset Name",
			inferenceSpec: &InferenceSpec{
				Preset: &PresetSpec{
					PresetMeta: PresetMeta{
						Name: ModelName("Invalid-Preset-Name"),
					},
				},
			},
			errContent: "Unsupported preset name",
			expectErrs: true,
		},
		{
			name: "Preset and Template Set",
			inferenceSpec: &InferenceSpec{
				Preset: &PresetSpec{
					PresetMeta: PresetMeta{
						Name: ModelName("falcon-7b"),
					},
				},
				Template: &v1.PodTemplateSpec{}, // Assuming a non-nil TemplateSpec implies it's set
			},
			errContent: "preset and template cannot be set at the same time",
			expectErrs: true,
		},
		{
			name: "Private Access Without Image",
			inferenceSpec: &InferenceSpec{
				Preset: &PresetSpec{
					PresetMeta: PresetMeta{
						Name:       ModelName("llama-2-7b"),
						AccessMode: "private",
					},
					PresetOptions: PresetOptions{},
				},
			},
			errContent: "When AccessMode is private, an image must be provided",
			expectErrs: true,
		},
		{
			name: "Valid Preset",
			inferenceSpec: &InferenceSpec{
				Preset: &PresetSpec{
					PresetMeta: PresetMeta{
						Name:       ModelName("falcon-7b"),
						AccessMode: "public",
					},
				},
			},
			errContent: "",
			expectErrs: false,
		},
	}

	// Run the tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			errs := tc.inferenceSpec.validateCreate()
			hasErrs := errs != nil
			if hasErrs != tc.expectErrs {
				t.Errorf("validateCreate() errors = %v, expectErrs %v", errs, tc.expectErrs)
			}

			// If there is an error and errContent is not empty, check that the error contains the expected content.
			if hasErrs && tc.errContent != "" {
				errMsg := errs.Error()
				if !strings.Contains(errMsg, tc.errContent) {
					t.Errorf("validateCreate() error message = %v, expected to contain = %v", errMsg, tc.errContent)
				}
			}
		})
	}
}

// TestGetSupportedSKUs tests the getSupportedSKUs function.
func TestGetSupportedSKUs(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		gpuConfigs     map[string]GPUConfig
		expectedResult string
	}{
		{
			name:           "no SKUs supported",
			gpuConfigs:     map[string]GPUConfig{},
			expectedResult: "",
		},
		{
			name: "one SKU supported",
			gpuConfigs: map[string]GPUConfig{
				"standard_nc6": {SKU: "standard_nc6"},
			},
			expectedResult: "standard_nc6",
		},
		{
			name: "multiple SKUs supported",
			gpuConfigs: map[string]GPUConfig{
				"standard_nc6":  {SKU: "standard_nc6"},
				"standard_nc12": {SKU: "standard_nc12"},
			},
			expectedResult: "standard_nc6, standard_nc12",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			SupportedGPUConfigs = tc.gpuConfigs

			// Act
			result := getSupportedSKUs()

			// Assert
			if result != tc.expectedResult {
				t.Errorf("getSupportedSKUs() = %v, want %v", result, tc.expectedResult)
			}
		})
	}
}

// TestIsValidPreset tests the isValidPreset function.
func TestIsValidPreset(t *testing.T) {
	tests := []struct {
		name        string
		preset      string
		expectValid bool
	}{
		{
			name:        "valid preset",
			preset:      "falcon-7b",
			expectValid: true,
		},
		{
			name:        "invalid preset",
			preset:      "nonexistent-preset",
			expectValid: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if valid := isValidPreset(test.preset); valid != test.expectValid {
				t.Errorf("isValidPreset(%s) = %v, want %v", test.preset, valid, test.expectValid)
			}
		})
	}
}

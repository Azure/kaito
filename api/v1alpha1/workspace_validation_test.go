// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package v1alpha1

/*
import (
	"reflect"
	"sort"
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func pointerToInt(i int) *int {
	return &i
}

func TestResourceSpecValidateCreate(t *testing.T) {
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
				InstanceType: "Standard_NC6",
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
				InstanceType: "Standard_NC24ads_A100_v4",
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
				InstanceType: "Standard_invalid_sku",
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
			name: "Only Template set",
			resourceSpec: &ResourceSpec{
				InstanceType: "Standard_NV12s_v3",
				Count:        pointerToInt(1),
			},
			inferenceSpec: &InferenceSpec{
				Template: &v1.PodTemplateSpec{}, // Assuming a non-nil TemplateSpec implies it's set
			},
			errContent: "",
			expectErrs: false,
		},
		{
			name: "Invalid Preset",
			resourceSpec: &ResourceSpec{
				InstanceType: "Standard_NV12s_v3",
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
				InstanceType: "Standard_invalid_sku",
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
				InstanceType: "Standard_Nsku",
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
				InstanceType: "Standard_Dsku",
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

func TestResourceSpecValidateUpdate(t *testing.T) {

	tests := []struct {
		name        string
		newResource *ResourceSpec
		oldResource *ResourceSpec
		errContent  string // Content expected error to include, if any
		expectErrs  bool
	}{
		{
			name: "Immutable Count",
			newResource: &ResourceSpec{
				Count: pointerToInt(10),
			},
			oldResource: &ResourceSpec{
				Count: pointerToInt(5),
			},
			errContent: "field is immutable",
			expectErrs: true,
		},
		{
			name: "Immutable InstanceType",
			newResource: &ResourceSpec{
				InstanceType: "new_type",
			},
			oldResource: &ResourceSpec{
				InstanceType: "old_type",
			},
			errContent: "field is immutable",
			expectErrs: true,
		},
		{
			name: "Immutable LabelSelector",
			newResource: &ResourceSpec{
				LabelSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"key1": "value1"}},
			},
			oldResource: &ResourceSpec{
				LabelSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"key2": "value2"}},
			},
			errContent: "field is immutable",
			expectErrs: true,
		},
		{
			name: "Valid Update",
			newResource: &ResourceSpec{
				Count:         pointerToInt(5),
				InstanceType:  "same_type",
				LabelSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"key": "value"}},
			},
			oldResource: &ResourceSpec{
				Count:         pointerToInt(5),
				InstanceType:  "same_type",
				LabelSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"key": "value"}},
			},
			errContent: "",
			expectErrs: false,
		},
	}

	// Run the tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			errs := tc.newResource.validateUpdate(tc.oldResource)
			hasErrs := errs != nil
			if hasErrs != tc.expectErrs {
				t.Errorf("validateUpdate() errors = %v, expectErrs %v", errs, tc.expectErrs)
			}

			// If there is an error and errContent is not empty, check that the error contains the expected content.
			if hasErrs && tc.errContent != "" {
				errMsg := errs.Error()
				if !strings.Contains(errMsg, tc.errContent) {
					t.Errorf("validateUpdate() error message = %v, expected to contain = %v", errMsg, tc.errContent)
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
			name: "Only Template set",
			inferenceSpec: &InferenceSpec{
				Template: &v1.PodTemplateSpec{}, // Assuming a non-nil TemplateSpec implies it's set
			},
			errContent: "",
			expectErrs: false,
		},
		{
			name:          "Preset and Template Unset",
			inferenceSpec: &InferenceSpec{},
			errContent:    "Preset or Template must be specified",
			expectErrs:    true,
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
			errContent: "Preset and Template cannot be set at the same time",
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

func TestInferenceSpecValidateUpdate(t *testing.T) {
	tests := []struct {
		name         string
		newInference *InferenceSpec
		oldInference *InferenceSpec
		errContent   string // Content expected error to include, if any
		expectErrs   bool
	}{
		{
			name: "Preset Immutable",
			newInference: &InferenceSpec{
				Preset: &PresetSpec{
					PresetMeta: PresetMeta{
						Name: ModelName("new-preset"),
					},
				},
			},
			oldInference: &InferenceSpec{
				Preset: &PresetSpec{
					PresetMeta: PresetMeta{
						Name: ModelName("old-preset"),
					},
				},
			},
			errContent: "field is immutable",
			expectErrs: true,
		},
		{
			name: "Template Unset",
			newInference: &InferenceSpec{
				Template: nil,
			},
			oldInference: &InferenceSpec{
				Template: &v1.PodTemplateSpec{},
			},
			errContent: "field cannot be unset/set if it was set/unset",
			expectErrs: true,
		},
		{
			name: "Template Set",
			newInference: &InferenceSpec{
				Template: &v1.PodTemplateSpec{},
			},
			oldInference: &InferenceSpec{
				Template: nil,
			},
			errContent: "field cannot be unset/set if it was set/unset",
			expectErrs: true,
		},
		{
			name: "Valid Update",
			newInference: &InferenceSpec{
				Template: &v1.PodTemplateSpec{},
			},
			oldInference: &InferenceSpec{
				Template: &v1.PodTemplateSpec{},
			},
			errContent: "",
			expectErrs: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			errs := tc.newInference.validateUpdate(tc.oldInference)
			hasErrs := errs != nil
			if hasErrs != tc.expectErrs {
				t.Errorf("validateUpdate() errors = %v, expectErrs %v", errs, tc.expectErrs)
			}

			// If there is an error and errContent is not empty, check that the error contains the expected content.
			if hasErrs && tc.errContent != "" {
				errMsg := errs.Error()
				if !strings.Contains(errMsg, tc.errContent) {
					t.Errorf("validateUpdate() error message = %v, expected to contain = %v", errMsg, tc.errContent)
				}
			}
		})
	}
}

func TestGetSupportedSKUs(t *testing.T) {
	tests := []struct {
		name           string
		gpuConfigs     map[string]GPUConfig
		expectedResult []string // changed to a slice for deterministic ordering
	}{
		{
			name:           "no SKUs supported",
			gpuConfigs:     map[string]GPUConfig{},
			expectedResult: []string{""},
		},
		{
			name: "one SKU supported",
			gpuConfigs: map[string]GPUConfig{
				"Standard_NC6": {SKU: "Standard_NC6"},
			},
			expectedResult: []string{"Standard_NC6"},
		},
		{
			name: "multiple SKUs supported",
			gpuConfigs: map[string]GPUConfig{
				"Standard_NC6":  {SKU: "Standard_NC6"},
				"Standard_NC12": {SKU: "Standard_NC12"},
			},
			expectedResult: []string{"Standard_NC6", "Standard_NC12"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			SupportedGPUConfigs = tc.gpuConfigs

			resultSlice := strings.Split(getSupportedSKUs(), ", ")
			sort.Strings(resultSlice)

			// Sort the expectedResult for comparison
			expectedResultSlice := tc.expectedResult
			sort.Strings(expectedResultSlice)

			if !reflect.DeepEqual(resultSlice, expectedResultSlice) {
				t.Errorf("getSupportedSKUs() = %v, want %v", resultSlice, expectedResultSlice)
			}
		})
	}
}

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
*/

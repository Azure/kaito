// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package v1alpha1

import (
	"context"
	"github.com/azure/kaito/pkg/utils/plugin"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/azure/kaito/pkg/model"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const DefaultReleaseNamespace = "kaito-workspace"
const DefaultConfigMapName = "lora-params"

var gpuCountRequirement string
var totalGPUMemoryRequirement string
var perGPUMemoryRequirement string

type testModel struct{}

func (*testModel) GetInferenceParameters() *model.PresetParam {
	return &model.PresetParam{
		GPUCountRequirement:       gpuCountRequirement,
		TotalGPUMemoryRequirement: totalGPUMemoryRequirement,
		PerGPUMemoryRequirement:   perGPUMemoryRequirement,
	}
}
func (*testModel) GetTuningParameters() *model.PresetParam {
	return &model.PresetParam{
		GPUCountRequirement:       gpuCountRequirement,
		TotalGPUMemoryRequirement: totalGPUMemoryRequirement,
		PerGPUMemoryRequirement:   perGPUMemoryRequirement,
	}
}
func (*testModel) SupportDistributedInference() bool {
	return false
}
func (*testModel) SupportTuning() bool {
	return true
}

type testModelPrivate struct{}

func (*testModelPrivate) GetInferenceParameters() *model.PresetParam {
	return &model.PresetParam{
		ImageAccessMode:           "private",
		GPUCountRequirement:       gpuCountRequirement,
		TotalGPUMemoryRequirement: totalGPUMemoryRequirement,
		PerGPUMemoryRequirement:   perGPUMemoryRequirement,
	}
}
func (*testModelPrivate) GetTuningParameters() *model.PresetParam {
	return &model.PresetParam{
		ImageAccessMode:           "private",
		GPUCountRequirement:       gpuCountRequirement,
		TotalGPUMemoryRequirement: totalGPUMemoryRequirement,
		PerGPUMemoryRequirement:   perGPUMemoryRequirement,
	}
}
func (*testModelPrivate) SupportDistributedInference() bool {
	return false
}
func (*testModelPrivate) SupportTuning() bool {
	return true
}

func RegisterValidationTestModels() {
	var test testModel
	var testPrivate testModelPrivate
	plugin.KaitoModelRegister.Register(&plugin.Registration{
		Name:     "test-validation",
		Instance: &test,
	})
	plugin.KaitoModelRegister.Register(&plugin.Registration{
		Name:     "private-test-validation",
		Instance: &testPrivate,
	})
}

func pointerToInt(i int) *int {
	return &i
}

func defaultConfigMapManifest() *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultConfigMapName,
			Namespace: DefaultReleaseNamespace, // Replace this with the appropriate namespace variable if dynamic
		},
		Data: map[string]string{
			"training_config.yaml": `training_config:
  ModelConfig:
    torch_dtype: "bfloat16"
    local_files_only: true
    device_map: "auto"

  TokenizerParams:
    padding: true
    truncation: true

  QuantizationConfig:
    load_in_4bit: false

  LoraConfig:
    r: 16
    lora_alpha: 32
    target_modules: "query_key_value"
    lora_dropout: 0.05
    bias: "none"

  TrainingArguments:
    output_dir: "."
    num_train_epochs: 4
    auto_find_batch_size: true
    ddp_find_unused_parameters: false
    save_strategy: "epoch"

  DatasetConfig:
    shuffle_dataset: true
    train_test_split: 1

  DataCollator:
    mlm: true`,
		},
	}
}

func TestResourceSpecValidateCreate(t *testing.T) {
	RegisterValidationTestModels()
	tests := []struct {
		name                string
		resourceSpec        *ResourceSpec
		modelGPUCount       string
		modelPerGPUMemory   string
		modelTotalGPUMemory string
		preset              bool
		errContent          string // Content expect error to include, if any
		expectErrs          bool
	}{
		{
			name: "Valid resource",
			resourceSpec: &ResourceSpec{
				InstanceType: "Standard_ND96asr_v4",
				Count:        pointerToInt(1),
			},
			modelGPUCount:       "8",
			modelPerGPUMemory:   "19Gi",
			modelTotalGPUMemory: "152Gi",
			preset:              true,
			errContent:          "",
			expectErrs:          false,
		},
		{
			name: "Insufficient total GPU memory",
			resourceSpec: &ResourceSpec{
				InstanceType: "Standard_NC6",
				Count:        pointerToInt(1),
			},
			modelGPUCount:       "1",
			modelPerGPUMemory:   "0",
			modelTotalGPUMemory: "14Gi",
			preset:              true,
			errContent:          "Insufficient total GPU memory",
			expectErrs:          true,
		},

		{
			name: "Insufficient number of GPUs",
			resourceSpec: &ResourceSpec{
				InstanceType: "Standard_NC24ads_A100_v4",
				Count:        pointerToInt(1),
			},
			modelGPUCount:       "2",
			modelPerGPUMemory:   "15Gi",
			modelTotalGPUMemory: "30Gi",
			preset:              true,
			errContent:          "Insufficient number of GPUs",
			expectErrs:          true,
		},
		{
			name: "Insufficient per GPU memory",
			resourceSpec: &ResourceSpec{
				InstanceType: "Standard_NC6",
				Count:        pointerToInt(2),
			},
			modelGPUCount:       "1",
			modelPerGPUMemory:   "15Gi",
			modelTotalGPUMemory: "15Gi",
			preset:              true,
			errContent:          "Insufficient per GPU memory",
			expectErrs:          true,
		},

		{
			name: "Invalid SKU",
			resourceSpec: &ResourceSpec{
				InstanceType: "Standard_invalid_sku",
				Count:        pointerToInt(1),
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
			preset:     false,
			errContent: "",
			expectErrs: false,
		},
		{
			name: "N-Prefix SKU",
			resourceSpec: &ResourceSpec{
				InstanceType: "Standard_Nsku",
				Count:        pointerToInt(1),
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
			errContent: "",
			expectErrs: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var spec InferenceSpec

			if tc.preset {
				spec = InferenceSpec{
					Preset: &PresetSpec{
						PresetMeta: PresetMeta{
							Name: ModelName("test-validation"),
						},
					},
				}
			} else {
				spec = InferenceSpec{
					Template: &v1.PodTemplateSpec{}, // Assuming a non-nil TemplateSpec implies it's set
				}
			}

			gpuCountRequirement = tc.modelGPUCount
			totalGPUMemoryRequirement = tc.modelTotalGPUMemory
			perGPUMemoryRequirement = tc.modelPerGPUMemory

			errs := tc.resourceSpec.validateCreate(spec)
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
	RegisterValidationTestModels()
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
			errContent: "model is not registered",
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
						Name: ModelName("test-validation"),
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
						Name:       ModelName("test-validation"),
						AccessMode: "private",
					},
					PresetOptions: PresetOptions{},
				},
			},
			errContent: "When AccessMode is private, an image must be provided",
			expectErrs: true,
		},
		{
			name: "Private Preset With Public AccessMode",
			inferenceSpec: &InferenceSpec{
				Preset: &PresetSpec{
					PresetMeta: PresetMeta{
						Name: ModelName("private-test-validation"),
					},
					PresetOptions: PresetOptions{},
				},
			},
			errContent: "This preset only supports private AccessMode, AccessMode must be private to continue",
			expectErrs: true,
		},
		{
			name: "Valid Preset",
			inferenceSpec: &InferenceSpec{
				Preset: &PresetSpec{
					PresetMeta: PresetMeta{
						Name:       ModelName("test-validation"),
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
			// If the test expects an error, setup defer function to catch the panic.
			if tc.expectErrs {
				defer func() {
					if r := recover(); r != nil {
						// Check if the recovered panic matches the expected error content.
						if errContent, ok := r.(string); ok && strings.Contains(errContent, tc.errContent) {
							return
						}
						t.Errorf("unexpected panic: %v", r)
					}
				}()
			}
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

func TestWorkspaceValidateCreate(t *testing.T) {
	tests := []struct {
		name      string
		workspace *Workspace
		wantErr   bool
		errField  string
	}{
		{
			name:      "Neither Inference nor Tuning specified",
			workspace: &Workspace{},
			wantErr:   true,
			errField:  "neither",
		},
		{
			name: "Both Inference and Tuning specified",
			workspace: &Workspace{
				Inference: &InferenceSpec{},
				Tuning:    &TuningSpec{},
			},
			wantErr:  true,
			errField: "both",
		},
		{
			name: "Only Inference specified",
			workspace: &Workspace{
				Inference: &InferenceSpec{},
			},
			wantErr:  false,
			errField: "",
		},
		{
			name: "Only Tuning specified",
			workspace: &Workspace{
				Tuning: &TuningSpec{Input: &DataSource{}},
			},
			wantErr:  false,
			errField: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.workspace.validateCreate()
			if (errs != nil) != tt.wantErr {
				t.Errorf("validateCreate() error = %v, wantErr %v", errs, tt.wantErr)
			}
			if errs != nil && !strings.Contains(errs.Error(), tt.errField) {
				t.Errorf("validateCreate() expected error to contain field %s, but got %s", tt.errField, errs.Error())
			}
		})
	}
}

func TestWorkspaceValidateUpdate(t *testing.T) {
	tests := []struct {
		name         string
		oldWorkspace *Workspace
		newWorkspace *Workspace
		expectErrs   bool
		errFields    []string // Fields we expect to have errors
	}{
		{
			name:         "Inference toggled on",
			oldWorkspace: &Workspace{},
			newWorkspace: &Workspace{
				Inference: &InferenceSpec{},
			},
			expectErrs: true,
			errFields:  []string{"inference"},
		},
		{
			name: "Inference toggled off",
			oldWorkspace: &Workspace{
				Inference: &InferenceSpec{Preset: &PresetSpec{}},
			},
			newWorkspace: &Workspace{},
			expectErrs:   true,
			errFields:    []string{"inference"},
		},
		{
			name:         "Tuning toggled on",
			oldWorkspace: &Workspace{},
			newWorkspace: &Workspace{
				Tuning: &TuningSpec{Input: &DataSource{}},
			},
			expectErrs: true,
			errFields:  []string{"tuning"},
		},
		{
			name: "Tuning toggled off",
			oldWorkspace: &Workspace{
				Tuning: &TuningSpec{Input: &DataSource{}},
			},
			newWorkspace: &Workspace{},
			expectErrs:   true,
			errFields:    []string{"tuning"},
		},
		{
			name: "No toggling",
			oldWorkspace: &Workspace{
				Tuning: &TuningSpec{Input: &DataSource{}},
			},
			newWorkspace: &Workspace{
				Tuning: &TuningSpec{Input: &DataSource{}},
			},
			expectErrs: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.newWorkspace.validateUpdate(tt.oldWorkspace)
			hasErrs := errs != nil

			if hasErrs != tt.expectErrs {
				t.Errorf("validateUpdate() errors = %v, expectErrs %v", errs, tt.expectErrs)
			}

			if hasErrs {
				for _, field := range tt.errFields {
					if !strings.Contains(errs.Error(), field) {
						t.Errorf("validateUpdate() expected errors to contain field %s, but got %s", field, errs.Error())
					}
				}
			}
		})
	}
}

func TestTuningSpecValidateCreate(t *testing.T) {
	RegisterValidationTestModels()
	// Set ReleaseNamespace Env
	os.Setenv(DefaultReleaseNamespaceEnvVar, DefaultReleaseNamespace)
	defer os.Unsetenv(DefaultReleaseNamespaceEnvVar)

	// Create fake client with default ConfigMap
	scheme := runtime.NewScheme()
	_ = v1.AddToScheme(scheme)
	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(defaultConfigMapManifest()).Build()

	// Include client in ctx
	ctx := context.Background()
	ctx = context.WithValue(ctx, "clientKey", client)

	tests := []struct {
		name       string
		tuningSpec *TuningSpec
		wantErr    bool
		errFields  []string // Fields we expect to have errors
	}{
		{
			name: "All fields valid",
			tuningSpec: &TuningSpec{
				Input:  &DataSource{Name: "valid-input", HostPath: "valid-input"},
				Output: &DataDestination{HostPath: "valid-output"},
				Preset: &PresetSpec{PresetMeta: PresetMeta{Name: ModelName("test-validation")}},
				Method: TuningMethodLora,
				Config: DefaultConfigMapName,
			},
			wantErr:   false,
			errFields: nil,
		},
		{
			name: "Missing Input",
			tuningSpec: &TuningSpec{
				Output: &DataDestination{HostPath: "valid-output"},
				Preset: &PresetSpec{PresetMeta: PresetMeta{Name: ModelName("test-validation")}},
				Method: TuningMethodLora,
			},
			wantErr:   true,
			errFields: []string{"Input"},
		},
		{
			name: "Missing Output",
			tuningSpec: &TuningSpec{
				Input:  &DataSource{Name: "valid-input"},
				Preset: &PresetSpec{PresetMeta: PresetMeta{Name: ModelName("test-validation")}},
				Method: TuningMethodLora,
			},
			wantErr:   true,
			errFields: []string{"Output"},
		},
		{
			name: "Missing Preset",
			tuningSpec: &TuningSpec{
				Input:  &DataSource{Name: "valid-input"},
				Output: &DataDestination{HostPath: "valid-output"},
				Method: TuningMethodLora,
			},
			wantErr:   true,
			errFields: []string{"Preset"},
		},
		{
			name: "Invalid Preset",
			tuningSpec: &TuningSpec{
				Input:  &DataSource{Name: "valid-input"},
				Output: &DataDestination{HostPath: "valid-output"},
				Preset: &PresetSpec{PresetMeta: PresetMeta{Name: ModelName("invalid-preset")}},
				Method: TuningMethodLora,
			},
			wantErr:   true,
			errFields: []string{"presetName"},
		},
		{
			name: "Invalid Method",
			tuningSpec: &TuningSpec{
				Input:  &DataSource{Name: "valid-input"},
				Output: &DataDestination{HostPath: "valid-output"},
				Preset: &PresetSpec{PresetMeta: PresetMeta{Name: ModelName("test-validation")}},
				Method: "invalid-method",
			},
			wantErr:   true,
			errFields: []string{"Method"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.tuningSpec.validateCreate(ctx)
			hasErrs := errs != nil

			if hasErrs != tt.wantErr {
				t.Errorf("validateCreate() errors = %v, wantErr %v", errs, tt.wantErr)
			}

			if hasErrs {
				for _, field := range tt.errFields {
					if !strings.Contains(errs.Error(), field) {
						t.Errorf("validateCreate() expected errors to contain field %s, but got %s", field, errs.Error())
					}
				}
			}
		})
	}
}

func TestTuningSpecValidateUpdate(t *testing.T) {
	RegisterValidationTestModels()
	tests := []struct {
		name       string
		oldTuning  *TuningSpec
		newTuning  *TuningSpec
		expectErrs bool
		errFields  []string // Fields we expect to have errors
	}{
		{
			name: "No changes",
			oldTuning: &TuningSpec{
				Input:  &DataSource{Name: "input1"},
				Output: &DataDestination{HostPath: "path1"},
				Preset: &PresetSpec{PresetMeta: PresetMeta{Name: ModelName("test-validation")}},
				Method: TuningMethodLora,
			},
			newTuning: &TuningSpec{
				Input:  &DataSource{Name: "input1"},
				Output: &DataDestination{HostPath: "path1"},
				Preset: &PresetSpec{PresetMeta: PresetMeta{Name: ModelName("test-validation")}},
				Method: TuningMethodLora,
			},
			expectErrs: false,
		},
		{
			name: "Input changed",
			oldTuning: &TuningSpec{
				Input:  &DataSource{Name: "input", HostPath: "inputpath"},
				Output: &DataDestination{HostPath: "outputpath"},
			},
			newTuning: &TuningSpec{
				Input:  &DataSource{Name: "input", HostPath: "randompath"},
				Output: &DataDestination{HostPath: "outputpath"},
			},
			expectErrs: true,
			errFields:  []string{"HostPath"},
		},
		{
			name: "Output changed",
			oldTuning: &TuningSpec{
				Output: &DataDestination{HostPath: "path1"},
			},
			newTuning: &TuningSpec{
				Output: &DataDestination{HostPath: "path2"},
			},
			expectErrs: true,
			errFields:  []string{"Output"},
		},
		{
			name: "Preset changed",
			oldTuning: &TuningSpec{
				Preset: &PresetSpec{PresetMeta: PresetMeta{Name: ModelName("test-validation")}},
			},
			newTuning: &TuningSpec{
				Preset: &PresetSpec{PresetMeta: PresetMeta{Name: ModelName("invalid-preset")}},
			},
			expectErrs: true,
			errFields:  []string{"Preset"},
		},
		{
			name: "Method changed",
			oldTuning: &TuningSpec{
				Method: TuningMethodLora,
			},
			newTuning: &TuningSpec{
				Method: TuningMethodQLora,
			},
			expectErrs: true,
			errFields:  []string{"Method"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.newTuning.validateUpdate(tt.oldTuning)
			hasErrs := errs != nil

			if hasErrs != tt.expectErrs {
				t.Errorf("validateUpdate() errors = %v, expectErrs %v", errs, tt.expectErrs)
			}

			if hasErrs {
				for _, field := range tt.errFields {
					if !strings.Contains(errs.Error(), field) {
						t.Errorf("validateUpdate() expected errors to contain field %s, but got %s", field, errs.Error())
					}
				}
			}
		})
	}
}

func TestDataSourceValidateCreate(t *testing.T) {
	tests := []struct {
		name       string
		dataSource *DataSource
		wantErr    bool
		errField   string // The field we expect to have an error on
	}{
		{
			name: "URLs specified only",
			dataSource: &DataSource{
				URLs: []string{"http://example.com/data"},
			},
			wantErr: false,
		},
		{
			name: "HostPath specified only",
			dataSource: &DataSource{
				HostPath: "/data/path",
			},
			wantErr: false,
		},
		{
			name: "Image specified only",
			dataSource: &DataSource{
				Image: "data-image:latest",
			},
			wantErr: false,
		},
		{
			name:       "None specified",
			dataSource: &DataSource{},
			wantErr:    true,
			errField:   "Exactly one of URLs, HostPath, or Image must be specified",
		},
		{
			name: "URLs and HostPath specified",
			dataSource: &DataSource{
				URLs:     []string{"http://example.com/data"},
				HostPath: "/data/path",
			},
			wantErr:  true,
			errField: "Exactly one of URLs, HostPath, or Image must be specified",
		},
		{
			name: "All fields specified",
			dataSource: &DataSource{
				URLs:     []string{"http://example.com/data"},
				HostPath: "/data/path",
				Image:    "data-image:latest",
			},
			wantErr:  true,
			errField: "Exactly one of URLs, HostPath, or Image must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.dataSource.validateCreate()
			hasErrs := errs != nil

			if hasErrs != tt.wantErr {
				t.Errorf("validateCreate() error = %v, wantErr %v", errs, tt.wantErr)
			}

			if hasErrs && tt.errField != "" && !strings.Contains(errs.Error(), tt.errField) {
				t.Errorf("validateCreate() expected error to contain %s, but got %s", tt.errField, errs.Error())
			}
		})
	}
}

func TestDataSourceValidateUpdate(t *testing.T) {
	tests := []struct {
		name      string
		oldSource *DataSource
		newSource *DataSource
		wantErr   bool
		errFields []string // Fields we expect to have errors
	}{
		{
			name: "No changes",
			oldSource: &DataSource{
				URLs:             []string{"http://example.com/data1", "http://example.com/data2"},
				HostPath:         "/data/path",
				Image:            "data-image:latest",
				ImagePullSecrets: []string{"secret1", "secret2"},
			},
			newSource: &DataSource{
				URLs:             []string{"http://example.com/data2", "http://example.com/data1"}, // Note the different order, should not matter
				HostPath:         "/data/path",
				Image:            "data-image:latest",
				ImagePullSecrets: []string{"secret2", "secret1"}, // Note the different order, should not matter
			},
			wantErr: false,
		},
		{
			name: "Name changed",
			oldSource: &DataSource{
				Name: "original-dataset",
			},
			newSource: &DataSource{
				Name: "new-dataset",
			},
			wantErr:   true,
			errFields: []string{"Name"},
		},
		{
			name: "URLs changed",
			oldSource: &DataSource{
				URLs: []string{"http://example.com/old"},
			},
			newSource: &DataSource{
				URLs: []string{"http://example.com/new"},
			},
			wantErr:   true,
			errFields: []string{"URLs"},
		},
		{
			name: "HostPath changed",
			oldSource: &DataSource{
				HostPath: "/old/path",
			},
			newSource: &DataSource{
				HostPath: "/new/path",
			},
			wantErr:   true,
			errFields: []string{"HostPath"},
		},
		{
			name: "Image changed",
			oldSource: &DataSource{
				Image: "old-image:latest",
			},
			newSource: &DataSource{
				Image: "new-image:latest",
			},
			wantErr:   true,
			errFields: []string{"Image"},
		},
		{
			name: "ImagePullSecrets changed",
			oldSource: &DataSource{
				ImagePullSecrets: []string{"old-secret"},
			},
			newSource: &DataSource{
				ImagePullSecrets: []string{"new-secret"},
			},
			wantErr:   true,
			errFields: []string{"ImagePullSecrets"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.newSource.validateUpdate(tt.oldSource, true)
			hasErrs := errs != nil

			if hasErrs != tt.wantErr {
				t.Errorf("validateUpdate() error = %v, wantErr %v", errs, tt.wantErr)
			}

			if hasErrs {
				for _, field := range tt.errFields {
					if !strings.Contains(errs.Error(), field) {
						t.Errorf("validateUpdate() expected errors to contain field %s, but got %s", field, errs.Error())
					}
				}
			}
		})
	}
}

func TestDataDestinationValidateCreate(t *testing.T) {
	tests := []struct {
		name            string
		dataDestination *DataDestination
		wantErr         bool
		errField        string // The field we expect to have an error on
	}{
		{
			name:            "No fields specified",
			dataDestination: &DataDestination{},
			wantErr:         true,
			errField:        "At least one of HostPath or Image must be specified",
		},
		{
			name: "HostPath specified only",
			dataDestination: &DataDestination{
				HostPath: "/data/path",
			},
			wantErr: false,
		},
		{
			name: "Image specified only",
			dataDestination: &DataDestination{
				Image: "data-image:latest",
			},
			wantErr: false,
		},
		{
			name: "Both fields specified",
			dataDestination: &DataDestination{
				HostPath: "/data/path",
				Image:    "data-image:latest",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.dataDestination.validateCreate()
			hasErrs := errs != nil

			if hasErrs != tt.wantErr {
				t.Errorf("validateCreate() error = %v, wantErr %v", errs, tt.wantErr)
			}

			if hasErrs && tt.errField != "" && !strings.Contains(errs.Error(), tt.errField) {
				t.Errorf("validateCreate() expected error to contain %s, but got %s", tt.errField, errs.Error())
			}
		})
	}
}

func TestDataDestinationValidateUpdate(t *testing.T) {
	tests := []struct {
		name      string
		oldDest   *DataDestination
		newDest   *DataDestination
		wantErr   bool
		errFields []string // Fields we expect to have errors
	}{
		{
			name: "No changes",
			oldDest: &DataDestination{
				HostPath:        "/data/old",
				Image:           "old-image:latest",
				ImagePushSecret: "old-secret",
			},
			newDest: &DataDestination{
				HostPath:        "/data/old",
				Image:           "old-image:latest",
				ImagePushSecret: "old-secret",
			},
			wantErr: false,
		},
		{
			name: "HostPath changed",
			oldDest: &DataDestination{
				HostPath: "/data/old",
			},
			newDest: &DataDestination{
				HostPath: "/data/new",
			},
			wantErr:   true,
			errFields: []string{"HostPath"},
		},
		{
			name: "Image changed",
			oldDest: &DataDestination{
				Image: "old-image:latest",
			},
			newDest: &DataDestination{
				Image: "new-image:latest",
			},
			wantErr:   true,
			errFields: []string{"Image"},
		},
		{
			name: "ImagePushSecret changed",
			oldDest: &DataDestination{
				ImagePushSecret: "old-secret",
			},
			newDest: &DataDestination{
				ImagePushSecret: "new-secret",
			},
			wantErr:   true,
			errFields: []string{"ImagePushSecret"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.newDest.validateUpdate(tt.oldDest)
			hasErrs := errs != nil

			if hasErrs != tt.wantErr {
				t.Errorf("validateUpdate() error = %v, wantErr %v", errs, tt.wantErr)
			}

			if hasErrs {
				for _, field := range tt.errFields {
					if !strings.Contains(errs.Error(), field) {
						t.Errorf("validateUpdate() expected errors to contain field %s, but got %s", field, errs.Error())
					}
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

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tuning

import (
	"context"
	"os"
	"testing"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/pkg/model"
	"github.com/azure/kaito/pkg/utils/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/pointer"
)

// Mocking the SupportedGPUConfigs to be used in test scenarios.
var mockSupportedGPUConfigs = map[string]kaitov1alpha1.GPUConfig{
	"sku1": {GPUCount: 2},
	"sku2": {GPUCount: 4},
	"sku3": {GPUCount: 0},
}

func TestGetInstanceGPUCount(t *testing.T) {
	kaitov1alpha1.SupportedGPUConfigs = mockSupportedGPUConfigs
	testcases := map[string]struct {
		sku              string
		expectedGPUCount int
	}{
		"SKU Exists With Multiple GPUs": {
			sku:              "sku1",
			expectedGPUCount: 2,
		},
		"SKU Exists With Zero GPUs": {
			sku:              "sku3",
			expectedGPUCount: 0,
		},
		"SKU Does Not Exist": {
			sku:              "sku_unknown",
			expectedGPUCount: 1,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			result := getInstanceGPUCount(tc.sku)
			assert.Equal(t, tc.expectedGPUCount, result)
		})
	}
}

func TestGetTuningImageInfo(t *testing.T) {
	// Setting up test environment
	originalRegistryName := os.Getenv("PRESET_REGISTRY_NAME")
	defer func() {
		os.Setenv("PRESET_REGISTRY_NAME", originalRegistryName) // Reset after tests
	}()

	testcases := map[string]struct {
		registryName string
		wObj         *kaitov1alpha1.Workspace
		presetObj    *model.PresetParam
		expected     string
	}{
		"Valid Registry and Parameters": {
			registryName: "testregistry",
			wObj: &kaitov1alpha1.Workspace{
				Tuning: &kaitov1alpha1.TuningSpec{
					Preset: &kaitov1alpha1.PresetSpec{
						PresetMeta: kaitov1alpha1.PresetMeta{
							Name: "testpreset",
						},
					},
				},
			},
			presetObj: &model.PresetParam{
				Tag: "latest",
			},
			expected: "testregistry/kaito-tuning-testpreset:latest",
		},
		"Empty Registry Name": {
			registryName: "",
			wObj: &kaitov1alpha1.Workspace{
				Tuning: &kaitov1alpha1.TuningSpec{
					Preset: &kaitov1alpha1.PresetSpec{
						PresetMeta: kaitov1alpha1.PresetMeta{
							Name: "testpreset",
						},
					},
				},
			},
			presetObj: &model.PresetParam{
				Tag: "latest",
			},
			expected: "/kaito-tuning-testpreset:latest",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			os.Setenv("PRESET_REGISTRY_NAME", tc.registryName)
			result := GetTuningImageInfo(context.Background(), tc.wObj, tc.presetObj)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetDataSrcImageInfo(t *testing.T) {
	testcases := map[string]struct {
		wObj            *kaitov1alpha1.Workspace
		expectedImage   string
		expectedSecrets []corev1.LocalObjectReference
	}{
		"Multiple Image Pull Secrets": {
			wObj: &kaitov1alpha1.Workspace{
				Tuning: &kaitov1alpha1.TuningSpec{
					Input: &kaitov1alpha1.DataSource{
						Image:            "kaito/data-source",
						ImagePullSecrets: []string{"secret1", "secret2"},
					},
				},
			},
			expectedImage: "kaito/data-source",
			expectedSecrets: []corev1.LocalObjectReference{
				{Name: "secret1"},
				{Name: "secret2"},
			},
		},
		"No Image Pull Secrets": {
			wObj: &kaitov1alpha1.Workspace{
				Tuning: &kaitov1alpha1.TuningSpec{
					Input: &kaitov1alpha1.DataSource{
						Image: "kaito/data-source",
					},
				},
			},
			expectedImage:   "kaito/data-source",
			expectedSecrets: []corev1.LocalObjectReference{},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			resultImage, resultSecrets := GetDataSrcImageInfo(context.Background(), tc.wObj)
			assert.Equal(t, tc.expectedImage, resultImage)
			assert.Equal(t, tc.expectedSecrets, resultSecrets)
		})
	}
}

func TestEnsureTuningConfigMap(t *testing.T) {
	testcases := map[string]struct {
		callMocks     func(c *test.MockClient)
		workspaceObj  *kaitov1alpha1.Workspace
		expectedError string
	}{
		"Config already exists in workspace namespace": {
			callMocks: func(c *test.MockClient) {
				os.Setenv("RELEASE_NAMESPACE", "release-namespace")
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&corev1.ConfigMap{}), mock.Anything).Return(nil)
			},
			workspaceObj: &kaitov1alpha1.Workspace{
				Tuning: &kaitov1alpha1.TuningSpec{
					ConfigTemplate: "config-template",
				},
			},
			expectedError: "",
		},
		"Error finding release namespace": {
			callMocks: func(c *test.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&corev1.ConfigMap{}), mock.Anything).Return(errors.NewNotFound(schema.GroupResource{}, "config-template"))
			},
			workspaceObj: &kaitov1alpha1.Workspace{
				Tuning: &kaitov1alpha1.TuningSpec{
					ConfigTemplate: "config-template",
				},
			},
			expectedError: "failed to get ConfigMap from template namespace:  \"config-template\" not found",
		},
		"Config doesn't exist in template namespace": {
			callMocks: func(c *test.MockClient) {
				os.Setenv("RELEASE_NAMESPACE", "release-namespace")
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&corev1.ConfigMap{}), mock.Anything).Return(errors.NewNotFound(schema.GroupResource{}, "config-template"))
			},
			workspaceObj: &kaitov1alpha1.Workspace{
				Tuning: &kaitov1alpha1.TuningSpec{
					ConfigTemplate: "config-template",
				},
			},
			expectedError: "failed to get ConfigMap from template namespace:  \"config-template\" not found",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			mockClient := test.NewClient()
			tc.callMocks(mockClient)
			tc.workspaceObj.SetNamespace("workspace-namespace")
			err := EnsureTuningConfigMap(context.Background(), tc.workspaceObj, nil, mockClient)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestHandleImageDataSource(t *testing.T) {
	testcases := map[string]struct {
		workspaceObj              *kaitov1alpha1.Workspace
		expectedInitContainerName string
		expectedVolumeName        string
		expectedVolumeMountPath   string
	}{
		"Handle Image Data Source": {
			workspaceObj: &kaitov1alpha1.Workspace{
				Resource: kaitov1alpha1.ResourceSpec{
					Count: pointer.Int(1),
				},
				Tuning: &kaitov1alpha1.TuningSpec{
					Input: &kaitov1alpha1.DataSource{
						Image: "data-image",
					},
				},
			},
			expectedInitContainerName: "data-extractor",
			expectedVolumeName:        "data-volume",
			expectedVolumeMountPath:   "/mnt/data",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			initContainers, volumes, volumeMounts := handleImageDataSource(context.Background(), tc.workspaceObj)

			assert.Len(t, initContainers, 1)
			assert.Equal(t, tc.expectedInitContainerName, initContainers[0].Name)
			assert.Equal(t, tc.workspaceObj.Tuning.Input.Image, initContainers[0].Image)
			assert.Contains(t, initContainers[0].Command[2], "cp -r /data/* /mnt/data")

			assert.Len(t, volumes, 1)
			assert.Equal(t, tc.expectedVolumeName, volumes[0].Name)

			assert.Len(t, volumeMounts, 1)
			assert.Equal(t, tc.expectedVolumeMountPath, volumeMounts[0].MountPath)
		})
	}
}

func TestPrepareTuningParameters(t *testing.T) {
	ctx := context.TODO()

	testcases := map[string]struct {
		name                 string
		workspaceObj         *kaitov1alpha1.Workspace
		modelCommand         string
		tuningObj            *model.PresetParam
		expectedCommands     []string
		expectedRequirements corev1.ResourceRequirements
	}{
		"Basic Tuning Parameters Setup": {
			workspaceObj: &kaitov1alpha1.Workspace{
				Resource: kaitov1alpha1.ResourceSpec{
					InstanceType: "gpu-instance-type",
				},
			},
			modelCommand: "model-command",
			tuningObj: &model.PresetParam{
				BaseCommand:         "python train.py",
				TorchRunParams:      map[string]string{},
				TorchRunRdzvParams:  map[string]string{},
				GPUCountRequirement: "2",
			},
			expectedCommands: []string{"/bin/sh", "-c", "python train.py --num_processes=1 model-command"},
			expectedRequirements: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceName("nvidia.com/gpu"): resource.MustParse("2"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceName("nvidia.com/gpu"): resource.MustParse("2"),
				},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			commands, resources := prepareTuningParameters(ctx, tc.workspaceObj, tc.modelCommand, tc.tuningObj)
			assert.Equal(t, tc.expectedCommands, commands)
			assert.Equal(t, tc.expectedRequirements.Requests, resources.Requests)
			assert.Equal(t, tc.expectedRequirements.Limits, resources.Limits)
		})
	}
}

func TestPrepareDataSource_ImageSource(t *testing.T) {
	ctx := context.TODO()

	workspaceObj := &kaitov1alpha1.Workspace{
		Tuning: &kaitov1alpha1.TuningSpec{
			Input: &kaitov1alpha1.DataSource{
				Image: "custom/data-loader-image",
			},
		},
	}

	// Expected outputs from mocked functions
	expectedVolumes := []corev1.Volume{
		{
			Name: "data-volume",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{}, // Assume we expect an EmptyDir
			},
		},
	}
	expectedVolumeMounts := []corev1.VolumeMount{{Name: "data-volume", MountPath: "/mnt/data"}}
	expectedImagePullSecrets := []corev1.LocalObjectReference{}
	expectedInitContainers := []corev1.Container{
		{
			Name:         "data-extractor",
			Image:        "custom/data-loader-image",
			Command:      []string{"sh", "-c", "ls -la /data && cp -r /data/* /mnt/data && ls -la /mnt/data"},
			VolumeMounts: expectedVolumeMounts,
		},
	}

	initContainers, imagePullSecrets, volumes, volumeMounts, err := prepareDataSource(ctx, workspaceObj, nil)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedInitContainers, initContainers)
	assert.Equal(t, expectedVolumes, volumes)
	assert.Equal(t, expectedVolumeMounts, volumeMounts)
	assert.Equal(t, expectedImagePullSecrets, imagePullSecrets)
}

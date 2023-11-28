// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package v1alpha1

import "strings"

type GPUConfig struct {
	SKU         string
	SupportedOS []string
	GPUDriver   string
	GPUCount    int
	GPUMem      int
}

type PresetRequirements struct {
	MinGPUCount     int
	MinMemoryPerGPU int // in GB
	MinTotalMemory  int // in GB
}

var PresetRequirementsMap = map[string]PresetRequirements{
	"falcon-7b":           {MinGPUCount: 1, MinMemoryPerGPU: 0, MinTotalMemory: 15},
	"falcon-7b-instruct":  {MinGPUCount: 1, MinMemoryPerGPU: 0, MinTotalMemory: 15},
	"falcon-40b":          {MinGPUCount: 2, MinMemoryPerGPU: 0, MinTotalMemory: 90},
	"falcon-40b-instruct": {MinGPUCount: 2, MinMemoryPerGPU: 0, MinTotalMemory: 90},

	"Mistral-7B-v0.1": {MinGPUCount: 1, MinMemoryPerGPU: 0, MinTotalMemory: 16}
	"Mistral-7B-Instruct-v0.1": {MinGPUCount: 1, MinMemoryPerGPU: 0, MinTotalMemory: 16}

	"llama-2-7b":  {MinGPUCount: 1, MinMemoryPerGPU: 14, MinTotalMemory: 14},
	"llama-2-13b": {MinGPUCount: 2, MinMemoryPerGPU: 15, MinTotalMemory: 30},
	"llama-2-70b": {MinGPUCount: 8, MinMemoryPerGPU: 19, MinTotalMemory: 152},

	"llama-2-7b-chat":  {MinGPUCount: 1, MinMemoryPerGPU: 14, MinTotalMemory: 14},
	"llama-2-13b-chat": {MinGPUCount: 2, MinMemoryPerGPU: 15, MinTotalMemory: 30},
	"llama-2-70b-chat": {MinGPUCount: 8, MinMemoryPerGPU: 19, MinTotalMemory: 152},
}

// Helper function to check if a preset is valid
func isValidPreset(preset string) bool {
	_, exists := PresetRequirementsMap[preset]
	return exists
}

func getSupportedSKUs() string {
	skus := make([]string, 0, len(SupportedGPUConfigs))
	for sku := range SupportedGPUConfigs {
		skus = append(skus, sku)
	}
	return strings.Join(skus, ", ")
}

var SupportedGPUConfigs = map[string]GPUConfig{
	"Standard_NC6":      {SKU: "Standard_NC6", GPUCount: 1, GPUMem: 12, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia470CudaDriver"},
	"Standard_NC12":     {SKU: "Standard_NC12", GPUCount: 2, GPUMem: 24, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia470CudaDriver"},
	"Standard_NC24":     {SKU: "Standard_NC24", GPUCount: 4, GPUMem: 48, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia470CudaDriver"},
	"Standard_NC24r":    {SKU: "Standard_NC24r", GPUCount: 4, GPUMem: 48, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia470CudaDriver"},
	"Standard_NV6":      {SKU: "Standard_NV6", GPUCount: 1, GPUMem: 8, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	"Standard_NV12":     {SKU: "Standard_NV12", GPUCount: 2, GPUMem: 16, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	"Standard_NV24":     {SKU: "Standard_NV24", GPUCount: 4, GPUMem: 32, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	"Standard_NV12s_v3": {SKU: "Standard_NV12s_v3", GPUCount: 1, GPUMem: 8, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	"Standard_NV24s_v3": {SKU: "Standard_NV24s_v3", GPUCount: 2, GPUMem: 16, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	"Standard_NV48s_v3": {SKU: "Standard_NV48s_v3", GPUCount: 4, GPUMem: 32, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	// "Standard_NV24r":     {SKU: "Standard_NV24r", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	"Standard_ND6s":      {SKU: "Standard_ND6s", GPUCount: 1, GPUMem: 24, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"Standard_ND12s":     {SKU: "Standard_ND12s", GPUCount: 2, GPUMem: 48, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"Standard_ND24s":     {SKU: "Standard_ND24s", GPUCount: 4, GPUMem: 96, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"Standard_ND24rs":    {SKU: "Standard_ND24rs", GPUCount: 4, GPUMem: 96, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"Standard_NC6s_v2":   {SKU: "Standard_NC6s_v2", GPUCount: 1, GPUMem: 16, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"Standard_NC12s_v2":  {SKU: "Standard_NC12s_v2", GPUCount: 2, GPUMem: 32, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"Standard_NC24s_v2":  {SKU: "Standard_NC24s_v2", GPUCount: 4, GPUMem: 64, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"Standard_NC24rs_v2": {SKU: "Standard_NC24rs_v2", GPUCount: 4, GPUMem: 64, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"Standard_NC6s_v3":   {SKU: "Standard_NC6s_v3", GPUCount: 1, GPUMem: 16, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"Standard_NC12s_v3":  {SKU: "Standard_NC12s_v3", GPUCount: 2, GPUMem: 32, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"Standard_NC24s_v3":  {SKU: "Standard_NC24s_v3", GPUCount: 4, GPUMem: 64, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"Standard_NC24rs_v3": {SKU: "Standard_NC24rs_v3", GPUCount: 4, GPUMem: 64, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	// "Standard_ND40s_v3":          {SKU: "Standard_ND40s_v3", GPUCount: x, GPUMem: x, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"Standard_ND40rs_v2":    {SKU: "Standard_ND40rs_v2", GPUCount: 8, GPUMem: 256, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"Standard_NC4as_T4_v3":  {SKU: "Standard_NC4as_T4_v3", GPUCount: 1, GPUMem: 16, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"Standard_NC8as_T4_v3":  {SKU: "Standard_NC8as_T4_v3", GPUCount: 1, GPUMem: 16, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"Standard_NC16as_T4_v3": {SKU: "Standard_NC16as_T4_v3", GPUCount: 1, GPUMem: 16, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"Standard_NC64as_T4_v3": {SKU: "Standard_NC64as_T4_v3", GPUCount: 4, GPUMem: 64, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"Standard_ND96asr_v4":   {SKU: "Standard_ND96asr_v4", GPUCount: 8, GPUMem: 320, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	// "Standard_ND112asr_A100_v4":  {SKU: "Standard_ND112asr_A100_v4", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	// "Standard_ND120asr_A100_v4":  {SKU: "Standard_ND120asr_A100_v4", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"Standard_ND96amsr_A100_v4": {SKU: "Standard_ND96amsr_A100_v4", GPUCount: 8, GPUMem: 640, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	// "Standard_ND112amsr_A100_v4": {SKU: "Standard_ND112amsr_A100_v4", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	// "Standard_ND120amsr_A100_v4": {SKU: "Standard_ND120amsr_A100_v4", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"Standard_NC24ads_A100_v4": {SKU: "Standard_NC24ads_A100_v4", GPUCount: 1, GPUMem: 80, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"Standard_NC48ads_A100_v4": {SKU: "Standard_NC48ads_A100_v4", GPUCount: 2, GPUMem: 160, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"Standard_NC96ads_A100_v4": {SKU: "Standard_NC96ads_A100_v4", GPUCount: 4, GPUMem: 320, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	// "Standard_NCads_A100_v4":   {SKU: "Standard_NCads_A100_v4", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	/*GPU Mem based on A10-24 Spec - TODO: Need to confirm GPU Mem*/
	// "Standard_NC8ads_A10_v4":  {SKU: "Standard_NC8ads_A10_v4", GPUCount: 1, GPUMem: 24, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	// "Standard_NC16ads_A10_v4": {SKU: "Standard_NC16ads_A10_v4", GPUCount: 1, GPUMem: 24, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	// "Standard_NC32ads_A10_v4": {SKU: "Standard_NC32ads_A10_v4", GPUCount: 2, GPUMem: 48, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	/* SKUs with GPU Partition are treated as 1 GPU - https://learn.microsoft.com/en-us/azure/virtual-machines/nvA10v5-series*/
	"Standard_NV6ads_A10_v5":   {SKU: "Standard_NV6ads_A10_v5", GPUCount: 1, GPUMem: 4, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	"Standard_NV12ads_A10_v5":  {SKU: "Standard_NV12ads_A10_v5", GPUCount: 1, GPUMem: 8, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	"Standard_NV18ads_A10_v5":  {SKU: "Standard_NV18ads_A10_v5", GPUCount: 1, GPUMem: 12, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	"Standard_NV36ads_A10_v5":  {SKU: "Standard_NV36ads_A10_v5", GPUCount: 1, GPUMem: 24, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	"Standard_NV36adms_A10_v5": {SKU: "Standard_NV36adms_A10_v5", GPUCount: 1, GPUMem: 24, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	"Standard_NV72ads_A10_v5":  {SKU: "Standard_NV72ads_A10_v5", GPUCount: 2, GPUMem: 48, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	// "Standard_ND96ams_v4":      {SKU: "Standard_ND96ams_v4", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	// "Standard_ND96ams_A100_v4": {SKU: "Standard_ND96ams_A100_v4", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
}

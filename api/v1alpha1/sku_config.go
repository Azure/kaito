/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	var skus []string
	for sku := range SupportedGPUConfigs {
		skus = append(skus, sku)
	}
	return strings.Join(skus, ", ")
}

var SupportedGPUConfigs = map[string]GPUConfig{
	"standard_nc6":      {SKU: "standard_nc6", GPUCount: 1, GPUMem: 12, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia470CudaDriver"},
	"standard_nc12":     {SKU: "standard_nc12", GPUCount: 2, GPUMem: 24, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia470CudaDriver"},
	"standard_nc24":     {SKU: "standard_nc24", GPUCount: 4, GPUMem: 48, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia470CudaDriver"},
	"standard_nc24r":    {SKU: "standard_nc24r", GPUCount: 4, GPUMem: 48, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia470CudaDriver"},
	"standard_nv6":      {SKU: "standard_nv6", GPUCount: 1, GPUMem: 8, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	"standard_nv12":     {SKU: "standard_nv12", GPUCount: 2, GPUMem: 16, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	"standard_nv24":     {SKU: "standard_nv24", GPUCount: 4, GPUMem: 32, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	"standard_nv12s_v3": {SKU: "standard_nv12s_v3", GPUCount: 1, GPUMem: 8, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	"standard_nv24s_v3": {SKU: "standard_nv24s_v3", GPUCount: 2, GPUMem: 16, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	"standard_nv48s_v3": {SKU: "standard_nv48s_v3", GPUCount: 4, GPUMem: 32, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	// "standard_nv24r":     {SKU: "standard_nv24r", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	"standard_nd6s":      {SKU: "standard_nd6s", GPUCount: 1, GPUMem: 24, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"standard_nd12s":     {SKU: "standard_nd12s", GPUCount: 2, GPUMem: 48, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"standard_nd24s":     {SKU: "standard_nd24s", GPUCount: 4, GPUMem: 96, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"standard_nd24rs":    {SKU: "standard_nd24rs", GPUCount: 4, GPUMem: 96, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"standard_nc6s_v2":   {SKU: "standard_nc6s_v2", GPUCount: 1, GPUMem: 16, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"standard_nc12s_v2":  {SKU: "standard_nc12s_v2", GPUCount: 2, GPUMem: 32, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"standard_nc24s_v2":  {SKU: "standard_nc24s_v2", GPUCount: 4, GPUMem: 64, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"standard_nc24rs_v2": {SKU: "standard_nc24rs_v2", GPUCount: 4, GPUMem: 64, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"standard_nc6s_v3":   {SKU: "standard_nc6s_v3", GPUCount: 1, GPUMem: 16, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"standard_nc12s_v3":  {SKU: "standard_nc12s_v3", GPUCount: 2, GPUMem: 32, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"standard_nc24s_v3":  {SKU: "standard_nc24s_v3", GPUCount: 4, GPUMem: 64, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"standard_nc24rs_v3": {SKU: "standard_nc24rs_v3", GPUCount: 4, GPUMem: 64, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	// "standard_nd40s_v3":          {SKU: "standard_nd40s_v3", GPUCount: x, GPUMem: x, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"standard_nd40rs_v2":    {SKU: "standard_nd40rs_v2", GPUCount: 8, GPUMem: 256, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"standard_nc4as_t4_v3":  {SKU: "standard_nc4as_t4_v3", GPUCount: 1, GPUMem: 16, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"standard_nc8as_t4_v3":  {SKU: "standard_nc8as_t4_v3", GPUCount: 1, GPUMem: 16, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"standard_nc16as_t4_v3": {SKU: "standard_nc16as_t4_v3", GPUCount: 1, GPUMem: 16, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"standard_nc64as_t4_v3": {SKU: "standard_nc64as_t4_v3", GPUCount: 4, GPUMem: 64, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"standard_nd96asr_v4":   {SKU: "standard_nd96asr_v4", GPUCount: 8, GPUMem: 320, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	// "standard_nd112asr_a100_v4":  {SKU: "standard_nd112asr_a100_v4", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	// "standard_nd120asr_a100_v4":  {SKU: "standard_nd120asr_a100_v4", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"standard_nd96amsr_a100_v4": {SKU: "standard_nd96amsr_a100_v4", GPUCount: 8, GPUMem: 640, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	// "standard_nd112amsr_a100_v4": {SKU: "standard_nd112amsr_a100_v4", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	// "standard_nd120amsr_a100_v4": {SKU: "standard_nd120amsr_a100_v4", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"standard_nc24ads_a100_v4": {SKU: "standard_nc24ads_a100_v4", GPUCount: 1, GPUMem: 80, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"standard_nc48ads_a100_v4": {SKU: "standard_nc48ads_a100_v4", GPUCount: 2, GPUMem: 160, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	"standard_nc96ads_a100_v4": {SKU: "standard_nc96ads_a100_v4", GPUCount: 4, GPUMem: 320, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	// "standard_ncads_a100_v4":   {SKU: "standard_ncads_a100_v4", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	/*GPU Mem based on A10-24 Spec - TODO: Need to confirm GPU Mem*/
	// "standard_nc8ads_a10_v4":  {SKU: "standard_nc8ads_a10_v4", GPUCount: 1, GPUMem: 24, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	// "standard_nc16ads_a10_v4": {SKU: "standard_nc16ads_a10_v4", GPUCount: 1, GPUMem: 24, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	// "standard_nc32ads_a10_v4": {SKU: "standard_nc32ads_a10_v4", GPUCount: 2, GPUMem: 48, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	/* SKUs with GPU Partition are treated as 1 GPU - https://learn.microsoft.com/en-us/azure/virtual-machines/nva10v5-series*/
	"standard_nv6ads_a10_v5":   {SKU: "standard_nv6ads_a10_v5", GPUCount: 1, GPUMem: 4, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	"standard_nv12ads_a10_v5":  {SKU: "standard_nv12ads_a10_v5", GPUCount: 1, GPUMem: 8, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	"standard_nv18ads_a10_v5":  {SKU: "standard_nv18ads_a10_v5", GPUCount: 1, GPUMem: 12, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	"standard_nv36ads_a10_v5":  {SKU: "standard_nv36ads_a10_v5", GPUCount: 1, GPUMem: 24, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	"standard_nv36adms_a10_v5": {SKU: "standard_nv36adms_a10_v5", GPUCount: 1, GPUMem: 24, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	"standard_nv72ads_a10_v5":  {SKU: "standard_nv72ads_a10_v5", GPUCount: 2, GPUMem: 48, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver"},
	// "standard_nd96ams_v4":      {SKU: "standard_nd96ams_v4", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
	// "standard_nd96ams_a100_v4": {SKU: "standard_nd96ams_a100_v4", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver"},
}

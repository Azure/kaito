// sku_config.go

package v1alpha1

type GPUConfig struct {
	SKU         string
	SupportedOS []string
	GPUDriver   string
	GPUCount    int
	GPUMem      int
	Counts      map[string]int
}

// List of supported preset names
var validPresets = map[string]bool{
	"falcon-7b":           true,
	"falcon-7b-instruct":  true,
	"falcon-40b":          true,
	"falcon-40b-instruct": true,
	"llama-2-7b":          true,
	"llama-2-13b":         true,
	"llama-2-70b":         true,
	"llama-2-7b-chat":     true,
	"llama-2-13b-chat":    true,
	"llama-2-70b-chat":    true,
}

// Helper function to check if a preset is valid
func isValidPreset(preset string) bool {
	return validPresets[preset]
}

var SupportedGPUConfigs = map[string]GPUConfig{
	"standard_nc6":      {SKU: "standard_nc6", GPUCount: 1, GPUMem: 12, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia470CudaDriver", Counts: map[string]int{}},
	"standard_nc12":     {SKU: "standard_nc12", GPUCount: 2, GPUMem: 24, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia470CudaDriver", Counts: map[string]int{}},
	"standard_nc24":     {SKU: "standard_nc24", GPUCount: 4, GPUMem: 48, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia470CudaDriver", Counts: map[string]int{}},
	"standard_nc24r":    {SKU: "standard_nc24r", GPUCount: 4, GPUMem: 48, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia470CudaDriver", Counts: map[string]int{}},
	"standard_nv6":      {SKU: "standard_nv6", GPUCount: 1, GPUMem: 8, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nv12":     {SKU: "standard_nv12", GPUCount: 2, GPUMem: 16, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nv24":     {SKU: "standard_nv24", GPUCount: 4, GPUMem: 32, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nv12s_v3": {SKU: "standard_nv12s_v3", GPUCount: 1, GPUMem: 8, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nv24s_v3": {SKU: "standard_nv24s_v3", GPUCount: 2, GPUMem: 16, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nv48s_v3": {SKU: "standard_nv48s_v3", GPUCount: 4, GPUMem: 32, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	// "standard_nv24r":     {SKU: "standard_nv24r", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nd6s":      {SKU: "standard_nd6s", GPUCount: 1, GPUMem: 24, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nd12s":     {SKU: "standard_nd12s", GPUCount: 2, GPUMem: 48, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nd24s":     {SKU: "standard_nd24s", GPUCount: 4, GPUMem: 96, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nd24rs":    {SKU: "standard_nd24rs", GPUCount: 4, GPUMem: 96, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc6s_v2":   {SKU: "standard_nc6s_v2", GPUCount: 1, GPUMem: 16, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc12s_v2":  {SKU: "standard_nc12s_v2", GPUCount: 2, GPUMem: 32, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc24s_v2":  {SKU: "standard_nc24s_v2", GPUCount: 4, GPUMem: 64, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc24rs_v2": {SKU: "standard_nc24rs_v2", GPUCount: 4, GPUMem: 64, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc6s_v3":   {SKU: "standard_nc6s_v3", GPUCount: 1, GPUMem: 16, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc12s_v3":  {SKU: "standard_nc12s_v3", GPUCount: 2, GPUMem: 32, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc24s_v3":  {SKU: "standard_nc24s_v3", GPUCount: 4, GPUMem: 64, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc24rs_v3": {SKU: "standard_nc24rs_v3", GPUCount: 4, GPUMem: 64, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	// "standard_nd40s_v3":          {SKU: "standard_nd40s_v3", GPUCount: x, GPUMem: x, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nd40rs_v2":    {SKU: "standard_nd40rs_v2", GPUCount: 8, GPUMem: 256, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc4as_t4_v3":  {SKU: "standard_nc4as_t4_v3", GPUCount: 1, GPUMem: 16, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc8as_t4_v3":  {SKU: "standard_nc8as_t4_v3", GPUCount: 1, GPUMem: 16, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc16as_t4_v3": {SKU: "standard_nc16as_t4_v3", GPUCount: 1, GPUMem: 16, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc64as_t4_v3": {SKU: "standard_nc64as_t4_v3", GPUCount: 4, GPUMem: 64, SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nd96asr_v4":   {SKU: "standard_nd96asr_v4", GPUCount: 8, GPUMem: 320, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	// "standard_nd112asr_a100_v4":  {SKU: "standard_nd112asr_a100_v4", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	// "standard_nd120asr_a100_v4":  {SKU: "standard_nd120asr_a100_v4", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nd96amsr_a100_v4": {SKU: "standard_nd96amsr_a100_v4", GPUCount: 8, GPUMem: 640, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	// "standard_nd112amsr_a100_v4": {SKU: "standard_nd112amsr_a100_v4", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	// "standard_nd120amsr_a100_v4": {SKU: "standard_nd120amsr_a100_v4", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc24ads_a100_v4": {SKU: "standard_nc24ads_a100_v4", GPUCount: 1, GPUMem: 80, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc48ads_a100_v4": {SKU: "standard_nc48ads_a100_v4", GPUCount: 2, GPUMem: 160, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc96ads_a100_v4": {SKU: "standard_nc96ads_a100_v4", GPUCount: 4, GPUMem: 320, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	// "standard_ncads_a100_v4":   {SKU: "standard_ncads_a100_v4", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	/*GPU Mem based on A10-24 Spec*/
	"standard_nc8ads_a10_v4":  {SKU: "standard_nc8ads_a10_v4", GPUCount: 1, GPUMem: 24, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nc16ads_a10_v4": {SKU: "standard_nc16ads_a10_v4", GPUCount: 1, GPUMem: 24, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nc32ads_a10_v4": {SKU: "standard_nc32ads_a10_v4", GPUCount: 2, GPUMem: 48, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	/* SKUs with GPU Partition are treated as 1 GPU - https://learn.microsoft.com/en-us/azure/virtual-machines/nva10v5-series*/
	"standard_nv6ads_a10_v5":   {SKU: "standard_nv6ads_a10_v5", GPUCount: 1, GPUMem: 4, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nv12ads_a10_v5":  {SKU: "standard_nv12ads_a10_v5", GPUCount: 1, GPUMem: 8, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nv18ads_a10_v5":  {SKU: "standard_nv18ads_a10_v5", GPUCount: 1, GPUMem: 12, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nv36ads_a10_v5":  {SKU: "standard_nv36ads_a10_v5", GPUCount: 1, GPUMem: 24, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nv36adms_a10_v5": {SKU: "standard_nv36adms_a10_v5", GPUCount: 1, GPUMem: 24, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nv72ads_a10_v5":  {SKU: "standard_nv72ads_a10_v5", GPUCount: 2, GPUMem: 48, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	// "standard_nd96ams_v4":      {SKU: "standard_nd96ams_v4", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	// "standard_nd96ams_a100_v4": {SKU: "standard_nd96ams_a100_v4", GPUCount: x, GPUMem: x, SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
}

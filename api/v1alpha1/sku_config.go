// sku_config.go

package v1alpha1

type GPUConfig struct {
	SKU         string
	SupportedOS []string
	GPUDriver   string
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
	"standard_nc6":               {SKU: "standard_nc6", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia470CudaDriver", Counts: map[string]int{}},
	"standard_nc12":              {SKU: "standard_nc12", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia470CudaDriver", Counts: map[string]int{}},
	"standard_nc24":              {SKU: "standard_nc24", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia470CudaDriver", Counts: map[string]int{}},
	"standard_nc24r":             {SKU: "standard_nc24r", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia470CudaDriver", Counts: map[string]int{}},
	"standard_nv6":               {SKU: "standard_nv6", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nv12":              {SKU: "standard_nv12", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nv12s_v3":          {SKU: "standard_nv12s_v3", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nv24":              {SKU: "standard_nv24", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nv24s_v3":          {SKU: "standard_nv24s_v3", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nv24r":             {SKU: "standard_nv24r", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nv48s_v3":          {SKU: "standard_nv48s_v3", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nd6s":              {SKU: "standard_nd6s", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nd12s":             {SKU: "standard_nd12s", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nd24s":             {SKU: "standard_nd24s", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nd24rs":            {SKU: "standard_nd24rs", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc6s_v2":           {SKU: "standard_nc6s_v2", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc12s_v2":          {SKU: "standard_nc12s_v2", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc24s_v2":          {SKU: "standard_nc24s_v2", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc24rs_v2":         {SKU: "standard_nc24rs_v2", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc6s_v3":           {SKU: "standard_nc6s_v3", SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc12s_v3":          {SKU: "standard_nc12s_v3", SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc24s_v3":          {SKU: "standard_nc24s_v3", SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc24rs_v3":         {SKU: "standard_nc24rs_v3", SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nd40s_v3":          {SKU: "standard_nd40s_v3", SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nd40rs_v2":         {SKU: "standard_nd40rs_v2", SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc4as_t4_v3":       {SKU: "standard_nc4as_t4_v3", SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc8as_t4_v3":       {SKU: "standard_nc8as_t4_v3", SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc16as_t4_v3":      {SKU: "standard_nc16as_t4_v3", SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc64as_t4_v3":      {SKU: "standard_nc64as_t4_v3", SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nd96asr_v4":        {SKU: "standard_nd96asr_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nd112asr_a100_v4":  {SKU: "standard_nd112asr_a100_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nd120asr_a100_v4":  {SKU: "standard_nd120asr_a100_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nd96amsr_a100_v4":  {SKU: "standard_nd96amsr_a100_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nd112amsr_a100_v4": {SKU: "standard_nd112amsr_a100_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nd120amsr_a100_v4": {SKU: "standard_nd120amsr_a100_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc24ads_a100_v4":   {SKU: "standard_nc24ads_a100_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc48ads_a100_v4":   {SKU: "standard_nc48ads_a100_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc96ads_a100_v4":   {SKU: "standard_nc96ads_a100_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_ncads_a100_v4":     {SKU: "standard_ncads_a100_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nc8ads_a10_v4":     {SKU: "standard_nc8ads_a10_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nc16ads_a10_v4":    {SKU: "standard_nc16ads_a10_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nc32ads_a10_v4":    {SKU: "standard_nc32ads_a10_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nv6ads_a10_v5":     {SKU: "standard_nv6ads_a10_v5", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nv12ads_a10_v5":    {SKU: "standard_nv12ads_a10_v5", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nv18ads_a10_v5":    {SKU: "standard_nv18ads_a10_v5", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nv36ads_a10_v5":    {SKU: "standard_nv36ads_a10_v5", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nv36adms_a10_v5":   {SKU: "standard_nv36adms_a10_v5", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nv72ads_a10_v5":    {SKU: "standard_nv72ads_a10_v5", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Counts: map[string]int{}},
	"standard_nd96ams_v4":        {SKU: "standard_nd96ams_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
	"standard_nd96ams_a100_v4":   {SKU: "standard_nd96ams_a100_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Counts: map[string]int{}},
}

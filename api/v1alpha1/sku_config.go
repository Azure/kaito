// sku_config.go

package v1alpha1

type GPUConfig struct {
	SKU         string
	SupportedOS []string
	GPUDriver   string
	Count       int
}

var SupportedGPUConfigs = map[string]GPUConfig{
	"standard_nc6":               {SKU: "standard_nc6", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia470CudaDriver", Count: 0},
	"standard_nc12":              {SKU: "standard_nc12", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia470CudaDriver", Count: 0},
	"standard_nc24":              {SKU: "standard_nc24", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia470CudaDriver", Count: 0},
	"standard_nc24r":             {SKU: "standard_nc24r", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia470CudaDriver", Count: 0},
	"standard_nv6":               {SKU: "standard_nv6", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Count: 0},
	"standard_nv12":              {SKU: "standard_nv12", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Count: 0},
	"standard_nv12s_v3":          {SKU: "standard_nv12s_v3", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Count: 0},
	"standard_nv24":              {SKU: "standard_nv24", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Count: 0},
	"standard_nv24s_v3":          {SKU: "standard_nv24s_v3", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Count: 0},
	"standard_nv24r":             {SKU: "standard_nv24r", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Count: 0},
	"standard_nv48s_v3":          {SKU: "standard_nv48s_v3", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Count: 0},
	"standard_nd6s":              {SKU: "standard_nd6s", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nd12s":             {SKU: "standard_nd12s", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nd24s":             {SKU: "standard_nd24s", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nd24rs":            {SKU: "standard_nd24rs", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nc6s_v2":           {SKU: "standard_nc6s_v2", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nc12s_v2":          {SKU: "standard_nc12s_v2", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nc24s_v2":          {SKU: "standard_nc24s_v2", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nc24rs_v2":         {SKU: "standard_nc24rs_v2", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nc6s_v3":           {SKU: "standard_nc6s_v3", SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nc12s_v3":          {SKU: "standard_nc12s_v3", SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nc24s_v3":          {SKU: "standard_nc24s_v3", SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nc24rs_v3":         {SKU: "standard_nc24rs_v3", SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nd40s_v3":          {SKU: "standard_nd40s_v3", SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nd40rs_v2":         {SKU: "standard_nd40rs_v2", SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nc4as_t4_v3":       {SKU: "standard_nc4as_t4_v3", SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nc8as_t4_v3":       {SKU: "standard_nc8as_t4_v3", SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nc16as_t4_v3":      {SKU: "standard_nc16as_t4_v3", SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nc64as_t4_v3":      {SKU: "standard_nc64as_t4_v3", SupportedOS: []string{"Mariner", "Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nd96asr_v4":        {SKU: "standard_nd96asr_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nd112asr_a100_v4":  {SKU: "standard_nd112asr_a100_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nd120asr_a100_v4":  {SKU: "standard_nd120asr_a100_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nd96amsr_a100_v4":  {SKU: "standard_nd96amsr_a100_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nd112amsr_a100_v4": {SKU: "standard_nd112amsr_a100_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nd120amsr_a100_v4": {SKU: "standard_nd120amsr_a100_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nc24ads_a100_v4":   {SKU: "standard_nc24ads_a100_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nc48ads_a100_v4":   {SKU: "standard_nc48ads_a100_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nc96ads_a100_v4":   {SKU: "standard_nc96ads_a100_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_ncads_a100_v4":     {SKU: "standard_ncads_a100_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nc8ads_a10_v4":     {SKU: "standard_nc8ads_a10_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Count: 0},
	"standard_nc16ads_a10_v4":    {SKU: "standard_nc16ads_a10_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Count: 0},
	"standard_nc32ads_a10_v4":    {SKU: "standard_nc32ads_a10_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Count: 0},
	"standard_nv6ads_a10_v5":     {SKU: "standard_nv6ads_a10_v5", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Count: 0},
	"standard_nv12ads_a10_v5":    {SKU: "standard_nv12ads_a10_v5", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Count: 0},
	"standard_nv18ads_a10_v5":    {SKU: "standard_nv18ads_a10_v5", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Count: 0},
	"standard_nv36ads_a10_v5":    {SKU: "standard_nv36ads_a10_v5", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Count: 0},
	"standard_nv36adms_a10_v5":   {SKU: "standard_nv36adms_a10_v5", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Count: 0},
	"standard_nv72ads_a10_v5":    {SKU: "standard_nv72ads_a10_v5", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia510GridDriver", Count: 0},
	"standard_nd96ams_v4":        {SKU: "standard_nd96ams_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
	"standard_nd96ams_a100_v4":   {SKU: "standard_nd96ams_a100_v4", SupportedOS: []string{"Ubuntu"}, GPUDriver: "Nvidia525CudaDriver", Count: 0},
}

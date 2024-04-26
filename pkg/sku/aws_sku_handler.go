// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sku

var _ CloudSKUHandler = &AwsSKUHandler{}

type AwsSKUHandler struct {
	supportedSKUs map[string]GPUConfig
}

func NewAwsSKUHandler() *AwsSKUHandler {
	return &AwsSKUHandler{
		// Reference: https://aws.amazon.com/ec2/instance-types/
		supportedSKUs: map[string]GPUConfig{
			"p2.xlarge":     {SKU: "p2.xlarge", GPUCount: 1, GPUMem: 12, GPUModel: "NVIDIA K80"},
			"p2.8xlarge":    {SKU: "p2.8xlarge", GPUCount: 8, GPUMem: 96, GPUModel: "NVIDIA K80"},
			"p2.16xlarge":   {SKU: "p2.16xlarge", GPUCount: 16, GPUMem: 192, GPUModel: "NVIDIA K80"},
			"p3.2xlarge":    {SKU: "p3.2xlarge", GPUCount: 1, GPUMem: 16, GPUModel: "NVIDIA V100"},
			"p3.8xlarge":    {SKU: "p3.8xlarge", GPUCount: 4, GPUMem: 64, GPUModel: "NVIDIA V100"},
			"p3.16xlarge":   {SKU: "p3.16xlarge", GPUCount: 8, GPUMem: 128, GPUModel: "NVIDIA V100"},
			"p3dn.24xlarge": {SKU: "p3dn.24xlarge", GPUCount: 8, GPUMem: 256, GPUModel: "NVIDIA V100"},
			"p4d.24xlarge":  {SKU: "p4d.24xlarge", GPUCount: 8, GPUMem: 320, GPUModel: "NVIDIA A100"},
			"p4de.24xlarge": {SKU: "p4de.24xlarge", GPUCount: 8, GPUMem: 640, GPUModel: "NVIDIA A100"},
			"p5.48xlarge":   {SKU: "p5.48xlarge", GPUCount: 8, GPUMem: 640, GPUModel: "NVIDIA H100"},
			"g6.xlarge":     {SKU: "g6.xlarge", GPUCount: 1, GPUMem: 24, GPUModel: "NVIDIA L4"},
			"g6.2xlarge":    {SKU: "g6.2xlarge", GPUCount: 1, GPUMem: 24, GPUModel: "NVIDIA L4"},
			"g6.4xlarge":    {SKU: "g6.4xlarge", GPUCount: 1, GPUMem: 24, GPUModel: "NVIDIA L4"},
			"g6.8xlarge":    {SKU: "g6.8xlarge", GPUCount: 1, GPUMem: 24, GPUModel: "NVIDIA L4"},
			"g6.16xlarge":   {SKU: "g6.16xlarge", GPUCount: 1, GPUMem: 24, GPUModel: "NVIDIA L4"},
			"gr6.4xlarge":   {SKU: "gr6.4xlarge", GPUCount: 1, GPUMem: 24, GPUModel: "NVIDIA L4"},
			"gr6.8xlarge":   {SKU: "gr6.8xlarge", GPUCount: 1, GPUMem: 24, GPUModel: "NVIDIA L4"},
			"g6.12xlarge":   {SKU: "g6.12xlarge", GPUCount: 4, GPUMem: 96, GPUModel: "NVIDIA L4"},
			"g6.24xlarge":   {SKU: "g6.24xlarge", GPUCount: 4, GPUMem: 96, GPUModel: "NVIDIA L4"},
			"g6.48xlarge":   {SKU: "g6.48xlarge", GPUCount: 8, GPUMem: 192, GPUModel: "NVIDIA L4"},
			"g5g.xlarge":    {SKU: "g5g.xlarge", GPUCount: 1, GPUMem: 16, GPUModel: "NVIDIA T4"},
			"g5g.2xlarge":   {SKU: "g5g.2xlarge", GPUCount: 1, GPUMem: 16, GPUModel: "NVIDIA T4"},
			"g5g.4xlarge":   {SKU: "g5g.4xlarge", GPUCount: 1, GPUMem: 16, GPUModel: "NVIDIA T4"},
			"g5g.8xlarge":   {SKU: "g5g.8xlarge", GPUCount: 1, GPUMem: 16, GPUModel: "NVIDIA T4"},
			"g5g.16xlarge":  {SKU: "g5g.16xlarge", GPUCount: 2, GPUMem: 32, GPUModel: "NVIDIA T4"},
			"g5g.metal":     {SKU: "g5g.metal", GPUCount: 2, GPUMem: 32, GPUModel: "NVIDIA T4"},
			"g5.xlarge":     {SKU: "g5.xlarge", GPUCount: 1, GPUMem: 24, GPUModel: "NVIDIA A10G"},
			"g5.2xlarge":    {SKU: "g5.2xlarge", GPUCount: 1, GPUMem: 24, GPUModel: "NVIDIA A10G"},
			"g5.4xlarge":    {SKU: "g5.4xlarge", GPUCount: 1, GPUMem: 24, GPUModel: "NVIDIA A10G"},
			"g5.8xlarge":    {SKU: "g5.8xlarge", GPUCount: 1, GPUMem: 24, GPUModel: "NVIDIA A10G"},
			"g5.12xlarge":   {SKU: "g5.12xlarge", GPUCount: 4, GPUMem: 96, GPUModel: "NVIDIA A10G"},
			"g5.16xlarge":   {SKU: "g5.16xlarge", GPUCount: 1, GPUMem: 24, GPUModel: "NVIDIA A10G"},
			"g5.24xlarge":   {SKU: "g5.24xlarge", GPUCount: 4, GPUMem: 96, GPUModel: "NVIDIA A10G"},
			"g5.48xlarge":   {SKU: "g5.48xlarge", GPUCount: 8, GPUMem: 192, GPUModel: "NVIDIA A10G"},
			"g4dn.xlarge":   {SKU: "g4dn.xlarge", GPUCount: 1, GPUMem: 16, GPUModel: "NVIDIA T4"},
			"g4dn.2xlarge":  {SKU: "g4dn.2xlarge", GPUCount: 1, GPUMem: 16, GPUModel: "NVIDIA T4"},
			"g4dn.4xlarge":  {SKU: "g4dn.4xlarge", GPUCount: 1, GPUMem: 16, GPUModel: "NVIDIA T4"},
			"g4dn.8xlarge":  {SKU: "g4dn.8xlarge", GPUCount: 1, GPUMem: 16, GPUModel: "NVIDIA T4"},
			"g4dn.16xlarge": {SKU: "g4dn.16xlarge", GPUCount: 1, GPUMem: 16, GPUModel: "NVIDIA T4"},
			"g4dn.12xlarge": {SKU: "g4dn.12xlarge", GPUCount: 4, GPUMem: 64, GPUModel: "NVIDIA T4"},
			"g4dn.metal":    {SKU: "g4dn.metal", GPUCount: 8, GPUMem: 128, GPUModel: "NVIDIA T4"},
			"g4ad.xlarge":   {SKU: "g4ad.xlarge", GPUCount: 1, GPUMem: 8, GPUModel: "AMD Radeon Pro V520"},
			"g4ad.2xlarge":  {SKU: "g4ad.2xlarge", GPUCount: 1, GPUMem: 8, GPUModel: "AMD Radeon Pro V520"},
			"g4ad.4xlarge":  {SKU: "g4ad.4xlarge", GPUCount: 1, GPUMem: 8, GPUModel: "AMD Radeon Pro V520"},
			"g4ad.8xlarge":  {SKU: "g4ad.8xlarge", GPUCount: 2, GPUMem: 16, GPUModel: "AMD Radeon Pro V520"},
			"g4ad.16xlarge": {SKU: "g4ad.16xlarge", GPUCount: 4, GPUMem: 32, GPUModel: "AMD Radeon Pro V520"},
			"g3s.xlarge":    {SKU: "g3s.xlarge", GPUCount: 1, GPUMem: 8, GPUModel: "NVIDIA M60"},
			"g3s.4xlarge":   {SKU: "g3s.4xlarge", GPUCount: 1, GPUMem: 8, GPUModel: "NVIDIA M60"},
			"g3s.8xlarge":   {SKU: "g3s.8xlarge", GPUCount: 2, GPUMem: 16, GPUModel: "NVIDIA M60"},
			"g3s.16xlarge":  {SKU: "g3s.16xlarge", GPUCount: 4, GPUMem: 32, GPUModel: "NVIDIA M60"},
			//accelerator optimized
			"trn1.2xlarge":   {SKU: "trn1.2xlarge", GPUCount: 1, GPUMem: 32, GPUModel: "AWS Trainium accelerators"},
			"trn1.32xlarge":  {SKU: "trn1.32xlarge", GPUCount: 16, GPUMem: 512, GPUModel: "AWS Trainium accelerators"},
			"trn1n.32xlarge": {SKU: "trn1n.32xlarge", GPUCount: 16, GPUMem: 512, GPUModel: "AWS Trainium accelerators"},
			"inf2.xlarge":    {SKU: "inf2.xlarge", GPUCount: 1, GPUMem: 32, GPUModel: "AWS Inferentia2 accelerators"},
			"inf2.8xlarge":   {SKU: "inf2.8xlarge", GPUCount: 1, GPUMem: 32, GPUModel: "AWS Inferentia2 accelerators"},
			"inf2.24xlarge":  {SKU: "inf2.24xlarge", GPUCount: 6, GPUMem: 192, GPUModel: "AWS Inferentia2 accelerators"},
			"inf2.48xlarge":  {SKU: "inf2.48xlarge", GPUCount: 12, GPUMem: 384, GPUModel: "AWS Inferentia2 accelerators"},
		},
	}
}

func (a *AwsSKUHandler) GetSupportedSKUs() []string {
	return GetMapKeys(a.supportedSKUs)
}

func (a *AwsSKUHandler) GetGPUConfigs() map[string]GPUConfig {
	return a.supportedSKUs
}

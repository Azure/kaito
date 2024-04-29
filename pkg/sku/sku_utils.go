// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sku

func GetMapKeys(m map[string]GPUConfig) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

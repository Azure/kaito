// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package utils

import (
	"fmt"
	"github.com/aws/karpenter-core/pkg/apis/v1alpha5"
	"github.com/azure/kaito/pkg/sku"
	"github.com/azure/kaito/pkg/utils/consts"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/pkg/apis"
	"os"
)

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// SearchMap performs a search for a key in a map[string]interface{}.
func SearchMap(m map[string]interface{}, key string) (value interface{}, exists bool) {
	if val, ok := m[key]; ok {
		return val, true
	}
	return nil, false
}

// SearchRawExtension performs a search for a key within a runtime.RawExtension.
func SearchRawExtension(raw runtime.RawExtension, key string) (interface{}, bool, error) {
	var data map[string]interface{}
	if err := yaml.Unmarshal(raw.Raw, &data); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal runtime.RawExtension: %w", err)
	}

	result, found := data[key]
	if !found {
		return nil, false, nil
	}

	return result, true, nil
}

func MergeConfigMaps(baseMap, overrideMap map[string]string) map[string]string {
	merged := make(map[string]string)
	for k, v := range baseMap {
		merged[k] = v
	}

	// Override with values from overrideMap
	for k, v := range overrideMap {
		merged[k] = v
	}

	return merged
}

func BuildCmdStr(baseCommand string, runParams map[string]string) string {
	updatedBaseCommand := baseCommand
	for key, value := range runParams {
		if value == "" {
			updatedBaseCommand = fmt.Sprintf("%s --%s", updatedBaseCommand, key)
		} else {
			updatedBaseCommand = fmt.Sprintf("%s --%s=%s", updatedBaseCommand, key, value)
		}
	}

	return updatedBaseCommand
}

func ShellCmd(command string) []string {
	return []string{
		"/bin/sh",
		"-c",
		command,
	}
}

func GetReleaseNamespace() (string, error) {
	// Path to the namespace file inside a Kubernetes pod
	namespaceFilePath := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

	// Attempt to read the namespace from the file
	if content, err := ioutil.ReadFile(namespaceFilePath); err == nil {
		return string(content), nil
	}

	// Fallback: Read the namespace from an environment variable
	if namespace, exists := os.LookupEnv(consts.DefaultReleaseNamespaceEnvVar); exists {
		return namespace, nil
	}
	return "", fmt.Errorf("failed to determine release namespace from file %s and env var %s", namespaceFilePath, consts.DefaultReleaseNamespaceEnvVar)
}

func GetSKUHandler() (sku.CloudSKUHandler, error) {
	// Get the cloud provider from the environment
	provider := os.Getenv("CLOUD_PROVIDER")

	if provider == "" {
		return nil, apis.ErrMissingField("CLOUD_PROVIDER environment variable must be set")
	}
	// Select the correct SKU handler based on the cloud provider
	skuHandler := sku.GetCloudSKUHandler(provider)
	if skuHandler == nil {
		return nil, apis.ErrInvalidValue(fmt.Sprintf("Unsupported cloud provider %s", provider), "CLOUD_PROVIDER")
	}

	return skuHandler, nil
}

func GetGPUCountFromWorkspaceMachines(machineList *v1alpha5.MachineList) string {
	// Iterate through the Machine objects and get the GPU count from capacity
	for _, item := range machineList.Items {
		capacity := item.Status.Capacity
		gpuCountStr, exists := capacity["nvidia.com/gpu"]
		if !exists {
			fmt.Printf("Failed to find GPU capacity in Machine object %v", item)
			continue
		}

		fmt.Printf("Detected SKU GPU Count. Number of GPUs available: %s\n", gpuCountStr)
		return gpuCountStr.String()
	}
	return ""
}

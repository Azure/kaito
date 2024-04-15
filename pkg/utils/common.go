// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package utils

import (
	"fmt"
	"io/ioutil"
	"os"
)

const (
	// WorkspaceFinalizer is used to make sure that workspace controller handles garbage collection.
	WorkspaceFinalizer            = "workspace.finalizer.kaito.sh"
	DefaultReleaseNamespaceEnvVar = "RELEASE_NAMESPACE"
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
	if namespace, exists := os.LookupEnv(DefaultReleaseNamespaceEnvVar); exists {
		return namespace, nil
	}
	return "", fmt.Errorf("failed to determine release namespace from file %s and env var %s", namespaceFilePath, DefaultReleaseNamespaceEnvVar)
}

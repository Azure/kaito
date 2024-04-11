// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package utils

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	// WorkspaceFinalizer is used to make sure that workspace controller handles garbage collection.
	WorkspaceFinalizer = "workspace.finalizer.kaito.sh"
)

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
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

// GetFieldNamesFromStruct returns a slice of json tag names for fields in a given struct type.
func GetFieldNamesFromStruct(section any) []string {
	t := reflect.TypeOf(section)
	fieldNames := make([]string, 0, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		// If the field has a json tag, use the tag name, excluding any options like ",omitempty"
		if jsonTag != "" {
			tagName := strings.Split(jsonTag, ",")[0] // Get the name part of the json tag
			fieldNames = append(fieldNames, tagName)
		} else {
			// If there's no json tag, use the field name
			fieldNames = append(fieldNames, field.Name)
		}
	}

	return fieldNames
}

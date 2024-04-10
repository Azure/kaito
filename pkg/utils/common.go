// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

// common.go contains general utility functions applicable across the application.
package utils

import "fmt"

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

func ShellCmd(command string) []string {
	return []string{
		"/bin/sh",
		"-c",
		command,
	}
}

func BuildCmdStr(baseCommand string, runParams map[string]string) string {
	updatedBaseCommand := baseCommand
	for key, value := range runParams {
		updatedBaseCommand = fmt.Sprintf("%s --%s=%s", updatedBaseCommand, key, value)
	}

	return updatedBaseCommand
}

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package utils

import (
	"fmt"
	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const (
	DefaultVolumeMountPath = "/dev/shm"
)

func ConfigVolume(wObj *kaitov1alpha1.Workspace) ([]corev1.Volume, []corev1.VolumeMount) {
	volume := []corev1.Volume{}
	volumeMount := []corev1.VolumeMount{}

	// Signifies multinode inference requirement
	if *wObj.Resource.Count > 1 {
		// Append share memory volume to any existing volumes
		volume = append(volume, corev1.Volume{
			Name: "dshm",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium: "Memory",
				},
			},
		})

		volumeMount = append(volumeMount, corev1.VolumeMount{
			Name:      volume[0].Name,
			MountPath: DefaultVolumeMountPath,
		})
	}

	return volume, volumeMount
}

func ShellCmd(command string) []string {
	return []string{
		"/bin/sh",
		"-c",
		command,
	}
}

func BuildCmdStr(baseCommand string, torchRunParams map[string]string) string {
	updatedBaseCommand := baseCommand
	for key, value := range torchRunParams {
		updatedBaseCommand = fmt.Sprintf("%s --%s=%s", updatedBaseCommand, key, value)
	}

	return updatedBaseCommand
}

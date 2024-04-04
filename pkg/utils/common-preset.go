// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package utils

import (
	"fmt"
	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/pkg/tuning"
	corev1 "k8s.io/api/core/v1"
	"strconv"
)

const (
	DefaultVolumeMountPath = "/dev/shm"
)

func ConfigSHMVolume(wObj *kaitov1alpha1.Workspace) (corev1.Volume, corev1.VolumeMount) {
	volume := corev1.Volume{}
	volumeMount := corev1.VolumeMount{}

	// Signifies multinode inference requirement
	if *wObj.Resource.Count > 1 {
		// Append share memory volume to any existing volumes
		volume = corev1.Volume{
			Name: "dshm",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium: "Memory",
				},
			},
		}

		volumeMount = corev1.VolumeMount{
			Name:      volume.Name,
			MountPath: DefaultVolumeMountPath,
		}
	}

	return volume, volumeMount
}

func ConfigDataVolume() ([]corev1.Volume, []corev1.VolumeMount) {
	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount
	volumes = append(volumes, corev1.Volume{
		Name: "data-volume",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})

	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      "data-volume",
		MountPath: "/data",
	})
	return volumes, volumeMounts
}

func GetInstanceGPUCount(wObj *kaitov1alpha1.Workspace) int {
	sku := wObj.Resource.InstanceType
	gpuConfig, exists := kaitov1alpha1.SupportedGPUConfigs[sku]
	if !exists {
		numProcesses, err := strconv.Atoi(tuning.DefaultNumProcesses)
		if err != nil {
			return 1
		}
		return numProcesses
	}
	return gpuConfig.GPUCount
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

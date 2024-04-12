// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package utils

import (
	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
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

func ConfigDataVolume(hostPath string) ([]corev1.Volume, []corev1.VolumeMount) {
	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount
	var volumeSource corev1.VolumeSource
	if hostPath != "" {
		volumeSource = corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: hostPath,
			},
		}
	} else {
		volumeSource = corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		}
	}
	volumes = append(volumes, corev1.Volume{
		Name:         "data-volume",
		VolumeSource: volumeSource,
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
		return 1
	}
	return gpuConfig.GPUCount
}

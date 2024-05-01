// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package utils

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	DefaultVolumeMountPath    = "/dev/shm"
	DefaultConfigMapMountPath = "/mnt/config"
	DefaultDataVolumePath     = "/mnt/data"
	DefaultResultsVolumePath  = "/mnt/results"
)

func ConfigResultsVolume() (corev1.Volume, corev1.VolumeMount) {
	sharedWorkspaceVolume := corev1.Volume{
		Name: "results-volume",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
	sharedVolumeMount := corev1.VolumeMount{
		Name: "results-volume",
		// TODO: Override output path if specified in trainingconfig
		MountPath: DefaultResultsVolumePath,
	}
	return sharedWorkspaceVolume, sharedVolumeMount
}

func ConfigImagePushSecretVolume(imagePushSecret string) (corev1.Volume, corev1.VolumeMount) {
	volume := corev1.Volume{
		Name: "docker-config",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: imagePushSecret,
				Items: []corev1.KeyToPath{
					{
						Key:  ".dockerconfigjson",
						Path: "config.json",
					},
				},
			},
		},
	}

	volumeMount := corev1.VolumeMount{
		Name:      "docker-config",
		MountPath: "/root/.docker/config.json",
		SubPath:   "config.json", // Mount only the config.json file
	}

	return volume, volumeMount
}

func ConfigSHMVolume(instanceCount int) (corev1.Volume, corev1.VolumeMount) {
	volume := corev1.Volume{}
	volumeMount := corev1.VolumeMount{}

	// Signifies multinode inference requirement
	if instanceCount > 1 {
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

func ConfigCMVolume(cmName string) (corev1.Volume, corev1.VolumeMount) {
	volume := corev1.Volume{
		Name: "config-volume",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cmName,
				},
			},
		},
	}

	volumeMount := corev1.VolumeMount{
		Name:      volume.Name,
		MountPath: DefaultConfigMapMountPath,
	}
	return volume, volumeMount
}

func ConfigDataVolume(hostPath *string) ([]corev1.Volume, []corev1.VolumeMount) {
	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount
	var volumeSource corev1.VolumeSource
	if hostPath != nil {
		volumeSource = corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: *hostPath,
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
		MountPath: DefaultDataVolumePath,
	})
	return volumes, volumeMounts
}

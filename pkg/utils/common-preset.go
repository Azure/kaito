// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package utils

import (
	"fmt"
	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/pkg/tuning"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"os"
	"reflect"
)

const (
	DefaultVolumeMountPath        = "/dev/shm"
	DefaultReleaseNamespaceEnvVar = "RELEASE_NAMESPACE"
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
	return "", fmt.Errorf("failed to determine release namespace from file %s and env var %s", namespaceFilePath, ReleaseNamespaceEnvVar)
}

// GetSectionNamesFromStruct returns a slice of yaml tag names for fields in the TrainingConfig struct.
func GetSectionNamesFromStruct(trainingConfig tuning.TrainingConfig) []string {
	t := reflect.TypeOf(trainingConfig)
	sectionNames := make([]string, 0, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		yamlTag := field.Tag.Get("yaml")
		// If the field has a yaml tag, use the tag name
		if yamlTag != "" {
			sectionNames = append(sectionNames, yamlTag)
		}
	}

	return sectionNames
}

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	InferenceModeCustomTemplate kaitov1alpha1.ModelImageAccessMode = "customTemplate"
)

var (
	// PollInterval defines the interval time for a poll operation.
	PollInterval = 250 * time.Millisecond
	// PollTimeout defines the time after which the poll operation times out.
	PollTimeout = 60 * time.Second
)

func GetEnv(envVar string) string {
	env := os.Getenv(envVar)
	if env == "" {
		fmt.Printf("%s is not set or is empty", envVar)
		return ""
	}
	return env
}

func GenerateWorkspaceManifest(name, namespace, imageName string, resourceCount int, instanceType string,
	labelSelector *metav1.LabelSelector, preferredNodes []string, presetName kaitov1alpha1.ModelName,
	inferenceMode kaitov1alpha1.ModelImageAccessMode, imagePullSecret []string,
	podTemplate *corev1.PodTemplateSpec) *kaitov1alpha1.Workspace {

	workspace := &kaitov1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Resource: kaitov1alpha1.ResourceSpec{
			Count:          to.Ptr(resourceCount),
			InstanceType:   instanceType,
			LabelSelector:  labelSelector,
			PreferredNodes: preferredNodes,
		},
	}

	var workspaceInference kaitov1alpha1.InferenceSpec
	if inferenceMode == kaitov1alpha1.ModelImageAccessModePublic ||
		inferenceMode == kaitov1alpha1.ModelImageAccessModePrivate {
		workspaceInference.Preset = &kaitov1alpha1.PresetSpec{
			PresetMeta: kaitov1alpha1.PresetMeta{
				Name:       presetName,
				AccessMode: inferenceMode,
			},
			PresetOptions: kaitov1alpha1.PresetOptions{
				Image:            imageName,
				ImagePullSecrets: imagePullSecret,
			},
		}
	}
	if inferenceMode == InferenceModeCustomTemplate {
		workspaceInference.Template = podTemplate
	}

	workspace.Inference = workspaceInference

	return workspace
}

func GeneratePodTemplate(name, namespace, image string, labels map[string]string) *corev1.PodTemplateSpec {
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            name,
					Image:           image,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Command:         []string{"/bin/sleep", "10000"},
				},
			},
		},
	}
}

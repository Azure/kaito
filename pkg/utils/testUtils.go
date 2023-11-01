// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package utils

import (
	"github.com/aws/karpenter-core/pkg/apis/v1alpha5"
	"github.com/azure/kaito/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	MockWorkspace = &v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testWorkspace",
			Namespace: "kaito",
		},
		Resource: v1alpha1.ResourceSpec{
			InstanceType: "Standard_NC12s_v3",
		},
		Inference: v1alpha1.InferenceSpec{
			Preset: &v1alpha1.PresetSpec{
				PresetMeta: v1alpha1.PresetMeta{
					Name: "llama-2-7b-chat",
				},
			},
		},
	}
)

var (
	machineLabels = map[string]string{
		"karpenter.sh/provisioner-name": "default",
		"kaito.sh/workspace":            "none",
	}
)

var (
	MockMachine = v1alpha5.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "testmachine",
			Labels: machineLabels,
		},
		Spec: v1alpha5.MachineSpec{
			MachineTemplateRef: &v1alpha5.MachineTemplateRef{
				Name: "test-machine",
			},
			Requirements: []corev1.NodeSelectorRequirement{
				{
					Key:      corev1.LabelInstanceTypeStable,
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{"Standard_NC12s_v3"},
				},
				{
					Key:      "karpenter.sh/provisioner-name",
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{"default"},
				},
			},
		},
	}
)

func NewTestScheme() *runtime.Scheme {
	testScheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(testScheme)
	return testScheme
}

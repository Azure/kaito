// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package test

import (
	"github.com/aws/karpenter-core/pkg/apis/v1alpha5"
	"github.com/kaito-project/kaito/api/v1alpha1"
	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/karpenter/pkg/apis/v1beta1"
)

const (
	LabelKeyNvidia    = "accelerator"
	LabelValueNvidia  = "nvidia"
	CapacityNvidiaGPU = "nvidia.com/gpu"
)

var ValidStrength string = "0.5"

var (
	MockWorkspaceDistributedModel = &v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testWorkspace",
			Namespace: "kaito",
		},
		Resource: v1alpha1.ResourceSpec{
			Count:        &gpuNodeCount,
			InstanceType: "Standard_NC12s_v3",
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"apps": "test",
				},
			},
		},
		Inference: &v1alpha1.InferenceSpec{
			Preset: &v1alpha1.PresetSpec{
				PresetMeta: v1alpha1.PresetMeta{
					Name: "test-distributed-model",
				},
			},
		},
	}
	MockWorkspaceWithPreferredNodes = &v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testWorkspace",
			Namespace: "kaito",
		},
		Resource: v1alpha1.ResourceSpec{
			Count:        &gpuNodeCount,
			InstanceType: "Standard_NC12s_v3",
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"apps": "test",
				},
			},
			PreferredNodes: []string{"node-p1", "node-p2"},
		},
		Inference: &v1alpha1.InferenceSpec{
			Preset: &v1alpha1.PresetSpec{
				PresetMeta: v1alpha1.PresetMeta{
					Name: "test-distributed-model",
				},
			},
		},
	}
)

var (
	MockRAGEngine = &v1alpha1.RAGEngine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testRAGEngine",
			Namespace: "kaito",
		},
		Spec: &v1alpha1.RAGEngineSpec{
			Compute: &v1alpha1.ResourceSpec{
				Count:        &gpuNodeCount,
				InstanceType: "Standard_NC12s_v3",
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"ragengine.kaito.io/name": "testRAGEngine",
					},
				},
			},
			Embedding: &v1alpha1.EmbeddingSpec{
				Local: &v1alpha1.LocalEmbeddingSpec{
					ModelID: "BAAI/bge-small-en-v1.5",
				},
			},
		},
	}
)
var (
	MockRAGEngineDistributedModel = &v1alpha1.RAGEngine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testRAGEngine",
			Namespace: "kaito",
		},
		Spec: &v1alpha1.RAGEngineSpec{
			Compute: &v1alpha1.ResourceSpec{
				Count:        &gpuNodeCount,
				InstanceType: "Standard_NC12s_v3",
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"apps": "test",
					},
				},
			},
		},
	}
)

var (
	MockWorkspaceWithPreset = &v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testWorkspace",
			Namespace: "kaito",
		},
		Resource: v1alpha1.ResourceSpec{
			Count:        &gpuNodeCount,
			InstanceType: "Standard_NC12s_v3",
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"apps": "test",
				},
			},
		},
		Inference: &v1alpha1.InferenceSpec{
			Preset: &v1alpha1.PresetSpec{
				PresetMeta: v1alpha1.PresetMeta{
					Name: "test-model",
				},
			},
		},
	}
)

var MockWorkspaceWithPresetHash = "89ae127050ec264a5ce84db48ef7226574cdf1299e6bd27fe90b927e34cc8adb"

var (
	MockRAGEngineWithPreset = &v1alpha1.RAGEngine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testRAGEngine",
			Namespace: "kaito",
		},
		Spec: &v1alpha1.RAGEngineSpec{
			Compute: &v1alpha1.ResourceSpec{
				Count:        &gpuNodeCount,
				InstanceType: "Standard_NC12s_v3",
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"ragengine.kaito.io/name": "testRAGEngine",
					},
				},
			},
			Embedding: &v1alpha1.EmbeddingSpec{
				Local: &v1alpha1.LocalEmbeddingSpec{
					ModelID: "BAAI/bge-small-en-v1.5",
				},
			},
			InferenceService: &v1alpha1.InferenceServiceSpec{
				URL: "http://localhost:5000/chat",
			},
		},
	}
)

var MockRAGEngineWithPresetHash = "14485768c1b67a529a71e3c87d9f2e6c1ed747534dea07e268e93475a5e21e"

var (
	MockWorkspaceWithDeleteOldCR = v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testWorkspace",
			Namespace: "kaito",
			Annotations: map[string]string{
				"workspace.kaito.io/hash":     "1171dc5d15043c92e684c8f06689eb241763a735181fdd2b59c8bd8fd6eecdd4",
				"workspace.kaito.io/revision": "1",
			},
		},
		Resource: v1alpha1.ResourceSpec{
			Count:        &gpuNodeCount,
			InstanceType: "Standard_NC12s_v3",
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"workspace.kaito.io/name": "testWorkspace",
				},
			},
		},
		Inference: &v1alpha1.InferenceSpec{
			Preset: &v1alpha1.PresetSpec{
				PresetMeta: v1alpha1.PresetMeta{
					Name: "test-model-DeleteOldCR", // presetMeta name is changed
				},
			},
		},
	}
)

var (
	MockRAGEngineWithDeleteOldCR = v1alpha1.RAGEngine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testRAGEngine",
			Namespace: "kaito",
			Annotations: map[string]string{
				"workspace.kaito.io/hash":     "14485768c1b67a529a71e3c87d9f2e6c1ed747534dea07e268e93475a5e21e",
				"workspace.kaito.io/revision": "1",
			},
		},
		Spec: &v1alpha1.RAGEngineSpec{
			Compute: &v1alpha1.ResourceSpec{
				Count:        &gpuNodeCount,
				InstanceType: "Standard_NC12s_v3",
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"ragengine.kaito.io/name": "testRAGEngine",
					},
				},
			},
			Embedding: &v1alpha1.EmbeddingSpec{
				Local: &v1alpha1.LocalEmbeddingSpec{
					ModelID: "BAAI/bge-small-en-v1.5",
				},
			},
		},
	}
)

var (
	MockWorkspaceFailToCreateCR = v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testWorkspace-failedtocreateCR",
			Namespace: "kaito",
			Annotations: map[string]string{
				"workspace.kaito.io/revision": "1",
			},
		},
		Resource: v1alpha1.ResourceSpec{
			Count:        &gpuNodeCount,
			InstanceType: "Standard_NC12s_v3",
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"workspace.kaito.io/name": "testWorkspace",
				},
			},
		},
		Inference: &v1alpha1.InferenceSpec{
			Preset: &v1alpha1.PresetSpec{
				PresetMeta: v1alpha1.PresetMeta{
					Name: "test-model",
				},
			},
		},
	}
)

var (
	MockRAGEngineFailToCreateCR = v1alpha1.RAGEngine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testRAGEngine-failedtocreateCR",
			Namespace: "kaito",
			Annotations: map[string]string{
				"ragengine.kaito.io/revision": "1",
			},
		},
		Spec: &v1alpha1.RAGEngineSpec{
			Compute: &v1alpha1.ResourceSpec{
				Count:        &gpuNodeCount,
				InstanceType: "Standard_NC12s_v3",
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"ragengine.kaito.io/name": "testRAGEngine",
					},
				},
			}},
	}
)

var (
	MockWorkspaceSuccessful = v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testWorkspace-successful",
			Namespace: "kaito",
			Annotations: map[string]string{
				"workspace.kaito.io/revision": "0",
			},
		},
		Resource: v1alpha1.ResourceSpec{
			Count:        &gpuNodeCount,
			InstanceType: "Standard_NC12s_v3",
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"workspace.kaito.io/name": "testWorkspace",
				},
			},
		},
		Inference: &v1alpha1.InferenceSpec{
			Preset: &v1alpha1.PresetSpec{
				PresetMeta: v1alpha1.PresetMeta{
					Name: "test-model",
				},
			},
		},
	}
)

var (
	MockRAGEngineSuccessful = v1alpha1.RAGEngine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testRAGEngine-successful",
			Namespace: "kaito",
			Annotations: map[string]string{
				"ragengine.kaito.io/revision": "0",
			},
		},
		Spec: &v1alpha1.RAGEngineSpec{
			Compute: &v1alpha1.ResourceSpec{
				Count:        &gpuNodeCount,
				InstanceType: "Standard_NC12s_v3",
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"ragengine.kaito.io/name": "testRAGEngine",
					},
				},
			}},
	}
)

var (
	MockWorkspaceWithComputeHash = v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testWorkspace",
			Namespace: "kaito",
			Annotations: map[string]string{
				"workspace.kaito.io/hash":     "1171dc5d15043c92e684c8f06689eb241763a735181fdd2b59c8bd8fd6eecdd4",
				"workspace.kaito.io/revision": "1",
			},
		},
		Resource: v1alpha1.ResourceSpec{
			Count:        &gpuNodeCount,
			InstanceType: "Standard_NC12s_v3",
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"workspace.kaito.io/name": "testWorkspace",
				},
			},
		},
		Inference: &v1alpha1.InferenceSpec{
			Preset: &v1alpha1.PresetSpec{
				PresetMeta: v1alpha1.PresetMeta{
					Name: "test-model",
				},
			},
		},
	}
)

var (
	MockRAGEngineWithComputeHash = v1alpha1.RAGEngine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testRAGEngine",
			Namespace: "kaito",
			Annotations: map[string]string{
				"ragengine.kaito.io/hash":     "7985249e078eb041e38c10c3637032b2d352616c609be8542a779460d3ff1d67",
				"ragengine.kaito.io/revision": "1",
			},
		},
		Spec: &v1alpha1.RAGEngineSpec{
			Compute: &v1alpha1.ResourceSpec{
				Count:        &gpuNodeCount,
				InstanceType: "Standard_NC12s_v3",
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"ragengine.kaito.io/name": "testRAGEngine",
					},
				},
			}},
	}
)

var (
	MockWorkspaceUpdateCR = v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testWorkspace",
			Namespace: "kaito",
			Annotations: map[string]string{
				"workspace.kaito.io/hash":     "1171dc5d15043c92e684c8f06689eb241763a735181fdd2b59c8bd8fd6eecdd4",
				"workspace.kaito.io/revision": "1",
			},
		},
		Resource: v1alpha1.ResourceSpec{
			Count:        &gpuNodeCount,
			InstanceType: "Standard_NC12s_v3",
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"workspace.kaito.io/name": "testWorkspace",
				},
			},
		},
		Inference: &v1alpha1.InferenceSpec{
			Preset: &v1alpha1.PresetSpec{
				PresetMeta: v1alpha1.PresetMeta{
					Name: "test-model",
				},
			},
		},
	}
)

var (
	MockWorkspaceWithUpdatedDeployment = v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testWorkspace",
			Namespace: "kaito",
			Annotations: map[string]string{
				"workspace.kaito.io/hash":     "1171dc5d15043c92e684c8f06689eb241763a735181fdd2b59c8bd8fd6eecdd4",
				"workspace.kaito.io/revision": "1",
			},
		},
		Resource: v1alpha1.ResourceSpec{
			Count:        &gpuNodeCount,
			InstanceType: "Standard_NC12s_v3",
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"workspace.kaito.io/name": "testWorkspace",
				},
			},
		},
		Inference: &v1alpha1.InferenceSpec{
			Preset: &v1alpha1.PresetSpec{
				PresetMeta: v1alpha1.PresetMeta{
					Name: "test-model",
				},
			},
			Adapters: []v1alpha1.AdapterSpec{
				{
					Source: &v1alpha1.DataSource{
						Name:  "Adapter-1",
						Image: "fake.kaito.com/kaito-image:0.0.1",
					},
					Strength: &ValidStrength,
				},
			},
		},
	}
)

var (
	MockRAGEngineWithUpdatedDeployment = v1alpha1.RAGEngine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testRAGEngine",
			Namespace: "kaito",
			Annotations: map[string]string{
				"ragengine.kaito.io/hash":     "7985249e078eb041e38c10c3637032b2d352616c609be8542a779460d3ff1d67",
				"ragengine.kaito.io/revision": "1",
			},
		},
		Spec: &v1alpha1.RAGEngineSpec{
			Compute: &v1alpha1.ResourceSpec{
				Count:        &gpuNodeCount,
				InstanceType: "Standard_NC12s_v3",
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"ragengine.kaito.io/name": "testRAGEngine",
					},
				},
			}},
	}
)

var (
	numRep                = int32(1)
	MockDeploymentUpdated = appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "testWorkspace",
			Namespace:   "kaito",
			Annotations: map[string]string{v1alpha1.WorkspaceRevisionAnnotation: "1"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &numRep,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "test-app",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "nginx:latest",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "ENV_VAR_NAME",
									Value: "ENV_VAR_VALUE",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "volume-name",
									MountPath: "/mount/path",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "volume-name",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas: 1,
		},
	}
	MockDeploymentWithAnnotationsAndContainer1 = appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "nginx:latest",
						},
					},
				},
			},
		},
	}
	MockDeploymentWithAnnotationsAndContainer2 = appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container2",
							Image: "nginx:latest",
						},
					},
				},
			},
		},
	}
)

var (
	MockWorkspaceWithInferenceTemplate = &v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testWorkspace",
			Namespace: "kaito",
		},
		Resource: v1alpha1.ResourceSpec{
			Count:        &gpuNodeCount,
			InstanceType: "Standard_NC12s_v3",
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"apps": "test",
				},
			},
		},
		Inference: &v1alpha1.InferenceSpec{
			Template: &corev1.PodTemplateSpec{},
		},
	}
)

var (
	MockNodeList = &corev1.NodeList{
		Items: nodes,
	}
)

var (
	nodes = []corev1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node1",
				Labels: map[string]string{
					corev1.LabelInstanceTypeStable: "Standard_NC12s_v3",
					LabelKeyNvidia:                 LabelValueNvidia,
				},
			},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{
					{
						Type:   corev1.NodeReady,
						Status: corev1.ConditionTrue,
					},
				},
				Capacity: corev1.ResourceList{
					CapacityNvidiaGPU: resource.MustParse("1"),
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node2",
				Labels: map[string]string{
					corev1.LabelInstanceTypeStable: "Wrong_Instance_Type",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node3",
			},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{
					{
						Type:   corev1.NodeReady,
						Status: corev1.ConditionFalse,
					},
				},
			},
		},
	}
)

var (
	gpuNodeCount = 1
)

var (
	machineLabels = map[string]string{
		"karpenter.sh/provisioner-name": "default",
		"kaito.sh/workspace":            "none",
	}

	nodeClaimLabels = map[string]string{
		"karpenter.sh/nodepool": "kaito",
		"kaito.sh/workspace":    "none",
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

	MockNodeClaim = v1beta1.NodeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "testnodeclaim",
			Labels: nodeClaimLabels,
		},
		Spec: v1beta1.NodeClaimSpec{
			Requirements: []v1beta1.NodeSelectorRequirementWithMinValues{
				{
					NodeSelectorRequirement: corev1.NodeSelectorRequirement{
						Key:      corev1.LabelInstanceTypeStable,
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{"Standard_NC12s_v3"},
					},
					MinValues: lo.ToPtr(1),
				},
			},
		},
	}
)

var (
	MockMachineList = &v1alpha5.MachineList{
		Items: []v1alpha5.Machine{
			MockMachine,
		},
	}

	MockNodeClaimList = &v1beta1.NodeClaimList{
		Items: []v1beta1.NodeClaim{
			MockNodeClaim,
		},
	}
)

var (
	Adapters1 = []v1alpha1.AdapterSpec{
		{
			Source: &v1alpha1.DataSource{
				Name:  "Adapter-1",
				Image: "fake.kaito.com/kaito-image:0.0.1",
			},
			Strength: &ValidStrength,
		},
	}
	Adapters2 = []v1alpha1.AdapterSpec{
		{
			Source: &v1alpha1.DataSource{
				Name:  "Adapter-1",
				Image: "fake.kaito.com/kaito-image:0.0.1",
			},
			Strength: &ValidStrength,
		},
		{
			Source: &v1alpha1.DataSource{
				Name:  "Adapter-2",
				Image: "fake.kaito.com/kaito-image:0.0.2",
			},
			Strength: &ValidStrength,
		},
	}
	Adapters3 = []v1alpha1.AdapterSpec{
		{
			Source: &v1alpha1.DataSource{
				Name:  "Adapter-2",
				Image: "fake.kaito.com/kaito-image:0.0.2",
			},
			Strength: &ValidStrength,
		},
		{
			Source: &v1alpha1.DataSource{
				Name:  "Adapter-1",
				Image: "fake.kaito.com/kaito-image:0.0.1",
			},
			Strength: &ValidStrength,
		},
	}
	Adapters4 = []v1alpha1.AdapterSpec{
		{
			Source: &v1alpha1.DataSource{
				Name:  "Adapter-1",
				Image: "fake.kaito.com/kaito-image:0.0.1",
			},
			Strength: &ValidStrength,
		},
		{
			Source: &v1alpha1.DataSource{
				Name:  "Adapter-3",
				Image: "fake.kaito.com/kaito-image:0.0.3",
			},
			Strength: &ValidStrength,
		},
	}
)

func NewTestScheme() *runtime.Scheme {
	testScheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(testScheme)
	return testScheme
}

func NotFoundError() error {
	return &apierrors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonNotFound}}
}

func IsAlreadyExistsError() error {
	return &apierrors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonAlreadyExists}}
}

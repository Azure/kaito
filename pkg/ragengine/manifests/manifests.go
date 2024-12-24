// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package manifests

import (
	"context"

	"k8s.io/apimachinery/pkg/util/intstr"

	kaitov1alpha1 "github.com/kaito-project/kaito/api/v1alpha1"
	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var controller = true

func GenerateRAGDeploymentManifest(ctx context.Context, ragEngineObj *kaitov1alpha1.RAGEngine, revisionNum string, imageName string,
	imagePullSecretRefs []corev1.LocalObjectReference, replicas int, commands []string, containerPorts []corev1.ContainerPort,
	livenessProbe, readinessProbe *corev1.Probe, resourceRequirements corev1.ResourceRequirements,
	tolerations []corev1.Toleration, volumes []corev1.Volume, volumeMount []corev1.VolumeMount) *appsv1.Deployment {

	nodeRequirements := make([]corev1.NodeSelectorRequirement, 0, len(ragEngineObj.Spec.Compute.LabelSelector.MatchLabels))
	for key, value := range ragEngineObj.Spec.Compute.LabelSelector.MatchLabels {
		nodeRequirements = append(nodeRequirements, corev1.NodeSelectorRequirement{
			Key:      key,
			Operator: corev1.NodeSelectorOpIn,
			Values:   []string{value},
		})
	}

	selector := map[string]string{
		kaitov1alpha1.LabelRAGEngineName: ragEngineObj.Name,
	}
	labelselector := &v1.LabelSelector{
		MatchLabels: selector,
	}
	initContainers := []corev1.Container{}

	envs := RAGSetEnv(ragEngineObj)

	return &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      ragEngineObj.Name,
			Namespace: ragEngineObj.Namespace,
			OwnerReferences: []v1.OwnerReference{
				{
					APIVersion: kaitov1alpha1.GroupVersion.String(),
					Kind:       "RAGEngine",
					UID:        ragEngineObj.UID,
					Name:       ragEngineObj.Name,
					Controller: &controller,
				},
			},
			Annotations: map[string]string{
				kaitov1alpha1.RAGEngineRevisionAnnotation: revisionNum,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: lo.ToPtr(int32(replicas)),
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 0,
					},
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 1,
					},
				}, // Configuration for rolling updates: allows no extra pods during the update and permits at most one unavailable pod at a time。
			},
			Selector: labelselector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: selector,
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: imagePullSecretRefs,
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: nodeRequirements,
									},
								},
							},
						},
					},
					InitContainers: initContainers,
					Containers: []corev1.Container{
						{
							Name:           ragEngineObj.Name,
							Image:          imageName,
							Command:        commands,
							Resources:      resourceRequirements,
							LivenessProbe:  livenessProbe,
							ReadinessProbe: readinessProbe,
							Ports:          containerPorts,
							VolumeMounts:   volumeMount,
							Env:            envs,
						},
					},
					Tolerations: tolerations,
					Volumes:     volumes,
				},
			},
		},
	}
}

func RAGSetEnv(ragEngineObj *kaitov1alpha1.RAGEngine) []corev1.EnvVar {
	var envs []corev1.EnvVar
	var embeddingType string
	if ragEngineObj.Spec.Embedding.Local != nil {
		embeddingType = "local"
		if ragEngineObj.Spec.Embedding.Local.ModelID != "" {
			modelID := ragEngineObj.Spec.Embedding.Local.ModelID
			modelIDEnv := corev1.EnvVar{
				Name:  "MODEL_ID",
				Value: modelID,
			}
			envs = append(envs, modelIDEnv)
		}
		if ragEngineObj.Spec.Embedding.Local.ModelAccessSecret != "" {
			accessSecret := ragEngineObj.Spec.Embedding.Local.ModelAccessSecret
			accessSecretEnv := corev1.EnvVar{
				Name:  "ACCESS_SECRET",
				Value: accessSecret,
			}
			envs = append(envs, accessSecretEnv)
		}
	} else if ragEngineObj.Spec.Embedding.Remote != nil {
		embeddingType = "remote"
		// TODO: Model ID Env
	}
	embeddingTypeEnv := corev1.EnvVar{
		Name:  "EMBEDDING_TYPE",
		Value: embeddingType,
	}
	envs = append(envs, embeddingTypeEnv)

	stoageEnv := corev1.EnvVar{
		Name:  "VECTOR_DB_TYPE",
		Value: "faiss", // TODO: get storage done
	}
	envs = append(envs, stoageEnv)
	inferenceServiceURL := ragEngineObj.Spec.InferenceService.URL
	inferenceServiceURLEnv := corev1.EnvVar{
		Name:  "LLM_INFERENCE_URL",
		Value: inferenceServiceURL,
	}
	envs = append(envs, inferenceServiceURLEnv)

	if ragEngineObj.Spec.InferenceService.AccessSecret != "" {
		accessSecretEnv := corev1.EnvVar{
			Name:  "INFERENCE_ACCESS_SECRET",
			Value: ragEngineObj.Spec.InferenceService.AccessSecret,
		}
		envs = append(envs, accessSecretEnv)
	}
	return envs
}

func GenerateRAGServiceManifest(ctx context.Context, ragObj *kaitov1alpha1.RAGEngine, serviceName string, serviceType corev1.ServiceType) *corev1.Service {
	selector := map[string]string{
		kaitov1alpha1.LabelRAGEngineName: ragObj.Name,
	}

	servicePorts := []corev1.ServicePort{
		{
			Name:       "http",
			Protocol:   corev1.ProtocolTCP,
			Port:       80,
			TargetPort: intstr.FromInt32(5000),
		},
	}

	return &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      serviceName,
			Namespace: ragObj.Namespace,
			OwnerReferences: []v1.OwnerReference{
				{
					APIVersion: kaitov1alpha1.GroupVersion.String(),
					Kind:       "RAGEngine",
					UID:        ragObj.UID,
					Name:       ragObj.Name,
					Controller: &controller,
				},
			},
		},
		Spec: corev1.ServiceSpec{
			Type:                     serviceType,
			Ports:                    servicePorts,
			Selector:                 selector,
			PublishNotReadyAddresses: true,
		},
	}
}

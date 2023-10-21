package resources

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/util/intstr"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func GenerateServiceManifest(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace, serviceType corev1.ServiceType, isStatefulSet bool) *corev1.Service {
	klog.InfoS("GenerateServiceManifest", "workspace", klog.KObj(workspaceObj), "serviceType", serviceType)

	selector := make(map[string]string)
	for k, v := range workspaceObj.Resource.LabelSelector.MatchLabels {
		selector[k] = v
	}
	// If statefulset, modify the selector to select the pod with index 0 as the endpoint
	if isStatefulSet {
		podNameForIndex0 := fmt.Sprintf("%s-0", workspaceObj.Name)
		selector["statefulset.kubernetes.io/pod-name"] = podNameForIndex0
	}

	return &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      workspaceObj.Name,
			Namespace: workspaceObj.Namespace,
			OwnerReferences: []v1.OwnerReference{
				{
					APIVersion: kaitov1alpha1.GroupVersion.String(),
					Kind:       "Workspace",
					UID:        workspaceObj.UID,
					Name:       workspaceObj.Name,
				},
			},
		},
		Spec: corev1.ServiceSpec{
			Type: serviceType,
			Ports: []corev1.ServicePort{
				// HTTP API Port
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(5000),
				},
				// Torch NCCL Port
				{
					Name:       "torch",
					Protocol:   corev1.ProtocolTCP,
					Port:       29500,
					TargetPort: intstr.FromInt(29500),
				},
			},
			Selector: selector,
			// Added this to allow pods to discover each other
			// (DNS Resolution) During their initialization phase
			PublishNotReadyAddresses: true,
		},
	}
}

func GenerateStatefulSetManifest(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace, imageName string,
	replicas int, commands []string, containerPorts []corev1.ContainerPort,
	livenessProbe, readinessProbe *corev1.Probe, resourceRequirements corev1.ResourceRequirements,
	tolerations []corev1.Toleration, volumes []corev1.Volume, volumeMount []corev1.VolumeMount) *appsv1.StatefulSet {

	klog.InfoS("GenerateStatefulSetManifest", "workspace", klog.KObj(workspaceObj), "image", imageName)

	// Gather label requirements from workspaceObj's label selector
	labelRequirements := make([]v1.LabelSelectorRequirement, 0, len(workspaceObj.Resource.LabelSelector.MatchLabels))
	for key, value := range workspaceObj.Resource.LabelSelector.MatchLabels {
		labelRequirements = append(labelRequirements, v1.LabelSelectorRequirement{
			Key:      key,
			Operator: v1.LabelSelectorOpIn,
			Values:   []string{value},
		})
	}

	return &appsv1.StatefulSet{
		ObjectMeta: v1.ObjectMeta{
			Name:      workspaceObj.Name,
			Namespace: workspaceObj.Namespace,
			OwnerReferences: []v1.OwnerReference{
				{
					APIVersion: kaitov1alpha1.GroupVersion.String(),
					Kind:       "Workspace",
					UID:        workspaceObj.UID,
					Name:       workspaceObj.Name,
				},
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:            lo.ToPtr(int32(replicas)),
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Selector:            workspaceObj.Resource.LabelSelector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: workspaceObj.Resource.LabelSelector.MatchLabels,
				},
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: &v1.LabelSelector{
										MatchExpressions: labelRequirements,
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:           workspaceObj.Name,
							Image:          imageName,
							Command:        commands,
							Resources:      resourceRequirements,
							LivenessProbe:  livenessProbe,
							ReadinessProbe: readinessProbe,
							Ports:          containerPorts,
							VolumeMounts:   volumeMount,
						},
					},
					Tolerations: tolerations,
					Volumes:     volumes,
				},
			},
		},
	}
}

func GenerateDeploymentManifest(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace, imageName string,
	replicas int, commands []string, containerPorts []corev1.ContainerPort,
	livenessProbe, readinessProbe *corev1.Probe, resourceRequirements corev1.ResourceRequirements,
	tolerations []corev1.Toleration, volumes []corev1.Volume, volumeMount []corev1.VolumeMount) *appsv1.Deployment {

	klog.InfoS("GenerateDeploymentManifest", "workspace", klog.KObj(workspaceObj), "image", imageName)

	// Gather label requirements from workspaceObj's label selector
	labelRequirements := make([]v1.LabelSelectorRequirement, 0, len(workspaceObj.Resource.LabelSelector.MatchLabels))
	for key, value := range workspaceObj.Resource.LabelSelector.MatchLabels {
		labelRequirements = append(labelRequirements, v1.LabelSelectorRequirement{
			Key:      key,
			Operator: v1.LabelSelectorOpIn,
			Values:   []string{value},
		})
	}

	return &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      workspaceObj.Name,
			Namespace: workspaceObj.Namespace,
			OwnerReferences: []v1.OwnerReference{
				{
					APIVersion: kaitov1alpha1.GroupVersion.String(),
					Kind:       "Workspace",
					UID:        workspaceObj.UID,
					Name:       workspaceObj.Name,
				},
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: lo.ToPtr(int32(replicas)),
			Selector: workspaceObj.Resource.LabelSelector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: workspaceObj.Resource.LabelSelector.MatchLabels,
				},
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: &v1.LabelSelector{
										MatchExpressions: labelRequirements,
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:           workspaceObj.Name,
							Image:          imageName,
							Command:        commands,
							Resources:      resourceRequirements,
							LivenessProbe:  livenessProbe,
							ReadinessProbe: readinessProbe,
							Ports:          containerPorts,
							VolumeMounts:   volumeMount,
						},
					},
					Tolerations: tolerations,
					Volumes:     volumes,
				},
			},
		},
	}
}

func GenerateDeploymentManifestWithPodTemplate(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace) *appsv1.Deployment {
	klog.InfoS("GenerateDeploymentManifestWithPodTemplate", "workspace", klog.KObj(workspaceObj))

	templateCopy := workspaceObj.Inference.Template.DeepCopy()
	if templateCopy.ObjectMeta.Labels == nil {
		templateCopy.ObjectMeta.Labels = make(map[string]string)
	}

	// Gather label requirements from workspaceObj's label selector
	labelRequirements := make([]v1.LabelSelectorRequirement, 0, len(workspaceObj.Resource.LabelSelector.MatchLabels))
	for key, value := range workspaceObj.Resource.LabelSelector.MatchLabels {
		labelRequirements = append(labelRequirements, v1.LabelSelectorRequirement{
			Key:      key,
			Operator: v1.LabelSelectorOpIn,
			Values:   []string{value},
		})
		templateCopy.ObjectMeta.Labels[key] = value
	}
	// Overwrite affinity
	templateCopy.Spec.Affinity = &corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
				{
					LabelSelector: &v1.LabelSelector{
						MatchExpressions: labelRequirements,
					},
					TopologyKey: "kubernetes.io/hostname",
				},
			},
		},
	}

	return &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      workspaceObj.Name,
			Namespace: workspaceObj.Namespace,
			OwnerReferences: []v1.OwnerReference{
				{
					APIVersion: kaitov1alpha1.GroupVersion.String(),
					Kind:       "Workspace",
					UID:        workspaceObj.UID,
					Name:       workspaceObj.Name,
				},
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: lo.ToPtr(int32(*workspaceObj.Resource.Count)),
			Selector: workspaceObj.Resource.LabelSelector,
			Template: *templateCopy,
		},
	}

}

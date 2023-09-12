package k8sresources

import (
	"context"

	kdmv1alpha1 "github.com/kdm/api/v1alpha1"
	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func GenerateDeploymentManifest(ctx context.Context, workspaceObj *kdmv1alpha1.Workspace, imageName string,
	replicas int, commands []string, containerPorts []corev1.ContainerPort,
	livenessProbe, readinessProbe *corev1.Probe, resourceRequirements corev1.ResourceRequirements,
	volumeMount []corev1.VolumeMount, tolerations []corev1.Toleration, volumes []corev1.Volume) *appsv1.Deployment {
	klog.InfoS("GenerateDeploymentManifest", "workspace", klog.KObj(workspaceObj), "image", imageName)

	return &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      workspaceObj.Name,
			Namespace: workspaceObj.Namespace,
			OwnerReferences: []v1.OwnerReference{
				{
					APIVersion: kdmv1alpha1.GroupVersion.String(),
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
					Tolerations:  tolerations,
					Volumes:      volumes,
					NodeSelector: workspaceObj.Resource.LabelSelector.MatchLabels,
				},
			},
		},
	}
}

func GenerateStatefulSetManifest(ctx context.Context, workspaceObj *kdmv1alpha1.Workspace, imageName string,
	replicas int, commands []string, containerPorts []corev1.ContainerPort,
	livenessProbe, readinessProbe *corev1.Probe, resourceRequirements corev1.ResourceRequirements,
	volumeMount []corev1.VolumeMount, tolerations []corev1.Toleration, volumes []corev1.Volume) *appsv1.StatefulSet {

	klog.InfoS("GenerateStatefulSetManifest", "workspace", klog.KObj(workspaceObj), "image", imageName)

	// Gather label requirements from workspaceObj's label selector
	var labelRequirements []v1.LabelSelectorRequirement
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
					APIVersion: kdmv1alpha1.GroupVersion.String(),
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

package k8sresources

import (
	"context"

	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateDeployment(ctx context.Context, deploymentObj *appsv1.Deployment, kubeClient client.Client) error {
	klog.InfoS("CreateDeployment", "service", klog.KObj(deploymentObj))
	return retry.OnError(retry.DefaultBackoff, func(err error) bool {
		return true
	}, func() error {
		return kubeClient.Create(ctx, deploymentObj, &client.CreateOptions{})
	})
}

func GenerateDeploymentManifest(ctx context.Context, depName, namespace, imageName string,
	replicas int, labelSelector *v1.LabelSelector, commands []string, containerPorts []corev1.ContainerPort,
	livenessProbe, readinessProbe *corev1.Probe, resourceRequirements corev1.ResourceRequirements,
	volumeMount []corev1.VolumeMount, tolerations []corev1.Toleration, volumes []corev1.Volume) *appsv1.Deployment {
	klog.InfoS("GenerateDeploymentManifest", "deployment", depName, "namespace", namespace, "image", imageName)

	return &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      depName,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: lo.ToPtr(int32(replicas)),
			Selector: labelSelector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: labelSelector.MatchLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:           depName,
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
					NodeSelector: labelSelector.MatchLabels,
				},
			},
		},
	}
}

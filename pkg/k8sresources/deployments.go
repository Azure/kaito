package k8sresources

import (
	"context"

	kdmv1alpha1 "github.com/kdm/api/v1alpha1"
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

func GetDeployment(ctx context.Context, name, namespace string, kubeClient client.Client) (*appsv1.Deployment, error) {
	klog.InfoS("GetDeployment", "deploymentName", name, "deploymentNamespace", namespace)

	dep := &appsv1.Deployment{}
	err := retry.OnError(retry.DefaultBackoff, func(err error) bool {
		return true
	}, func() error {
		return kubeClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, dep, &client.GetOptions{})
	})

	if err != nil {
		return nil, err
	}

	return dep, nil
}

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
					APIVersion: kdmv1alpha1.SchemeBuilder.GroupVersion.String(),
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

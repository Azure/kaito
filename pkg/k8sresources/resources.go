package k8sresources

import (
	"context"
	"fmt"
	kdmv1alpha1 "github.com/kdm/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func configVolume(wObj *kdmv1alpha1.Workspace) ([]corev1.Volume, []corev1.VolumeMount) {
	// TODO check if preset exists, template shouldn't.
	volume := wObj.Inference.Preset.Volume
	if volume == nil {
		volume = []corev1.Volume{}
	}
	volumeMount := []corev1.VolumeMount{}

	// Signifies multinode inference requirement
	if *wObj.Resource.Count > 1 {
		// Append share memory volume to any existing volumes
		volume = append(volume, corev1.Volume{
			Name: "dshm",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium: "Memory",
				},
			},
		})

		volumeMount = append(volumeMount, corev1.VolumeMount{
			Name:      volume[0].Name,
			MountPath: "/dev/shm",
		})
	}

	return volume, volumeMount
}

func CreateResource(ctx context.Context, resource client.Object, kubeClient client.Client) error {
	// Log the creation attempt.
	switch r := resource.(type) {
	case *appsv1.Deployment:
		klog.InfoS("CreateDeployment", "service", klog.KObj(r))
	case *appsv1.StatefulSet:
		klog.InfoS("CreateStatefulSet", "service", klog.KObj(r))
	}

	// Create the resource.
	return retry.OnError(retry.DefaultBackoff, func(err error) bool {
		return true
	}, func() error {
		return kubeClient.Create(ctx, resource, &client.CreateOptions{})
	})
}

func GetResource(ctx context.Context, name, namespace string, kubeClient client.Client, resource client.Object) error {
	// Log the retrieval attempt.
	resourceType := fmt.Sprintf("%T", resource)
	klog.InfoS(fmt.Sprintf("Get%s", resourceType), "resourceName", name, "resourceNamespace", namespace)

	err := retry.OnError(retry.DefaultBackoff, func(err error) bool {
		return true
	}, func() error {
		return kubeClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, resource, &client.GetOptions{})
	})

	return err
}

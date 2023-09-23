package k8sresources

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateResource(ctx context.Context, resource client.Object, kubeClient client.Client) error {
	// Log the creation attempt.
	switch r := resource.(type) {

	case *appsv1.Deployment:
		klog.InfoS("CreateDeployment", "deployment", klog.KObj(r))
	case *appsv1.StatefulSet:
		klog.InfoS("CreateStatefulSet", "statefulset", klog.KObj(r))
	case *corev1.Service:
		klog.InfoS("CreateService", "service", klog.KObj(r))
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

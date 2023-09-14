package k8sresources

import (
	"context"
	"fmt"

	kdmv1alpha1 "github.com/kdm/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateService(ctx context.Context, serviceObj *v1.Service, kubeClient client.Client) error {
	klog.InfoS("CreateService", "service", klog.KObj(serviceObj))
	return retry.OnError(retry.DefaultBackoff, func(err error) bool {
		return true
	}, func() error {
		return kubeClient.Create(ctx, serviceObj, &client.CreateOptions{})
	})
}

func GenerateServiceManifest(ctx context.Context, workspaceObj *kdmv1alpha1.Workspace) *v1.Service {
	klog.InfoS("GenerateServiceManifest", "workspace", klog.KObj(workspaceObj), "serviceType", v1.ServiceTypeLoadBalancer)

	selector := make(map[string]string)
	for k, v := range workspaceObj.Resource.LabelSelector.MatchLabels {
		selector[k] = v
	}
	podNameForIndex0 := fmt.Sprintf("%s-0", workspaceObj.Name)
	selector["statefulset.kubernetes.io/pod-name"] = podNameForIndex0

	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workspaceObj.Name,
			Namespace: workspaceObj.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: kdmv1alpha1.GroupVersion.String(),
					Kind:       "Workspace",
					UID:        workspaceObj.UID,
					Name:       workspaceObj.Name,
				},
			},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeLoadBalancer,
			Ports: []v1.ServicePort{
				{
					Protocol:   v1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(5000),
				},
			},
			Selector: selector,
		},
	}
}

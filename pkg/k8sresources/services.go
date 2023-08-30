package k8sresources

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateLoadBalancerService(ctx context.Context, serviceObj *v1.Service, kubeClient client.Client) error {
	klog.InfoS("CreateLoadBalancerService", "service", klog.KObj(serviceObj))
	return retry.OnError(retry.DefaultBackoff, func(err error) bool {
		return true
	}, func() error {
		return kubeClient.Create(ctx, serviceObj, &client.CreateOptions{})
	})
}

func GenerateLoadBalancerService(ctx context.Context, serviceName, namespace string, serviceType v1.ServiceType, labelSelector map[string]string) *v1.Service {
	klog.InfoS("GenerateLoadBalancerService", "serviceName", serviceName, "namespace", "serviceType", serviceType)

	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
		},
		Spec: v1.ServiceSpec{
			Type: serviceType, //v1.ServiceTypeLoadBalancer,
			Ports: []v1.ServicePort{
				{
					Protocol:   v1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(5000),
				},
			},
			Selector: labelSelector,
		},
	}
}

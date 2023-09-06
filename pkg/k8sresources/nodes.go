package k8sresources

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	LabelKeyNvidia               = "accelerator"
	LabelValueNvidia             = "nvidia"
	CapacityNvidiaGPU            = "nvidia.com/gpu"
	LabelKeyCustomGPUProvisioner = "gpu-provisioner.sh/machine-type"
	DADIDaemonSetName            = "teleportinstall"
	GPUProvisionerNamespace      = "gpu-provisioner"
	GPUString                    = "gpu"
)

// GetNode get kubernetes node object with a provided name
func GetNode(ctx context.Context, nodeName string, kubeClient client.Client) (*corev1.Node, error) {
	klog.InfoS("GetNode", "nodeName", nodeName)
	node := &corev1.Node{}

	err := kubeClient.Get(ctx, client.ObjectKey{Name: nodeName}, node, &client.GetOptions{})
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, fmt.Errorf("no node has been found with nodeName %s", nodeName)
	}
	return node, nil
}

// ListNodes get list of kubernetes nodes
func ListNodes(ctx context.Context, kubeClient client.Client, options *client.ListOptions) (*corev1.NodeList, error) {
	klog.InfoS("ListNodes", "listOptions", options)
	nodeList := &corev1.NodeList{}

	err := kubeClient.List(ctx, nodeList, options)
	if err != nil {
		return nil, err
	}

	return nodeList, nil
}

// UpdateNodeWithLabel update the node object with the label key/value
func UpdateNodeWithLabel(ctx context.Context, nodeName, labelKey, labelValue string, kubeClient client.Client) error {
	klog.InfoS("UpdateNodeWithLabel", "nodeName", nodeName, "labelKey", labelKey, "labelValue", labelValue)

	// get fresh node object
	freshNode, err := GetNode(ctx, nodeName, kubeClient)
	if err != nil {
		klog.ErrorS(err, "cannot get node", "node", nodeName)
		return err
	}

	freshNode.Labels = lo.Assign(freshNode.Labels, map[string]string{labelKey: labelValue})
	opt := &client.UpdateOptions{}

	err = kubeClient.Update(ctx, freshNode, opt)
	if err != nil {
		klog.ErrorS(err, "cannot update node label", "node", nodeName, labelKey, labelValue)
		return err
	}
	return nil
}

func CheckNvidiaPlugin(ctx context.Context, nodeObj *corev1.Node) bool {
	klog.InfoS("CheckNvidiaPlugin", "node", klog.KObj(nodeObj))
	// check if label accelerator=nvidia exists in the node
	var foundLabel, foundCapacity bool
	if nvidiaLabelVal, found := nodeObj.Labels[LabelKeyNvidia]; found {
		if nvidiaLabelVal == LabelValueNvidia {
			foundLabel = true
		}
	}

	// check Status.Capacity.nvidia.com/gpu has value
	capacity := nodeObj.Status.Capacity
	if capacity != nil && !capacity.Name(CapacityNvidiaGPU, "").IsZero() {
		foundCapacity = true
	}

	if foundLabel && foundCapacity {
		return true
	}
	return false
}

func CheckDADIPlugin(ctx context.Context, nodeObj *corev1.Node, kubeClient client.Client) error {
	klog.InfoS("CheckDADIPlugin", "node", klog.KObj(nodeObj))
	if customLabel, found := nodeObj.Labels[LabelKeyCustomGPUProvisioner]; found {
		if customLabel != GPUString {
			return nil
		}
	}
	return checkDaemonSetPodForNode(ctx, DADIDaemonSetName, nodeObj.Name, kubeClient)
}

func checkDaemonSetPodForNode(ctx context.Context, daemonSetName, nodeName string, kubeClient client.Client) error {
	klog.InfoS("checkDaemonSetPodForNode", "daemonSetName", daemonSetName, "nodeName", nodeName)
	podList := &corev1.PodList{}

	listOpt := &client.ListOptions{
		Namespace: GPUProvisionerNamespace,
		FieldSelector: fields.SelectorFromSet(fields.Set{
			"spec.nodeName": nodeName,
		}),
	}
	err := retry.OnError(retry.DefaultRetry, func(err error) bool {
		return true
	}, func() error {
		return kubeClient.List(ctx, podList, listOpt)
	})
	if err != nil {
		klog.ErrorS(err, "cannot get pods for daemonset plugin", "daemonset-name", daemonSetName, "daemonset-namespace", GPUProvisionerNamespace, "node", nodeName)
		return err
	}
	// check ownerReference is the required daemonset
	if len(podList.Items) == 0 {
		return fmt.Errorf("no pods have been found running on the node %s", nodeName)
	}

	for p := range podList.Items {
		if podList.Items[p].OwnerReferences[0].Kind == "DaemonSet" &&
			podList.Items[p].OwnerReferences[0].Name == DADIDaemonSetName {
			return nil
		}
	}
	return fmt.Errorf("%s daemonset's pod for the node %s is not running", daemonSetName, nodeName)
}

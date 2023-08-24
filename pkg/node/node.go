package node

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
	NvidiaDaemonSetName          = "nvidia-device-plugin-daemonset"
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
	klog.InfoS("ListNodes")
	nodeList := &corev1.NodeList{}

	err := kubeClient.List(ctx, nodeList, options)
	if err != nil {
		return nil, err
	}

	return nodeList, nil
}

func EnsureNodePlugins(ctx context.Context, nodeObj *corev1.Node, kubeClient client.Client) error {
	klog.InfoS("EnsureNodePlugins", "node", klog.KObj(nodeObj))
	//does node have vhd installed
	foundNvidiaPlugin, err := checkAndInstallNvidiaPlugin(ctx, nodeObj, kubeClient)
	if err != nil {
		return err
	}
	//does node have the custom label for DADI
	foundDADIPlugin, err := checkAndInstallDADI(ctx, nodeObj, kubeClient)
	if err != nil {
		return err
	}
	if foundNvidiaPlugin && foundDADIPlugin {
		return nil
	}
	return fmt.Errorf("one or more plugin is not ready, Nvidia=%t, DADI=%t", foundNvidiaPlugin, foundDADIPlugin)
}

func checkAndInstallNvidiaPlugin(ctx context.Context, nodeObj *corev1.Node, kubeClient client.Client) (bool, error) {
	klog.InfoS("checkAndInstallNvidiaPlugin", "node", klog.KObj(nodeObj))
	// check if label accelerator=nvidia exists in the node
	var foundLabel, foundCapacity bool
	if nvidiaLabelVal, found := nodeObj.Labels[LabelKeyNvidia]; found {
		if nvidiaLabelVal == LabelValueNvidia {
			foundLabel = true
		} else {
			nodeObj.Labels = lo.Assign(nodeObj.Labels, map[string]string{LabelKeyCustomGPUProvisioner: GPUString})
			err := kubeClient.Patch(ctx, nodeObj, client.MergeFrom(nodeObj))
			if err != nil {
				klog.ErrorS(err, "cannot update node with custom label to enable Nvidia plugin", "node", nodeObj.Name, LabelKeyCustomGPUProvisioner, GPUString)
				return false, err
			}
		}
	}

	// check Status.Capacity.nvidia.com/gpu has value
	capacity := nodeObj.Status.Capacity
	if capacity != nil && !capacity.Name(CapacityNvidiaGPU, "").IsZero() {
		foundCapacity = true
	}

	if foundLabel && foundCapacity {
		return true, nil
	}

	klog.ErrorS(fmt.Errorf("nvidia plugin cannot be installed"), "node", nodeObj.Name)
	return false, nil
}

func checkAndInstallDADI(ctx context.Context, nodeObj *corev1.Node, kubeClient client.Client) (bool, error) {
	klog.InfoS("checkAndInstallDADI", "node", klog.KObj(nodeObj))
	if customLabel, found := nodeObj.Labels[LabelKeyCustomGPUProvisioner]; found {
		if customLabel != GPUString {
			nodeObj.Labels = lo.Assign(nodeObj.Labels, map[string]string{LabelKeyCustomGPUProvisioner: GPUString})
			err := kubeClient.Patch(ctx, nodeObj, client.MergeFrom(nodeObj))
			if err != nil {
				klog.ErrorS(err, "cannot update node with custom label to enable DADI plugin", "node", nodeObj.Name, LabelKeyCustomGPUProvisioner, GPUString)
				return false, err
			}
		}
	}

	return checkDaemonSetPodForNode(ctx, DADIDaemonSetName, nodeObj.Name, kubeClient)
}

func checkDaemonSetPodForNode(ctx context.Context, daemonSetName, nodeName string, kubeClient client.Client) (bool, error) {
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
		return false, err
	}
	// check ownerReference is the required daemonset
	for p := range podList.Items {
		if podList.Items[p].OwnerReferences[0].Kind == "DaemonSet" &&
			podList.Items[p].OwnerReferences[0].Name == DADIDaemonSetName {
			return true, nil
		}
	}
	return false, fmt.Errorf("%s daemonset's pod for the node %s is not running", daemonSetName, nodeName)
}

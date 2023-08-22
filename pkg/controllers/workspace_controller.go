package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	kdmv1alpha1 "github.com/kdm/api/v1alpha1"
	"github.com/kdm/pkg/node"
	"github.com/samber/lo"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	NvidiaLabelKey            = "accelerator"
	NvidiaLabelValue          = "nvidia"
	CapacityNvidiaGPU         = "nvidia.com/gpu"
	GPUProvisionerCustomLabel = "gpu-provisioner.sh/machine-type"
	DADIDaemonSetName         = "teleportinstall"
	DADIDaemonSetNamespace    = "gpu-provisioner"
)

type WorkspaceReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

func (c *WorkspaceReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	klog.InfoS("Reconciling", "workspace", req.NamespacedName)
	workspaceObj := &kdmv1alpha1.Workspace{}
	if err := c.Client.Get(ctx, req.NamespacedName, workspaceObj); err != nil {
		klog.ErrorS(err, "failed to get workspace", "workspace", req.Name)
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// Handle deleting workspace, garbage collect all the resources.
	if !workspaceObj.DeletionTimestamp.IsZero() {
		klog.InfoS("the workspace is in the process of being deleted", "workspace", klog.KObj(workspaceObj))
		return c.garbageCollectWorkspace(ctx, workspaceObj)
	}

	// Read ResourceSpec
	err := c.applyWorkspaceResource(ctx, workspaceObj)
	if err != nil {
		return reconcile.Result{}, err
	}
	// TODO apply InferenceSpec
	// TODO apply TrainingSpec
	// TODO Update workspace status
	c.updateWorkspaceCondition(ctx, workspaceObj, kdmv1alpha1.WorkspaceConditionTypeReady, metav1.ConditionTrue, "Done", "Done")

	return reconcile.Result{}, nil
}

func (c *WorkspaceReconciler) applyWorkspaceResource(ctx context.Context, wObj *kdmv1alpha1.Workspace) error {
	klog.InfoS("applyWorkspaceResource", "workspace", klog.KObj(wObj))
	// Check CandidateNodes, if the count, instance type
	candidateNodeCount, err := c.validateCandidateNodes(ctx, wObj)
	if err != nil {
		return err
	}

	remainingNodeCount := lo.FromPtr(wObj.Resource.Count) - candidateNodeCount
	// if current node Count == workspace count, then all good and return
	if remainingNodeCount <= 0 {
		klog.InfoS("number of candidate nodes, is equal or greater than required workspace count", "workspace.Count", lo.FromPtr(wObj.Resource.Count))
		return nil
	} else {
		klog.InfoS("need to create more nodes", "remainingNodeCount", remainingNodeCount)
		klog.InfoS("Will start building a machine")
		// TODO Create machine CR

	}

	// TODO Add nodes names to the WorkspaceStatus.WorkerNodes[]

	return nil
}

func (c *WorkspaceReconciler) validateCandidateNodes(ctx context.Context, wObj *kdmv1alpha1.Workspace) (int, error) {
	klog.InfoS("validateCandidateNodes", "workspace", klog.KObj(wObj))
	validNodesCount := 0
	nodeList := wObj.Resource.CandidateNodes

	if nodeList == nil {
		klog.InfoS("CandidateNodes is empty")
		return validNodesCount, nil
	}
	klog.InfoS("found candidate node list", "length", len(nodeList))

	var foundInstanceType, foundLabels, foundVHD, foundDADI bool
NodeListLoop:
	for index := range nodeList {
		nodeObj, err := node.GetNode(ctx, nodeList[index], c.Client)
		if err != nil {
			klog.ErrorS(err, "cannot get node", "name", nodeList[index])
			continue NodeListLoop
		}
		// check if node has the required instanceType
		if instanceTypeLabel, found := nodeObj.Labels[corev1.LabelInstanceTypeStable]; found {
			if instanceTypeLabel != wObj.Resource.InstanceType {
				klog.InfoS("node has instance type which does not match the workspace instance type", "node",
					nodeObj.Name, "InstanceType", wObj.Resource.InstanceType)
				foundInstanceType = false
				continue NodeListLoop
			}
			klog.InfoS("node instance type matches the workspace one", "node",
				nodeObj.Name, "InstanceType", wObj.Resource.InstanceType)
			foundInstanceType = true
		}

		// check if node has the required label selectors
		if wObj.Resource.LabelSelector != nil && len(wObj.Resource.LabelSelector.MatchLabels) != 0 {
			foundAllLabels := false
			for k, v := range wObj.Resource.LabelSelector.MatchLabels {
				if nodeObj.Labels[k] != v {
					klog.InfoS("workspace has label selector which does not match or exist on node", "workspace",
						wObj.Name, "node", nodeObj.Name)
					foundAllLabels = false
					continue NodeListLoop
				}
				klog.InfoS("node Label Selector matches the workspace one", "node", nodeObj.Name)
				foundAllLabels = true
			}
			if foundAllLabels {
				foundLabels = true
			}
		} else {
			klog.InfoS("no Label Selector sets for the workspace", "workspace", wObj.Name)
			foundLabels = true
		}

		//does node have vhd installed
		foundVHD = c.IsNvidiaDriverInstalled(ctx, nodeObj)

		//does node have the custom label for DADI
		foundDADI, err = c.checkAndInstallDADI(ctx, nodeObj)
		if err != nil {
			continue NodeListLoop
		}

		if foundInstanceType && foundLabels && foundVHD && foundDADI {
			klog.InfoS("found a candidate node", "name", nodeObj.Name)
			validNodesCount++
		}
	}
	klog.InfoS("found valid nodes", "count", validNodesCount)
	return validNodesCount, nil
}

func (c *WorkspaceReconciler) IsNvidiaDriverInstalled(ctx context.Context, nodeObj *corev1.Node) bool {
	// check if label accelerator=nvidia exists in the node
	var foundLabel, foundCapacity bool
	if nvidiaLabelVal, found := nodeObj.Labels[NvidiaLabelKey]; found {
		if nvidiaLabelVal == NvidiaLabelValue {
			//klog.InfoS("nvidia accelerator label has been found", "node", nodeObj.Name)
			foundLabel = true
		}
	}

	// check Status.Capacity.nvidia.com/gpu has value
	capacity := nodeObj.Status.Capacity
	if capacity != nil && !capacity.Name(CapacityNvidiaGPU, "").IsZero() {
		//klog.InfoS("nvidia GPU capacity value found greater than 0", "node", nodeObj.Name, CapacityNvidiaGPU, capacity.Name(CapacityNvidiaGPU, "").Value())
		foundCapacity = true
	}

	if foundLabel && foundCapacity {
		return true
	}
	klog.ErrorS(fmt.Errorf("nvidia plugin is not installed"), "node", nodeObj.Name, CapacityNvidiaGPU)

	return false
}

func (c *WorkspaceReconciler) checkAndInstallDADI(ctx context.Context, nodeObj *corev1.Node) (bool, error) {
	var installed bool
	ds := &apps.DaemonSet{}

	err := c.Client.Get(ctx, client.ObjectKey{Name: DADIDaemonSetName, Namespace: DADIDaemonSetNamespace}, ds, &client.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "cannot get DADI daemonset plugin", "daemonset-name", DADIDaemonSetName, "daemonset-namespace", DADIDaemonSetNamespace)
		return false, err
	}

	if ds.Status.NumberAvailable < 0 {
		klog.ErrorS(err, "DADI daemonset plugin is not running", "daemonset-name", DADIDaemonSetName, "daemonset-namespace", DADIDaemonSetNamespace)
		return false, err
	}

	if customLabel, found := nodeObj.Labels[GPUProvisionerCustomLabel]; found {
		if customLabel == "gpu" {
			klog.InfoS("the custom gpu-provisioner label has been found", "node", nodeObj.Name, GPUProvisionerCustomLabel, "gpu")
			installed = true
		}
	} else {
		nodeObj.Labels = lo.Assign(nodeObj.Labels, map[string]string{GPUProvisionerCustomLabel: "gpu"})
		err := c.Client.Update(ctx, nodeObj, &client.UpdateOptions{})
		if err != nil {
			klog.ErrorS(err, "cannot update node with custom label to enable DADI plugin", "node", nodeObj.Name, GPUProvisionerCustomLabel, "gpu")
			return false, err
		}
		installed = true
	}
	return installed, nil
}

func (c *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c.Recorder = mgr.GetEventRecorderFor("Workspace")
	return ctrl.NewControllerManagedBy(mgr).
		For(&kdmv1alpha1.Workspace{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(c)
}

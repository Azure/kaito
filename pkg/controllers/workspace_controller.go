package controllers

import (
	"context"

	"github.com/go-logr/logr"
	kdmv1alpha1 "github.com/kdm/api/v1alpha1"
	"github.com/kdm/pkg/node"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type WorkspaceReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=kdm.io,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kdm.io,resources=workspaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kdm.io,resources=workspaces/finalizers,verbs=update

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

	wObj := workspaceObj.DeepCopy()
	// Add the finalizer to the workspace
	if err := c.addFinalizer(ctx, wObj); err != nil {
		klog.ErrorS(err, "failed to add the finalizer to workspace", "workspace", klog.KObj(wObj))
		return ctrl.Result{}, err
	}

	// Read ResourceSpec
	err := c.applyWorkspaceResource(ctx, wObj)
	if err != nil {
		return reconcile.Result{}, err
	}
	// TODO apply InferenceSpec
	// TODO apply TrainingSpec
	// TODO Update workspace status
	c.updateWorkspaceCondition(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeReady, metav1.ConditionTrue, "Done", "Done")

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
		klog.InfoS("number of candidate nodes, %d, is equal or greater than required workspace count, %d ",
			candidateNodeCount, lo.FromPtr(wObj.Resource.Count))
		return nil
	}
	klog.InfoS("Will start building a machine")

	// Create machine CR

	// Add nodes names to the WorkspaceStatus.WorkerNodes[]

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
	var foundNode, foundInstanceType, foundLabels, foundVHD, foundDADI bool
	for index := range nodeList {
		nodeObj, err := node.GetNode(ctx, nodeList[index], c.Client)
		if err != nil {
			klog.ErrorS(err, "cannot get node with name %s", nodeList[index])
			continue
		}
		foundNode = true
		// check if node has the required instanceType
		nodeLabels := nodeObj.Labels
		if instanceTypeLabel, found := nodeLabels[corev1.LabelInstanceTypeStable]; found {
			if instanceTypeLabel != wObj.Resource.InstanceType {
				klog.InfoS("node %s has instance type %s which does not match the workspace instance type (%s)",
					nodeObj.Name, instanceTypeLabel, wObj.Resource.InstanceType)
				foundInstanceType = false
				continue
			}
			foundInstanceType = true
		}
		// check if node has the required label selectors
		for k, v := range wObj.Resource.LabelSelector.MatchLabels {
			if nodeLabels[k] != v {
				klog.InfoS("workspace %s has label selector ,%s=%s, which does not match or exist on node %s ",
					wObj.Name, k, v, nodeObj.Name)
				foundLabels = false
				continue
			}
			foundLabels = true
		}

		// TODO
		//does node have vhd installed
		foundVHD = true
		// TODO
		//does node have the custom label for DADI
		foundDADI = true
		if foundNode && foundInstanceType && foundLabels && foundVHD && foundDADI {
			validNodesCount++
		}
	}

	return validNodesCount, nil
}

func (c *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kdmv1alpha1.Workspace{}).
		Complete(c)
}

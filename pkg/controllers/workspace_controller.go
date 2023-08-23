package controllers

import (
	"context"

	"github.com/go-logr/logr"
	kdmv1alpha1 "github.com/kdm/api/v1alpha1"
	"github.com/kdm/pkg/node"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type WorkspaceReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

func (c *WorkspaceReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	workspaceObj := &kdmv1alpha1.Workspace{}
	if err := c.Client.Get(ctx, req.NamespacedName, workspaceObj); err != nil {
		if !errors.IsNotFound(err) {
			klog.ErrorS(err, "failed to get workspace", "workspace", req.Name)
		}
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}
	klog.InfoS("Reconciling", "workspace", req.NamespacedName)
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
	// Check CandidateNodes
	var validNodeList []*corev1.Node

	validCandidateNodeList, err := c.validateCandidateNodes(ctx, wObj)
	if err != nil {
		return err
	}
	validNodeList = append(validNodeList, validCandidateNodeList...)

	// Check the current cluster nodes if they match the labelSelector and instanceType
	validCurrentClusterNodeList, err := c.validateCurrentClusterNodes(ctx, wObj)
	if err != nil {
		return err
	}

	validNodeList = append(validNodeList, validCurrentClusterNodeList...)

	// remove duplicate nodes
	validNodeList = lo.FindUniques(validNodeList)

	validNodeCount := len(validNodeList)
	// subtract all valid nodes from the desired count
	remainingNodeCount := lo.FromPtr(wObj.Resource.Count) - validNodeCount

	// if current node Count == workspace count, then all good and return
	if remainingNodeCount <= 0 {
		klog.InfoS("number of candidate nodes, is equal or greater than required workspace count", "workspace.Count", lo.FromPtr(wObj.Resource.Count))
		return nil
	} else {
		klog.InfoS("need to create more nodes", "remainingNodeCount", remainingNodeCount)
		klog.InfoS("Will start building a machine")
		// TODO Create machine CR

	}

	// Ensure all node plugins are running successfully
	for i := range validNodeList {
		err = node.EnsureNodePlugins(ctx, validNodeList[i], c.Client)
		if err != nil {
			return err
		}
	}
	// Add the valid nodes names to the WorkspaceStatus.WorkerNodes
	err = c.updateWorkspaceStatusWithNodeList(ctx, wObj, validNodeList)
	if err != nil {
		return err
	}

	return nil
}

func (c *WorkspaceReconciler) validateCandidateNodes(ctx context.Context, wObj *kdmv1alpha1.Workspace) ([]*corev1.Node, error) {
	klog.InfoS("validateCandidateNodes", "workspace", klog.KObj(wObj))
	var validNodeList []*corev1.Node
	nodeList := wObj.Resource.CandidateNodes
	if nodeList == nil {
		klog.InfoS("CandidateNodes is empty")
		return nil, nil
	}

	klog.InfoS("found candidate node list", "length", len(nodeList))

	var foundInstanceType, foundLabel bool
	for index := range nodeList {
		nodeObj, err := node.GetNode(ctx, nodeList[index], c.Client)
		if err != nil {
			klog.ErrorS(err, "cannot get node", "name", nodeList[index])
			continue
		}
		foundLabel = c.validateNodeLabelSelector(ctx, wObj, nodeObj)
		foundInstanceType = c.validateNodeInstanceType(ctx, wObj, nodeObj)

		if foundInstanceType && foundLabel {
			klog.InfoS("found a candidate node", "name", nodeObj.Name)
			validNodeList = append(validNodeList, nodeObj)
		}
	}

	klog.InfoS("found valid nodes", "count", len(validNodeList))
	return validNodeList, nil
}

func (c *WorkspaceReconciler) validateCurrentClusterNodes(ctx context.Context, wObj *kdmv1alpha1.Workspace) ([]*corev1.Node, error) {
	klog.InfoS("validateCurrentClusterNodes", "workspace", klog.KObj(wObj))
	var validNodeList []*corev1.Node
	opt := &client.ListOptions{}
	if wObj.Resource.LabelSelector != nil && len(wObj.Resource.LabelSelector.MatchLabels) != 0 {
		opt.LabelSelector = labels.SelectorFromSet(wObj.Resource.LabelSelector.MatchLabels)

	} else {
		klog.InfoS("no Label Selector sets for the workspace", "workspace", wObj.Name)
		return nil, nil
	}

	nodeList, err := node.ListNodes(ctx, c.Client, opt)
	if err != nil {
		return nil, err
	}
	if nodeList == nil {
		klog.InfoS("CandidateNodes is empty")
		return nil, nil
	}

	klog.InfoS("found candidate node list", "length", len(nodeList.Items))

	var foundInstanceType bool
	for index := range nodeList.Items {
		nodeObj := nodeList.Items[index]
		// Check if current node exists in the candidateNodes list
		if _, found := lo.Find(wObj.Resource.CandidateNodes, func(item string) bool {
			return item == nodeObj.Name
		}); found {
			continue
		}
		foundInstanceType = c.validateNodeInstanceType(ctx, wObj, lo.ToPtr(nodeObj))

		if foundInstanceType {
			klog.InfoS("found a current valid node", "name", nodeObj.Name)
			validNodeList = append(validNodeList, lo.ToPtr(nodeObj))
		}
	}

	klog.InfoS("found current valid nodes", "count", len(validNodeList))
	return validNodeList, nil
}

// check if node has the required label selectors
func (c *WorkspaceReconciler) validateNodeLabelSelector(ctx context.Context, wObj *kdmv1alpha1.Workspace, nodeObj *corev1.Node) bool {
	if wObj.Resource.LabelSelector != nil && len(wObj.Resource.LabelSelector.MatchLabels) != 0 {
		for k, v := range wObj.Resource.LabelSelector.MatchLabels {
			if nodeObj.Labels[k] != v {
				klog.InfoS("workspace has label selector which does not match or exist on node", "workspace",
					wObj.Name, "node", nodeObj.Name)
				return false
			}
		}
	} else {
		klog.InfoS("no Label Selector sets for the workspace", "workspace", wObj.Name)
		return true
	}
	return true
}

// check if node has the required instanceType
func (c *WorkspaceReconciler) validateNodeInstanceType(ctx context.Context, wObj *kdmv1alpha1.Workspace, nodeObj *corev1.Node) bool {
	if instanceTypeLabel, found := nodeObj.Labels[corev1.LabelInstanceTypeStable]; found {
		if instanceTypeLabel != wObj.Resource.InstanceType {
			klog.InfoS("node has instance type which does not match the workspace instance type", "node",
				nodeObj.Name, "InstanceType", wObj.Resource.InstanceType)
			return false
		}
	}
	klog.InfoS("node instance type matches the workspace one", "node",
		nodeObj.Name, "InstanceType", wObj.Resource.InstanceType)
	return true
}

func (c *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c.Recorder = mgr.GetEventRecorderFor("Workspace")
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Pod{}, "spec.nodeName", func(rawObj client.Object) []string {
		pod := rawObj.(*corev1.Pod)
		return []string{pod.Spec.NodeName}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&kdmv1alpha1.Workspace{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(c)
}

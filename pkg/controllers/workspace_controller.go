package controllers

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-logr/logr"
	kdmv1alpha1 "github.com/kdm/api/v1alpha1"
	"github.com/kdm/pkg/inference"
	"github.com/kdm/pkg/k8sresources"
	"github.com/kdm/pkg/machine"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/apis/core"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var torchRunParams = map[string]string{
	"max_seq_len": "128",
}

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
		return c.deleteWorkspace(ctx, workspaceObj)
	}

	return c.addOrUpdateWorkspace(ctx, workspaceObj)

}

func (c *WorkspaceReconciler) addOrUpdateWorkspace(ctx context.Context, wObj *kdmv1alpha1.Workspace) (reconcile.Result, error) {
	// Read ResourceSpec
	err := c.applyWorkspaceResource(ctx, wObj)
	if err != nil {
		return reconcile.Result{}, err
	}
	// TODO apply TrainingSpec

	if wObj.GetAnnotations() != nil {
		err := c.applyAnnotations(ctx, wObj)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	err = c.applyInference(ctx, wObj)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = c.setWorkspaceStatusCondition(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeReady, metav1.ConditionTrue, "workspaceReady", "workspace is ready")
	if err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (c *WorkspaceReconciler) deleteWorkspace(ctx context.Context, wObj *kdmv1alpha1.Workspace) (reconcile.Result, error) {
	klog.InfoS("deleteWorkspace", "workspace", klog.KObj(wObj))
	// TODO delete workspace, machine(s), training and inference (deployment, service) obj ( ok to delete machines? which will delete nodes??)
	err := c.setWorkspaceStatusCondition(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeDeleting, metav1.ConditionTrue, "workspaceDeleted", "workspace is being deleted")
	if err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
		return reconcile.Result{}, err
	}

	return c.garbageCollectWorkspace(ctx, wObj)
}

func (c *WorkspaceReconciler) applyWorkspaceResource(ctx context.Context, wObj *kdmv1alpha1.Workspace) error {
	klog.InfoS("applyWorkspaceResource", "workspace", klog.KObj(wObj))
	validNodeList := []*corev1.Node{}

	// Check the current cluster nodes if they match the labelSelector and instanceType
	validCurrentClusterNodeList, err := c.validateCurrentClusterNodes(ctx, wObj)
	if err != nil {
		return err
	}

	// Check preferredNodes
	preferredList := wObj.Resource.PreferredNodes
	if preferredList == nil {
		klog.InfoS("PreferredNodes list is empty")
	} else {
		for n := range preferredList {
			lo.Find(validNodeList, func(nodeItem *corev1.Node) bool {
				if nodeItem.Name == preferredList[n] {
					validNodeList = append(validNodeList, nodeItem)
					return true
				}
				// else do nothing for now
				return false
			})
		}
	}
	// TODO check nodes in the WorkspaceStatus.WorkerNodes.

	for n := range validCurrentClusterNodeList {
		if len(validNodeList) == lo.FromPtr(wObj.Resource.Count) {
			break
		}
		_, found := lo.Find(validNodeList, func(nodeItem *corev1.Node) bool {
			return nodeItem.Name == validCurrentClusterNodeList[n].Name
		})
		if !found {
			validNodeList = append(validNodeList, validCurrentClusterNodeList[n])
		}
	}

	validNodeCount := len(validNodeList)
	// subtract all valid nodes from the desired count
	remainingNodeCount := lo.FromPtr(wObj.Resource.Count) - validNodeCount

	// if current valid nodes Count == workspace count, then all good and return
	if remainingNodeCount == 0 {
		klog.InfoS("number of existing nodes are equal to the required workspace count", "workspace.Count", lo.FromPtr(wObj.Resource.Count))
	} else {
		klog.InfoS("need to create more nodes", "NodeCount", remainingNodeCount)
		for i := 0; i < remainingNodeCount; i++ {
			newNode, err := c.createAndValidateNode(ctx, wObj)
			if err != nil {
				return err
			}
			validNodeList = append(validNodeList, newNode)
		}
	}

	// Ensure all nodes plugins are running successfully
	for i := range validNodeList {
		if err := c.setWorkspaceStatusCondition(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeInstallNodePlugins, metav1.ConditionFalse, "installNodePlugins",
			fmt.Sprintf("installing plugins for node %s", validNodeList[i].Name)); err != nil {
			klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
			return err
		}
		err = c.ensureNodePlugins(ctx, wObj, validNodeList[i])
		if err != nil {
			if err := c.setWorkspaceStatusCondition(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeInstallNodePlugins, metav1.ConditionFalse,
				"installNodePlugins", err.Error()); err != nil {
				klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
				return err
			}
			return err
		}
	}

	// Add the valid nodes names to the WorkspaceStatus.WorkerNodes
	err = c.updateWorkspaceStatusWithNodeList(ctx, wObj, validNodeList)
	if err != nil {
		return err
	}

	err = c.setWorkspaceStatusCondition(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeResourceProvisioned, metav1.ConditionTrue,
		"workspaceResourceDeployedSuccess", "workspace resource is ready")
	if err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
		return err
	}

	return nil
}

func (c *WorkspaceReconciler) validateCurrentClusterNodes(ctx context.Context, wObj *kdmv1alpha1.Workspace) ([]*corev1.Node, error) {
	klog.InfoS("validateCurrentClusterNodes", "workspace", klog.KObj(wObj))
	var validCurrentNodeList []*corev1.Node
	opt := &client.ListOptions{}
	if wObj.Resource.LabelSelector != nil && len(wObj.Resource.LabelSelector.MatchLabels) != 0 {
		opt.LabelSelector = labels.SelectorFromSet(wObj.Resource.LabelSelector.MatchLabels)

	} else {
		klog.InfoS("no Label Selector sets for the workspace", "workspace", wObj.Name)
		return nil, nil
	}

	nodeList, err := k8sresources.ListNodes(ctx, c.Client, opt)
	if err != nil {
		return nil, err
	}
	if len(nodeList.Items) == 0 {
		klog.InfoS("no current nodes match the workspace resource spec", "workspace", wObj.Name)
		return nil, nil
	}

	var foundInstanceType bool
	for index := range nodeList.Items {
		nodeObj := nodeList.Items[index]
		foundInstanceType = c.validateNodeInstanceType(ctx, wObj, lo.ToPtr(nodeObj))
		if foundInstanceType {
			klog.InfoS("found a current valid node", "name", nodeObj.Name)
			validCurrentNodeList = append(validCurrentNodeList, lo.ToPtr(nodeObj))
		}
	}

	klog.InfoS("found current valid nodes", "count", len(validCurrentNodeList))
	return validCurrentNodeList, nil
}

// check if node has the required instanceType
func (c *WorkspaceReconciler) validateNodeInstanceType(ctx context.Context, wObj *kdmv1alpha1.Workspace, nodeObj *corev1.Node) bool {
	klog.InfoS("validateNodeInstanceType", "workspace", klog.KObj(wObj))

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

func (c *WorkspaceReconciler) createAndValidateNode(ctx context.Context, wObj *kdmv1alpha1.Workspace) (*corev1.Node, error) {
	newMachine := machine.GenerateMachineManifest(ctx, wObj)

	if err := c.setWorkspaceStatusCondition(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeMachineProvisioned, metav1.ConditionFalse,
		"machineProvisioning", fmt.Sprintf("machine %s is getting provisioned", newMachine.Name)); err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
		return nil, err
	}

	err := machine.CreateMachine(ctx, newMachine, c.Client)
	if err != nil {
		klog.ErrorS(err, "failed to create machine", "machine", newMachine.Name)
		if err := c.setWorkspaceStatusCondition(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeMachineProvisioned, metav1.ConditionFalse,
			"machineFailedProvision", err.Error()); err != nil {
			klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
			return nil, err
		}
		return nil, err
	}
	klog.InfoS("a new machine has been created", "machine", newMachine.Name)

	err = c.setWorkspaceStatusCondition(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeMachineProvisioned, metav1.ConditionTrue,
		"machineProvisionSuccess", "machine has been provisioned successfully")
	if err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
		return nil, err
	}

	if err := c.setWorkspaceStatusCondition(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeMachineStatus, metav1.ConditionFalse,
		"checkMachineStatus", fmt.Sprintf("checking machine %s status", newMachine.Name)); err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
		return nil, err
	}
	// check machine status until it's ready
	if err := machine.CheckMachineStatus(ctx, newMachine, c.Client); err != nil {
		if err := c.setWorkspaceStatusCondition(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeMachineStatus, metav1.ConditionFalse,
			"checkMachineStatusFailed", err.Error()); err != nil {
			klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
			return nil, err
		}
		return nil, err
	}

	if err := c.setWorkspaceStatusCondition(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeMachineStatus, metav1.ConditionTrue,
		"checkMachineStatusSuccess", "check machine status successfully"); err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
		return nil, err
	}

	nodeName := newMachine.Status.NodeName
	if nodeName == "" {
		// TODO retry get machine
	}
	nodeObj, err := k8sresources.GetNode(ctx, nodeName, c.Client)
	if err != nil {
		if err := c.setWorkspaceStatusCondition(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeMachineStatus, metav1.ConditionFalse,
			"checkMachineStatusFailed", err.Error()); err != nil {
			klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
			return nil, err
		}
		return nil, err
	}

	err = c.setWorkspaceStatusCondition(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeMachineStatus, metav1.ConditionTrue,
		"installNodePluginsSuccess", "machines plugins have been installed successfully")
	if err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
		return nil, err
	}
	return nodeObj, nil
}

func (c *WorkspaceReconciler) ensureNodePlugins(ctx context.Context, wObj *kdmv1alpha1.Workspace, nodeObj *corev1.Node) error {
	klog.InfoS("EnsureNodePlugins", "node", klog.KObj(nodeObj))

	var foundNvidiaPlugin bool

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if nodeObj == nil {
				return errors.NewNotFound(core.Resource("nodes"), nodeObj.Name)
			}

			//Nvidia Plugin
			foundNvidiaPlugin = k8sresources.CheckNvidiaPlugin(ctx, nodeObj)
			if !foundNvidiaPlugin {
				err := k8sresources.UpdateNodeWithLabel(ctx, nodeObj.Name, k8sresources.LabelKeyNvidia, k8sresources.LabelValueNvidia, c.Client)
				if err != nil {
					if errors.IsNotFound(err) {
						klog.ErrorS(err, "nvidia plugin cannot be installed, node not found", "node", nodeObj.Name)
						return err
					}
					if err := c.setConditionInstallNodePluginsToUnknown(ctx, wObj, nodeObj); err != nil {
						return err
					}
					time.Sleep(1 * time.Second)
					continue
				}
			}

			//DADI plugin
			err := k8sresources.CheckDADIPlugin(ctx, nodeObj, c.Client)
			if err != nil {
				err = k8sresources.UpdateNodeWithLabel(ctx, nodeObj.Name, k8sresources.LabelKeyCustomGPUProvisioner, k8sresources.GPUString, c.Client)
				if err != nil {
					if errors.IsNotFound(err) {
						klog.ErrorS(err, "DADI plugin cannot be installed, node not found", "node", nodeObj.Name)
						return err
					}
					if err := c.setConditionInstallNodePluginsToUnknown(ctx, wObj, nodeObj); err != nil {
						return err
					}
					time.Sleep(1 * time.Second)
					continue
				}
			}

			// Update status
			if err := c.setWorkspaceStatusCondition(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeInstallNodePlugins, metav1.ConditionTrue,
				"InstallNodePluginsSuccess", "node plugins have been installed"); err != nil {
				klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
				return err
			}
			return nil
		}
	}
}

func (c *WorkspaceReconciler) setConditionInstallNodePluginsToUnknown(ctx context.Context, wObj *kdmv1alpha1.Workspace, nodeObj *corev1.Node) error {
	klog.InfoS("setConditionInstallNodePluginsToUnknown", "workspace", klog.KObj(wObj), "node", klog.KObj(nodeObj))
	err := c.setWorkspaceStatusCondition(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeInstallNodePlugins, metav1.ConditionUnknown, "InstallNodePluginsWaiting",
		fmt.Sprintf("waiting for plugins to get installed on node %s", nodeObj.Name))
	if err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
		return err
	}
	return nil
}

func (c *WorkspaceReconciler) applyAnnotations(ctx context.Context, wObj *kdmv1alpha1.Workspace) error {
	klog.InfoS("applyAnnotations", "workspace", klog.KObj(wObj))
	serviceType := corev1.ServiceTypeClusterIP
	wAnnotation := wObj.GetAnnotations()

	if len(wAnnotation) != 0 {
		_, found := lo.FindKey(wAnnotation, kdmv1alpha1.ServiceTypeLoadBalancer)
		if found {
			serviceType = corev1.ServiceTypeLoadBalancer
		}
	}

	//TODO generate more strong random service name
	serviceObj := k8sresources.GenerateLoadBalancerService(ctx, fmt.Sprint(wObj.Name, "-scv-", rand.Intn(100_000)), wObj.Namespace, serviceType, wObj.Resource.LabelSelector.MatchLabels)
	err := k8sresources.CreateLoadBalancerService(ctx, serviceObj, c.Client)
	if err != nil {
		return err
	}

	klog.InfoS("a service has been created for workspace", "workspace", klog.KObj(wObj), "serviceType", serviceType)
	return nil
}

func (c *WorkspaceReconciler) applyInference(ctx context.Context, wObj *kdmv1alpha1.Workspace) error {
	klog.InfoS("applyInference", "service", klog.KObj(wObj))

	// TODO check if preset exists, template shouldn't.
	volume := wObj.Inference.Preset.Volume
	if volume == nil {
		volume = []corev1.Volume{}
	}

	presetName := wObj.Inference.Preset.Name
	var err error
	switch presetName {
	case kdmv1alpha1.PresetSetModelllama2A:
		err = inference.CreateLLAMA2APresetModel(ctx, wObj.Name, wObj.Namespace, wObj.Resource.LabelSelector, volume, torchRunParams, c.Client)
	case kdmv1alpha1.PresetSetModelllama2B:
		err = inference.CreateLLAMA2BPresetModel(ctx, wObj.Name, wObj.Namespace, wObj.Resource.LabelSelector, volume, torchRunParams, c.Client)
	case kdmv1alpha1.PresetSetModelllama2C:
		err = inference.CreateLLAMA2CPresetModel(ctx, wObj.Name, wObj.Namespace, wObj.Resource.LabelSelector, volume, torchRunParams, c.Client)
	default:
		err = fmt.Errorf("preset model %s is not supported", presetName)
		klog.ErrorS(err, "no inference has been created")
	}
	if err != nil {
		return err
	}
	return nil
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

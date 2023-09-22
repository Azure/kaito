package controllers

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	"strconv"
	"time"

	"github.com/aws/karpenter-core/pkg/apis/v1alpha5"
	"github.com/go-logr/logr"
	kdmv1alpha1 "github.com/kdm/api/v1alpha1"
	"github.com/kdm/pkg/inference"
	"github.com/kdm/pkg/k8sresources"
	"github.com/kdm/pkg/machine"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/apis/core"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	PresetModelllama2AWorldSize = 1
	PresetModelllama2BWorldSize = 4
	PresetModelllama2CWorldSize = 8
)

var torchRunParams = map[string]string{}

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
		if err := c.updateStatusConditionIfNotMatch(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeReady, metav1.ConditionFalse,
			"workspaceFailed", err.Error()); err != nil {
			klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
			return reconcile.Result{}, err
		}
		// if error is	due to machine instance types unavailability, stop reconcile.
		if err.Error() == machine.ErrorInstanceTypesUnavailable {
			return reconcile.Result{Requeue: false}, err
		}
		return reconcile.Result{}, err
	}

	if wObj.GetAnnotations() != nil {
		if err := c.applyAnnotations(ctx, wObj); err != nil {
			if err := c.updateStatusConditionIfNotMatch(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeReady, metav1.ConditionFalse,
				"workspaceFailed", err.Error()); err != nil {
				klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, err
		}
	}

	if err = c.applyInference(ctx, wObj); err != nil {
		if err := c.updateStatusConditionIfNotMatch(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeReady, metav1.ConditionFalse,
			"workspaceFailed", err.Error()); err != nil {
			klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, err
	}

	// TODO apply TrainingSpec
	if err = c.updateStatusConditionIfNotMatch(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeReady, metav1.ConditionTrue,
		"workspaceReady", "workspace is ready"); err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (c *WorkspaceReconciler) deleteWorkspace(ctx context.Context, wObj *kdmv1alpha1.Workspace) (reconcile.Result, error) {
	klog.InfoS("deleteWorkspace", "workspace", klog.KObj(wObj))
	// TODO delete workspace, machine(s), training and inference (deployment, service) obj ( ok to delete machines? which will delete nodes??)
	err := c.updateStatusConditionIfNotMatch(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeDeleting, metav1.ConditionTrue, "workspaceDeleted", "workspace is being deleted")
	if err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
		return reconcile.Result{}, err
	}

	return c.garbageCollectWorkspace(ctx, wObj)
}

// applyWorkspaceResource applies workspace resource spec.
func (c *WorkspaceReconciler) applyWorkspaceResource(ctx context.Context, wObj *kdmv1alpha1.Workspace) error {
	klog.InfoS("applyWorkspaceResource", "workspace", klog.KObj(wObj))
	validNodeList := []*corev1.Node{}

	machinesProvisioningCount, err := machine.CheckOngoingProvisioningMachines(ctx, wObj, c.Client)
	if err != nil {
		return err
	}

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
	remainingNodeCount := lo.FromPtr(wObj.Resource.Count) - validNodeCount - machinesProvisioningCount

	// if current valid nodes Count == workspace count, then all good and return
	if remainingNodeCount == 0 {
		klog.InfoS("number of existing nodes are equal to the required workspace count", "workspace.Count", lo.FromPtr(wObj.Resource.Count))
	} else {
		klog.InfoS("need to create more nodes", "NodeCount", remainingNodeCount)
		if err := c.updateStatusConditionIfNotMatch(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeMachineStatus, metav1.ConditionUnknown,
			"CreateMachinePending", fmt.Sprintf("creating %d machines", remainingNodeCount)); err != nil {
			klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
			return err
		}

		for i := 0; i < remainingNodeCount; i++ {
			newNode, err := c.createAndValidateNode(ctx, wObj)
			if err != nil {
				if err := c.updateStatusConditionIfNotMatch(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeResourceStatus, metav1.ConditionFalse,
					"workspaceResourceStatusFailed", err.Error()); err != nil {
					klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
					return err
				}
				return err
			}
			validNodeList = append(validNodeList, newNode)
		}
	}

	// Ensure all nodes plugins are running successfully
	for i := range validNodeList {
		err = c.ensureNodePlugins(ctx, wObj, validNodeList[i])
		if err != nil {
			if err := c.updateStatusConditionIfNotMatch(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeResourceStatus, metav1.ConditionFalse,
				"workspaceResourceStatusFailed", err.Error()); err != nil {
				klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
				return err
			}
			return err
		}
	}

	if err = c.updateStatusConditionIfNotMatch(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeMachineStatus, metav1.ConditionTrue,
		"installNodePluginsSuccess", "machines plugins have been installed successfully"); err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
		return err
	}

	// Add the valid nodes names to the WorkspaceStatus.WorkerNodes
	err = c.updateStatusNodeListIfNotMatch(ctx, wObj, validNodeList)
	if err != nil {
		if err := c.updateStatusConditionIfNotMatch(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeResourceStatus, metav1.ConditionFalse,
			"workspaceResourceStatusFailed", err.Error()); err != nil {
			klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
			return err
		}
		return err
	}

	if err = c.updateStatusConditionIfNotMatch(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeResourceStatus, metav1.ConditionTrue,
		"workspaceResourceStatusSuccess", "workspace resource is ready"); err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
		return err
	}

	return nil
}

// validateCurrentClusterNodes checks if the current cluster nodes match the labelSelector and instanceType.
func (c *WorkspaceReconciler) validateCurrentClusterNodes(ctx context.Context, wObj *kdmv1alpha1.Workspace) ([]*corev1.Node, error) {
	klog.InfoS("validateCurrentClusterNodes", "workspace", klog.KObj(wObj))
	var validCurrentNodeList []*corev1.Node

	nodeList, err := k8sresources.ListNodes(ctx, c.Client, wObj.Resource.LabelSelector.MatchLabels)
	if err != nil {
		return nil, err
	}
	if len(nodeList.Items) == 0 {
		klog.InfoS("no current nodes match the workspace resource spec", "workspace", wObj.Name)
		return nil, nil
	}

	for index := range nodeList.Items {
		nodeObj := nodeList.Items[index]
		foundInstanceType := c.validateNodeInstanceType(ctx, wObj, lo.ToPtr(nodeObj))
		_, statusRunning := lo.Find(nodeObj.Status.Conditions, func(condition corev1.NodeCondition) bool {
			return condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue
		})

		if foundInstanceType && statusRunning {
			klog.InfoS("found a current valid node", "name", nodeObj.Name)
			validCurrentNodeList = append(validCurrentNodeList, lo.ToPtr(nodeObj))
		}
	}

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

// createAndValidateNode creates a new machine and validates status.
func (c *WorkspaceReconciler) createAndValidateNode(ctx context.Context, wObj *kdmv1alpha1.Workspace) (*corev1.Node, error) {
	klog.InfoS("createAndValidateNode", "workspace", klog.KObj(wObj))
	newMachine := machine.GenerateMachineManifest(ctx, wObj)

	if err := machine.CreateMachine(ctx, newMachine, c.Client); err != nil {
		klog.ErrorS(err, "failed to create machine", "machine", newMachine.Name)
		if err := c.updateStatusConditionIfNotMatch(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeMachineStatus, metav1.ConditionFalse,
			"machineFailedCreation", err.Error()); err != nil {
			klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
			return nil, err
		}
		return nil, err
	}
	klog.InfoS("a new machine has been created", "machine", newMachine.Name)

	// check machine status until it is ready
	err := machine.CheckMachineStatus(ctx, newMachine, c.Client)
	if err != nil {
		if err := c.updateStatusConditionIfNotMatch(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeMachineStatus, metav1.ConditionFalse,
			"checkMachineStatusFailed", err.Error()); err != nil {
			klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
			return nil, err
		}
		return nil, err
	}

	// get the node object from the machine status nodeName.
	return k8sresources.GetNode(ctx, newMachine.Status.NodeName, c.Client)
}

// ensureNodePlugins ensures node plugins (Nvidia and DADI) are installed.
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
						if err := c.updateStatusConditionIfNotMatch(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeMachineStatus, metav1.ConditionFalse,
							"checkMachineStatusFailed", err.Error()); err != nil {
							klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
							return err
						}
						return err
					}
					if err := c.updateStatusConditionIfNotMatch(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeMachineStatus, metav1.ConditionUnknown, "InstallNodePluginsWaiting",
						fmt.Sprintf("waiting for nvidia plugins to get installed on node %s", nodeObj.Name)); err != nil {
						return err
					}
					time.Sleep(1 * time.Second)
					continue
				}
			}

			//DADI plugin
			if err := k8sresources.CheckDADIPlugin(ctx, nodeObj, c.Client); err != nil {
				if err := k8sresources.UpdateNodeWithLabel(ctx, nodeObj.Name, k8sresources.LabelKeyCustomGPUProvisioner,
					k8sresources.GPUString, c.Client); err != nil {
					if errors.IsNotFound(err) {
						klog.ErrorS(err, "DADI plugin cannot be installed, node not found", "node", nodeObj.Name)
						if err := c.updateStatusConditionIfNotMatch(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeMachineStatus, metav1.ConditionFalse,
							"checkMachineStatusFailed", err.Error()); err != nil {
							klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
							return err
						}
						return err
					}
					if err := c.updateStatusConditionIfNotMatch(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeMachineStatus, metav1.ConditionUnknown, "InstallNodePluginsWaiting",
						fmt.Sprintf("waiting for dadi plugins to get installed on node %s", nodeObj.Name)); err != nil {
						return err
					}
					time.Sleep(1 * time.Second)
					continue
				}
			}

			return nil
		}
	}
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

	existingSVC := &corev1.Service{}
	err := k8sresources.GetResource(ctx, wObj.Name, wObj.Namespace, c.Client, existingSVC)
	// TODO: Add a check to ensure service found is of expected spec
	if err != nil {
		if !errors.IsNotFound(err) {
			klog.InfoS("a service already exists for workspace", "workspace", klog.KObj(wObj), "serviceType", serviceType)
			return nil
		}
	}
	serviceObj := k8sresources.GenerateServiceManifest(wObj)
	err = k8sresources.CreateResource(ctx, serviceObj, c.Client)
	if err != nil {
		return err
	}

	klog.InfoS("a service has been created for workspace", "workspace", klog.KObj(wObj), "serviceType", serviceType)
	return nil
}

func (c *WorkspaceReconciler) setTorchParams(ctx context.Context, wObj *kdmv1alpha1.Workspace, presetName kdmv1alpha1.PresetModelName) error {
	existingService := &corev1.Service{}
	err := k8sresources.GetResource(ctx, wObj.Name, wObj.Namespace, c.Client, existingService)
	if err != nil {
		return err
	}

	worldSize := 1
	switch presetName {
	case kdmv1alpha1.PresetSetModelllama2A:
		worldSize = PresetModelllama2AWorldSize
	case kdmv1alpha1.PresetSetModelllama2B:
		worldSize = PresetModelllama2BWorldSize
	case kdmv1alpha1.PresetSetModelllama2C:
		worldSize = PresetModelllama2CWorldSize
	default:
		err = fmt.Errorf("preset model %s is not supported", presetName)
		klog.ErrorS(err, "no inference has been created")
	}

	nodes := *wObj.Resource.Count
	torchRunParams["nnodes"] = strconv.Itoa(nodes)
	torchRunParams["nproc_per_node"] = strconv.Itoa(worldSize / nodes)
	if nodes > 1 {
		torchRunParams["node_rank"] = "$(echo $HOSTNAME | grep -o '[^-]*$')"
		torchRunParams["master_addr"] = existingService.Spec.ClusterIP
		torchRunParams["master_port"] = "29500"
	}
	return nil
}

// applyInference applies inference spec.
func (c *WorkspaceReconciler) applyInference(ctx context.Context, wObj *kdmv1alpha1.Workspace) error {
	klog.InfoS("applyInference", "workspace", klog.KObj(wObj))

	presetName := wObj.Inference.Preset.Name
	existingObj := &appsv1.StatefulSet{}
	err := k8sresources.GetResource(ctx, wObj.Name, wObj.Namespace, c.Client, existingObj)
	// TODO: Add a check to ensure statefulset found is of expected spec
	if err != nil {
		if !errors.IsNotFound(err) {
			klog.InfoS("a clientObject already exists for workspace", "workspace", klog.KObj(wObj))
			return nil
		}
		if err := c.updateStatusConditionIfNotMatch(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeInferenceStatus, metav1.ConditionFalse,
			"WorkspaceInferenceStatusFailed", err.Error()); err != nil {
			klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
			return err
		}
		return err
	}

	if err := c.setTorchParams(ctx, wObj, presetName); err != nil {
		klog.ErrorS(err, "failed to update torch params", "workspace", wObj)
		return err
	}

	switch presetName {
	case kdmv1alpha1.PresetSetModelllama2A:
		err = inference.CreateLLAMA2APresetModel(ctx, wObj, torchRunParams, c.Client)
	case kdmv1alpha1.PresetSetModelllama2B:
		err = inference.CreateLLAMA2BPresetModel(ctx, wObj, torchRunParams, c.Client)
	case kdmv1alpha1.PresetSetModelllama2C:
		err = inference.CreateLLAMA2CPresetModel(ctx, wObj, torchRunParams, c.Client)
	default:
		err = fmt.Errorf("preset model %s is not supported", presetName)
		klog.ErrorS(err, "no inference has been created")
	}
	if err != nil {
		if err := c.updateStatusConditionIfNotMatch(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeInferenceStatus, metav1.ConditionFalse,
			"WorkspaceInferenceStatusFailed", err.Error()); err != nil {
			klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
			return err
		}
		return err
	}

	if err := c.updateStatusConditionIfNotMatch(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeInferenceStatus, metav1.ConditionTrue,
		"WorkspaceInferenceStatusSuccess", "Inference has been deployed successfully"); err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
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
		Watches(
			&v1alpha5.Machine{}, c.watchMachines()).
		WithOptions(controller.Options{MaxConcurrentReconciles: 5}).
		Complete(c)
}

// watches	for machine with label LabelCreatedByWorkspace equals workspace name.
func (c *WorkspaceReconciler) watchMachines() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(
		func(ctx context.Context, o client.Object) []reconcile.Request {
			machineObj := o.(*v1alpha5.Machine)
			return []reconcile.Request{
				{
					NamespacedName: client.ObjectKey{
						Name:      machineObj.Name,
						Namespace: machineObj.Namespace,
					},
				},
			}
		})
}

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package controllers

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/utils/clock"

	"github.com/aws/karpenter-core/pkg/apis/v1alpha5"
	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/pkg/inference"
	"github.com/azure/kaito/pkg/machine"
	"github.com/azure/kaito/pkg/model"
	"github.com/azure/kaito/pkg/resources"
	"github.com/azure/kaito/pkg/utils"
	"github.com/azure/kaito/pkg/utils/plugin"
	"github.com/go-logr/logr"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	gpuSkuPrefix             = "Standard_N"
	nodePluginInstallTimeout = 60 * time.Second
)

type WorkspaceReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

func (c *WorkspaceReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	workspaceObj := &kaitov1alpha1.Workspace{}
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
	} else {
		// Ensure finalizer
		if !controllerutil.ContainsFinalizer(workspaceObj, utils.WorkspaceFinalizer) {
			controllerutil.AddFinalizer(workspaceObj, utils.WorkspaceFinalizer)
			updateCopy := workspaceObj.DeepCopy()
			if updateErr := c.Update(ctx, updateCopy, &client.UpdateOptions{}); updateErr != nil {
				klog.ErrorS(updateErr, "failed to ensure the finalizer to the workspace",
					"workspace", klog.KObj(updateCopy))
				return ctrl.Result{}, updateErr
			}
		}
	}

	if workspaceObj.Inference.Preset != nil {
		if !plugin.KaitoModelRegister.Has(string(workspaceObj.Inference.Preset.Name)) {
			return reconcile.Result{}, fmt.Errorf("The preset model name %s is not registered for workspace %s/%s", string(workspaceObj.Inference.Preset.Name), workspaceObj.Namespace, workspaceObj.Name)
		}
	}

	return c.addOrUpdateWorkspace(ctx, workspaceObj)
}

func (c *WorkspaceReconciler) addOrUpdateWorkspace(ctx context.Context, wObj *kaitov1alpha1.Workspace) (reconcile.Result, error) {
	// Read ResourceSpec
	err := c.applyWorkspaceResource(ctx, wObj)
	if err != nil {
		if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeReady, metav1.ConditionFalse,
			"workspaceFailed", err.Error()); updateErr != nil {
			klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
			return reconcile.Result{}, updateErr
		}
		// if error is	due to machine instance types unavailability, stop reconcile.
		if err.Error() == machine.ErrorInstanceTypesUnavailable {
			return reconcile.Result{Requeue: false}, err
		}
		return reconcile.Result{}, err
	}

	if err := c.ensureService(ctx, wObj); err != nil {
		if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeReady, metav1.ConditionFalse,
			"workspaceFailed", err.Error()); updateErr != nil {
			klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
			return reconcile.Result{}, updateErr
		}
		return reconcile.Result{}, err
	}

	if err = c.applyInference(ctx, wObj); err != nil {
		if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeReady, metav1.ConditionFalse,
			"workspaceFailed", err.Error()); updateErr != nil {
			klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
			return reconcile.Result{}, updateErr
		}
		return reconcile.Result{}, err
	}

	// TODO apply TrainingSpec
	if err = c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeReady, metav1.ConditionTrue,
		"workspaceReady", "workspace is ready"); err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", klog.KObj(wObj))
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (c *WorkspaceReconciler) deleteWorkspace(ctx context.Context, wObj *kaitov1alpha1.Workspace) (reconcile.Result, error) {
	klog.InfoS("deleteWorkspace", "workspace", klog.KObj(wObj))
	// TODO delete workspace, machine(s), training and inference (deployment, service) obj ( ok to delete machines? which will delete nodes??)
	err := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeDeleting, metav1.ConditionTrue, "workspaceDeleted", "workspace is being deleted")
	if err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", klog.KObj(wObj))
		return reconcile.Result{}, err
	}

	return c.garbageCollectWorkspace(ctx, wObj)
}

func selectWorkspaceNodes(qualified []*corev1.Node, preferred []string, previous []string, count int) []*corev1.Node {

	sort.Slice(qualified, func(i, j int) bool {
		iPreferred := utils.Contains(preferred, qualified[i].Name)
		jPreferred := utils.Contains(preferred, qualified[j].Name)

		if iPreferred && !jPreferred {
			return true
		} else if !iPreferred && jPreferred {
			return false
		} else { // either all are preferred, or none is preferred
			iPrevious := utils.Contains(previous, qualified[i].Name)
			jPrevious := utils.Contains(previous, qualified[j].Name)

			if iPrevious && !jPrevious {
				return true
			} else if !iPrevious && jPrevious {
				return false
			} else { // either all are previous, or none is previous
				_, iCreatedByKaito := qualified[i].Labels["kaito.sh/machine-type"]
				_, jCreatedByKaito := qualified[j].Labels["kaito.sh/machine-type"]

				// Choose node created by gpu-provisioner since it is more likely to be empty to use.
				if iCreatedByKaito && !jCreatedByKaito {
					return true
				} else if !iCreatedByKaito && jCreatedByKaito {
					return false
				} else {
					return qualified[i].Name < qualified[j].Name
				}
			}
		}
	})

	if len(qualified) <= count {
		return qualified

	}
	return qualified[0:count]
}

// applyWorkspaceResource applies workspace resource spec.
func (c *WorkspaceReconciler) applyWorkspaceResource(ctx context.Context, wObj *kaitov1alpha1.Workspace) error {

	// Wait for pending machines if any before we decide whether to create new machine or not.
	if err := machine.WaitForPendingMachines(ctx, wObj, c.Client); err != nil {
		return err
	}

	// Find all nodes that match the labelSelector and instanceType, they are not necessarily created by machines.
	validNodes, err := c.getAllQualifiedNodes(ctx, wObj)
	if err != nil {
		return err
	}

	selectedNodes := selectWorkspaceNodes(validNodes, wObj.Resource.PreferredNodes, wObj.Status.WorkerNodes, lo.FromPtr(wObj.Resource.Count))

	newNodesCount := lo.FromPtr(wObj.Resource.Count) - len(selectedNodes)

	if newNodesCount > 0 {
		klog.InfoS("need to create more nodes", "NodeCount", newNodesCount)
		if err := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeMachineStatus, metav1.ConditionUnknown,
			"CreateMachinePending", fmt.Sprintf("creating %d machines", newNodesCount)); err != nil {
			klog.ErrorS(err, "failed to update workspace status", "workspace", klog.KObj(wObj))
			return err
		}

		for i := 0; i < newNodesCount; i++ {
			newNode, err := c.createAndValidateNode(ctx, wObj)
			if err != nil {
				if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeResourceStatus, metav1.ConditionFalse,
					"workspaceResourceStatusFailed", err.Error()); updateErr != nil {
					klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
					return updateErr
				}
				return err
			}
			selectedNodes = append(selectedNodes, newNode)
		}
	}

	// Ensure all gpu plugins are running successfully.
	if strings.Contains(wObj.Resource.InstanceType, gpuSkuPrefix) { // GPU skus
		for i := range selectedNodes {
			err = c.ensureNodePlugins(ctx, wObj, selectedNodes[i])
			if err != nil {
				if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeResourceStatus, metav1.ConditionFalse,
					"workspaceResourceStatusFailed", err.Error()); updateErr != nil {
					klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
					return updateErr
				}
				return err
			}
		}
	}

	if err = c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeMachineStatus, metav1.ConditionTrue,
		"installNodePluginsSuccess", "machines plugins have been installed successfully"); err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", klog.KObj(wObj))
		return err
	}

	// Add the valid nodes names to the WorkspaceStatus.WorkerNodes.
	err = c.updateStatusNodeListIfNotMatch(ctx, wObj, selectedNodes)
	if err != nil {
		if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeResourceStatus, metav1.ConditionFalse,
			"workspaceResourceStatusFailed", err.Error()); updateErr != nil {
			klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
			return updateErr
		}
		return err
	}

	if err = c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeResourceStatus, metav1.ConditionTrue,
		"workspaceResourceStatusSuccess", "workspace resource is ready"); err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", klog.KObj(wObj))
		return err
	}

	return nil
}

// getAllQualifiedNodes returns all nodes that match the labelSelector and instanceType.
func (c *WorkspaceReconciler) getAllQualifiedNodes(ctx context.Context, wObj *kaitov1alpha1.Workspace) ([]*corev1.Node, error) {
	var qualifiedNodes []*corev1.Node

	nodeList, err := resources.ListNodes(ctx, c.Client, wObj.Resource.LabelSelector.MatchLabels)
	if err != nil {
		return nil, err
	}
	if len(nodeList.Items) == 0 {
		klog.InfoS("no current nodes match the workspace resource spec", "workspace", klog.KObj(wObj))
		return nil, nil
	}

	for index := range nodeList.Items {
		nodeObj := nodeList.Items[index]
		foundInstanceType := c.validateNodeInstanceType(ctx, wObj, lo.ToPtr(nodeObj))
		_, statusRunning := lo.Find(nodeObj.Status.Conditions, func(condition corev1.NodeCondition) bool {
			return condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue
		})

		if foundInstanceType && statusRunning {
			qualifiedNodes = append(qualifiedNodes, lo.ToPtr(nodeObj))
		}
	}

	return qualifiedNodes, nil
}

// check if node has the required instanceType
func (c *WorkspaceReconciler) validateNodeInstanceType(ctx context.Context, wObj *kaitov1alpha1.Workspace, nodeObj *corev1.Node) bool {
	if instanceTypeLabel, found := nodeObj.Labels[corev1.LabelInstanceTypeStable]; found {
		if instanceTypeLabel != wObj.Resource.InstanceType {
			return false
		}
	}
	return true
}

// createAndValidateNode creates a new machine and validates status.
func (c *WorkspaceReconciler) createAndValidateNode(ctx context.Context, wObj *kaitov1alpha1.Workspace) (*corev1.Node, error) {
	var machineOSDiskSize string
	if wObj.Inference.Preset != nil && wObj.Inference.Preset.Name != "" {
		presetName := string(wObj.Inference.Preset.Name)
		machineOSDiskSize = plugin.KaitoModelRegister.MustGet(presetName).GetInferenceParameters().DiskStorageRequirement
	}
	if machineOSDiskSize == "" {
		machineOSDiskSize = "0" // The default OS size is used
	}

Retry_withdifferentname:
	newMachine := machine.GenerateMachineManifest(ctx, machineOSDiskSize, wObj)

	if err := machine.CreateMachine(ctx, newMachine, c.Client); err != nil {
		if apierrors.IsAlreadyExists(err) {
			klog.InfoS("There exists a machine with the same name, retry with a different name", "machine", klog.KObj(newMachine))
			goto Retry_withdifferentname
		} else {

			klog.ErrorS(err, "failed to create machine", "machine", newMachine.Name)
			if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeMachineStatus, metav1.ConditionFalse,
				"machineFailedCreation", err.Error()); updateErr != nil {
				klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
				return nil, updateErr
			}
			return nil, err
		}
	}

	// check machine status until it is ready
	err := machine.CheckMachineStatus(ctx, newMachine, c.Client)
	if err != nil {
		if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeMachineStatus, metav1.ConditionFalse,
			"checkMachineStatusFailed", err.Error()); updateErr != nil {
			klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
			return nil, updateErr
		}
		return nil, err
	}

	// get the node object from the machine status nodeName.
	return resources.GetNode(ctx, newMachine.Status.NodeName, c.Client)
}

// ensureNodePlugins ensures node plugins are installed.
func (c *WorkspaceReconciler) ensureNodePlugins(ctx context.Context, wObj *kaitov1alpha1.Workspace, nodeObj *corev1.Node) error {
	timeClock := clock.RealClock{}
	tick := timeClock.NewTicker(nodePluginInstallTimeout)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C():
			return fmt.Errorf("node plugin installation timed out. node %s is not ready", nodeObj.Name)
		default:
			//Nvidia Plugin
			if found := resources.CheckNvidiaPlugin(ctx, nodeObj); !found {
				if err := resources.UpdateNodeWithLabel(ctx, nodeObj.Name, resources.LabelKeyNvidia, resources.LabelValueNvidia, c.Client); err != nil {
					if apierrors.IsNotFound(err) {
						klog.ErrorS(err, "nvidia plugin cannot be installed, node not found", "node", nodeObj.Name)
						if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeMachineStatus, metav1.ConditionFalse,
							"checkMachineStatusFailed", err.Error()); updateErr != nil {
							klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
							return updateErr
						}
						return err
					}
				}
				time.Sleep(1 * time.Second)
			}
			return nil
		}
	}
}

func (c *WorkspaceReconciler) ensureService(ctx context.Context, wObj *kaitov1alpha1.Workspace) error {
	serviceType := corev1.ServiceTypeClusterIP
	wAnnotation := wObj.GetAnnotations()

	if len(wAnnotation) != 0 {
		val, found := wAnnotation[kaitov1alpha1.AnnotationEnableLB]
		if found && val == "True" {
			serviceType = corev1.ServiceTypeLoadBalancer
		}
	}

	existingSVC := &corev1.Service{}
	err := resources.GetResource(ctx, wObj.Name, wObj.Namespace, c.Client, existingSVC)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
	} else {
		return nil
	}

	if wObj.Inference.Preset != nil {
		presetName := string(wObj.Inference.Preset.Name)
		model := plugin.KaitoModelRegister.MustGet(presetName)
		serviceObj := resources.GenerateServiceManifest(ctx, wObj, serviceType, model.SupportDistributedInference())
		err = resources.CreateResource(ctx, serviceObj, c.Client)
		if err != nil {
			return err
		}
		if model.SupportDistributedInference() {
			headlessService := resources.GenerateHeadlessServiceManifest(ctx, wObj)
			err = resources.CreateResource(ctx, headlessService, c.Client)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *WorkspaceReconciler) updateInferenceParamFromWorkspace(ctx context.Context, wObj *kaitov1alpha1.Workspace, inferenceParam *model.PresetInferenceParam) {
	inferenceParam.ImageAccessMode = string(wObj.Inference.Preset.PresetMeta.ImageAccessMode)
	if inferenceParam.ImageAccessMode == "private" && wObj.Inference.Preset.PresetOptions.Image != "" {
		inferenceParam.Image = wObj.Inference.Preset.PresetOptions.Image

		imagePullSecretRefs := []corev1.LocalObjectReference{}
		for _, secretName := range wObj.Inference.Preset.PresetOptions.ImagePullSecrets {
			imagePullSecretRefs = append(imagePullSecretRefs, corev1.LocalObjectReference{Name: secretName})
		}
		inferenceParam.ImagePullSecrets = imagePullSecretRefs
	}
}

// applyInference applies inference spec.
func (c *WorkspaceReconciler) applyInference(ctx context.Context, wObj *kaitov1alpha1.Workspace) error {
	var err error
	func() {
		if wObj.Inference.Template != nil {
			var workloadObj client.Object
			// TODO: handle update
			workloadObj, err = inference.CreateTemplateInference(ctx, wObj, c.Client)
			if err != nil {
				return
			}
			if err = resources.CheckResourceStatus(workloadObj, c.Client, time.Duration(10)*time.Minute); err != nil {
				return
			}
		} else if wObj.Inference.Preset != nil {
			presetName := string(wObj.Inference.Preset.Name)
			model := plugin.KaitoModelRegister.MustGet(presetName)

			inferenceParam := model.GetInferenceParameters()

			c.updateInferenceParamFromWorkspace(ctx, wObj, inferenceParam)
			// TODO: we only do create if it does not exist for preset model. Need to document it.

			var existingObj client.Object
			if model.SupportDistributedInference() {
				existingObj = &appsv1.StatefulSet{}
			} else {
				existingObj = &appsv1.Deployment{}

			}

			if err = resources.GetResource(ctx, wObj.Name, wObj.Namespace, c.Client, existingObj); err == nil {
				klog.InfoS("An inference workload already exists for workspace", "workspace", klog.KObj(wObj))
				if err = resources.CheckResourceStatus(existingObj, c.Client, inferenceParam.DeploymentTimeout); err != nil {
					return
				}
			} else if apierrors.IsNotFound(err) {
				var workloadObj client.Object
				// Need to create a new workload
				workloadObj, err = inference.CreatePresetInference(ctx, wObj, inferenceParam, model.SupportDistributedInference(), c.Client)
				if err != nil {
					return
				}
				if err = resources.CheckResourceStatus(workloadObj, c.Client, inferenceParam.DeploymentTimeout); err != nil {
					return
				}
			}
		}
	}()

	if err != nil {
		if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeInferenceStatus, metav1.ConditionFalse,
			"WorkspaceInferenceStatusFailed", err.Error()); updateErr != nil {
			klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
			return updateErr
		} else {
			return err

		}
	}

	if err := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeInferenceStatus, metav1.ConditionTrue,
		"WorkspaceInferenceStatusSuccess", "Inference has been deployed successfully"); err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", klog.KObj(wObj))
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
		For(&kaitov1alpha1.Workspace{}).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.StatefulSet{}).
		Watches(&v1alpha5.Machine{}, c.watchMachines()).
		WithOptions(controller.Options{MaxConcurrentReconciles: 5}).
		Complete(c)
}

// watches for machine with labels indicating workspace name.
func (c *WorkspaceReconciler) watchMachines() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(
		func(ctx context.Context, o client.Object) []reconcile.Request {
			machineObj := o.(*v1alpha5.Machine)
			name, ok := machineObj.Labels[kaitov1alpha1.LabelWorkspaceName]
			if !ok {
				return nil
			}
			namespace, ok := machineObj.Labels[kaitov1alpha1.LabelWorkspaceNamespace]
			if !ok {
				return nil
			}
			return []reconcile.Request{
				{
					NamespacedName: client.ObjectKey{
						Name:      name,
						Namespace: namespace,
					},
				},
			}
		})
}

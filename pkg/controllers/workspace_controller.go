// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/kaito-project/kaito/pkg/featuregates"
	"github.com/kaito-project/kaito/pkg/nodeclaim"
	"github.com/kaito-project/kaito/pkg/tuning"
	"github.com/kaito-project/kaito/pkg/utils/consts"
	batchv1 "k8s.io/api/batch/v1"
	"sigs.k8s.io/karpenter/pkg/apis/v1beta1"

	"github.com/aws/karpenter-core/pkg/apis/v1alpha5"
	"github.com/go-logr/logr"
	kaitov1alpha1 "github.com/kaito-project/kaito/api/v1alpha1"
	"github.com/kaito-project/kaito/pkg/inference"
	"github.com/kaito-project/kaito/pkg/machine"
	"github.com/kaito-project/kaito/pkg/resources"
	"github.com/kaito-project/kaito/pkg/utils"
	"github.com/kaito-project/kaito/pkg/utils/plugin"
	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
	"knative.dev/pkg/apis"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	nodePluginInstallTimeout = 60 * time.Second
	WorkspaceHashAnnotation  = "workspace.kaito.io/hash"
	WorkspaceNameLabel       = "workspace.kaito.io/name"
	revisionHashSuffix       = 5
)

type WorkspaceReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

func NewWorkspaceReconciler(client client.Client, scheme *runtime.Scheme, log logr.Logger, Recorder record.EventRecorder) *WorkspaceReconciler {
	return &WorkspaceReconciler{
		Client:   client,
		Scheme:   scheme,
		Log:      log,
		Recorder: Recorder,
	}
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

	if err := c.ensureFinalizer(ctx, workspaceObj); err != nil {
		return reconcile.Result{}, err
	}
	// Handle deleting workspace, garbage collect all the resources.
	if !workspaceObj.DeletionTimestamp.IsZero() {
		return c.deleteWorkspace(ctx, workspaceObj)
	}

	if err := c.syncControllerRevision(ctx, workspaceObj); err != nil {
		return reconcile.Result{}, err
	}

	if workspaceObj.Inference != nil && workspaceObj.Inference.Preset != nil {
		if !plugin.KaitoModelRegister.Has(string(workspaceObj.Inference.Preset.Name)) {
			return reconcile.Result{}, fmt.Errorf("the preset model name %s is not registered for workspace %s/%s",
				string(workspaceObj.Inference.Preset.Name), workspaceObj.Namespace, workspaceObj.Name)
		}
	}

	result, err := c.addOrUpdateWorkspace(ctx, workspaceObj)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (c *WorkspaceReconciler) ensureFinalizer(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace) error {
	if !controllerutil.ContainsFinalizer(workspaceObj, consts.WorkspaceFinalizer) {
		patch := client.MergeFrom(workspaceObj.DeepCopy())
		controllerutil.AddFinalizer(workspaceObj, consts.WorkspaceFinalizer)
		if err := c.Client.Patch(ctx, workspaceObj, patch); err != nil {
			klog.ErrorS(err, "failed to ensure the finalizer to the workspace", "workspace", klog.KObj(workspaceObj))
			return err
		}
	}
	return nil
}

func (c *WorkspaceReconciler) addOrUpdateWorkspace(ctx context.Context, wObj *kaitov1alpha1.Workspace) (reconcile.Result, error) {
	// Read ResourceSpec
	err := c.applyWorkspaceResource(ctx, wObj)
	if err != nil {
		if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeSucceeded, metav1.ConditionFalse,
			"workspaceFailed", err.Error()); updateErr != nil {
			klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
			return reconcile.Result{}, updateErr
		}
		// If the error is due to machine/nodeClaim instance types unavailability, stop reconcile.
		if err.Error() == consts.ErrorInstanceTypesUnavailable {
			return reconcile.Result{Requeue: false}, err
		}
		return reconcile.Result{}, err
	}

	if wObj.Tuning != nil {
		if err = c.applyTuning(ctx, wObj); err != nil {
			if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeSucceeded, metav1.ConditionFalse,
				"workspaceFailed", err.Error()); updateErr != nil {
				klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
				return reconcile.Result{}, updateErr
			}
			return reconcile.Result{}, err
		}
		// Only mark workspace succeeded when job completes.
		job := &batchv1.Job{}
		if err = resources.GetResource(ctx, wObj.Name, wObj.Namespace, c.Client, job); err == nil {
			if job.Status.Succeeded > 0 {
				if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeSucceeded, metav1.ConditionTrue,
					"workspaceSucceeded", "workspace succeeds"); updateErr != nil {
					klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
					return reconcile.Result{}, updateErr
				}
			} else { // The job is still running
				var readyPod int32
				if job.Status.Ready != nil {
					readyPod = *job.Status.Ready
				}
				if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeSucceeded, metav1.ConditionFalse,
					"workspacePending", fmt.Sprintf("workspace has not completed, tuning job has %d active pod, %d ready pod", job.Status.Active, readyPod)); updateErr != nil {
					klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
					return reconcile.Result{}, updateErr
				}
			}
		} else {
			klog.ErrorS(err, "failed to get job resource", "workspace", klog.KObj(wObj))
			return reconcile.Result{}, err
		}
	} else if wObj.Inference != nil {
		if err := c.ensureService(ctx, wObj); err != nil {
			if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeSucceeded, metav1.ConditionFalse,
				"workspaceFailed", err.Error()); updateErr != nil {
				klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
				return reconcile.Result{}, updateErr
			}
			return reconcile.Result{}, err
		}
		if err = c.applyInference(ctx, wObj); err != nil {
			if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeSucceeded, metav1.ConditionFalse,
				"workspaceFailed", err.Error()); updateErr != nil {
				klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
				return reconcile.Result{}, updateErr
			}
			return reconcile.Result{}, err
		}

		if err = c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeSucceeded, metav1.ConditionTrue,
			"workspaceSucceeded", "workspace succeeds"); err != nil {
			klog.ErrorS(err, "failed to update workspace status", "workspace", klog.KObj(wObj))
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (c *WorkspaceReconciler) deleteWorkspace(ctx context.Context, wObj *kaitov1alpha1.Workspace) (reconcile.Result, error) {
	klog.InfoS("deleteWorkspace", "workspace", klog.KObj(wObj))
	err := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeDeleting, metav1.ConditionTrue, "workspaceDeleted", "workspace is being deleted")
	if err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", klog.KObj(wObj))
		return reconcile.Result{}, err
	}

	return c.garbageCollectWorkspace(ctx, wObj)
}
func (c *WorkspaceReconciler) syncControllerRevision(ctx context.Context, wObj *kaitov1alpha1.Workspace) error {
	currentHash := computeHash(wObj)
	annotations := wObj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	} // nil checking.

	revisionNum := int64(1)

	revisions := &appsv1.ControllerRevisionList{}
	if err := c.List(ctx, revisions, client.InNamespace(wObj.Namespace), client.MatchingLabels{WorkspaceNameLabel: wObj.Name}); err != nil {
		return fmt.Errorf("failed to list revisions: %w", err)
	}
	sort.Slice(revisions.Items, func(i, j int) bool {
		return revisions.Items[i].Revision < revisions.Items[j].Revision
	})

	var latestRevision *appsv1.ControllerRevision

	jsonData, err := marshalSelectedFields(wObj)
	if err != nil {
		return fmt.Errorf("failed to marshal revision data: %w", err)
	}

	if len(revisions.Items) > 0 {
		latestRevision = &revisions.Items[len(revisions.Items)-1]

		revisionNum = latestRevision.Revision + 1
	}
	newRevision := &appsv1.ControllerRevision{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", wObj.Name, currentHash[:revisionHashSuffix]),
			Namespace: wObj.Namespace,
			Annotations: map[string]string{
				WorkspaceHashAnnotation: currentHash,
			},
			Labels: map[string]string{
				WorkspaceNameLabel: wObj.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(wObj, kaitov1alpha1.GroupVersion.WithKind("Workspace")),
			},
		},
		Revision: revisionNum,
		Data:     runtime.RawExtension{Raw: jsonData},
	}

	annotations[WorkspaceHashAnnotation] = currentHash
	wObj.SetAnnotations(annotations)
	controllerRevision := &appsv1.ControllerRevision{}
	if err := c.Get(ctx, types.NamespacedName{
		Name:      newRevision.Name,
		Namespace: newRevision.Namespace,
	}, controllerRevision); err != nil {
		if errors.IsNotFound(err) {

			if err := c.Create(ctx, newRevision); err != nil {
				return fmt.Errorf("failed to create new ControllerRevision: %w", err)
			} else {
				annotations[kaitov1alpha1.WorkspaceRevisionAnnotation] = strconv.FormatInt(revisionNum, 10)
			}

			if len(revisions.Items) > consts.MaxRevisionHistoryLimit {
				if err := c.Delete(ctx, &revisions.Items[0]); err != nil {
					return fmt.Errorf("failed to delete old revision: %w", err)
				}
			}
		} else {
			return fmt.Errorf("failed to get controller revision: %w", err)
		}
	} else {
		if controllerRevision.Annotations[WorkspaceHashAnnotation] != newRevision.Annotations[WorkspaceHashAnnotation] {
			return fmt.Errorf("revision name conflicts, the hash values are different")
		}
		annotations[kaitov1alpha1.WorkspaceRevisionAnnotation] = strconv.FormatInt(controllerRevision.Revision, 10)
	}
	annotations[WorkspaceHashAnnotation] = currentHash
	wObj.SetAnnotations(annotations)

	if err := c.Update(ctx, wObj); err != nil {
		return fmt.Errorf("failed to update Workspace annotations: %w", err)
	}
	return nil
}

func marshalSelectedFields(wObj *kaitov1alpha1.Workspace) ([]byte, error) {
	partialMap := map[string]interface{}{
		"resource":  wObj.Resource,
		"inference": wObj.Inference,
		"tuning":    wObj.Tuning,
	}

	jsonData, err := json.Marshal(partialMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal selected fields: %w", err)
	}

	return jsonData, nil
}

func computeHash(w *kaitov1alpha1.Workspace) string {
	hasher := sha256.New()
	encoder := json.NewEncoder(hasher)
	encoder.Encode(w.Resource)
	encoder.Encode(w.Inference)
	encoder.Encode(w.Tuning)
	return hex.EncodeToString(hasher.Sum(nil))
}

// applyWorkspaceResource applies workspace resource spec.
func (c *WorkspaceReconciler) applyWorkspaceResource(ctx context.Context, wObj *kaitov1alpha1.Workspace) error {

	// Wait for pending machines if any before we decide whether to create new machine or not.
	if err := machine.WaitForPendingMachines(ctx, wObj, c.Client); err != nil {
		return err
	}

	if featuregates.FeatureGates[consts.FeatureFlagKarpenter] {
		// Wait for pending nodeClaims if any before we decide whether to create new node or not.
		if err := nodeclaim.WaitForPendingNodeClaims(ctx, wObj, c.Client); err != nil {
			return err
		}
	}

	// Find all nodes that meet the requirements, they are not necessarily created by machines/nodeClaims.
	validNodes, err := c.getAllQualifiedNodes(ctx, wObj)
	if err != nil {
		return err
	}

	selectedNodes := utils.SelectNodes(validNodes, wObj.Resource.PreferredNodes, wObj.Status.WorkerNodes, lo.FromPtr(wObj.Resource.Count))

	newNodesCount := lo.FromPtr(wObj.Resource.Count) - len(selectedNodes)

	if newNodesCount > 0 {
		klog.InfoS("need to create more nodes", "NodeCount", newNodesCount)
		if featuregates.FeatureGates[consts.FeatureFlagKarpenter] {
			if err := c.updateStatusConditionIfNotMatch(ctx, wObj,
				kaitov1alpha1.ConditionTypeNodeClaimStatus, metav1.ConditionUnknown,
				"CreateNodeClaimPending", fmt.Sprintf("creating %d nodeClaims", newNodesCount)); err != nil {
				klog.ErrorS(err, "failed to update workspace status", "workspace", klog.KObj(wObj))
				return err
			}
		} else if err := c.updateStatusConditionIfNotMatch(ctx, wObj,
			kaitov1alpha1.ConditionTypeMachineStatus, metav1.ConditionUnknown,
			"CreateMachinePending", fmt.Sprintf("creating %d machines", newNodesCount)); err != nil {
			klog.ErrorS(err, "failed to update workspace status", "workspace", klog.KObj(wObj))
			return err
		}

		for i := 0; i < newNodesCount; i++ {
			newNode, err := c.createAndValidateNode(ctx, wObj)
			if err != nil {
				if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.ConditionTypeResourceStatus, metav1.ConditionFalse,
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
	if strings.Contains(wObj.Resource.InstanceType, consts.GpuSkuPrefix) { // GPU skus
		for i := range selectedNodes {
			err = c.ensureNodePlugins(ctx, wObj, selectedNodes[i])
			if err != nil {
				if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.ConditionTypeResourceStatus, metav1.ConditionFalse,
					"workspaceResourceStatusFailed", err.Error()); updateErr != nil {
					klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
					return updateErr
				}
				return err
			}
		}
	}

	if featuregates.FeatureGates[consts.FeatureFlagKarpenter] {
		if err = c.updateStatusConditionIfNotMatch(ctx, wObj,
			kaitov1alpha1.ConditionTypeNodeClaimStatus, metav1.ConditionTrue,
			"installNodePluginsSuccess", "nodeClaim plugins have been installed successfully"); err != nil {
			klog.ErrorS(err, "failed to update workspace status", "workspace", klog.KObj(wObj))
			return err
		}
	} else if err = c.updateStatusConditionIfNotMatch(ctx, wObj,
		kaitov1alpha1.ConditionTypeMachineStatus, metav1.ConditionTrue,
		"installNodePluginsSuccess", "machines plugins have been installed successfully"); err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", klog.KObj(wObj))
		return err
	}

	// Add the valid nodes names to the WorkspaceStatus.WorkerNodes.
	err = c.updateStatusNodeListIfNotMatch(ctx, wObj, selectedNodes)
	if err != nil {
		if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.ConditionTypeResourceStatus, metav1.ConditionFalse,
			"workspaceResourceStatusFailed", err.Error()); updateErr != nil {
			klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
			return updateErr
		}
		return err
	}

	if err = c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.ConditionTypeResourceStatus, metav1.ConditionTrue,
		"workspaceResourceStatusSuccess", "workspace resource is ready"); err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", klog.KObj(wObj))
		return err
	}

	return nil
}

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

	preferredNodeSet := sets.New(wObj.Resource.PreferredNodes...)

	for index := range nodeList.Items {
		nodeObj := nodeList.Items[index]
		// skip nodes that are being deleted
		if nodeObj.DeletionTimestamp != nil {
			continue
		}

		// skip nodes that are not ready
		_, statusRunning := lo.Find(nodeObj.Status.Conditions, func(condition corev1.NodeCondition) bool {
			return condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue
		})
		if !statusRunning {
			continue
		}

		// match the preferred node
		if preferredNodeSet.Has(nodeObj.Name) {
			qualifiedNodes = append(qualifiedNodes, lo.ToPtr(nodeObj))
			continue
		}

		// match the instanceType
		if nodeObj.Labels[corev1.LabelInstanceTypeStable] == wObj.Resource.InstanceType {
			qualifiedNodes = append(qualifiedNodes, lo.ToPtr(nodeObj))
		}
	}

	return qualifiedNodes, nil
}

// createAndValidateNode creates a new node and validates status.
func (c *WorkspaceReconciler) createAndValidateNode(ctx context.Context, wObj *kaitov1alpha1.Workspace) (*corev1.Node, error) {
	var nodeOSDiskSize string
	if wObj.Inference != nil && wObj.Inference.Preset != nil && wObj.Inference.Preset.Name != "" {
		presetName := string(wObj.Inference.Preset.Name)
		nodeOSDiskSize = plugin.KaitoModelRegister.MustGet(presetName).
			GetInferenceParameters().DiskStorageRequirement
	}
	if nodeOSDiskSize == "" {
		nodeOSDiskSize = "0" // The default OS size is used
	}

	if featuregates.FeatureGates[consts.FeatureFlagKarpenter] {
		return c.CreateNodeClaim(ctx, wObj, nodeOSDiskSize)
	} else {
		return c.CreateMachine(ctx, wObj, nodeOSDiskSize)
	}
}

func (c *WorkspaceReconciler) CreateMachine(ctx context.Context, wObj *kaitov1alpha1.Workspace, nodeOSDiskSize string) (*corev1.Node, error) {
RetryWithDifferentName:
	newMachine := machine.GenerateMachineManifest(ctx, nodeOSDiskSize, wObj)

	if err := machine.CreateMachine(ctx, newMachine, c.Client); err != nil {
		if apierrors.IsAlreadyExists(err) {
			klog.InfoS("A machine exists with the same name, retry with a different name", "machine", klog.KObj(newMachine))
			goto RetryWithDifferentName
		} else {

			klog.ErrorS(err, "failed to create machine", "machine", newMachine.Name)
			if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.ConditionTypeMachineStatus, metav1.ConditionFalse,
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
		if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.ConditionTypeMachineStatus, metav1.ConditionFalse,
			"checkMachineStatusFailed", err.Error()); updateErr != nil {
			klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
			return nil, updateErr
		}
		return nil, err
	}

	// get the node object from the machine status nodeName.
	return resources.GetNode(ctx, newMachine.Status.NodeName, c.Client)
}

func (c *WorkspaceReconciler) CreateNodeClaim(ctx context.Context, wObj *kaitov1alpha1.Workspace, nodeOSDiskSize string) (*corev1.Node, error) {
RetryWithDifferentName:
	newNodeClaim := nodeclaim.GenerateNodeClaimManifest(ctx, nodeOSDiskSize, wObj)

	if err := nodeclaim.CreateNodeClaim(ctx, newNodeClaim, c.Client); err != nil {
		if apierrors.IsAlreadyExists(err) {
			klog.InfoS("There exists a nodeClaim with the same name, retry with a different name", "nodeClaim", klog.KObj(newNodeClaim))
			goto RetryWithDifferentName
		} else {

			klog.ErrorS(err, "failed to create nodeClaim", "nodeClaim", newNodeClaim.Name)
			if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.ConditionTypeNodeClaimStatus, metav1.ConditionFalse,
				"nodeClaimFailedCreation", err.Error()); updateErr != nil {
				klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
				return nil, updateErr
			}
			return nil, err
		}
	}

	// check nodeClaim status until it is ready
	err := nodeclaim.CheckNodeClaimStatus(ctx, newNodeClaim, c.Client)
	if err != nil {
		if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.ConditionTypeNodeClaimStatus, metav1.ConditionFalse,
			"checkNodeClaimStatusFailed", err.Error()); updateErr != nil {
			klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
			return nil, updateErr
		}
		return nil, err
	}

	// get the node object from the nodeClaim status nodeName.
	return resources.GetNode(ctx, newNodeClaim.Status.NodeName, c.Client)
}

// ensureNodePlugins ensures node plugins are installed.
func (c *WorkspaceReconciler) ensureNodePlugins(ctx context.Context, wObj *kaitov1alpha1.Workspace, nodeObj *corev1.Node) error {
	timeClock := clock.RealClock{}
	tick := timeClock.NewTicker(consts.NodePluginInstallTimeout)
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
						if featuregates.FeatureGates[consts.FeatureFlagKarpenter] {
							if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.ConditionTypeNodeClaimStatus, metav1.ConditionFalse,
								"checkNodeClaimStatusFailed", err.Error()); updateErr != nil {
								klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
								return updateErr
							}
						} else {
							if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.ConditionTypeMachineStatus, metav1.ConditionFalse,
								"checkMachineStatusFailed", err.Error()); updateErr != nil {
								klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
								return updateErr
							}
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

// getPresetName returns the preset name from wObj if available
func getPresetName(wObj *kaitov1alpha1.Workspace) string {
	if wObj.Inference != nil && wObj.Inference.Preset != nil {
		return string(wObj.Inference.Preset.Name)
	}
	if wObj.Tuning != nil && wObj.Tuning.Preset != nil {
		return string(wObj.Tuning.Preset.Name)
	}
	return ""
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

	if presetName := getPresetName(wObj); presetName != "" {
		model := plugin.KaitoModelRegister.MustGet(presetName)
		serviceObj := resources.GenerateServiceManifest(ctx, wObj, serviceType, model.SupportDistributedInference())
		if err := resources.CreateResource(ctx, serviceObj, c.Client); err != nil {
			return err
		}
		if model.SupportDistributedInference() {
			headlessService := resources.GenerateHeadlessServiceManifest(ctx, wObj)
			if err := resources.CreateResource(ctx, headlessService, c.Client); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *WorkspaceReconciler) applyTuning(ctx context.Context, wObj *kaitov1alpha1.Workspace) error {
	var err error
	func() {
		if wObj.Tuning.Preset != nil {
			presetName := string(wObj.Tuning.Preset.Name)
			model := plugin.KaitoModelRegister.MustGet(presetName)

			tuningParam := model.GetTuningParameters()
			existingObj := &batchv1.Job{}
			revisionNum := wObj.Annotations[kaitov1alpha1.WorkspaceRevisionAnnotation]
			if err = resources.GetResource(ctx, wObj.Name, wObj.Namespace, c.Client, existingObj); err == nil {
				klog.InfoS("A tuning workload already exists for workspace", "workspace", klog.KObj(wObj))

				if existingObj.Annotations[kaitov1alpha1.WorkspaceRevisionAnnotation] != revisionNum {
					deletePolicy := metav1.DeletePropagationForeground
					if err := c.Delete(ctx, existingObj, &client.DeleteOptions{
						PropagationPolicy: &deletePolicy,
					}); err != nil {
						return
					}

					var workloadObj client.Object
					workloadObj, err = tuning.CreatePresetTuning(ctx, wObj, revisionNum, tuningParam, c.Client)
					if err != nil {
						return
					}
					existingObj = workloadObj.(*batchv1.Job)
				}

				if err = resources.CheckResourceStatus(existingObj, c.Client, tuningParam.ReadinessTimeout); err != nil {
					return
				}
			} else if apierrors.IsNotFound(err) {
				var workloadObj client.Object
				// Need to create a new workload
				workloadObj, err = tuning.CreatePresetTuning(ctx, wObj, revisionNum, tuningParam, c.Client)
				if err != nil {
					return
				}
				if err = resources.CheckResourceStatus(workloadObj, c.Client, tuningParam.ReadinessTimeout); err != nil {
					return
				}
			}
		}
	}()

	if err != nil {
		if updateErr := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeTuningJobStatus, metav1.ConditionFalse,
			"WorkspaceTuningJobStatusFailed", err.Error()); updateErr != nil {
			klog.ErrorS(updateErr, "failed to update workspace status", "workspace", klog.KObj(wObj))
			return updateErr
		}
		return err
	}

	if err := c.updateStatusConditionIfNotMatch(ctx, wObj, kaitov1alpha1.WorkspaceConditionTypeTuningJobStatus, metav1.ConditionTrue,
		"WorkspaceTuningJobStatusStarted", "Tuning job has started"); err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", klog.KObj(wObj))
		return err
	}

	return nil
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
		} else if wObj.Inference != nil && wObj.Inference.Preset != nil {
			presetName := string(wObj.Inference.Preset.Name)
			model := plugin.KaitoModelRegister.MustGet(presetName)

			inferenceParam := model.GetInferenceParameters()

			var existingObj client.Object
			if model.SupportDistributedInference() {
				existingObj = &appsv1.StatefulSet{}
			} else {
				existingObj = &appsv1.Deployment{}

			}
			revisionStr := wObj.Annotations[kaitov1alpha1.WorkspaceRevisionAnnotation]
			if err = resources.GetResource(ctx, wObj.Name, wObj.Namespace, c.Client, existingObj); err == nil {
				klog.InfoS("An inference workload already exists for workspace", "workspace", klog.KObj(wObj))
				if !model.SupportDistributedInference() {
					deployment := existingObj.(*appsv1.Deployment)
					if deployment.Annotations[kaitov1alpha1.WorkspaceRevisionAnnotation] != revisionStr {
						var volumes []corev1.Volume
						var volumeMounts []corev1.VolumeMount
						shmVolume, shmVolumeMount := utils.ConfigSHMVolume(*wObj.Resource.Count)
						if shmVolume.Name != "" {
							volumes = append(volumes, shmVolume)
						}
						if shmVolumeMount.Name != "" {
							volumeMounts = append(volumeMounts, shmVolumeMount)
						}

						if len(wObj.Inference.Adapters) > 0 {
							adapterVolume, adapterVolumeMount := utils.ConfigAdapterVolume()
							volumes = append(volumes, adapterVolume)
							volumeMounts = append(volumeMounts, adapterVolumeMount)
						}
						initContainers, envs := resources.GenerateInitContainers(wObj, volumeMounts)
						spec := &deployment.Spec

						spec.Template.Spec.InitContainers = initContainers
						spec.Template.Spec.Containers[0].Env = envs
						spec.Template.Spec.Containers[0].VolumeMounts = volumeMounts
						deployment.Annotations[kaitov1alpha1.WorkspaceRevisionAnnotation] = revisionStr
						spec.Template.Spec.Volumes = volumes

						if err := c.Update(ctx, deployment); err != nil {
							return
						}
					}
				}
				if err = resources.CheckResourceStatus(existingObj, c.Client, inferenceParam.ReadinessTimeout); err != nil {
					return
				}
			} else if apierrors.IsNotFound(err) {
				var workloadObj client.Object
				// Need to create a new workload
				workloadObj, err = inference.CreatePresetInference(ctx, wObj, revisionStr, inferenceParam, model.SupportDistributedInference(), c.Client)
				if err != nil {
					return
				}
				if err = resources.CheckResourceStatus(workloadObj, c.Client, inferenceParam.ReadinessTimeout); err != nil {
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
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Pod{},
		"spec.nodeName", func(rawObj client.Object) []string {
			pod := rawObj.(*corev1.Pod)
			return []string{pod.Spec.NodeName}
		}); err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).
		For(&kaitov1alpha1.Workspace{}).
		Owns(&appsv1.ControllerRevision{}).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&batchv1.Job{}).
		Watches(&v1alpha5.Machine{}, c.watchMachines()).
		WithOptions(controller.Options{MaxConcurrentReconciles: 5})

	if featuregates.FeatureGates[consts.FeatureFlagKarpenter] {
		builder.
			Watches(&v1beta1.NodeClaim{}, c.watchNodeClaims()) // watches for nodeClaim with labels indicating workspace name.
	}
	return builder.Complete(c)
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
			_, conditionFound := lo.Find(machineObj.GetConditions(), func(condition apis.Condition) bool {
				return condition.Type == apis.ConditionReady &&
					condition.Status == v1.ConditionTrue
			})
			if conditionFound && machineObj.DeletionTimestamp.IsZero() {
				// No need to reconcile workspace if the machine is in READY state unless machine is deleted.
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

// watches for nodeClaim with labels indicating workspace name.
func (c *WorkspaceReconciler) watchNodeClaims() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(
		func(ctx context.Context, o client.Object) []reconcile.Request {
			nodeClaimObj := o.(*v1beta1.NodeClaim)
			name, ok := nodeClaimObj.Labels[kaitov1alpha1.LabelWorkspaceName]
			if !ok {
				return nil
			}
			namespace, ok := nodeClaimObj.Labels[kaitov1alpha1.LabelWorkspaceNamespace]
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

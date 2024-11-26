// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package controllers

import (
	"context"

	kaitov1alpha1 "github.com/kaito-project/kaito/api/v1alpha1"
	"github.com/kaito-project/kaito/pkg/featuregates"
	"github.com/kaito-project/kaito/pkg/utils/consts"
	"github.com/kaito-project/kaito/pkg/utils/machine"
	"github.com/kaito-project/kaito/pkg/utils/nodeclaim"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// garbageCollectWorkspace remove finalizer associated with workspace object.
func (c *WorkspaceReconciler) garbageCollectWorkspace(ctx context.Context, wObj *kaitov1alpha1.Workspace) (ctrl.Result, error) {
	klog.InfoS("garbageCollectWorkspace", "workspace", klog.KObj(wObj))

	// Check if there are any machines associated with this workspace.
	mList, err := machine.ListMachines(ctx, wObj, c.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	// We should delete all the machines that are created by this workspace
	for i := range mList.Items {
		if deleteErr := c.Delete(ctx, &mList.Items[i], &client.DeleteOptions{}); deleteErr != nil {
			klog.ErrorS(deleteErr, "failed to delete the machine", "machine", klog.KObj(&mList.Items[i]))
			return ctrl.Result{}, deleteErr
		}
	}

	if featuregates.FeatureGates[consts.FeatureFlagKarpenter] {
		// Check if there are any nodeClaims associated with this workspace.
		ncList, err := nodeclaim.ListNodeClaim(ctx, wObj, c.Client)
		if err != nil {
			return ctrl.Result{}, err
		}

		// We should delete all the nodeClaims that are created by this workspace
		for i := range ncList.Items {
			if deleteErr := c.Delete(ctx, &ncList.Items[i], &client.DeleteOptions{}); deleteErr != nil {
				klog.ErrorS(deleteErr, "failed to delete the nodeClaim", "nodeClaim", klog.KObj(&ncList.Items[i]))
				return ctrl.Result{}, deleteErr
			}
		}
	}

	if controllerutil.RemoveFinalizer(wObj, consts.WorkspaceFinalizer) {
		if updateErr := c.Update(ctx, wObj, &client.UpdateOptions{}); updateErr != nil {
			klog.ErrorS(updateErr, "failed to remove the finalizer from the workspace",
				"workspace", klog.KObj(wObj))
			return ctrl.Result{}, updateErr
		}
	}

	klog.InfoS("successfully removed the workspace finalizers",
		"workspace", klog.KObj(wObj))
	return ctrl.Result{}, nil
}

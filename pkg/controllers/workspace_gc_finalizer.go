package controllers

import (
	"context"

	kdmv1alpha1 "github.com/kdm/api/v1alpha1"
	"github.com/kdm/pkg/utils"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// garbageCollectWorkspace remove finalizer associated with workspace object.
func (c *WorkspaceReconciler) garbageCollectWorkspace(ctx context.Context, wObj *kdmv1alpha1.Workspace) (ctrl.Result, error) {
	klog.InfoS("garbageCollectWorkspace", "workspace", klog.KObj(wObj))

	staleWObj := wObj.DeepCopy()
	staleWObj.SetFinalizers(nil)
	if updateErr := c.Update(ctx, staleWObj, &client.UpdateOptions{}); updateErr != nil {
		klog.ErrorS(updateErr, "failed to remove the finalizer from the workspace",
			"workspace", klog.KObj(wObj), "workspace", klog.KObj(staleWObj))
		return ctrl.Result{}, updateErr
	}
	klog.InfoS("successfully removed the workspace finalizers",
		"workspace", klog.KObj(wObj))
	controllerutil.RemoveFinalizer(wObj, utils.WorkspaceFinalizer)
	return ctrl.Result{}, nil
}

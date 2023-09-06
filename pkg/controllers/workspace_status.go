package controllers

import (
	"context"
	"fmt"

	kdmv1alpha1 "github.com/kdm/api/v1alpha1"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

// updateWorkspaceStatus updates workspace status.
func (c *WorkspaceReconciler) updateWorkspaceStatus(ctx context.Context, wObj *kdmv1alpha1.Workspace) error {
	klog.InfoS("updateWorkspaceStatus", "workspace", klog.KObj(wObj))

	return retry.OnError(retry.DefaultRetry,
		func(err error) bool {
			return apierrors.IsServiceUnavailable(err) || apierrors.IsServerTimeout(err) || apierrors.IsTooManyRequests(err)
		},
		func() error {
			return c.Client.Status().Update(ctx, wObj)
		})
}

func (c *WorkspaceReconciler) setWorkspaceStatusCondition(ctx context.Context, wObj *kdmv1alpha1.Workspace, cType kdmv1alpha1.ConditionType,
	cStatus metav1.ConditionStatus, cReason, cMessage string) error {
	klog.InfoS("setWorkspaceStatusCondition", "workspace", klog.KObj(wObj), "conditionType", cType, "status", cStatus, "reason", cReason, "message", cMessage)
	cObj := metav1.Condition{
		Type:               string(cType),
		Status:             cStatus,
		Reason:             cReason,
		ObservedGeneration: wObj.GetGeneration(),
		Message:            cMessage,
	}
	meta.SetStatusCondition(&wObj.Status.Conditions, cObj)
	return c.updateWorkspaceStatus(ctx, wObj)
}

func (c *WorkspaceReconciler) setConditionMachineProvisionedToUnknown(ctx context.Context, wObj *kdmv1alpha1.Workspace, nodeObj *corev1.Node) error {
	klog.InfoS("setConditionMachineProvisionedToUnknown", "workspace", klog.KObj(wObj), "node", klog.KObj(nodeObj))
	err := c.setWorkspaceStatusCondition(ctx, wObj, kdmv1alpha1.WorkspaceConditionTypeMachineStatus, metav1.ConditionUnknown, "InstallNodePluginsWaiting",
		fmt.Sprintf("waiting for plugins to get installed on node %s", nodeObj.Name))
	if err != nil {
		klog.ErrorS(err, "failed to update workspace status", "workspace", wObj)
		return err
	}
	return nil
}

func (c *WorkspaceReconciler) updateWorkspaceStatusWithNodeList(ctx context.Context, wObj *kdmv1alpha1.Workspace, validNodeList []*corev1.Node) error {
	klog.InfoS("updateWorkspaceStatusWithNodeList", "workspace", klog.KObj(wObj))
	nodeNameList := lo.Map(validNodeList, func(v *corev1.Node, _ int) string {
		return v.Name
	})
	wObj.Status.WorkerNodes = nodeNameList
	return c.updateWorkspaceStatus(ctx, wObj)
}

package controllers

import (
	"context"

	kdmv1alpha1 "github.com/kdm/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

// updateWorkspaceStatus updates workspace status.
func (c *WorkspaceReconciler) setWorkspaceStatus(ctx context.Context, wObj *kdmv1alpha1.Workspace) error {
	klog.InfoS("updateWorkspaceStatus", "workspace", klog.KObj(wObj))

	return retry.OnError(retry.DefaultRetry,
		func(err error) bool {
			return apierrors.IsServiceUnavailable(err) || apierrors.IsServerTimeout(err) || apierrors.IsTooManyRequests(err)
		},
		func() error {
			return c.Client.Status().Update(ctx, wObj)
		})
}

func (c *WorkspaceReconciler) updateWorkspaceCondition(ctx context.Context, wObj *kdmv1alpha1.Workspace, cType kdmv1alpha1.ConditionType, cStatus metav1.ConditionStatus, cReason, cMessage string) {
	klog.InfoS("setWorkspaceCondition", "workspace", klog.KObj(wObj), "conditionType", cType, "status", cStatus)
	cObj := metav1.Condition{
		Type:               string(cType),
		Status:             cStatus,
		Reason:             cReason,
		ObservedGeneration: wObj.GetGeneration(),
		Message:            cMessage,
	}
	meta.SetStatusCondition(&wObj.Status.Conditions, cObj)
	c.setWorkspaceStatus(ctx, wObj)
}

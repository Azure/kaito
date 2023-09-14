package controllers

import (
	"context"
	"reflect"
	"sort"

	kdmv1alpha1 "github.com/kdm/api/v1alpha1"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

func (c *WorkspaceReconciler) updateWorkspaceStatus(ctx context.Context, wObj *kdmv1alpha1.Workspace) error {
	return retry.OnError(retry.DefaultRetry,
		func(err error) bool {
			return apierrors.IsServiceUnavailable(err) || apierrors.IsServerTimeout(err) || apierrors.IsTooManyRequests(err)
		},
		func() error {
			return c.Client.Status().Update(ctx, wObj)
		})
}

func (c *WorkspaceReconciler) updateStatusConditionIfNotMatch(ctx context.Context, wObj *kdmv1alpha1.Workspace, cType kdmv1alpha1.ConditionType,
	cStatus metav1.ConditionStatus, cReason, cMessage string) error {
	if curCondition := meta.FindStatusCondition(wObj.Status.Conditions, string(cType)); curCondition != nil {
		if curCondition.Status == cStatus && curCondition.Reason == cReason && curCondition.Message == cMessage {
			// Nonthing to change
			return nil
		}
	}
	klog.InfoS("updateStatusCondition", "workspace", klog.KObj(wObj), "conditionType", cType, "status", cStatus, "reason", cReason, "message", cMessage)
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

func (c *WorkspaceReconciler) updateStatusNodeListIfNotMatch(ctx context.Context, wObj *kdmv1alpha1.Workspace, validNodeList []*corev1.Node) error {
	nodeNameList := lo.Map(validNodeList, func(v *corev1.Node, _ int) string {
		return v.Name
	})
	sort.Strings(wObj.Status.WorkerNodes)
	sort.Strings(nodeNameList)
	if reflect.DeepEqual(wObj.Status.WorkerNodes, nodeNameList) {
		return nil
	}
	klog.InfoS("updateStatusNodeList", "workspace", klog.KObj(wObj))
	wObj.Status.WorkerNodes = nodeNameList
	return c.updateWorkspaceStatus(ctx, wObj)
}

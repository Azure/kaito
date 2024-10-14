// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package controllers

import (
	"context"
	"fmt"
	"reflect"
	"sort"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func updateObjStatus(ctx context.Context, c client.Client, name *client.ObjectKey, objType string, condition *metav1.Condition, workerNodes []string) error {
	return retry.OnError(retry.DefaultRetry,
		func(err error) bool {
			return apierrors.IsServiceUnavailable(err) || apierrors.IsServerTimeout(err) || apierrors.IsTooManyRequests(err)
		},
		func() error {
			var obj client.Object
			var conditions *[]metav1.Condition
			var workerNodesField *[]string
			switch objType {
			case "workspace":
				ragObj := &kaitov1alpha1.Workspace{}
				obj = ragObj
				conditions = &ragObj.Status.Conditions
				workerNodesField = &ragObj.Status.WorkerNodes
			case "ragengine":
				wObj := &kaitov1alpha1.RAGEngine{}
				obj = wObj
				conditions = &wObj.Status.Conditions
				workerNodesField = &wObj.Status.WorkerNodes
			default:
				return fmt.Errorf("unsupported object type: %s", objType)
			}

			if err := c.Get(ctx, *name, obj); err != nil {
				if !errors.IsNotFound(err) {
					return err
				}
				return nil
			}
			if condition != nil {
				meta.SetStatusCondition(conditions, *condition)
			}
			if workerNodes != nil {
				*workerNodesField = workerNodes
			}
			return c.Status().Update(ctx, obj)
		})
}

func updateStatusConditionIfNotMatch(ctx context.Context, obj client.Object, c *WorkspaceReconciler,
	name *client.ObjectKey, currentStatus kaitov1alpha1.WorkspaceStatus, cType kaitov1alpha1.ConditionType,
	cStatus metav1.ConditionStatus, objType string, cReason, cMessage string) error {
	if curCondition := meta.FindStatusCondition(currentStatus.Conditions, string(cType)); curCondition != nil {
		if curCondition.Status == cStatus && curCondition.Reason == cReason && curCondition.Message == cMessage {
			// Nonthing to change
			return nil
		}
	}
	klog.InfoS("updateStatusCondition", objType, klog.KObj(obj), "conditionType", cType, "status", cStatus, "reason", cReason, "message", cMessage)
	cObj := metav1.Condition{
		Type:               string(cType),
		Status:             cStatus,
		Reason:             cReason,
		ObservedGeneration: obj.GetGeneration(),
		Message:            cMessage,
	}
	return updateObjStatus(ctx, c.Client, name, objType, &cObj, nil)
}

func (c *WorkspaceReconciler) updateStatusNodeListIfNotMatch(ctx context.Context, wObj *kaitov1alpha1.Workspace, validNodeList []*corev1.Node) error {
	nodeNameList := lo.Map(validNodeList, func(v *corev1.Node, _ int) string {
		return v.Name
	})
	sort.Strings(wObj.Status.WorkerNodes)
	sort.Strings(nodeNameList)
	if reflect.DeepEqual(wObj.Status.WorkerNodes, nodeNameList) {
		return nil
	}
	klog.InfoS("updateStatusNodeList", "workspace", klog.KObj(wObj))
	return updateObjStatus(ctx, c.Client, &client.ObjectKey{Name: wObj.Name, Namespace: wObj.Namespace}, "workspace", nil, nodeNameList)
}

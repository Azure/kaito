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

func updateObjStatus(ctx context.Context, c client.Client, inputObj interface{}, name *client.ObjectKey, condition *metav1.Condition, workerNodes []string) error {
	return retry.OnError(retry.DefaultRetry,
		func(err error) bool {
			return apierrors.IsServiceUnavailable(err) || apierrors.IsServerTimeout(err) || apierrors.IsTooManyRequests(err)
		},
		func() error {
			var obj client.Object
			var conditions *[]metav1.Condition
			var workerNodesField *[]string

			switch inputObj.(type) {
			case *kaitov1alpha1.Workspace:
				wObj := &kaitov1alpha1.Workspace{}
				obj = wObj
				conditions = &wObj.Status.Conditions
				workerNodesField = &wObj.Status.WorkerNodes
			case *kaitov1alpha1.RAGEngine:
				ragObj := &kaitov1alpha1.RAGEngine{}
				obj = ragObj
				conditions = &ragObj.Status.Conditions
				workerNodesField = &ragObj.Status.WorkerNodes
			default:
				return fmt.Errorf("unsupported object type: %T", obj)
			}
			// After obj is fetched, *conditions and *workerNodesField will reflect the updated values
			// because they are pointers to the corresponding fields in obj.
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

func updateStatusConditionIfNotMatch(ctx context.Context, obj client.Object, inputObj interface{}, client client.Client,
	name *client.ObjectKey, existingConditions []metav1.Condition, cType kaitov1alpha1.ConditionType,
	cStatus metav1.ConditionStatus, cReason, cMessage string) error {
	if curCondition := meta.FindStatusCondition(existingConditions, string(cType)); curCondition != nil {
		if curCondition.Status == cStatus && curCondition.Reason == cReason && curCondition.Message == cMessage {
			// Nonthing to change
			return nil
		}
	}
	klog.InfoS("updateStatusCondition", "object", klog.KObj(obj), "conditionType", cType, "status", cStatus, "reason", cReason, "message", cMessage)
	cObj := metav1.Condition{
		Type:               string(cType),
		Status:             cStatus,
		Reason:             cReason,
		ObservedGeneration: obj.GetGeneration(),
		Message:            cMessage,
	}
	return updateObjStatus(ctx, client, inputObj, name, &cObj, nil)
}

func updateStatusNodeListIfNotMatch(ctx context.Context, inputObj interface{}, client client.Client,
	name *client.ObjectKey, validNodeList []*corev1.Node, workerNodes []string) error {
	nodeNameList := lo.Map(validNodeList, func(v *corev1.Node, _ int) string {
		return v.Name
	})
	sort.Strings(workerNodes)
	sort.Strings(nodeNameList)
	if reflect.DeepEqual(workerNodes, nodeNameList) {
		return nil
	}
	klog.InfoS("updateStatusNodeList", "object", name.String())
	return updateObjStatus(ctx, client, inputObj, name, nil, nodeNameList)
}

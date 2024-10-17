// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package controllers

import (
	"context"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *RAGEngineReconciler) updateRAGEngineStatus(ctx context.Context, name *client.ObjectKey, condition *metav1.Condition, workerNodes []string) error {
	return retry.OnError(retry.DefaultRetry,
		func(err error) bool {
			return apierrors.IsServiceUnavailable(err) || apierrors.IsServerTimeout(err) || apierrors.IsTooManyRequests(err)
		},
		func() error {
			// Read the latest version to avoid update conflict.
			ragObj := &kaitov1alpha1.RAGEngine{}
			if err := c.Client.Get(ctx, *name, ragObj); err != nil {
				if !errors.IsNotFound(err) {
					return err
				}
				return nil
			}
			if condition != nil {
				meta.SetStatusCondition(&ragObj.Status.Conditions, *condition)
			}
			if workerNodes != nil {
				ragObj.Status.WorkerNodes = workerNodes
			}
			return c.Client.Status().Update(ctx, ragObj)
		})
}

func (c *RAGEngineReconciler) updateStatusConditionIfNotMatch(ctx context.Context, ragObj *kaitov1alpha1.RAGEngine, cType kaitov1alpha1.ConditionType,
	cStatus metav1.ConditionStatus, cReason, cMessage string) error {
	if curCondition := meta.FindStatusCondition(ragObj.Status.Conditions, string(cType)); curCondition != nil {
		if curCondition.Status == cStatus && curCondition.Reason == cReason && curCondition.Message == cMessage {
			// Nonthing to change
			return nil
		}
	}
	klog.InfoS("updateStatusCondition", "ragengine", klog.KObj(ragObj), "conditionType", cType, "status", cStatus, "reason", cReason, "message", cMessage)
	cObj := metav1.Condition{
		Type:               string(cType),
		Status:             cStatus,
		Reason:             cReason,
		ObservedGeneration: ragObj.GetGeneration(),
		Message:            cMessage,
	}
	return c.updateRAGEngineStatus(ctx, &client.ObjectKey{Name: ragObj.Name, Namespace: ragObj.Namespace}, &cObj, nil)
}

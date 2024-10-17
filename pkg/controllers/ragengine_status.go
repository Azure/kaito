// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package controllers

import (
	"context"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
	return updateObjStatus(ctx, c.Client, &client.ObjectKey{Name: ragObj.Name, Namespace: ragObj.Namespace}, "RAGEngine", &cObj, nil)
}

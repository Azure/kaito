// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package controllers

import (
	"context"
	"errors"
	"testing"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/pkg/utils/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestUpdateRAGEngineStatus(t *testing.T) {
	t.Run("Should update ragengine status successfully", func(t *testing.T) {
		mockClient := test.NewClient()
		reconciler := &RAGEngineReconciler{
			Client: mockClient,
			Scheme: test.NewTestScheme(),
		}
		ctx := context.Background()
		ragengine := test.MockRAGEngineDistributedModel
		condition := metav1.Condition{
			Type:    "TestCondition",
			Status:  metav1.ConditionStatus("True"),
			Reason:  "TestReason",
			Message: "TestMessage",
		}
		workerNodes := []string{"node1", "node2"}

		mockClient.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&kaitov1alpha1.RAGEngine{}), mock.Anything).Return(nil)
		mockClient.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&kaitov1alpha1.RAGEngine{}), mock.Anything).Return(nil)

		err := reconciler.updateRAGEngineStatus(ctx, &client.ObjectKey{Name: ragengine.Name, Namespace: ragengine.Namespace}, &condition, workerNodes)
		assert.Nil(t, err)
	})

	t.Run("Should return error when Get operation fails", func(t *testing.T) {
		mockClient := test.NewClient()
		reconciler := &RAGEngineReconciler{
			Client: mockClient,
			Scheme: test.NewTestScheme(),
		}
		ctx := context.Background()
		ragengine := test.MockRAGEngineDistributedModel
		condition := metav1.Condition{
			Type:    "TestCondition",
			Status:  metav1.ConditionStatus("True"),
			Reason:  "TestReason",
			Message: "TestMessage",
		}
		workerNodes := []string{"node1", "node2"}

		mockClient.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&kaitov1alpha1.RAGEngine{}), mock.Anything).Return(errors.New("Get operation failed"))

		err := reconciler.updateRAGEngineStatus(ctx, &client.ObjectKey{Name: ragengine.Name, Namespace: ragengine.Namespace}, &condition, workerNodes)
		assert.NotNil(t, err)
	})

	t.Run("Should return nil when ragengine is not found", func(t *testing.T) {
		mockClient := test.NewClient()
		reconciler := &RAGEngineReconciler{
			Client: mockClient,
			Scheme: test.NewTestScheme(),
		}
		ctx := context.Background()
		ragengine := test.MockRAGEngineDistributedModel
		condition := metav1.Condition{
			Type:    "TestCondition",
			Status:  metav1.ConditionStatus("True"),
			Reason:  "TestReason",
			Message: "TestMessage",
		}
		workerNodes := []string{"node1", "node2"}

		mockClient.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&kaitov1alpha1.RAGEngine{}), mock.Anything).Return(apierrors.NewNotFound(schema.GroupResource{}, "ragengine"))

		err := reconciler.updateRAGEngineStatus(ctx, &client.ObjectKey{Name: ragengine.Name, Namespace: ragengine.Namespace}, &condition, workerNodes)
		assert.Nil(t, err)
	})
}

func TestRAGEngineUpdateStatusConditionIfNotMatch(t *testing.T) {
	t.Run("Should not update when condition matches", func(t *testing.T) {
		mockClient := test.NewClient()
		reconciler := &RAGEngineReconciler{
			Client: mockClient,
			Scheme: test.NewTestScheme(),
		}
		ctx := context.Background()
		ragengine := test.MockRAGEngineDistributedModel
		conditionType := kaitov1alpha1.ConditionType("TestCondition")
		conditionStatus := metav1.ConditionStatus("True")
		conditionReason := "TestReason"
		conditionMessage := "TestMessage"

		ragengine.Status.Conditions = []metav1.Condition{
			{
				Type:    string(conditionType),
				Status:  conditionStatus,
				Reason:  conditionReason,
				Message: conditionMessage,
			},
		}

		err := reconciler.updateStatusConditionIfNotMatch(ctx, ragengine, conditionType, conditionStatus, conditionReason, conditionMessage)
		assert.Nil(t, err)
	})

	t.Run("Should update when condition does not match", func(t *testing.T) {
		mockClient := test.NewClient()
		reconciler := &RAGEngineReconciler{
			Client: mockClient,
			Scheme: test.NewTestScheme(),
		}
		ctx := context.Background()
		ragengine := test.MockRAGEngineDistributedModel
		conditionType := kaitov1alpha1.ConditionType("TestCondition")
		conditionStatus := metav1.ConditionStatus("True")
		conditionReason := "TestReason"
		conditionMessage := "TestMessage"

		ragengine.Status.Conditions = []metav1.Condition{
			{
				Type:    string(conditionType),
				Status:  conditionStatus,
				Reason:  conditionReason,
				Message: "DifferentMessage",
			},
		}
		mockClient.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&kaitov1alpha1.RAGEngine{}), mock.Anything).Return(nil)
		mockClient.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&kaitov1alpha1.RAGEngine{}), mock.Anything).Return(nil)

		err := reconciler.updateStatusConditionIfNotMatch(ctx, ragengine, conditionType, conditionStatus, conditionReason, conditionMessage)
		assert.Nil(t, err)
	})

	t.Run("Should update when condition is not found", func(t *testing.T) {
		mockClient := test.NewClient()
		reconciler := &RAGEngineReconciler{
			Client: mockClient,
			Scheme: test.NewTestScheme(),
		}
		ctx := context.Background()
		ragengine := test.MockRAGEngineDistributedModel
		conditionType := kaitov1alpha1.ConditionType("TestCondition")
		conditionStatus := metav1.ConditionStatus("True")
		conditionReason := "TestReason"
		conditionMessage := "TestMessage"

		ragengine.Status.Conditions = []metav1.Condition{
			{
				Type:    "DifferentCondition",
				Status:  conditionStatus,
				Reason:  conditionReason,
				Message: conditionMessage,
			},
		}
		mockClient.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&kaitov1alpha1.RAGEngine{}), mock.Anything).Return(nil)
		mockClient.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&kaitov1alpha1.RAGEngine{}), mock.Anything).Return(nil)

		err := reconciler.updateStatusConditionIfNotMatch(ctx, ragengine, conditionType, conditionStatus, conditionReason, conditionMessage)
		assert.Nil(t, err)
	})
}

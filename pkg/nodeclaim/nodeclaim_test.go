// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package nodeclaim

import (
	"context"
	"errors"
	"os"
	"testing"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/pkg/utils/consts"
	"github.com/azure/kaito/pkg/utils/test"
	"github.com/stretchr/testify/mock"
	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/karpenter/pkg/apis/v1beta1"
)

func TestCreateNodeClaim(t *testing.T) {
	testcases := map[string]struct {
		callMocks           func(c *test.MockClient)
		nodeClaimConditions apis.Conditions
		expectedError       error
	}{
		"NodeClaim creation fails": {
			callMocks: func(c *test.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(errors.New("failed to create nodeClaim"))
			},
			expectedError: errors.New("failed to create nodeClaim"),
		},
		"NodeClaim creation fails because SKU is not available": {
			callMocks: func(c *test.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
			},
			nodeClaimConditions: apis.Conditions{
				{
					Type:    v1beta1.Launched,
					Status:  corev1.ConditionFalse,
					Message: ErrorInstanceTypesUnavailable,
				},
			},
			expectedError: errors.New(ErrorInstanceTypesUnavailable),
		},
		"A nodeClaim is successfully created": {
			callMocks: func(c *test.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
			},
			nodeClaimConditions: apis.Conditions{
				{
					Type:   apis.ConditionReady,
					Status: corev1.ConditionTrue,
				},
			},
			expectedError: nil,
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			mockClient := test.NewClient()
			tc.callMocks(mockClient)

			mockNodeClaim := &test.MockNodeClaim
			mockNodeClaim.Status.Conditions = tc.nodeClaimConditions

			err := CreateNodeClaim(context.Background(), mockNodeClaim, mockClient)
			if tc.expectedError == nil {
				assert.Check(t, err == nil, "Not expected to return error")
			} else {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			}
		})
	}
}

func TestWaitForPendingNodeClaims(t *testing.T) {
	testcases := map[string]struct {
		callMocks           func(c *test.MockClient)
		nodeClaimConditions apis.Conditions
		expectedError       error
	}{
		"Fail to list nodeClaims because associated nodeClaims cannot be retrieved": {
			callMocks: func(c *test.MockClient) {
				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaimList{}), mock.Anything).Return(errors.New("failed to retrieve nodeClaims"))
			},
			expectedError: errors.New("failed to retrieve nodeClaims"),
		},
		"Fail to list nodeClaims because nodeClaim status cannot be retrieved": {
			callMocks: func(c *test.MockClient) {
				nodeClaimList := test.MockNodeClaimList
				relevantMap := c.CreateMapWithType(nodeClaimList)
				c.CreateOrUpdateObjectInMap(&test.MockNodeClaim)

				//insert nodeClaim objects into the map
				for _, obj := range test.MockNodeClaimList.Items {
					m := obj
					objKey := client.ObjectKeyFromObject(&m)

					relevantMap[objKey] = &m
				}

				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaimList{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(errors.New("fail to get nodeClaim"))
			},
			nodeClaimConditions: apis.Conditions{
				{
					Type:   v1beta1.Initialized,
					Status: corev1.ConditionFalse,
				},
			},
			expectedError: errors.New("fail to get nodeClaim"),
		},
		"Successfully waits for all pending nodeClaims": {
			callMocks: func(c *test.MockClient) {
				nodeClaimList := test.MockNodeClaimList
				relevantMap := c.CreateMapWithType(nodeClaimList)
				c.CreateOrUpdateObjectInMap(&test.MockNodeClaim)

				//insert nodeClaim objects into the map
				for _, obj := range test.MockNodeClaimList.Items {
					m := obj
					objKey := client.ObjectKeyFromObject(&m)

					relevantMap[objKey] = &m
				}

				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaimList{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
			},
			nodeClaimConditions: apis.Conditions{
				{
					Type:   apis.ConditionReady,
					Status: corev1.ConditionTrue,
				},
			},
			expectedError: nil,
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			mockClient := test.NewClient()
			tc.callMocks(mockClient)

			mockNodeClaim := &v1beta1.NodeClaim{}

			mockClient.UpdateCb = func(key types.NamespacedName) {
				mockClient.GetObjectFromMap(mockNodeClaim, key)
				mockNodeClaim.Status.Conditions = tc.nodeClaimConditions
				mockClient.CreateOrUpdateObjectInMap(mockNodeClaim)
			}

			err := WaitForPendingNodeClaims(context.Background(), test.MockWorkspaceWithPreset, mockClient)
			if tc.expectedError == nil {
				assert.Check(t, err == nil, "Not expected to return error")
			} else {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			}
		})
	}
}

func TestGenerateNodeClaimManifest(t *testing.T) {
	t.Run("Should generate a nodeClaim object from the given workspace when cloud provider set to azure", func(t *testing.T) {
		mockWorkspace := test.MockWorkspaceWithPreset
		os.Setenv("CLOUD_PROVIDER", consts.AzureCloudName)
		nodeClaim := GenerateNodeClaimManifest(context.Background(), "0", mockWorkspace)

		assert.Check(t, nodeClaim != nil, "NodeClaim must not be nil")
		assert.Equal(t, nodeClaim.Namespace, mockWorkspace.Namespace, "NodeClaim must have same namespace as workspace")
		assert.Equal(t, nodeClaim.Labels[kaitov1alpha1.LabelWorkspaceName], mockWorkspace.Name, "label must have same workspace name as workspace")
		assert.Equal(t, nodeClaim.Labels[kaitov1alpha1.LabelWorkspaceNamespace], mockWorkspace.Namespace, "label must have same workspace namespace as workspace")
		assert.Equal(t, nodeClaim.Labels[LabelNodePool], KaitoNodePoolName, "label must have same labels as workspace label selector")
		assert.Equal(t, nodeClaim.Annotations[v1beta1.DoNotDisruptAnnotationKey], "true", "label must have do not disrupt annotation")
		assert.Equal(t, len(nodeClaim.Spec.Requirements), 4, " NodeClaim must have 4 NodeSelector Requirements")
		assert.Equal(t, nodeClaim.Spec.Requirements[1].NodeSelectorRequirement.Values[0], mockWorkspace.Resource.InstanceType, "NodeClaim must have same instance type as workspace")
		assert.Equal(t, nodeClaim.Spec.Requirements[2].NodeSelectorRequirement.Key, corev1.LabelOSStable, "NodeClaim must have OS label")
		assert.Check(t, nodeClaim.Spec.NodeClassRef != nil, "NodeClaim must have NodeClassRef")
		assert.Equal(t, nodeClaim.Spec.NodeClassRef.Kind, "AKSNodeClass", "NodeClaim must have 'AKSNodeClass' kind")
	})

	t.Run("Should generate a nodeClaim object from the given workspace when cloud provider set to aws", func(t *testing.T) {
		mockWorkspace := test.MockWorkspaceWithPreset
		os.Setenv("CLOUD_PROVIDER", "aws")
		nodeClaim := GenerateNodeClaimManifest(context.Background(), "0", mockWorkspace)

		assert.Check(t, nodeClaim != nil, "NodeClaim must not be nil")
		assert.Equal(t, nodeClaim.Namespace, mockWorkspace.Namespace, "NodeClaim must have same namespace as workspace")
		assert.Equal(t, nodeClaim.Labels[kaitov1alpha1.LabelWorkspaceName], mockWorkspace.Name, "label must have same workspace name as workspace")
		assert.Equal(t, nodeClaim.Labels[kaitov1alpha1.LabelWorkspaceNamespace], mockWorkspace.Namespace, "label must have same workspace namespace as workspace")
		assert.Equal(t, nodeClaim.Labels[LabelNodePool], KaitoNodePoolName, "label must have same labels as workspace label selector")
		assert.Equal(t, nodeClaim.Annotations[v1beta1.DoNotDisruptAnnotationKey], "true", "label must have do not disrupt annotation")
		assert.Equal(t, len(nodeClaim.Spec.Requirements), 3, " NodeClaim must have 3 NodeSelector Requirements")
		assert.Check(t, nodeClaim.Spec.NodeClassRef != nil, "NodeClaim must have NodeClassRef")
		assert.Equal(t, nodeClaim.Spec.NodeClassRef.Kind, "EC2NodeClass", "NodeClaim must have 'EC2NodeClass' kind")
	})
}

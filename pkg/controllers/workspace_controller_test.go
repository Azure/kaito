// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"

	azurev1alpha2 "github.com/Azure/karpenter-provider-azure/pkg/apis/v1alpha2"
	"github.com/aws/karpenter-core/pkg/apis/v1alpha5"
	awsv1beta1 "github.com/aws/karpenter-provider-aws/pkg/apis/v1beta1"
	"github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/pkg/featuregates"
	"github.com/azure/kaito/pkg/machine"
	"github.com/azure/kaito/pkg/nodeclaim"
	"github.com/azure/kaito/pkg/utils/consts"
	"github.com/azure/kaito/pkg/utils/test"
	"github.com/stretchr/testify/mock"
	"gotest.tools/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/karpenter/pkg/apis/v1beta1"
)

func TestSelectWorkspaceNodes(t *testing.T) {
	test.RegisterTestModel()
	testcases := map[string]struct {
		qualified             []*corev1.Node
		preferred             []string
		previous              []string
		count                 int
		karpenterFeatureGates bool
		expected              []string
	}{
		"two qualified nodes, need one": {
			qualified: []*corev1.Node{
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node1",
					},
				},
			},
			preferred: []string{},
			previous:  []string{},
			count:     1,
			expected:  []string{"node1"},
		},

		"three qualified nodes, prefer two of them": {
			qualified: []*corev1.Node{
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node1",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node3",
					},
				},
			},
			preferred: []string{"node3", "node2"},
			previous:  []string{},
			count:     2,
			expected:  []string{"node2", "node3"},
		},

		"three qualified nodes, two of them are selected previously, need two": {
			qualified: []*corev1.Node{
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node1",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node3",
					},
				},
			},
			preferred: []string{},
			previous:  []string{"node3", "node2"},
			count:     2,
			expected:  []string{"node2", "node3"},
		},

		"three qualified nodes, one preferred, one previous, need two": {
			qualified: []*corev1.Node{
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node1",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node3",
					},
				},
			},
			preferred: []string{"node3"},
			previous:  []string{"node2"},
			count:     2,
			expected:  []string{"node2", "node3"},
		},

		"three qualified nodes, one preferred, one previous, need one": {
			qualified: []*corev1.Node{
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node1",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node3",
					},
				},
			},
			preferred: []string{"node3"},
			previous:  []string{"node2"},
			count:     1,
			expected:  []string{"node3"},
		},

		"three qualified nodes, one is created by gpu-provisioner, need one": {
			qualified: []*corev1.Node{
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node1",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node3",
						Labels: map[string]string{
							machine.LabelGPUProvisionerCustom: consts.GPUString,
						},
					},
				},
			},
			preferred: []string{},
			previous:  []string{},
			count:     1,
			expected:  []string{"node3"},
		},
		"three qualified nodes, one is created by gpu-provisioner, one is preferred, one is previous, need two": {
			qualified: []*corev1.Node{
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node1",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node3",
						Labels: map[string]string{
							machine.LabelGPUProvisionerCustom: consts.GPUString,
						},
					},
				},
			},
			preferred: []string{"node2"},
			previous:  []string{"node1"},
			count:     2,
			expected:  []string{"node1", "node2"},
		},
		"three qualified nodes, one is created by gpu-provisioner, one is preferred, one is previous, need three": {
			qualified: []*corev1.Node{
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node1",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node3",
						Labels: map[string]string{
							machine.LabelGPUProvisionerCustom: consts.GPUString,
						},
					},
				},
			},
			preferred:             []string{"node2"},
			previous:              []string{"node1"},
			count:                 3,
			karpenterFeatureGates: false,
			expected:              []string{"node1", "node2", "node3"},
		},
		"three qualified nodes, one is created by gpu-provisioner (machine), the other created by karpenter (nodeClaim), one is preferred, need two": {
			qualified: []*corev1.Node{
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node1",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
						Labels: map[string]string{
							nodeclaim.LabelNodePool: nodeclaim.KaitoNodePoolName,
						},
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node3",
						Labels: map[string]string{
							machine.LabelGPUProvisionerCustom: consts.GPUString,
						},
					},
				},
			},
			preferred:             []string{"node1"},
			previous:              []string{},
			count:                 2,
			karpenterFeatureGates: true,
			expected:              []string{"node1", "node3"},
		},
		"three qualified nodes, one is created by  by karpenter (nodeClaim), two is preferred, need two": {
			qualified: []*corev1.Node{
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node1",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "node3",
						Labels: map[string]string{
							nodeclaim.LabelNodePool: nodeclaim.KaitoNodePoolName,
						},
					},
				},
			},
			preferred:             []string{"node1"},
			previous:              []string{},
			count:                 2,
			karpenterFeatureGates: true,
			expected:              []string{"node1", "node3"},
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			reconciler := &WorkspaceReconciler{
				Scheme: test.NewTestScheme(),
			}
			featuregates.FeatureGates[consts.FeatureFlagKarpenter] = tc.karpenterFeatureGates

			selectedNodes := reconciler.selectWorkspaceNodes(tc.qualified, tc.preferred, tc.previous, tc.count)

			selectedNodesArray := []string{}

			for _, each := range selectedNodes {
				selectedNodesArray = append(selectedNodesArray, each.Name)
			}

			sort.Strings(selectedNodesArray)
			sort.Strings(tc.expected)

			if !reflect.DeepEqual(selectedNodesArray, tc.expected) {
				t.Errorf("%s: selected Nodes %+v are different from the expected %+v", k, selectedNodesArray, tc.expected)
			}
		})
	}
}

func TestCreateAndValidateMachineNode(t *testing.T) {
	test.RegisterTestModel()
	testcases := map[string]struct {
		callMocks             func(c *test.MockClient)
		objectConditions      apis.Conditions
		workspace             v1alpha1.Workspace
		karpenterFeatureGates bool
		expectedError         error
	}{
		"Node is not created because machine creation fails": {
			callMocks: func(c *test.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&v1alpha5.Machine{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha5.Machine{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
			},
			objectConditions: apis.Conditions{
				{
					Type:    v1alpha5.MachineLaunched,
					Status:  corev1.ConditionFalse,
					Message: machine.ErrorInstanceTypesUnavailable,
				},
			},
			workspace:     *test.MockWorkspaceWithPreset,
			expectedError: errors.New(machine.ErrorInstanceTypesUnavailable),
		},
		"A machine is successfully created": {
			callMocks: func(c *test.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&v1alpha5.Machine{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha5.Machine{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&corev1.Node{}), mock.Anything).Return(nil)
			},
			objectConditions: apis.Conditions{
				{
					Type:   apis.ConditionReady,
					Status: corev1.ConditionTrue,
				},
			},
			workspace:     *test.MockWorkspaceDistributedModel,
			expectedError: nil,
		},
		"An Azure nodeClaim is successfully created": {
			callMocks: func(c *test.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&azurev1alpha2.AKSNodeClass{}), mock.Anything).Return(nil)
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&corev1.Node{}), mock.Anything).Return(nil)
				os.Setenv("CLOUD_PROVIDER", consts.AzureCloudName)
			},
			objectConditions: apis.Conditions{
				{
					Type:   apis.ConditionReady,
					Status: corev1.ConditionTrue,
				},
			},
			workspace:             *test.MockWorkspaceDistributedModel,
			karpenterFeatureGates: true,
			expectedError:         nil,
		},
		"An AWS nodeClaim is successfully created": {
			callMocks: func(c *test.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&awsv1beta1.EC2NodeClass{}), mock.Anything).Return(nil)
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&corev1.Node{}), mock.Anything).Return(nil)
				os.Setenv("CLOUD_PROVIDER", "aws")
			},
			objectConditions: apis.Conditions{
				{
					Type:   apis.ConditionReady,
					Status: corev1.ConditionTrue,
				},
			},
			workspace:             *test.MockWorkspaceDistributedModel,
			karpenterFeatureGates: true,
			expectedError:         nil,
		},
		"Node is not created because nodeClaim creation fails": {
			callMocks: func(c *test.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&azurev1alpha2.AKSNodeClass{}), mock.Anything).Return(nil)
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				os.Setenv("CLOUD_PROVIDER", consts.AzureCloudName)
			},
			objectConditions: apis.Conditions{
				{
					Type:    v1beta1.Launched,
					Status:  corev1.ConditionFalse,
					Message: nodeclaim.ErrorInstanceTypesUnavailable,
				},
			},
			workspace:             *test.MockWorkspaceWithPreset,
			karpenterFeatureGates: true,
			expectedError:         errors.New(nodeclaim.ErrorInstanceTypesUnavailable),
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			mockClient := test.NewClient()
			mockMachine := &v1alpha5.Machine{}
			mockNodeClaim := &v1beta1.NodeClaim{}

			mockClient.UpdateCb = func(key types.NamespacedName) {
				mockClient.GetObjectFromMap(mockMachine, key)
				mockMachine.Status.Conditions = tc.objectConditions
				mockClient.CreateOrUpdateObjectInMap(mockMachine)

				if tc.karpenterFeatureGates {
					mockClient.GetObjectFromMap(mockNodeClaim, key)
					mockNodeClaim.Status.Conditions = tc.objectConditions
					mockClient.CreateOrUpdateObjectInMap(mockNodeClaim)
				}
			}

			tc.callMocks(mockClient)

			reconciler := &WorkspaceReconciler{
				Client: mockClient,
				Scheme: test.NewTestScheme(),
			}
			ctx := context.Background()
			featuregates.FeatureGates[consts.FeatureFlagKarpenter] = tc.karpenterFeatureGates

			node, err := reconciler.createAndValidateNode(ctx, &tc.workspace)
			if tc.expectedError == nil {
				assert.Check(t, err == nil, "Not expected to return error")
				assert.Check(t, node != nil, "Response node should not be nil")
			} else {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			}
		})
	}
}

func TestCreateAndValidateNodeClaimNode(t *testing.T) {
	test.RegisterTestModel()
	testcases := map[string]struct {
		callMocks             func(c *test.MockClient)
		karpenterFeatureGates bool
		nodeClaimConditions   apis.Conditions
		workspace             v1alpha1.Workspace
		expectedError         error
	}{
		"Node is not created because nodeClaim creation fails": {
			callMocks: func(c *test.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&azurev1alpha2.AKSNodeClass{}), mock.Anything).Return(nil)
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				os.Setenv("CLOUD_PROVIDER", consts.AzureCloudName)
			},
			karpenterFeatureGates: true,
			nodeClaimConditions: apis.Conditions{
				{
					Type:    v1beta1.Launched,
					Status:  corev1.ConditionFalse,
					Message: nodeclaim.ErrorInstanceTypesUnavailable,
				},
			},
			workspace:     *test.MockWorkspaceWithPreset,
			expectedError: errors.New(nodeclaim.ErrorInstanceTypesUnavailable),
		},
		"A nodeClaim is successfully created": {
			callMocks: func(c *test.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&azurev1alpha2.AKSNodeClass{}), mock.Anything).Return(nil)
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&corev1.Node{}), mock.Anything).Return(nil)
			},
			karpenterFeatureGates: true,
			nodeClaimConditions: apis.Conditions{
				{
					Type:   apis.ConditionReady,
					Status: corev1.ConditionTrue,
				},
			},
			workspace:     *test.MockWorkspaceDistributedModel,
			expectedError: nil,
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			mockClient := test.NewClient()
			mockNodeClaim := &v1beta1.NodeClaim{}

			mockClient.UpdateCb = func(key types.NamespacedName) {
				mockClient.GetObjectFromMap(mockNodeClaim, key)
				mockNodeClaim.Status.Conditions = tc.nodeClaimConditions
				mockClient.CreateOrUpdateObjectInMap(mockNodeClaim)
			}

			tc.callMocks(mockClient)

			featuregates.FeatureGates[consts.FeatureFlagKarpenter] = tc.karpenterFeatureGates

			reconciler := &WorkspaceReconciler{
				Client: mockClient,
				Scheme: test.NewTestScheme(),
			}
			ctx := context.Background()

			node, err := reconciler.createAndValidateNode(ctx, &tc.workspace)
			if tc.expectedError == nil {
				assert.Check(t, err == nil, "Not expected to return error")
				assert.Check(t, node != nil, "Response node should not be nil")
			} else {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			}
		})
	}
}

func TestEnsureService(t *testing.T) {
	test.RegisterTestModel()
	testcases := map[string]struct {
		callMocks     func(c *test.MockClient)
		expectedError error
	}{
		"Existing service is found for workspace": {
			callMocks: func(c *test.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&corev1.Service{}), mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		"Service creation fails": {
			callMocks: func(c *test.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&corev1.Service{}), mock.Anything).Return(test.NotFoundError())
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&corev1.Service{}), mock.Anything).Return(errors.New("cannot create service"))
			},
			expectedError: errors.New("cannot create service"),
		},
		"Successfully creates a new service": {
			callMocks: func(c *test.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&corev1.Service{}), mock.Anything).Return(test.NotFoundError())
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&corev1.Service{}), mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			mockClient := test.NewClient()
			tc.callMocks(mockClient)

			reconciler := &WorkspaceReconciler{
				Client: mockClient,
				Scheme: test.NewTestScheme(),
			}
			ctx := context.Background()

			err := reconciler.ensureService(ctx, test.MockWorkspaceDistributedModel)
			if tc.expectedError == nil {
				assert.Check(t, err == nil, "Not expected to return error")
			} else {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			}
		})
	}

}

func TestApplyInferenceWithPreset(t *testing.T) {
	test.RegisterTestModel()
	testcases := map[string]struct {
		callMocks     func(c *test.MockClient)
		workspace     v1alpha1.Workspace
		expectedError error
	}{
		"Fail to get inference because associated workload with workspace cannot be retrieved": {
			callMocks: func(c *test.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&appsv1.StatefulSet{}), mock.Anything).Return(errors.New("Failed to get resource"))

				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
			},
			workspace:     *test.MockWorkspaceDistributedModel,
			expectedError: errors.New("Failed to get resource"),
		},
		"Create preset inference because inference workload did not exist": {
			callMocks: func(c *test.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&appsv1.Deployment{}), mock.Anything).Return(test.NotFoundError()).Times(4)
				c.On("Get", mock.Anything, mock.Anything, mock.IsType(&appsv1.Deployment{}), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					depObj := &appsv1.Deployment{}
					key := client.ObjectKey{Namespace: "kaito", Name: "testWorkspace"}
					c.GetObjectFromMap(depObj, key)
					depObj.Status.ReadyReplicas = 1
					c.CreateOrUpdateObjectInMap(depObj)
				})
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&appsv1.Deployment{}), mock.Anything).Return(nil)

				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&corev1.Service{}), mock.Anything).Return(nil)

				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
			},
			workspace:     *test.MockWorkspaceWithPreset,
			expectedError: nil,
		},
		"Apply inference from existing workload": {
			callMocks: func(c *test.MockClient) {
				c.On("Get", mock.Anything, mock.Anything, mock.IsType(&appsv1.StatefulSet{}), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					depObj := &appsv1.StatefulSet{}
					key := client.ObjectKey{Namespace: "kaito", Name: "testWorkspace"}
					c.GetObjectFromMap(depObj, key)
					numRep := int32(1)
					depObj.Status.ReadyReplicas = numRep
					depObj.Spec.Replicas = &numRep
					c.CreateOrUpdateObjectInMap(depObj)
				})

				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
			},
			workspace:     *test.MockWorkspaceDistributedModel,
			expectedError: nil,
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			mockClient := test.NewClient()
			tc.callMocks(mockClient)

			reconciler := &WorkspaceReconciler{
				Client: mockClient,
				Scheme: test.NewTestScheme(),
			}
			ctx := context.Background()

			err := reconciler.applyInference(ctx, &tc.workspace)
			if tc.expectedError == nil {
				assert.Check(t, err == nil, "Not expected to return error")
			} else {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			}
		})
	}
}

func TestApplyInferenceWithTemplate(t *testing.T) {
	testcases := map[string]struct {
		callMocks     func(c *test.MockClient)
		workspace     v1alpha1.Workspace
		expectedError error
	}{
		"Fail to apply inference from workspace template": {
			callMocks: func(c *test.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&appsv1.Deployment{}), mock.Anything).Return(errors.New("Failed to create deployment"))
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
			},
			workspace:     *test.MockWorkspaceWithInferenceTemplate,
			expectedError: errors.New("Failed to create deployment"),
		},
		"Apply inference from workspace template": {
			callMocks: func(c *test.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&appsv1.Deployment{}), mock.Anything).Return(nil)
				c.On("Get", mock.Anything, mock.Anything, mock.IsType(&appsv1.Deployment{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
			},
			workspace:     *test.MockWorkspaceWithInferenceTemplate,
			expectedError: nil,
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			mockClient := test.NewClient()
			depObj := &appsv1.Deployment{}

			mockClient.UpdateCb = func(key types.NamespacedName) {
				mockClient.GetObjectFromMap(depObj, key)
				depObj.Status.ReadyReplicas = 1
				mockClient.CreateOrUpdateObjectInMap(depObj)
			}

			tc.callMocks(mockClient)

			reconciler := &WorkspaceReconciler{
				Client: mockClient,
				Scheme: test.NewTestScheme(),
			}
			ctx := context.Background()

			err := reconciler.applyInference(ctx, &tc.workspace)
			if tc.expectedError == nil {
				assert.Check(t, err == nil, "Not expected to return error")
			} else {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			}
		})
	}
}

func TestGetAllQualifiedNodes(t *testing.T) {
	testcases := map[string]struct {
		callMocks     func(c *test.MockClient)
		expectedError error
	}{
		"Fails to get qualified nodes because can't list nodes": {
			callMocks: func(c *test.MockClient) {
				c.On("List", mock.IsType(context.Background()), mock.IsType(&corev1.NodeList{}), mock.Anything).Return(errors.New("Failed to list nodes"))
			},
			expectedError: errors.New("Failed to list nodes"),
		},
		"Gets all qualified nodes": {
			callMocks: func(c *test.MockClient) {
				nodeList := test.MockNodeList
				deletedNode := corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node4",
						Labels: map[string]string{
							corev1.LabelInstanceTypeStable: "Standard_NC12s_v3",
						},
						DeletionTimestamp: &v1.Time{Time: time.Now()},
					},
				}
				nodeList.Items = append(nodeList.Items, deletedNode)

				relevantMap := c.CreateMapWithType(nodeList)
				//insert node objects into the map
				for _, obj := range test.MockNodeList.Items {
					n := obj
					objKey := client.ObjectKeyFromObject(&n)

					relevantMap[objKey] = &n
				}

				c.On("List", mock.IsType(context.Background()), mock.IsType(&corev1.NodeList{}), mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			mockClient := test.NewClient()
			mockWorkspace := test.MockWorkspaceDistributedModel
			reconciler := &WorkspaceReconciler{
				Client: mockClient,
				Scheme: test.NewTestScheme(),
			}
			ctx := context.Background()

			tc.callMocks(mockClient)

			nodes, err := reconciler.getAllQualifiedNodes(ctx, mockWorkspace)
			if tc.expectedError == nil {
				assert.Check(t, err == nil, "Not expected to return error")
				assert.Check(t, nodes != nil, "Response node array should not be nil")
				assert.Check(t, len(nodes) == 1, "One out of three nodes should be qualified")
			} else {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
				assert.Check(t, nodes == nil, "Response node array should be nil")
			}
		})
	}
}

func TestApplyWorkspaceResource(t *testing.T) {
	test.RegisterTestModel()
	testcases := map[string]struct {
		callMocks                   func(c *test.MockClient)
		karpenterFeatureGateEnabled bool
		expectedError               error
		workspace                   v1alpha1.Workspace
	}{
		"Fail to apply workspace because associated machines cannot be retrieved": {
			callMocks: func(c *test.MockClient) {
				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1alpha5.MachineList{}), mock.Anything).Return(errors.New("failed to retrieve machines"))
			},
			workspace:     *test.MockWorkspaceDistributedModel,
			expectedError: errors.New("failed to retrieve machines"),
		},
		"Fail to apply workspace because can't get qualified nodes": {
			callMocks: func(c *test.MockClient) {
				machineList := test.MockMachineList
				relevantMap := c.CreateMapWithType(machineList)
				c.CreateOrUpdateObjectInMap(&test.MockMachine)

				//insert machine objects into the map
				for _, obj := range machineList.Items {
					m := obj
					objKey := client.ObjectKeyFromObject(&m)

					relevantMap[objKey] = &m
				}

				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1alpha5.MachineList{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha5.Machine{}), mock.Anything).Return(nil)

				c.On("List", mock.IsType(context.Background()), mock.IsType(&corev1.NodeList{}), mock.Anything).Return(errors.New("failed to list nodes"))
			},
			workspace:     *test.MockWorkspaceDistributedModel,
			expectedError: errors.New("failed to list nodes"),
		},
		"Fail to apply workspace because associated nodeClaim cannot be retrieved": {
			callMocks: func(c *test.MockClient) {
				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1alpha5.MachineList{}), mock.Anything).Return(nil)
				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaimList{}), mock.Anything).Return(errors.New("failed to retrieve nodeClaims"))

			},
			karpenterFeatureGateEnabled: true,
			workspace:                   *test.MockWorkspaceDistributedModel,
			expectedError:               errors.New("failed to retrieve nodeClaims"),
		},
		"Fail to apply workspace with nodeClaims because can't get qualified nodes": {
			callMocks: func(c *test.MockClient) {
				nodeClaimList := test.MockNodeClaimList
				relevantMap := c.CreateMapWithType(nodeClaimList)
				c.CreateOrUpdateObjectInMap(&test.MockNodeClaim)

				//insert nodeClaim objects into the map
				for _, obj := range nodeClaimList.Items {
					m := obj
					objKey := client.ObjectKeyFromObject(&m)

					relevantMap[objKey] = &m
				}
				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1alpha5.MachineList{}), mock.Anything).Return(nil)
				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaimList{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)

				c.On("List", mock.IsType(context.Background()), mock.IsType(&corev1.NodeList{}), mock.Anything).Return(errors.New("failed to list nodes"))
			},
			karpenterFeatureGateEnabled: true,
			workspace:                   *test.MockWorkspaceDistributedModel,
			expectedError:               errors.New("failed to list nodes"),
		},
		"Successfully apply workspace resource with machine": {
			callMocks: func(c *test.MockClient) {
				nodeList := test.MockNodeList
				relevantMap := c.CreateMapWithType(nodeList)
				//insert node objects into the map
				for _, obj := range test.MockNodeList.Items {
					n := obj
					objKey := client.ObjectKeyFromObject(&n)

					relevantMap[objKey] = &n
				}

				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1alpha5.MachineList{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha5.Machine{}), mock.Anything).Return(nil)

				c.On("List", mock.IsType(context.Background()), mock.IsType(&corev1.NodeList{}), mock.Anything).Return(nil)

				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)

			},
			workspace:     *test.MockWorkspaceDistributedModel,
			expectedError: nil,
		},
		"Successfully apply workspace resource with nodeClaim": {
			callMocks: func(c *test.MockClient) {
				nodeList := test.MockNodeList
				relevantMap := c.CreateMapWithType(nodeList)
				//insert node objects into the map
				for _, obj := range nodeList.Items {
					n := obj
					objKey := client.ObjectKeyFromObject(&n)

					relevantMap[objKey] = &n
				}

				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1alpha5.MachineList{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha5.Machine{}), mock.Anything).Return(nil)

				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaimList{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)

				c.On("List", mock.IsType(context.Background()), mock.IsType(&corev1.NodeList{}), mock.Anything).Return(nil)

				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)

			},
			karpenterFeatureGateEnabled: true,
			workspace:                   *test.MockWorkspaceDistributedModel,
			expectedError:               nil,
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			mockClient := test.NewClient()
			tc.callMocks(mockClient)

			mockMachine := &v1alpha5.Machine{}
			mockNodeClaim := &v1beta1.NodeClaim{}

			mockClient.UpdateCb = func(key types.NamespacedName) {
				mockClient.GetObjectFromMap(mockMachine, key)
				mockMachine.Status.Conditions = apis.Conditions{
					{
						Type:   apis.ConditionReady,
						Status: corev1.ConditionTrue,
					},
				}
				mockClient.CreateOrUpdateObjectInMap(mockMachine)

				mockClient.GetObjectFromMap(mockNodeClaim, key)
				mockNodeClaim.Status.Conditions = apis.Conditions{
					{
						Type:   apis.ConditionReady,
						Status: corev1.ConditionTrue,
					},
				}
				mockClient.CreateOrUpdateObjectInMap(mockNodeClaim)
			}

			reconciler := &WorkspaceReconciler{
				Client: mockClient,
				Scheme: test.NewTestScheme(),
			}
			featuregates.FeatureGates[consts.FeatureFlagKarpenter] = tc.karpenterFeatureGateEnabled
			ctx := context.Background()

			err := reconciler.applyWorkspaceResource(ctx, &tc.workspace)
			if tc.expectedError == nil {
				assert.Check(t, err == nil, "Not expected to return error")
			} else {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			}
		})
	}
}

func TestUpdateControllerRevision(t *testing.T) {
	testcases := map[string]struct {
		callMocks     func(c *test.MockClient)
		workspace     v1alpha1.Workspace
		expectedError error
		verifyCalls   func(c *test.MockClient)
	}{
		"No new revision needed": {
			callMocks: func(c *test.MockClient) {
				c.On("List", mock.IsType(context.Background()), mock.Anything, mock.IsType(&appsv1.ControllerRevisionList{}), mock.Anything).Return(errors.New("should not be called"))
			},
			workspace:     test.MockWorkspaceWithComputeHash,
			expectedError: nil,
			verifyCalls: func(c *test.MockClient) {
				c.AssertNumberOfCalls(t, "List", 0)
				c.AssertNumberOfCalls(t, "Create", 0)
				c.AssertNumberOfCalls(t, "Update", 0)
				c.AssertNumberOfCalls(t, "Delete", 0)
			},
		},
		"Fail to create ControllerRevision": {
			callMocks: func(c *test.MockClient) {
				c.On("List", mock.IsType(context.Background()), mock.IsType(&appsv1.ControllerRevisionList{}), mock.Anything).Return(nil)
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&appsv1.ControllerRevision{}), mock.Anything).Return(errors.New("failed to create ControllerRevision"))
				c.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
			},
			workspace:     test.MockWorkspaceFailToCreateCR,
			expectedError: errors.New("failed to create new ControllerRevision: failed to create ControllerRevision"),
			verifyCalls: func(c *test.MockClient) {
				c.AssertNumberOfCalls(t, "List", 1)
				c.AssertNumberOfCalls(t, "Create", 1)
				c.AssertNumberOfCalls(t, "Update", 1)
				c.AssertNumberOfCalls(t, "Delete", 0)
			},
		},
		"Successfully create new ControllerRevision": {
			callMocks: func(c *test.MockClient) {
				c.On("List", mock.IsType(context.Background()), mock.IsType(&appsv1.ControllerRevisionList{}), mock.Anything).Return(nil)
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&appsv1.ControllerRevision{}), mock.Anything).Return(nil)
				c.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
			},
			workspace:     test.MockWorkspaceSuccessful,
			expectedError: nil,
			verifyCalls: func(c *test.MockClient) {
				c.AssertNumberOfCalls(t, "List", 1)
				c.AssertNumberOfCalls(t, "Create", 1)
				c.AssertNumberOfCalls(t, "Update", 1)
				c.AssertNumberOfCalls(t, "Delete", 0)
			},
		},
		"Successfully delete old ControllerRevision": {
			callMocks: func(c *test.MockClient) {
				revisions := &appsv1.ControllerRevisionList{}
				for i := 0; i <= consts.MaxRevisionHistoryLimit; i++ {
					revision := &appsv1.ControllerRevision{
						ObjectMeta: v1.ObjectMeta{
							Name: fmt.Sprintf("revision-%d", i),
						},
						Revision: int64(i),
					}
					revisions.Items = append(revisions.Items, *revision)
				}
				relevantMap := c.CreateMapWithType(revisions)

				for _, obj := range revisions.Items {
					m := obj
					objKey := client.ObjectKeyFromObject(&m)
					relevantMap[objKey] = &m
				}

				c.On("List", mock.IsType(context.Background()), mock.IsType(&appsv1.ControllerRevisionList{}), mock.Anything).Return(nil)
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&appsv1.ControllerRevision{}), mock.Anything).Return(nil)
				c.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.On("Delete", mock.IsType(context.Background()), mock.IsType(&appsv1.ControllerRevision{}), mock.Anything).Return(nil)
			},
			workspace:     test.MockWorkspaceWithDeleteOldCR,
			expectedError: nil,
			verifyCalls: func(c *test.MockClient) {
				c.AssertNumberOfCalls(t, "List", 1)
				c.AssertNumberOfCalls(t, "Create", 1)
				c.AssertNumberOfCalls(t, "Update", 1)
				c.AssertNumberOfCalls(t, "Delete", 1)
			},
		},
		"Deployment updated when adapters change": {
			callMocks: func(c *test.MockClient) {
				revisions := &appsv1.ControllerRevisionList{}
				jsonData, _ := json.Marshal(test.MockWorkspaceWithComputeHash)
				revision := &appsv1.ControllerRevision{
					ObjectMeta: v1.ObjectMeta{
						Name:        "revision-1",
						Annotations: test.MockWorkspaceWithComputeHash.Annotations,
					},
					Revision: int64(1),
					Data:     runtime.RawExtension{Raw: jsonData},
				}
				revisions.Items = append(revisions.Items, *revision)

				relevantMap := c.CreateMapWithType(revisions)

				for _, obj := range revisions.Items {
					m := obj
					objKey := client.ObjectKeyFromObject(&m)
					relevantMap[objKey] = &m
				}

				c.CreateOrUpdateObjectInMap(&test.MockDeployment)

				c.On("List", mock.IsType(context.Background()), mock.IsType(&appsv1.ControllerRevisionList{}), mock.Anything).Return(nil)
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&appsv1.ControllerRevision{}), mock.Anything).Return(nil)
				c.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)

				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&appsv1.Deployment{}), mock.Anything).Return(nil)

				c.On("Update", mock.IsType(context.Background()), mock.IsType(&appsv1.Deployment{}), mock.Anything).Return(nil)
			},
			workspace:     test.MockWorkspaceWithUpdatedDeployment,
			expectedError: nil,
			verifyCalls: func(c *test.MockClient) {
				c.AssertNumberOfCalls(t, "List", 1)
				c.AssertNumberOfCalls(t, "Create", 1)
				c.AssertNumberOfCalls(t, "Update", 2) // one for workspace and one for deployment
				c.AssertNumberOfCalls(t, "Get", 1)
				c.AssertNumberOfCalls(t, "Delete", 0)
			},
		},
	}
	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			mockClient := test.NewClient()
			tc.callMocks(mockClient)

			reconciler := &WorkspaceReconciler{
				Client: mockClient,
				Scheme: test.NewTestScheme(),
			}
			ctx := context.Background()

			err := reconciler.updateControllerRevision(ctx, &tc.workspace)
			if tc.expectedError == nil {
				assert.Check(t, err == nil, "Not expected to return error")
			} else {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			}
			if tc.verifyCalls != nil {
				tc.verifyCalls(mockClient)
			}
		})
	}
}

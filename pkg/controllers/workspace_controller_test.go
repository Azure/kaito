// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package controllers

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/aws/karpenter-core/pkg/apis/v1alpha5"
	"github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/pkg/machine"
	"github.com/azure/kaito/pkg/utils"
	"github.com/stretchr/testify/mock"
	"gotest.tools/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestSelectWorkspaceNodes(t *testing.T) {

	testcases := map[string]struct {
		qualified []*corev1.Node
		preferred []string
		previous  []string
		count     int
		expected  []string
	}{
		"two qualified nodes, need one": {
			qualified: []*corev1.Node{
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
					},
				},
				&corev1.Node{
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
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node1",
					},
				},
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
					},
				},
				&corev1.Node{
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
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node1",
					},
				},
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
					},
				},
				&corev1.Node{
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
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node1",
					},
				},
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
					},
				},
				&corev1.Node{
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
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node1",
					},
				},
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
					},
				},
				&corev1.Node{
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

		"three qualified nodes, one is created by kaito, need one": {
			qualified: []*corev1.Node{
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node1",
					},
				},
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
					},
				},
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node3",
						Labels: map[string]string{
							"kaito.sh/machine-type": "gpu",
						},
					},
				},
			},
			preferred: []string{},
			previous:  []string{},
			count:     1,
			expected:  []string{"node3"},
		},
		"three qualified nodes, one is created by kaito, one is preferred, one is previous, need two": {
			qualified: []*corev1.Node{
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node1",
					},
				},
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
					},
				},
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node3",
						Labels: map[string]string{
							"kaito.sh/machine-type": "gpu",
						},
					},
				},
			},
			preferred: []string{"node2"},
			previous:  []string{"node1"},
			count:     2,
			expected:  []string{"node1", "node2"},
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {

			selectedNodes := selectWorkspaceNodes(tc.qualified, tc.preferred, tc.previous, tc.count)

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

func TestCreateAndValidateNode(t *testing.T) {
	testcases := map[string]struct {
		callMocks         func(c *utils.MockClient)
		machineConditions apis.Conditions
		expectedError     error
	}{
		"Node is not created because machine creation fails": {
			callMocks: func(c *utils.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&v1alpha5.Machine{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha5.Machine{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
			},
			machineConditions: apis.Conditions{
				{
					Type:    v1alpha5.MachineLaunched,
					Status:  corev1.ConditionFalse,
					Message: machine.ErrorInstanceTypesUnavailable,
				},
			},
			expectedError: errors.New(machine.ErrorInstanceTypesUnavailable),
		},
		"A machine is successfully created": {
			callMocks: func(c *utils.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&v1alpha5.Machine{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha5.Machine{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&corev1.Node{}), mock.Anything).Return(nil)
			},
			machineConditions: apis.Conditions{
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
			mockClient := utils.NewClient()
			mockMachine := &v1alpha5.Machine{}

			mockClient.UpdateCb = func(key types.NamespacedName) {
				mockClient.GetObjectFromMap(mockMachine, key)
				mockMachine.Status.Conditions = tc.machineConditions
				mockClient.CreateOrUpdateObjectInMap(mockMachine)
			}

			tc.callMocks(mockClient)

			reconciler := &WorkspaceReconciler{
				Client: mockClient,
				Scheme: utils.NewTestScheme(),
			}
			ctx := context.Background()

			node, err := reconciler.createAndValidateNode(ctx, utils.MockWorkspace)
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
	testcases := map[string]struct {
		callMocks     func(c *utils.MockClient)
		expectedError error
	}{
		"Existing service is found for workspace": {
			callMocks: func(c *utils.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&corev1.Service{}), mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		"Service creation fails": {
			callMocks: func(c *utils.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&corev1.Service{}), mock.Anything).Return(utils.NotFoundError())
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&corev1.Service{}), mock.Anything).Return(errors.New("cannot create service"))
			},
			expectedError: errors.New("cannot create service"),
		},
		"Successfully creates a new service": {
			callMocks: func(c *utils.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&corev1.Service{}), mock.Anything).Return(utils.NotFoundError())
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&corev1.Service{}), mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			mockClient := utils.NewClient()
			mockClient.UpdateCb = func(key types.NamespacedName) {}

			tc.callMocks(mockClient)

			reconciler := &WorkspaceReconciler{
				Client: mockClient,
				Scheme: utils.NewTestScheme(),
			}
			ctx := context.Background()

			err := reconciler.ensureService(ctx, utils.MockWorkspace, false)
			if tc.expectedError == nil {
				assert.Check(t, err == nil, "Not expected to return error")
			} else {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			}
		})
	}

}

func TestApplyInferenceWithPreset(t *testing.T) {
	testcases := map[string]struct {
		callMocks     func(c *utils.MockClient)
		workspace     v1alpha1.Workspace
		expectedError error
	}{
		"Fail to get inference because preset model is not supported": {
			callMocks: func(c *utils.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
			},
			workspace:     *utils.MockWorkspaceWithBadPreset,
			expectedError: errors.New("preset model does-not-exist is not supported"),
		},
		"Fail to get inference because associated workload with workspace cannot be retrieved": {
			callMocks: func(c *utils.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&appsv1.StatefulSet{}), mock.Anything).Return(errors.New("Failed to get resource"))

				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
			},
			workspace:     *utils.MockWorkspace,
			expectedError: errors.New("Failed to get resource"),
		},
		"Create preset inference because inference workload did not exist": {
			callMocks: func(c *utils.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&appsv1.Deployment{}), mock.Anything).Return(utils.NotFoundError()).Times(4)
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
			workspace:     *utils.MockWorkspaceWithFalconPreset,
			expectedError: nil,
		},
		"Apply inference from existing workload": {
			callMocks: func(c *utils.MockClient) {
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
			workspace:     *utils.MockWorkspace,
			expectedError: nil,
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			mockClient := utils.NewClient()

			mockClient.UpdateCb = func(key types.NamespacedName) {}

			tc.callMocks(mockClient)

			reconciler := &WorkspaceReconciler{
				Client: mockClient,
				Scheme: utils.NewTestScheme(),
			}
			ctx := context.Background()

			err := reconciler.applyInference(ctx, &tc.workspace, false)
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
		callMocks     func(c *utils.MockClient)
		workspace     v1alpha1.Workspace
		expectedError error
	}{
		"Fail to apply inference from workspace template": {
			callMocks: func(c *utils.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&appsv1.Deployment{}), mock.Anything).Return(errors.New("Failed to create deployment"))
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
			},
			workspace:     *utils.MockWorkspaceWithInferenceTemplate,
			expectedError: errors.New("Failed to create deployment"),
		},
		"Apply inference from workspace template": {
			callMocks: func(c *utils.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&appsv1.Deployment{}), mock.Anything).Return(nil)
				c.On("Get", mock.Anything, mock.Anything, mock.IsType(&appsv1.Deployment{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
			},
			workspace:     *utils.MockWorkspaceWithInferenceTemplate,
			expectedError: nil,
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			mockClient := utils.NewClient()
			depObj := &appsv1.Deployment{}

			mockClient.UpdateCb = func(key types.NamespacedName) {
				mockClient.GetObjectFromMap(depObj, key)
				depObj.Status.ReadyReplicas = 1
				mockClient.CreateOrUpdateObjectInMap(depObj)
			}

			tc.callMocks(mockClient)

			reconciler := &WorkspaceReconciler{
				Client: mockClient,
				Scheme: utils.NewTestScheme(),
			}
			ctx := context.Background()

			err := reconciler.applyInference(ctx, &tc.workspace, false)
			if tc.expectedError == nil {
				assert.Check(t, err == nil, "Not expected to return error")
			} else {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			}
		})
	}

}

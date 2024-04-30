package controllers

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/karpenter-core/pkg/apis/v1alpha5"
	"github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/pkg/utils/test"
	"github.com/stretchr/testify/mock"
	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/version"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/karpenter/pkg/apis/v1beta1"
)

func TestGarbageCollectWorkspace(t *testing.T) {
	testcases := map[string]struct {
		callMocks     func(c *test.MockClient)
		k8sVersion    *version.Info
		expectedError error
	}{
		"Fails to delete workspace because associated machines cannot be retrieved": {
			callMocks: func(c *test.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)

				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1alpha5.MachineList{}), mock.Anything).Return(errors.New("failed to list machines"))
			},
			k8sVersion:    &version.Info{Major: "27"},
			expectedError: errors.New("failed to list machines"),
		},
		"Fails to delete workspace because associated machines cannot be deleted": {
			callMocks: func(c *test.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)

				machineList := test.MockMachineList
				relevantMap := c.CreateMapWithType(machineList)
				//insert machine objects into the map
				for _, obj := range test.MockMachineList.Items {
					m := obj
					objKey := client.ObjectKeyFromObject(&m)

					relevantMap[objKey] = &m
				}
				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1alpha5.MachineList{}), mock.Anything).Return(nil)
				c.On("Delete", mock.IsType(context.Background()), mock.IsType(&v1alpha5.Machine{}), mock.Anything).Return(errors.New("failed to delete machine"))

			},
			expectedError: errors.New("failed to delete machine"),
		},
		"Fails to delete workspace because associated nodeClaims cannot be retrieved": {
			callMocks: func(c *test.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)

				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1alpha5.MachineList{}), mock.Anything).Return(nil)
				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaimList{}), mock.Anything).Return(errors.New("failed to list nodeClaims"))
			},
			k8sVersion:    &version.Info{Major: "29"},
			expectedError: errors.New("failed to list nodeClaims"),
		},
		"Fails to delete workspace because associated nodeClaims cannot be deleted": {
			callMocks: func(c *test.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)

				nodeClaimList := test.MockNodeClaimList
				relevantMap := c.CreateMapWithType(nodeClaimList)
				//insert nodeClaim objects into the map
				for _, obj := range nodeClaimList.Items {
					m := obj
					objKey := client.ObjectKeyFromObject(&m)

					relevantMap[objKey] = &m
				}
				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1alpha5.MachineList{}), mock.Anything).Return(nil)

				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaimList{}), mock.Anything).Return(nil)
				c.On("Delete", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(errors.New("failed to delete nodeClaim"))

			},
			k8sVersion:    &version.Info{Major: "29"},
			expectedError: errors.New("failed to delete nodeClaim"),
		},
		"Delete workspace with associated machine objects because finalizer cannot be removed from workspace": {
			callMocks: func(c *test.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(errors.New("failed to update workspace"))

				machineList := test.MockMachineList
				relevantMap := c.CreateMapWithType(machineList)
				//insert machine objects into the map
				for _, obj := range machineList.Items {
					m := obj
					objKey := client.ObjectKeyFromObject(&m)

					relevantMap[objKey] = &m
				}
				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1alpha5.MachineList{}), mock.Anything).Return(nil)
				c.On("Delete", mock.IsType(context.Background()), mock.IsType(&v1alpha5.Machine{}), mock.Anything).Return(nil)
			},
			k8sVersion:    &version.Info{Major: "27"},
			expectedError: errors.New("failed to update workspace"),
		},
		"Successfully deletes workspace with associated machine objects and removes finalizer associated with workspace": {
			callMocks: func(c *test.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)

				machineList := test.MockMachineList
				relevantMap := c.CreateMapWithType(machineList)
				//insert machine objects into the map
				for _, obj := range test.MockMachineList.Items {
					m := obj
					objKey := client.ObjectKeyFromObject(&m)

					relevantMap[objKey] = &m
				}
				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1alpha5.MachineList{}), mock.Anything).Return(nil)
				c.On("Delete", mock.IsType(context.Background()), mock.IsType(&v1alpha5.Machine{}), mock.Anything).Return(nil)
			},
			k8sVersion:    &version.Info{Major: "27"},
			expectedError: nil,
		},
		"Delete workspace with associated nodeClaim objects because finalizer cannot be removed from workspace": {
			callMocks: func(c *test.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(errors.New("failed to update workspace"))

				nodeClaimList := test.MockNodeClaimList
				relevantMap := c.CreateMapWithType(nodeClaimList)
				//insert nodeClaim objects into the map
				for _, obj := range nodeClaimList.Items {
					m := obj
					objKey := client.ObjectKeyFromObject(&m)

					relevantMap[objKey] = &m
				}
				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1alpha5.MachineList{}), mock.Anything).Return(nil)

				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaimList{}), mock.Anything).Return(nil)
				c.On("Delete", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
			},
			k8sVersion:    &version.Info{Major: "29"},
			expectedError: errors.New("failed to update workspace"),
		},
		"Successfully deletes workspace with associated nodeClaim objects and removes finalizer associated with workspace": {
			callMocks: func(c *test.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)

				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaimList{}), mock.Anything).Return(nil)

				nodeClaimList := test.MockNodeClaimList
				relevantMap := c.CreateMapWithType(nodeClaimList)
				//insert nodeClaim objects into the map
				for _, obj := range nodeClaimList.Items {
					m := obj
					objKey := client.ObjectKeyFromObject(&m)

					relevantMap[objKey] = &m
				}
				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1alpha5.MachineList{}), mock.Anything).Return(nil)

				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaimList{}), mock.Anything).Return(nil)
				c.On("Delete", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
			},
			k8sVersion:    &version.Info{Major: "29"},
			expectedError: nil,
		},
		"Delete workspace with machine and nodeClaim objects because finalizer cannot be removed from workspace": {
			callMocks: func(c *test.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(errors.New("failed to update workspace"))

				machineList := test.MockMachineList
				relevantMachinesMap := c.CreateMapWithType(machineList)
				//insert machine objects into the map
				for _, obj := range machineList.Items {
					m := obj
					objKey := client.ObjectKeyFromObject(&m)

					relevantMachinesMap[objKey] = &m
				}
				nodeClaimList := test.MockNodeClaimList
				relevantNodeClaimsMap := c.CreateMapWithType(nodeClaimList)
				//insert nodeClaim objects into the map
				for _, obj := range nodeClaimList.Items {
					m := obj
					objKey := client.ObjectKeyFromObject(&m)

					relevantNodeClaimsMap[objKey] = &m
				}
				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1alpha5.MachineList{}), mock.Anything).Return(nil)
				c.On("Delete", mock.IsType(context.Background()), mock.IsType(&v1alpha5.Machine{}), mock.Anything).Return(nil)
				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaimList{}), mock.Anything).Return(nil)
				c.On("Delete", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
			},
			k8sVersion:    &version.Info{Major: "29"},
			expectedError: errors.New("failed to update workspace"),
		},
		"Successfully deletes workspace with machine and nodeClaim objects and removes finalizer associated with workspace": {
			callMocks: func(c *test.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.StatusMock.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)
				c.On("Update", mock.IsType(context.Background()), mock.IsType(&v1alpha1.Workspace{}), mock.Anything).Return(nil)

				machineList := test.MockMachineList
				relevantMachinesMap := c.CreateMapWithType(machineList)
				//insert machine objects into the map
				for _, obj := range machineList.Items {
					m := obj
					objKey := client.ObjectKeyFromObject(&m)

					relevantMachinesMap[objKey] = &m
				}
				nodeClaimList := test.MockNodeClaimList
				relevantNodeClaimsMap := c.CreateMapWithType(nodeClaimList)
				//insert nodeClaim objects into the map
				for _, obj := range nodeClaimList.Items {
					m := obj
					objKey := client.ObjectKeyFromObject(&m)

					relevantNodeClaimsMap[objKey] = &m
				}
				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1alpha5.MachineList{}), mock.Anything).Return(nil)
				c.On("Delete", mock.IsType(context.Background()), mock.IsType(&v1alpha5.Machine{}), mock.Anything).Return(nil)
				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaimList{}), mock.Anything).Return(nil)
				c.On("Delete", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
			},
			k8sVersion:    &version.Info{Major: "29"},
			expectedError: nil,
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			mockClient := test.NewClient()
			tc.callMocks(mockClient)

			reconciler := &WorkspaceReconciler{
				Client:            mockClient,
				Scheme:            test.NewTestScheme(),
				KubernetesVersion: tc.k8sVersion,
			}
			ctx := context.Background()

			_, err := reconciler.garbageCollectWorkspace(ctx, test.MockWorkspaceDistributedModel)
			if tc.expectedError == nil {
				assert.Check(t, err == nil, "Not expected to return error")
			} else {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			}
		})
	}
}

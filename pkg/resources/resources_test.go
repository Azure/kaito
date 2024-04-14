// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package resources

import (
	"context"
	"errors"
	"github.com/azure/kaito/pkg/utils/test"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	goassert "gotest.tools/assert"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func int32Ptr(i int32) *int32 {
	return &i
}

func TestCheckResourceStatus(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1.AddToScheme(scheme)
	t.Run("Should return nil for ready Deployment", func(t *testing.T) {
		// Create a deployment object for testing
		dep := &appsv1.Deployment{
			Status: appsv1.DeploymentStatus{
				ReadyReplicas: 3,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(3),
			},
		}

		cl := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(dep).Build()
		err := CheckResourceStatus(dep, cl, 2*time.Second)
		assert.Nil(t, err)
	})

	t.Run("Should return timeout error for non-ready Deployment", func(t *testing.T) {
		dep := &appsv1.Deployment{
			Status: appsv1.DeploymentStatus{
				ReadyReplicas: 0,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(1),
			},
		}

		cl := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(dep).Build()
		err := CheckResourceStatus(dep, cl, 1*time.Millisecond)
		assert.Error(t, err)
	})

	t.Run("Should return nil for ready StatefulSet", func(t *testing.T) {
		ss := &appsv1.StatefulSet{
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 3,
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: int32Ptr(3),
			},
		}

		cl := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ss).Build()
		err := CheckResourceStatus(ss, cl, 2*time.Second)
		assert.Nil(t, err)
	})

	t.Run("Should return timeout error for non-ready StatefulSet", func(t *testing.T) {
		ss := &appsv1.StatefulSet{
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 0,
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: int32Ptr(1),
			},
		}

		cl := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ss).Build()
		err := CheckResourceStatus(ss, cl, 1*time.Millisecond)
		assert.Error(t, err)
	})

	t.Run("Should return error for mocked client Get error", func(t *testing.T) {
		// This deployment won't be added to the fake client
		dep := &appsv1.Deployment{
			Status: appsv1.DeploymentStatus{
				ReadyReplicas: 0,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(1),
			},
		}

		// Create the fake client without adding the dep object
		cl := fake.NewClientBuilder().WithScheme(scheme).Build()

		err := CheckResourceStatus(dep, cl, 2*time.Second)
		assert.Error(t, err)
	})

	t.Run("Should return error for unsupported resource type", func(t *testing.T) {
		unsupportedResource := &appsv1.DaemonSet{}
		cl := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(unsupportedResource).Build()
		err := CheckResourceStatus(unsupportedResource, cl, 2*time.Second)
		assert.Error(t, err)
		assert.Equal(t, "unsupported resource type", err.Error())
	})
}

func TestCreateResource(t *testing.T) {
	testcases := map[string]struct {
		callMocks        func(c *test.MockClient)
		expectedResource client.Object
		expectedError    error
	}{
		"Resource creation fails with Deployment object": {
			callMocks: func(c *test.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&v1.Deployment{}), mock.Anything).Return(errors.New("Failed to create resource"))
			},
			expectedResource: &v1.Deployment{},
			expectedError:    errors.New("Failed to create resource"),
		},
		"Resource creation succeeds with Statefulset object": {
			callMocks: func(c *test.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&v1.StatefulSet{}), mock.Anything).Return(nil)
			},
			expectedResource: &v1.StatefulSet{},
			expectedError:    nil,
		},
		"Resource creation succeeds with Service object": {
			callMocks: func(c *test.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&corev1.Service{}), mock.Anything).Return(nil)
			},
			expectedResource: &corev1.Service{},
			expectedError:    nil,
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			mockClient := test.NewClient()
			tc.callMocks(mockClient)

			err := CreateResource(context.Background(), tc.expectedResource, mockClient)
			if tc.expectedError == nil {
				goassert.Check(t, err == nil, "Not expected to return error")
			} else {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			}
		})
	}
}

func TestGetResource(t *testing.T) {
	testcases := map[string]struct {
		callMocks     func(c *test.MockClient)
		expectedError error
	}{
		"GetResource fails": {
			callMocks: func(c *test.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&corev1.Node{}), mock.Anything).Return(errors.New("Failed to get resource"))
			},
			expectedError: errors.New("Failed to get resource"),
		},
		"Resource creation succeeds with Statefulset object": {
			callMocks: func(c *test.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&corev1.Node{}), mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			mockClient := test.NewClient()
			tc.callMocks(mockClient)

			err := GetResource(context.Background(), "fakeName", "fakeNamespace", mockClient, &corev1.Node{})
			if tc.expectedError == nil {
				goassert.Check(t, err == nil, "Not expected to return error")
			} else {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			}
		})
	}
}

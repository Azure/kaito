package inference

import (
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/stretchr/testify/assert"
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
		err := checkResourceStatus(dep, cl, 2*time.Second)
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
		err := checkResourceStatus(dep, cl, 1*time.Millisecond)
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
		err := checkResourceStatus(ss, cl, 2*time.Second)
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
		err := checkResourceStatus(ss, cl, 1*time.Millisecond)
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

		err := checkResourceStatus(dep, cl, 2*time.Second)
		assert.Error(t, err)
	})

	t.Run("Should return error for unsupported resource type", func(t *testing.T) {
		unsupportedResource := &appsv1.DaemonSet{}
		cl := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(unsupportedResource).Build()
		err := checkResourceStatus(unsupportedResource, cl, 2*time.Second)
		assert.Error(t, err)
		assert.Equal(t, "unsupported resource type", err.Error())
	})
}

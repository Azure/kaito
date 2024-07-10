// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package e2e

import (
	"time"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/test/e2e/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var DefaultStrength = "1.0"

var imageName = "e2e-adapter"
var fullImageName = utils.GetEnv("ADAPTER_REGISTRY") + "/" + imageName + ":0.0.1"

var validAdapters = []kaitov1alpha1.AdapterSpec{
	{
		Source: &kaitov1alpha1.DataSource{
			Name:  imageName,
			Image: fullImageName,
		},
		Strength: &DefaultStrength,
	},
}

var expectedInitContainers = []corev1.Container{
	{
		Name:  imageName,
		Image: fullImageName,
	},
}

func validateAdapters(workspaceObj *kaitov1alpha1.Workspace, expectedInitContainers []corev1.Container) {
	By("Checking the Adapters", func() {
		Eventually(func() bool {
			var err error
			var initContainers []corev1.Container

			dep := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      workspaceObj.Name,
					Namespace: workspaceObj.Namespace,
				},
			}
			err = TestingCluster.KubeClient.Get(ctx, client.ObjectKey{
				Namespace: workspaceObj.Namespace,
				Name:      workspaceObj.Name,
			}, dep)
			initContainers = dep.Spec.Template.Spec.InitContainers

			if err != nil {
				GinkgoWriter.Printf("Error fetching resource: %v\n", err)
				return false
			}

			if len(initContainers) != len(expectedInitContainers) {
				return false
			}
			initContainer, expectedInitContainer := initContainers[0], expectedInitContainers[0]

			// GinkgoWriter.Printf("Resource '%s' not ready. Ready replicas: %d\n", workspaceObj.Name, readyReplicas)
			return initContainer.Image == expectedInitContainer.Image && initContainer.Name == expectedInitContainer.Name
		}, 20*time.Minute, utils.PollInterval).Should(BeTrue(), "Failed to wait for adapter resource to be ready")
	})
}

var _ = Describe("Workspace Preset", func() {
	BeforeEach(func() {
		loadTestEnvVars()

		loadModelVersions()
	})

	It("should create a falcon workspace with adapter", func() {
		numOfNode := 1
		workspaceObj := createCustomWorkspaceWithAdapter(numOfNode)

		defer cleanupResources(workspaceObj)
		time.Sleep(30 * time.Second)

		validateMachineCreation(workspaceObj, numOfNode)
		validateResourceStatus(workspaceObj)

		time.Sleep(30 * time.Second)

		validateAssociatedService(workspaceObj)

		validateInferenceResource(workspaceObj, int32(numOfNode), false)

		validateWorkspaceReadiness(workspaceObj)

		validateAdapters(workspaceObj, expectedInitContainers)
	})

})

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package e2e

import (
	"fmt"
	"strings"
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
var fullImageName = utils.GetEnv("E2E_ACR_REGISTRY") + "/" + imageName + ":0.0.1"

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

func validateInitContainers(workspaceObj *kaitov1alpha1.Workspace, expectedInitContainers []corev1.Container) {
	By("Checking the InitContainers", func() {
		Eventually(func() bool {
			var err error
			var initContainers []corev1.Container

			dep := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      workspaceObj.Name,
					Namespace: workspaceObj.Namespace,
				},
			}
			err = utils.TestingCluster.KubeClient.Get(ctx, client.ObjectKey{
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
		}, 20*time.Minute, utils.PollInterval).Should(BeTrue(), "Failed to wait for initContainers to be ready")
	})
}

func validateAdapterAdded(workspaceObj *kaitov1alpha1.Workspace, deploymentName string, adapterName string) {
	By("Checking the Adapters", func() {
		Eventually(func() bool {
			coreClient, err := utils.GetK8sConfig()
			if err != nil {
				GinkgoWriter.Printf("Failed to create core client: %v\n", err)
				return false
			}

			namespace := workspaceObj.Namespace
			podName, err := utils.GetPodNameForDeployment(coreClient, namespace, deploymentName)
			if err != nil {
				GinkgoWriter.Printf("Failed to get pod name for deployment %s: %v\n", deploymentName, err)
				return false
			}

			logs, err := utils.GetPodLogs(coreClient, namespace, podName, "")
			if err != nil {
				GinkgoWriter.Printf("Failed to get logs from pod %s: %v\n", podName, err)
				return false
			}

			searchStringAdapter := fmt.Sprintf("Adapter added: %s", adapterName)
			searchStringModelSuccess := "Model loaded successfully"

			return strings.Contains(logs, searchStringAdapter) && strings.Contains(logs, searchStringModelSuccess)
		}, 20*time.Minute, utils.PollInterval).Should(BeTrue(), "Failed to wait for adapter resource to be ready")
	})
}

var _ = Describe("Workspace Preset", func() {
	BeforeEach(func() {
		loadTestEnvVars()

		loadModelVersions()
	})

	AfterEach(func() {
		if CurrentSpecReport().Failed() {
			utils.PrintPodLogsOnFailure(namespaceName, "")     // The Preset Pod
			utils.PrintPodLogsOnFailure("kaito-workspace", "") // The Kaito Workspace Pod
			utils.PrintPodLogsOnFailure("gpu-provisioner", "") // The gpu-provisioner Pod
			Fail("Fail threshold reached")
		}
	})

	It("should create a falcon workspace with adapter", func() {
		numOfNode := 1
		workspaceObj := createCustomWorkspaceWithAdapter(numOfNode)

		defer cleanupResources(workspaceObj)
		time.Sleep(30 * time.Second)

		if nodeProvisionerName == "azkarpenter" {
			utils.ValidateNodeClaimCreation(ctx, workspaceObj, numOfNode)
		} else {
			utils.ValidateMachineCreation(ctx, workspaceObj, numOfNode)
		}
		validateResourceStatus(workspaceObj)

		time.Sleep(30 * time.Second)

		validateAssociatedService(workspaceObj)

		validateInferenceResource(workspaceObj, int32(numOfNode), false)

		validateWorkspaceReadiness(workspaceObj)

		validateInitContainers(workspaceObj, expectedInitContainers)
		validateAdapterAdded(workspaceObj, workspaceObj.Name, imageName)
	})

})

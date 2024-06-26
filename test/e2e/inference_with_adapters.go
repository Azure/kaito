// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package e2e

import (
	"fmt"
	"os"
	"strconv"
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

var emptyAdapters = make([]kaitov1alpha1.AdapterSpec, 0)
var DefaultStrength = "1.0"

var validAdapters = []kaitov1alpha1.AdapterSpec{
	{
		Source: &kaitov1alpha1.DataSource{
			Name:  "falcon-7b-adapter",
			Image: "aimodelsregistrytest.azurecr.io/adapter-falcon-7b-dolly-oai-busybox:0.0.2",
		},
		Strength: &DefaultStrength,
	},
}

var expectedInitContainers = []corev1.Container{
	{
		Name:  "falcon-7b-adapter",
		Image: "aimodelsregistrytest.azurecr.io/adapter-falcon-7b-dolly-oai-busybox:0.0.2",
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
		var err error
		runLlama13B, err = strconv.ParseBool(os.Getenv("RUN_LLAMA_13B"))
		if err != nil {
			// Handle error or set a default value
			fmt.Print("Error: RUN_LLAMA_13B ENV Variable not set")
			runLlama13B = false
		}

		aiModelsRegistry = utils.GetEnv("AI_MODELS_REGISTRY")
		aiModelsRegistrySecret = utils.GetEnv("AI_MODELS_REGISTRY_SECRET")
		supportedModelsYamlPath = utils.GetEnv("SUPPORTED_MODELS_YAML_PATH")

		// Load stable model versions
		configs, err := utils.GetModelConfigInfo(supportedModelsYamlPath)
		if err != nil {
			fmt.Printf("Failed to load model configs: %v\n", err)
			os.Exit(1)
		}

		modelInfo, err = utils.ExtractModelVersion(configs)
		if err != nil {
			fmt.Printf("Failed to extract stable model versions: %v\n", err)
			os.Exit(1)
		}
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

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package e2e

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/aws/karpenter-core/pkg/apis/v1alpha5"
	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/test/e2e/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func createFalconWorkspaceWithPresetPublicMode(numOfNode int) *kaitov1alpha1.Workspace {
	workspaceObj := &kaitov1alpha1.Workspace{}
	By("Creating a workspace CR with Falcon 7B preset public mode", func() {
		uniqueID := fmt.Sprint("preset-", rand.Intn(1000))
		workspaceObj = utils.GenerateWorkspaceManifest(uniqueID, namespaceName, "", numOfNode, "Standard_NC12s_v3",
			&metav1.LabelSelector{
				MatchLabels: map[string]string{"kaito-workspace": "public-preset-e2e-test"},
			}, nil, kaitov1alpha1.PresetFalcon7BModel, kaitov1alpha1.ModelImageAccessModePublic, nil, nil)

		createAndValidateWorkspace(workspaceObj)
	})
	return workspaceObj
}

func createLlama7BWorkspaceWithPresetPrivateMode(registry, registrySecret, imageVersion string, numOfNode int) *kaitov1alpha1.Workspace {
	workspaceObj := &kaitov1alpha1.Workspace{}
	By("Creating a workspace CR with Llama 7B Chat preset private mode", func() {
		uniqueID := fmt.Sprint("preset-", rand.Intn(1000))
		workspaceObj = utils.GenerateWorkspaceManifest(uniqueID, namespaceName, fmt.Sprintf("%s/%s:%s", registry, kaitov1alpha1.PresetLlama2AChat, imageVersion),
			numOfNode, "Standard_NC12s_v3", &metav1.LabelSelector{
				MatchLabels: map[string]string{"kaito-workspace": "private-preset-e2e-test"},
			}, nil, kaitov1alpha1.PresetLlama2AChat, kaitov1alpha1.ModelImageAccessModePrivate, []string{registrySecret}, nil)

		createAndValidateWorkspace(workspaceObj)
	})
	return workspaceObj
}

func createLlama13BWorkspaceWithPresetPrivateMode(registry, registrySecret, imageVersion string, numOfNode int) *kaitov1alpha1.Workspace {
	workspaceObj := &kaitov1alpha1.Workspace{}
	By("Creating a workspace CR with Llama 13B Chat preset private mode", func() {
		uniqueID := fmt.Sprint("preset-", rand.Intn(1000))
		workspaceObj = utils.GenerateWorkspaceManifest(uniqueID, namespaceName, fmt.Sprintf("%s/%s:%s", registry, kaitov1alpha1.PresetLlama2BChat, imageVersion),
			numOfNode, "Standard_NC12s_v3", &metav1.LabelSelector{
				MatchLabels: map[string]string{"kaito-workspace": "private-preset-e2e-test"},
			}, nil, kaitov1alpha1.PresetLlama2BChat, kaitov1alpha1.ModelImageAccessModePrivate, []string{registrySecret}, nil)

		createAndValidateWorkspace(workspaceObj)
	})
	return workspaceObj
}

func createCustomWorkspaceWithPresetCustomMode(imageName string, numOfNode int) *kaitov1alpha1.Workspace {
	workspaceObj := &kaitov1alpha1.Workspace{}
	By("Creating a workspace CR with custom workspace mode", func() {
		uniqueID := fmt.Sprint("preset-", rand.Intn(1000))
		workspaceObj = utils.GenerateWorkspaceManifest(uniqueID, namespaceName, "",
			numOfNode, "Standard_D4s_v3", &metav1.LabelSelector{
				MatchLabels: map[string]string{"kaito-workspace": "private-preset-e2e-test"},
			}, nil, "", utils.InferenceModeCustomTemplate, nil, utils.GeneratePodTemplate(uniqueID, namespaceName, imageName, nil))

		createAndValidateWorkspace(workspaceObj)
	})
	return workspaceObj
}

func createAndValidateWorkspace(workspaceObj *kaitov1alpha1.Workspace) {
	By("Creating workspace", func() {
		Eventually(func() error {
			return TestingCluster.KubeClient.Create(ctx, workspaceObj, &client.CreateOptions{})
		}, utils.PollTimeout, utils.PollInterval).
			Should(Succeed(), "Failed to create workspace %s", workspaceObj.Name)

		By("Validating workspace creation", func() {
			err := TestingCluster.KubeClient.Get(ctx, client.ObjectKey{
				Namespace: workspaceObj.Namespace,
				Name:      workspaceObj.Name,
			}, workspaceObj, &client.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
}

func getAllValidMachines(workspaceObj *kaitov1alpha1.Workspace) (*v1alpha5.MachineList, error) {
	machineList := &v1alpha5.MachineList{}
	ls := labels.Set{
		kaitov1alpha1.LabelWorkspaceName:      workspaceObj.Name,
		kaitov1alpha1.LabelWorkspaceNamespace: workspaceObj.Namespace,
	}

	err := TestingCluster.KubeClient.List(ctx, machineList, &client.MatchingLabelsSelector{Selector: ls.AsSelector()})
	if err != nil {
		return nil, err
	}
	return machineList, nil
}

// Logic to validate machine creation
func validateMachineCreation(workspaceObj *kaitov1alpha1.Workspace, expectedCount int) {
	By("Checking machine created by the workspace CR", func() {
		Eventually(func() bool {
			machineList, err := getAllValidMachines(workspaceObj)
			if err != nil {
				fmt.Printf("Failed to get all valid machines: %v", err)
				return false
			}

			if len(machineList.Items) != expectedCount {
				return false
			}

			for _, machine := range machineList.Items {
				_, conditionFound := lo.Find(machine.GetConditions(), func(condition apis.Condition) bool {
					return condition.Type == apis.ConditionReady && condition.Status == v1.ConditionTrue
				})
				if !conditionFound {
					return false
				}
			}
			return true
		}, 20*time.Minute, utils.PollInterval).Should(BeTrue(), "Failed to wait for machine to be ready")
	})
}

// Logic to validate resource status
func validateResourceStatus(workspaceObj *kaitov1alpha1.Workspace) {
	By("Checking the resource status", func() {
		Eventually(func() bool {
			err := TestingCluster.KubeClient.Get(ctx, client.ObjectKey{
				Namespace: workspaceObj.Namespace,
				Name:      workspaceObj.Name,
			}, workspaceObj, &client.GetOptions{})

			if err != nil {
				return false
			}

			_, conditionFound := lo.Find(workspaceObj.Status.Conditions, func(condition metav1.Condition) bool {
				return condition.Type == string(kaitov1alpha1.WorkspaceConditionTypeResourceStatus) &&
					condition.Status == metav1.ConditionTrue
			})
			return conditionFound
		}, 10*time.Minute, utils.PollInterval).Should(BeTrue(), "Failed to wait for resource status to be ready")
	})
}

func validateAssociatedService(workspaceObj *kaitov1alpha1.Workspace) {
	serviceName := workspaceObj.Name
	serviceNamespace := workspaceObj.Namespace

	By(fmt.Sprintf("Checking for service %s in namespace %s", serviceName, serviceNamespace), func() {
		service := &v1.Service{}

		Eventually(func() bool {
			err := TestingCluster.KubeClient.Get(ctx, client.ObjectKey{
				Namespace: serviceNamespace,
				Name:      serviceName,
			}, service)

			if err != nil {
				if errors.IsNotFound(err) {
					GinkgoWriter.Printf("Service %s not found in namespace %s\n", serviceName, serviceNamespace)
				} else {
					GinkgoWriter.Printf("Error fetching service %s in namespace %s: %v\n", serviceName, serviceNamespace, err)
				}
				return false
			}

			GinkgoWriter.Printf("Found service: %s in namespace %s\n", serviceName, serviceNamespace)
			return true
		}, 10*time.Minute, utils.PollInterval).Should(BeTrue(), "Failed to wait for service to be created")
	})
}

// Logic to validate inference deployment
func validateInferenceResource(workspaceObj *kaitov1alpha1.Workspace, expectedReplicas int32, isStatefulSet bool) {
	By("Checking the inference resource", func() {
		Eventually(func() bool {
			var err error
			var readyReplicas int32

			if isStatefulSet {
				sts := &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      workspaceObj.Name,
						Namespace: workspaceObj.Namespace,
					},
				}
				err = TestingCluster.KubeClient.Get(ctx, client.ObjectKey{
					Namespace: workspaceObj.Namespace,
					Name:      workspaceObj.Name,
				}, sts)
				readyReplicas = sts.Status.ReadyReplicas

			} else {
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
				readyReplicas = dep.Status.ReadyReplicas
			}

			if err != nil {
				GinkgoWriter.Printf("Error fetching resource: %v\n", err)
				return false
			}

			if readyReplicas == expectedReplicas {
				return true
			}

			// GinkgoWriter.Printf("Resource '%s' not ready. Ready replicas: %d\n", workspaceObj.Name, readyReplicas)
			return false
		}, 20*time.Minute, utils.PollInterval).Should(BeTrue(), "Failed to wait for inference resource to be ready")
	})
}

// Logic to validate workspace readiness
func validateWorkspaceReadiness(workspaceObj *kaitov1alpha1.Workspace) {
	By("Checking the workspace status is ready", func() {
		Eventually(func() bool {
			err := TestingCluster.KubeClient.Get(ctx, client.ObjectKey{
				Namespace: workspaceObj.Namespace,
				Name:      workspaceObj.Name,
			}, workspaceObj, &client.GetOptions{})

			if err != nil {
				return false
			}

			_, conditionFound := lo.Find(workspaceObj.Status.Conditions, func(condition metav1.Condition) bool {
				return condition.Type == string(kaitov1alpha1.WorkspaceConditionTypeReady) &&
					condition.Status == metav1.ConditionTrue
			})
			return conditionFound
		}, 10*time.Minute, utils.PollInterval).Should(BeTrue(), "Failed to wait for workspace to be ready")
	})
}

func cleanupResources(workspaceObj *kaitov1alpha1.Workspace) {
	By("Cleaning up resources", func() {
		// delete workspace
		err := deleteWorkspace(workspaceObj)
		Expect(err).NotTo(HaveOccurred(), "Failed to delete workspace")
	})
}

func deleteWorkspace(workspaceObj *kaitov1alpha1.Workspace) error {
	By("Deleting workspace", func() {
		Eventually(func() error {
			// Check if the workspace exists
			err := TestingCluster.KubeClient.Get(ctx, client.ObjectKey{
				Namespace: workspaceObj.Namespace,
				Name:      workspaceObj.Name,
			}, workspaceObj)

			if errors.IsNotFound(err) {
				GinkgoWriter.Printf("Workspace %s does not exist, no need to delete\n", workspaceObj.Name)
				return nil
			}
			if err != nil {
				return fmt.Errorf("error checking if workspace %s exists: %v", workspaceObj.Name, err)
			}

			err = TestingCluster.KubeClient.Delete(ctx, workspaceObj, &client.DeleteOptions{})
			if err != nil {
				return fmt.Errorf("failed to delete workspace %s: %v", workspaceObj.Name, err)
			}
			return nil
		}, utils.PollTimeout, utils.PollInterval).Should(Succeed(), "Failed to delete workspace")
	})

	return nil
}

var runLlama13B bool
var aiModelsRegistry string
var aiModelsRegistrySecret string
var aiModelsImageVersion string

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
		aiModelsImageVersion = utils.GetEnv("AI_MODELS_IMAGE_VERSION")
	})

	It("should create a workspace with preset public mode successfully", func() {
		numOfNode := 1
		workspaceObj := createFalconWorkspaceWithPresetPublicMode(numOfNode)

		defer cleanupResources(workspaceObj)
		time.Sleep(30 * time.Second)

		validateMachineCreation(workspaceObj, numOfNode)
		validateResourceStatus(workspaceObj)

		time.Sleep(30 * time.Second)

		validateAssociatedService(workspaceObj)

		validateInferenceResource(workspaceObj, int32(numOfNode), false)

		validateWorkspaceReadiness(workspaceObj)
	})

	It("should create a llama 7b workspace with preset private mode successfully", func() {
		numOfNode := 1
		workspaceObj := createLlama7BWorkspaceWithPresetPrivateMode(aiModelsRegistry, aiModelsRegistrySecret, aiModelsImageVersion, numOfNode)

		defer cleanupResources(workspaceObj)
		time.Sleep(30 * time.Second)

		validateMachineCreation(workspaceObj, numOfNode)
		validateResourceStatus(workspaceObj)

		time.Sleep(30 * time.Second)

		validateAssociatedService(workspaceObj)

		validateInferenceResource(workspaceObj, int32(numOfNode), true)

		validateWorkspaceReadiness(workspaceObj)
	})

	It("should create a llama 13b workspace with preset private mode successfully", func() {
		if !runLlama13B {
			Skip("Skipping llama 13b workspace test")
		}
		numOfNode := 2
		workspaceObj := createLlama13BWorkspaceWithPresetPrivateMode(aiModelsRegistry, aiModelsRegistrySecret, aiModelsImageVersion, numOfNode)

		defer cleanupResources(workspaceObj)

		time.Sleep(30 * time.Second)
		validateMachineCreation(workspaceObj, numOfNode)
		validateResourceStatus(workspaceObj)

		time.Sleep(30 * time.Second)

		validateAssociatedService(workspaceObj)

		validateInferenceResource(workspaceObj, int32(numOfNode), true)

		validateWorkspaceReadiness(workspaceObj)
	})

	It("should create a custom template workspace successfully", func() {
		numOfNode := 1
		imageName := "nginx:latest"
		workspaceObj := createCustomWorkspaceWithPresetCustomMode(imageName, numOfNode)

		defer cleanupResources(workspaceObj)

		time.Sleep(30 * time.Second)
		validateMachineCreation(workspaceObj, numOfNode)
		validateResourceStatus(workspaceObj)

		time.Sleep(30 * time.Second)

		validateInferenceResource(workspaceObj, int32(numOfNode), false)

		validateWorkspaceReadiness(workspaceObj)
	})

})

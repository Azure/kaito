// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package e2e

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/aws/karpenter-core/pkg/apis/v1alpha5"
	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/test/e2e/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Workspace Preset", func() {

	It("should create a workspace with preset public mode successfully", func() {
		workspaceObj := &kaitov1alpha1.Workspace{}
		machineObj := v1alpha5.Machine{}
		By("Creating a workspace CR with Falcon 7B preset public mode", func() {
			workspaceObj = utils.GenerateWorkspaceManifest(fmt.Sprint("preset-", rand.Intn(1000)), namespaceName, "", 1, "Standard_NC12s_v3",
				&metav1.LabelSelector{
					MatchLabels: map[string]string{"kaito-workspace": "public-preset-e2e-test"},
				}, nil, kaitov1alpha1.PresetFalcon7BModel, kaitov1alpha1.ModelImageAccessModePublic, nil, nil)

			// Create workspace
			Eventually(func() error {
				return TestingCluster.KubeClient.Create(ctx, workspaceObj, &client.CreateOptions{})
			}, utils.PollTimeout, utils.PollInterval).
				Should(Succeed(), "Failed to create workspace %s", workspaceObj.Name)

			// Get workspace
			err := TestingCluster.KubeClient.Get(ctx, client.ObjectKey{
				Namespace: workspaceObj.Namespace,
				Name:      workspaceObj.Name,
			}, workspaceObj, &client.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
		})

		time.Sleep(30 * time.Second)

		By("Checking machine created by the workspace CR", func() {
			// get machine
			machineList := &v1alpha5.MachineList{}
			ls := labels.Set{
				kaitov1alpha1.LabelWorkspaceName:      workspaceObj.Name,
				kaitov1alpha1.LabelWorkspaceNamespace: workspaceObj.Namespace,
			}

			Eventually(func() bool {
				merr := TestingCluster.KubeClient.List(ctx, machineList, &client.MatchingLabelsSelector{Selector: ls.AsSelector()})
				Expect(merr).NotTo(HaveOccurred())
				Expect(len(machineList.Items)).To(Equal(1))

				machineObj = machineList.Items[0]

				_, conditionFound := lo.Find(machineObj.GetConditions(), func(condition apis.Condition) bool {
					return condition.Type == apis.ConditionReady &&
						condition.Status == v1.ConditionTrue
				})
				return conditionFound
			}, 20*time.Minute, utils.PollInterval).Should(BeTrue(), "Failed to wait for machine to be ready")
		})
		By("Checking the resource status", func() {
			Eventually(func() bool {
				err := TestingCluster.KubeClient.Get(ctx, client.ObjectKey{
					Namespace: workspaceObj.Namespace,
					Name:      workspaceObj.Name,
				}, workspaceObj, &client.GetOptions{})

				Expect(err).NotTo(HaveOccurred())

				_, conditionFound := lo.Find(workspaceObj.Status.Conditions, func(condition metav1.Condition) bool {
					return condition.Type == string(kaitov1alpha1.WorkspaceConditionTypeResourceStatus) &&
						condition.Status == metav1.ConditionTrue
				})
				return conditionFound
			}, 10*time.Minute, utils.PollInterval).Should(BeTrue(), "Failed to wait for resource status to be ready")
		})

		time.Sleep(30 * time.Second)

		By("Checking the inference deployment", func() {
			inferenceDep := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      workspaceObj.Name,
					Namespace: workspaceObj.Namespace,
				},
			}

			Eventually(func() bool {
				ierr := TestingCluster.KubeClient.Get(ctx, client.ObjectKey{
					Namespace: workspaceObj.Namespace,
					Name:      workspaceObj.Name,
				}, inferenceDep, &client.GetOptions{})

				Expect(ierr).NotTo(HaveOccurred())

				return inferenceDep.Status.ReadyReplicas == 1
			}, 20*time.Minute, utils.PollInterval).Should(BeTrue(), "Failed to wait for inference deployment to be ready")
		})

		By("Checking the workspace status is ready", func() {
			Eventually(func() bool {
				err := TestingCluster.KubeClient.Get(ctx, client.ObjectKey{
					Namespace: workspaceObj.Namespace,
					Name:      workspaceObj.Name,
				}, workspaceObj, &client.GetOptions{})

				Expect(err).NotTo(HaveOccurred())

				_, conditionFound := lo.Find(workspaceObj.Status.Conditions, func(condition metav1.Condition) bool {
					return condition.Type == string(kaitov1alpha1.WorkspaceConditionTypeReady) &&
						condition.Status == metav1.ConditionTrue
				})
				return conditionFound
			}, 10*time.Minute, utils.PollInterval).Should(BeTrue(), "Failed to wait for workspace to be ready")
		})

		// delete	workspace
		Eventually(func() error {
			return TestingCluster.KubeClient.Delete(ctx, workspaceObj, &client.DeleteOptions{})
		}, utils.PollTimeout, utils.PollInterval).Should(Succeed(), "Failed to delete workspace")
		// delete machine
		Eventually(func() error {
			return TestingCluster.KubeClient.Delete(ctx, &machineObj, &client.DeleteOptions{})
		}, utils.PollTimeout, utils.PollInterval).Should(Succeed(), "Failed to delete machine")

	})

})

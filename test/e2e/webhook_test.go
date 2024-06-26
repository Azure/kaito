// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package e2e

import (
	"fmt"
	"math/rand"
	"time"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/azure/kaito/test/e2e/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Workspace Validation Webhook", func() {
	It("should validate the workspace resource spec at creation ", func() {
		workspaceObj := utils.GenerateWorkspaceManifest(fmt.Sprint("webhook-", rand.Intn(1000)), namespaceName, "", 1, "Standard_Bad",
			&metav1.LabelSelector{
				MatchLabels: map[string]string{"kaito-workspace": "webhook-e2e-test"},
			}, nil, PresetFalcon7BModel, kaitov1alpha1.ModelImageAccessModePublic, nil, nil, emptyAdapters)

		By("Creating a workspace with invalid instancetype", func() {
			// Create workspace
			Eventually(func() error {
				return TestingCluster.KubeClient.Create(ctx, workspaceObj, &client.CreateOptions{})
			}, 20*time.Minute, utils.PollInterval).
				Should(HaveOccurred(), "Failed to create workspace %s", workspaceObj.Name)
		})
	})

	It("should validate the workspace inference spec at creation ", func() {
		workspaceObj := utils.GenerateWorkspaceManifest(fmt.Sprint("webhook-", rand.Intn(1000)), namespaceName, "", 1, "Standard_NC6",
			&metav1.LabelSelector{
				MatchLabels: map[string]string{"kaito-workspace": "webhook-e2e-test"},
			}, nil, "invalid-name", kaitov1alpha1.ModelImageAccessModePublic, nil, nil, emptyAdapters)

		By("Creating a workspace with invalid preset name", func() {
			// Create workspace
			Eventually(func() error {
				return TestingCluster.KubeClient.Create(ctx, workspaceObj, &client.CreateOptions{})
			}, utils.PollTimeout, utils.PollInterval).
				Should(HaveOccurred(), "Failed to create workspace %s", workspaceObj.Name)
		})
	})

	//TODO preset private mode
	//TODO custom template

	It("should validate the workspace resource spec at update ", func() {
		workspaceObj := utils.GenerateWorkspaceManifest(fmt.Sprint("webhook-", rand.Intn(1000)), namespaceName, "", 1, "Standard_NC12s_v3",
			&metav1.LabelSelector{
				MatchLabels: map[string]string{"kaito-workspace": "webhook-e2e-test"},
			}, nil, PresetFalcon7BModel, kaitov1alpha1.ModelImageAccessModePublic, nil, nil, emptyAdapters)

		By("Creating a valid workspace", func() {
			// Create workspace
			Eventually(func() error {
				return TestingCluster.KubeClient.Create(ctx, workspaceObj, &client.CreateOptions{})
			}, 20*time.Minute, utils.PollInterval).
				Should(Succeed(), "Failed to create workspace %s", workspaceObj.Name)
		})

		By("Updating the label selector", func() {
			updatedObj := workspaceObj
			updatedObj.Resource.LabelSelector = &metav1.LabelSelector{}
			// update workspace
			Eventually(func() error {
				return TestingCluster.KubeClient.Update(ctx, updatedObj, &client.UpdateOptions{})
			}, utils.PollTimeout, utils.PollInterval).
				Should(HaveOccurred(), "Failed to update workspace %s", updatedObj.Name)
		})

		By("Updating the InstanceType", func() {
			updatedObj := workspaceObj
			updatedObj.Resource.InstanceType = "Standard_NC12"
			// update workspace
			Eventually(func() error {
				return TestingCluster.KubeClient.Update(ctx, updatedObj, &client.UpdateOptions{})
			}, utils.PollTimeout, utils.PollInterval).
				Should(HaveOccurred(), "Failed to update workspace %s", updatedObj.Name)
		})

		//TODO custom template

		// delete	workspace
		Eventually(func() error {
			return TestingCluster.KubeClient.Delete(ctx, workspaceObj, &client.DeleteOptions{})
		}, utils.PollTimeout, utils.PollInterval).Should(Succeed(), "Failed to delete workspace")

	})

	It("should validate the workspace inference spec at update ", func() {
		workspaceObj := utils.GenerateWorkspaceManifest(fmt.Sprint("webhook-", rand.Intn(1000)), namespaceName, "", 1, "Standard_NC12s_v3",
			&metav1.LabelSelector{
				MatchLabels: map[string]string{"kaito-workspace": "webhook-e2e-test"},
			}, nil, PresetFalcon7BModel, kaitov1alpha1.ModelImageAccessModePublic, nil, nil, emptyAdapters)

		By("Creating a valid workspace", func() {
			// Create workspace
			Eventually(func() error {
				return TestingCluster.KubeClient.Create(ctx, workspaceObj, &client.CreateOptions{})
			}, 20*time.Minute, utils.PollInterval).
				Should(Succeed(), "Failed to create workspace %s", workspaceObj.Name)
		})

		By("Updating the preset spec", func() {
			updatedObj := workspaceObj
			updatedObj.Inference.Preset.Name = PresetFalcon40BModel
			// update workspace
			Eventually(func() error {
				return TestingCluster.KubeClient.Update(ctx, updatedObj, &client.UpdateOptions{})
			}, utils.PollTimeout, utils.PollInterval).
				Should(HaveOccurred(), "Failed to update workspace %s", updatedObj.Name)
		})

		//TODO custom template

		// delete	workspace
		Eventually(func() error {
			return TestingCluster.KubeClient.Delete(ctx, workspaceObj, &client.DeleteOptions{})
		}, utils.PollTimeout, utils.PollInterval).Should(Succeed(), "Failed to delete workspace")

	})

})

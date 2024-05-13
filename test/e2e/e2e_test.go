// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package e2e

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/azure/kaito/test/e2e/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ctx           = context.Background()
	namespaceName = fmt.Sprint(E2eNamespace, rand.Intn(100))
)

var _ = SynchronizedBeforeSuite(func() []byte {
	GetClusterClient(TestingCluster)
	gpuNamespace := os.Getenv("GPU_NAMESPACE")
	kaitoNamespace := os.Getenv("KAITO_NAMESPACE")

	//check gpu-provisioner deployment is up and running
	gpuProvisionerDeployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gpu-provisioner",
			Namespace: gpuNamespace,
		},
	}

	Eventually(func() error {
		return TestingCluster.KubeClient.Get(ctx, client.ObjectKey{
			Namespace: gpuProvisionerDeployment.Namespace,
			Name:      gpuProvisionerDeployment.Name,
		}, gpuProvisionerDeployment, &client.GetOptions{})
	}, utils.PollTimeout, utils.PollInterval).Should(Succeed(), "Failed to wait for	gpu-provisioner deployment")

	//check kaito-workspace deployment is up and running
	kaitoWorkspaceDeployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kaito-workspace",
			Namespace: kaitoNamespace,
		},
	}

	Eventually(func() error {
		return TestingCluster.KubeClient.Get(ctx, client.ObjectKey{
			Namespace: kaitoWorkspaceDeployment.Namespace,
			Name:      kaitoWorkspaceDeployment.Name,
		}, kaitoWorkspaceDeployment, &client.GetOptions{})
	}, utils.PollTimeout, utils.PollInterval).Should(Succeed(), "Failed to wait for	kaito-workspace deployment")

	// create testing namespace
	err := TestingCluster.KubeClient.Create(context.TODO(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	})
	Expect(err).NotTo(HaveOccurred())

	return nil
}, func(data []byte) {})

var _ = SynchronizedAfterSuite(func() {
	// delete testing namespace
	Eventually(func() error {
		return TestingCluster.KubeClient.Delete(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespaceName,
			},
		}, &client.DeleteOptions{})
	}, utils.PollTimeout, utils.PollInterval).Should(Succeed(), "Failed to delete namespace for e2e")

}, func() {})

func RunE2ETests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AI Toolchain Operator E2E Test Suite")
}

func TestE2E(t *testing.T) {
	RunE2ETests(t)
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

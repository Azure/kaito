// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package e2e

import (
	"testing"

	"github.com/azure/kaito/test/e2e/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var env common.Environment

func TestWorkspace(t *testing.T) {
	RegisterFailHandler(Fail)
	BeforeSuite(func() {
		env = *common.NewEnvironment(t)
	})
	RunSpecs(t, "Kaito Workspace")
}

var _ = BeforeEach(func() { env.BeforeEach() })
var _ = AfterEach(func() { env.Cleanup() })
var _ = AfterEach(func() { env.AfterEach() })

var _ = Describe("Kaito Workspace ", func() {
	It("should create one workspace successfully ", func() {
		env.ExpectCreated(&common.Workspace)
		env.EventuallyExpectCreatedWorkspaceCount("==", 1)
		env.EventuallyExpectWorkspaceReady(&common.Workspace)
	})

	It("should deploy other resources to the cluster", func() {
		//env.ExpectCreated(common.Machine)
		env.ExpectKaitoDeployment()
		env.ExpectKaitoReplicaSet()
		env.ExpectKaitoService()
	})
})

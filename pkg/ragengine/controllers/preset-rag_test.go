// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package controllers

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/kaito-project/kaito/pkg/utils/consts"
	"github.com/kaito-project/kaito/pkg/utils/test"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
)

func TestCreatePresetRAG(t *testing.T) {
	test.RegisterTestModel()

	testcases := map[string]struct {
		nodeCount      int
		callMocks      func(c *test.MockClient)
		expectedCmd    string
		expectedGPUReq string
		expectedImage  string
		expectedVolume string
	}{
		"test-rag-model": {
			nodeCount: 1,
			callMocks: func(c *test.MockClient) {
				c.On("Create", mock.IsType(context.TODO()), mock.IsType(&appsv1.Deployment{}), mock.Anything).Return(nil)
			},
			expectedCmd:   "/bin/sh -c python3 main.py",
			expectedImage: "mcr.microsoft.com/aks/kaito/kaito-rag-service:0.0.1",
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			os.Setenv("CLOUD_PROVIDER", consts.AzureCloudName)
			mockClient := test.NewClient()
			tc.callMocks(mockClient)

			ragEngineObj := test.MockRAGEngineWithPreset
			createdObject, _ := CreatePresetRAG(context.TODO(), ragEngineObj, "1", mockClient)

			workloadCmd := strings.Join((createdObject.(*appsv1.Deployment)).Spec.Template.Spec.Containers[0].Command, " ")

			if workloadCmd != tc.expectedCmd {
				t.Errorf("%s: main cmdline is not expected, got %s, expected %s", k, workloadCmd, tc.expectedCmd)
			}

			image := (createdObject.(*appsv1.Deployment)).Spec.Template.Spec.Containers[0].Image

			if image != tc.expectedImage {
				t.Errorf("%s: image is not expected, got %s, expected %s", k, image, tc.expectedImage)
			}
		})
	}
}

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package main

import (
	"os"
	"testing"

	"github.com/azure/kaito/pkg/utils/consts"
	"gotest.tools/assert"
)

func TestParseFeatureGates(t *testing.T) {
	tests := []struct {
		name          string
		featureGates  string
		expectedError bool
		expectedValue string
	}{
		{
			name:          "WithValidEnableFeatureGates",
			featureGates:  "karpenter=true",
			expectedError: false,
			expectedValue: "true",
		},
		{
			name:          "WithValidDisableFeatureGates",
			featureGates:  "karpenter=false",
			expectedError: false,
			expectedValue: "false",
		},
		{
			name:          "WithEmptyFeatureGates",
			featureGates:  "",
			expectedError: false,
		},
		{
			name:          "WithMultipleFeatureGates",
			featureGates:  "karpenter=true,feature2=false",
			expectedError: false,
			expectedValue: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make sure to unset the environment variable
			err := os.Unsetenv(consts.FeatureFlagEnableKarpenter)
			if err != nil {
				t.Error("Failed to unset environment variable")
				return
			}
			err = ParseFeatureGates(tt.featureGates)
			if (err != nil) != tt.expectedError {
				t.Errorf("ParseFeatureGates() error = %v, expectedError %v", err, tt.expectedError)
				return
			}

			value, _ := os.LookupEnv(consts.FeatureFlagEnableKarpenter)
			if tt.expectedError {
				assert.Check(t, err == nil, "Not expected to return error")
			} else {
				assert.Equal(t, tt.expectedValue, value)
			}
		})
	}
}

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package featuregates

import (
	"errors"

	"github.com/azure/kaito/pkg/utils/consts"
	cliflag "k8s.io/component-base/cli/flag"
)

var (
	// FeatureGates is a map that holds	the feature gates and their values for Kaito.
	FeatureGates = map[string]bool{
		consts.FeatureFlagKarpenter: false,
	}
)

// ParseAndValidateFeatureGates parses the feature gates flag and sets the environment variables for each feature.
func ParseAndValidateFeatureGates(featureGates string) error {
	gateMap := map[string]bool{}

	if err := cliflag.NewMapStringBool(&gateMap).Set(featureGates); err != nil {
		return err
	}
	if len(gateMap) == 0 {
		// no feature gates set
		return nil
	}
	if val, ok := gateMap[consts.FeatureFlagKarpenter]; ok {
		// set the environment variable to enable karpenter feature
		FeatureGates[consts.FeatureFlagKarpenter] = val
	} else {
		return errors.New("invalid feature gate")
	}
	// add more feature gates here
	return nil
}

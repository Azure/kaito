/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (

	//"github.com/aws/karpenter-core/pkg/test"
	"github.com/azure/kaito/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// var (
// 	MachineLabels = map[string]string{
// 		"karpenter.sh/provisioner-name": "default",
// 		"kaito.sh/workspace":            "none",
// 	}
// )

// var (
// 	Machine = test.Machine(v1alpha5.Machine{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:   "testmachine",
// 			Labels: MachineLabels,
// 		},
// 		Spec: v1alpha5.MachineSpec{
// 			MachineTemplateRef: &v1alpha5.MachineTemplateRef{
// 				Name: "test-machine",
// 			},
// 			Requirements: []v1.NodeSelectorRequirement{
// 				{
// 					Key:      v1.LabelInstanceTypeStable,
// 					Operator: v1.NodeSelectorOpIn,
// 					Values:   []string{"Standard_NC12s_v3"},
// 				},
// 				{
// 					Key:      "karpenter.sh/provisioner-name",
// 					Operator: v1.NodeSelectorOpIn,
// 					Values:   []string{"default"},
// 				},
// 			},
// 			Taints: []v1.Taint{
// 				{
// 					Key:    "sku",
// 					Value:  "gpu",
// 					Effect: v1.TaintEffectNoSchedule,
// 				},
// 			},
// 		},
// 	})
// )

var (
	Workspace = v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testWorkspace",
		},
		Resource: v1alpha1.ResourceSpec{
			InstanceType: "Standard_NC12s_v3",
		},
		Inference: v1alpha1.InferenceSpec{
			Preset: &v1alpha1.PresetSpec{
				PresetMeta: v1alpha1.PresetMeta{
					Name: "llama-2-7b-chat",
				},
			},
		},
	}
)

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package inferencetest

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/azure/kaito/pkg/inference"
	"github.com/azure/kaito/pkg/model"
	"github.com/azure/kaito/pkg/utils"
	"github.com/azure/kaito/pkg/utils/plugin"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	_ "github.com/azure/kaito/presets/models/falcon"
	_ "github.com/azure/kaito/presets/models/llama2"
	_ "github.com/azure/kaito/presets/models/llama2chat"
)

func TestCreatePresetInference(t *testing.T) {

	testcases := map[string]struct {
		nodeCount   int
		modelName   string
		callMocks   func(c *utils.MockClient)
		workload    string
		expectedCmd string
	}{

		"falcon-7b": {
			nodeCount: 1,
			modelName: "falcon-7b",
			callMocks: func(c *utils.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&appsv1.Deployment{}), mock.Anything).Return(nil)
			},
			workload:    "Deployment",
			expectedCmd: "/bin/sh -c accelerate launch --use_deepspeed --config_file=config.yaml --num_processes=1 --num_machines=1 --machine_rank=0 --gpu_ids=all inference-api.py",
		},
		"falcon-7b-instruct": {
			nodeCount: 1,
			modelName: "falcon-7b-instruct",
			callMocks: func(c *utils.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&appsv1.Deployment{}), mock.Anything).Return(nil)
			},
			workload:    "Deployment",
			expectedCmd: "/bin/sh -c accelerate launch --use_deepspeed --config_file=config.yaml --num_processes=1 --num_machines=1 --machine_rank=0 --gpu_ids=all inference-api.py",
		},
		"falcon-40b": {
			nodeCount: 1,
			modelName: "falcon-40b",
			callMocks: func(c *utils.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&appsv1.Deployment{}), mock.Anything).Return(nil)
			},
			workload:    "Deployment",
			expectedCmd: "/bin/sh -c accelerate launch --use_deepspeed --num_machines=1 --machine_rank=0 --gpu_ids=all --config_file=config.yaml --num_processes=1 inference-api.py",
		},
		"falcon-40b-instruct": {
			nodeCount: 1,
			modelName: "falcon-40b-instruct",
			callMocks: func(c *utils.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&appsv1.Deployment{}), mock.Anything).Return(nil)
			},
			workload:    "Deployment",
			expectedCmd: "/bin/sh -c accelerate launch --use_deepspeed --config_file=config.yaml --num_processes=1 --num_machines=1 --machine_rank=0 --gpu_ids=all inference-api.py",
		},

		"llama-7b-chat": {
			nodeCount: 1,
			modelName: "llama-2-7b-chat",
			callMocks: func(c *utils.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&corev1.Service{}), mock.Anything).Return(nil)
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&appsv1.StatefulSet{}), mock.Anything).Return(nil)
			},
			workload:    "StatefulSet",
			expectedCmd: "/bin/sh -c cd /workspace/llama/llama-2 && torchrun --nnodes=1 --nproc_per_node=1 --node_rank=$(echo $HOSTNAME | grep -o '[^-]*$') --master_addr=10.0.0.1 --master_port=29500 --max_restarts=3 --rdzv_id=job --rdzv_backend=c10d --rdzv_endpoint=testWorkspace-0.testWorkspace-headless.default.svc.cluster.local:29500 inference-api.py --max_seq_len=512 --max_batch_size=8",
		},
		"llama-13b-chat": {
			nodeCount: 1,
			modelName: "llama-2-13b-chat",
			callMocks: func(c *utils.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&corev1.Service{}), mock.Anything).Return(nil)
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&appsv1.StatefulSet{}), mock.Anything).Return(nil)
			},
			workload:    "StatefulSet",
			expectedCmd: "/bin/sh -c cd /workspace/llama/llama-2 && torchrun --nnodes=1 --nproc_per_node=2 --node_rank=$(echo $HOSTNAME | grep -o '[^-]*$') --master_addr=10.0.0.1 --master_port=29500 --max_restarts=3 --rdzv_id=job --rdzv_backend=c10d --rdzv_endpoint=testWorkspace-0.testWorkspace-headless.default.svc.cluster.local:29500 inference-api.py --max_seq_len=512 --max_batch_size=8",
		},
		"llama-70b-chat": {
			nodeCount: 2,
			modelName: "llama-2-70b-chat",
			callMocks: func(c *utils.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&corev1.Service{}), mock.Anything).Return(nil)
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&appsv1.StatefulSet{}), mock.Anything).Return(nil)
			},
			workload:    "StatefulSet",
			expectedCmd: "/bin/sh -c cd /workspace/llama/llama-2 && torchrun --nproc_per_node=4 --node_rank=$(echo $HOSTNAME | grep -o '[^-]*$') --master_addr=10.0.0.1 --master_port=29500 --nnodes=2 --rdzv_backend=c10d --rdzv_endpoint=testWorkspace-0.testWorkspace-headless.default.svc.cluster.local:29500 --max_restarts=3 --rdzv_id=job inference-api.py --max_seq_len=512 --max_batch_size=8",
		},
		"llama-7b": {
			nodeCount: 1,
			modelName: "llama-2-7b",
			callMocks: func(c *utils.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&corev1.Service{}), mock.Anything).Return(nil)
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&appsv1.StatefulSet{}), mock.Anything).Return(nil)
			},
			workload:    "StatefulSet",
			expectedCmd: "/bin/sh -c cd /workspace/llama/llama-2 && torchrun --nnodes=1 --nproc_per_node=1 --node_rank=$(echo $HOSTNAME | grep -o '[^-]*$') --master_addr=10.0.0.1 --master_port=29500 --max_restarts=3 --rdzv_id=job --rdzv_backend=c10d --rdzv_endpoint=testWorkspace-0.testWorkspace-headless.default.svc.cluster.local:29500 inference-api.py --max_seq_len=512 --max_batch_size=8",
		},
		"llama-13b": {
			nodeCount: 1,
			modelName: "llama-2-13b",
			callMocks: func(c *utils.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&corev1.Service{}), mock.Anything).Return(nil)
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&appsv1.StatefulSet{}), mock.Anything).Return(nil)
			},
			workload:    "StatefulSet",
			expectedCmd: "/bin/sh -c cd /workspace/llama/llama-2 && torchrun --node_rank=$(echo $HOSTNAME | grep -o '[^-]*$') --master_addr=10.0.0.1 --master_port=29500 --nnodes=1 --nproc_per_node=2 --max_restarts=3 --rdzv_id=job --rdzv_backend=c10d --rdzv_endpoint=testWorkspace-0.testWorkspace-headless.default.svc.cluster.local:29500 inference-api.py --max_batch_size=8 --max_seq_len=512",
		},
		"llama-70b": {
			nodeCount: 2,
			modelName: "llama-2-70b",
			callMocks: func(c *utils.MockClient) {
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&corev1.Service{}), mock.Anything).Return(nil)
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&appsv1.StatefulSet{}), mock.Anything).Return(nil)
			},
			workload:    "StatefulSet",
			expectedCmd: "/bin/sh -c cd /workspace/llama/llama-2 && torchrun --nproc_per_node=4 --node_rank=$(echo $HOSTNAME | grep -o '[^-]*$') --master_addr=10.0.0.1 --master_port=29500 --nnodes=2 --rdzv_backend=c10d --rdzv_endpoint=testWorkspace-0.testWorkspace-headless.default.svc.cluster.local:29500 --max_restarts=3 --rdzv_id=job inference-api.py --max_seq_len=512 --max_batch_size=8",
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			mockClient := utils.NewClient()
			tc.callMocks(mockClient)

			workspace := utils.MockWorkspaceWithPreset
			workspace.Resource.Count = &tc.nodeCount

			useHeadlessSvc := false

			var inferenceObj *model.PresetInferenceParam
			model := plugin.KaitoModelRegister.MustGet(tc.modelName)
			inferenceObj = model.GetInferenceParameters()

			if strings.HasPrefix(tc.modelName, "llama") {
				useHeadlessSvc = true
			}
			svc := &corev1.Service{
				ObjectMeta: v1.ObjectMeta{
					Name:      workspace.Name,
					Namespace: workspace.Namespace,
				},
				Spec: corev1.ServiceSpec{
					ClusterIP: "10.0.0.1",
				},
			}
			mockClient.CreateOrUpdateObjectInMap(svc)

			createdObject, _ := inference.CreatePresetInference(context.TODO(), workspace, inferenceObj, useHeadlessSvc, mockClient)
			createdWorkload := ""
			switch createdObject.(type) {
			case *appsv1.Deployment:
				createdWorkload = "Deployment"
			case *appsv1.StatefulSet:
				createdWorkload = "StatefulSet"
			}
			if tc.workload != createdWorkload {
				t.Errorf("%s: returned worklaod type is wrong", k)
			}

			var workloadCmd string
			if tc.workload == "Deployment" {
				workloadCmd = strings.Join((createdObject.(*appsv1.Deployment)).Spec.Template.Spec.Containers[0].Command, " ")

			} else {
				workloadCmd = strings.Join((createdObject.(*appsv1.StatefulSet)).Spec.Template.Spec.Containers[0].Command, " ")
			}

			mainCmd := strings.Split(workloadCmd, "--")[0]
			params := toParameterMap(strings.Split(workloadCmd, "--")[1:])

			expectedMaincmd := strings.Split(tc.expectedCmd, "--")[0]
			expectedParams := toParameterMap(strings.Split(workloadCmd, "--")[1:])

			if mainCmd != expectedMaincmd {
				t.Errorf("%s main cmdline is not expected, got %s, expect %s ", k, workloadCmd, tc.expectedCmd)
			}

			if !reflect.DeepEqual(params, expectedParams) {
				t.Errorf("%s parameters are not expected, got %s, expect %s ", k, params, expectedParams)
			}
		})
	}
}

func toParameterMap(in []string) map[string]string {
	ret := make(map[string]string)
	for _, each := range in {
		r := strings.Split(each, "=")
		k := r[0]
		var v string
		if len(r) == 1 {
			v = ""
		} else {
			v = r[1]
		}
		ret[k] = v
	}
	return ret
}

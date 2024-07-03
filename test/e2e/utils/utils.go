// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package utils

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/samber/lo"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	InferenceModeCustomTemplate kaitov1alpha1.ModelImageAccessMode = "customTemplate"
)

var (
	// PollInterval defines the interval time for a poll operation.
	PollInterval = 250 * time.Millisecond
	// PollTimeout defines the time after which the poll operation times out.
	PollTimeout = 60 * time.Second
)

func GetEnv(envVar string) string {
	env := os.Getenv(envVar)
	if env == "" {
		fmt.Printf("%s is not set or is empty", envVar)
		return ""
	}
	return env
}

// GenerateRandomString generates a random number between 0 and 1000 and returns it as a string.
func GenerateRandomString() string {
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator
	randomNumber := rand.Intn(1001)  // Generate a random number between 0 and 1000
	return fmt.Sprintf("%d", randomNumber)
}

func GetModelConfigInfo(configFilePath string) (map[string]interface{}, error) {
	var data map[string]interface{}

	yamlData, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML file: %w", err)
	}

	err = yaml.Unmarshal(yamlData, &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling YAML: %w", err)
	}

	return data, nil
}

func GetPodNameForJob(coreClient *kubernetes.Clientset, namespace, jobName string) (string, error) {
	podList, err := coreClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil {
		return "", err
	}

	if len(podList.Items) == 0 {
		return "", fmt.Errorf("no pods found for job %s", jobName)
	}

	return podList.Items[0].Name, nil
}

func GetPodLogs(coreClient *kubernetes.Clientset, namespace, podName, containerName string) (string, error) {
	req := coreClient.CoreV1().Pods(namespace).GetLogs(podName, &v1.PodLogOptions{Container: containerName})
	logs, err := req.Stream(context.Background())
	if err != nil {
		return "", err
	}
	defer logs.Close()

	buf := new(strings.Builder)
	_, err = io.Copy(buf, logs)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func CopySecret(original *corev1.Secret, targetNamespace string) *corev1.Secret {
	newSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      original.Name,
			Namespace: targetNamespace,
		},
		Data: original.Data,
		Type: original.Type,
	}
	return newSecret
}

func ExtractModelVersion(configs map[string]interface{}) (map[string]string, error) {
	modelsInfo := make(map[string]string)
	models, ok := configs["models"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("'models' key not found or is not a slice")
	}

	for _, modelItem := range models {
		model, ok := modelItem.(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("model item is not a map")
		}

		modelName, ok := model["name"].(string)
		if !ok {
			return nil, fmt.Errorf("model name is not a string or not found")
		}

		modelTag, ok := model["tag"].(string) // Using 'tag' as the version
		if !ok {
			return nil, fmt.Errorf("model version for %s is not a string or not found", modelName)
		}

		modelsInfo[modelName] = modelTag
	}

	return modelsInfo, nil
}

func GenerateInferenceWorkspaceManifest(name, namespace, imageName string, resourceCount int, instanceType string,
	labelSelector *metav1.LabelSelector, preferredNodes []string, presetName kaitov1alpha1.ModelName,
	accessMode kaitov1alpha1.ModelImageAccessMode, imagePullSecret []string,
	podTemplate *corev1.PodTemplateSpec, adapters []kaitov1alpha1.AdapterSpec) *kaitov1alpha1.Workspace {

	workspace := &kaitov1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Resource: kaitov1alpha1.ResourceSpec{
			Count:          lo.ToPtr(resourceCount),
			InstanceType:   instanceType,
			LabelSelector:  labelSelector,
			PreferredNodes: preferredNodes,
		},
	}

	var workspaceInference kaitov1alpha1.InferenceSpec
	if accessMode == kaitov1alpha1.ModelImageAccessModePublic ||
		accessMode == kaitov1alpha1.ModelImageAccessModePrivate {
		workspaceInference.Preset = &kaitov1alpha1.PresetSpec{
			PresetMeta: kaitov1alpha1.PresetMeta{
				Name:       presetName,
				AccessMode: accessMode,
			},
			PresetOptions: kaitov1alpha1.PresetOptions{
				Image:            imageName,
				ImagePullSecrets: imagePullSecret,
			},
		}
	}

	if adapters != nil {
		workspaceInference.Adapters = adapters
	}

	if accessMode == InferenceModeCustomTemplate {
		workspaceInference.Template = podTemplate
	}

	workspace.Inference = &workspaceInference

	return workspace
}

func GenerateTuningWorkspaceManifest(name, namespace, imageName string, resourceCount int, instanceType string,
	labelSelector *metav1.LabelSelector, preferredNodes []string, input *kaitov1alpha1.DataSource,
	output *kaitov1alpha1.DataDestination, preset *kaitov1alpha1.PresetSpec, method kaitov1alpha1.TuningMethod) *kaitov1alpha1.Workspace {

	workspace := &kaitov1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Resource: kaitov1alpha1.ResourceSpec{
			Count:          lo.ToPtr(resourceCount),
			InstanceType:   instanceType,
			LabelSelector:  labelSelector,
			PreferredNodes: preferredNodes,
		},
		Tuning: &kaitov1alpha1.TuningSpec{
			Method: method,
			Input:  input,
			Output: output,
			Preset: preset,
		},
	}

	return workspace
}

// GenerateE2ETuningConfigMapManifest generates a ConfigMap manifest for E2E tuning.
func GenerateE2ETuningConfigMapManifest(namespace string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "e2e-qlora-params-template",
			Namespace: namespace, // Same as workspace namespace
		},
		Data: map[string]string{
			"training_config.yaml": `training_config:
  ModelConfig:
	torch_dtype: "bfloat16"
	local_files_only: true
	device_map: "auto"
  
  QuantizationConfig:
	load_in_4bit: true
	bnb_4bit_quant_type: "nf4"
	bnb_4bit_compute_dtype: "bfloat16"
	bnb_4bit_use_double_quant: true
  
  LoraConfig:
	r: 8
	lora_alpha: 8
	lora_dropout: 0.0
	target_modules: ['k_proj', 'q_proj', 'v_proj', 'o_proj', "gate_proj", "down_proj", "up_proj"]
  
  TrainingArguments:
	output_dir: "/mnt/results"
	ddp_find_unused_parameters: false
	save_strategy: "epoch"
	per_device_train_batch_size: 1
	max_steps: 5  # Adding this line to limit training to 5 steps
  
  DataCollator:
	mlm: true
  
  DatasetConfig:
	shuffle_dataset: true
	train_test_split: 1`,
		},
	}
}

func GeneratePodTemplate(name, namespace, image string, labels map[string]string) *corev1.PodTemplateSpec {
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            name,
					Image:           image,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Command:         []string{"/bin/sleep", "10000"},
				},
			},
		},
	}
}

package tuning

import corev1 "k8s.io/api/core/v1"

const (
	DefaultNumProcesses = "1"
	DefaultNumMachines  = "1"
	DefaultMachineRank  = "0"
	DefaultGPUIds       = "all"
)

var (
	DefaultAccelerateParams = map[string]string{
		"config_file": "naive_mp_config.yaml",
	}

	DefaultImagePullSecrets = []corev1.LocalObjectReference{}
)

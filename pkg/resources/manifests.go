// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package resources

import (
	"context"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var controller = true

func GenerateHeadlessServiceManifest(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace) *corev1.Service {
	serviceName := fmt.Sprintf("%s-headless", workspaceObj.Name)
	selector := map[string]string{
		kaitov1alpha1.LabelWorkspaceName: workspaceObj.Name,
	}

	return &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      serviceName,
			Namespace: workspaceObj.Namespace,
			OwnerReferences: []v1.OwnerReference{
				{
					APIVersion: kaitov1alpha1.GroupVersion.String(),
					Kind:       "Workspace",
					UID:        workspaceObj.UID,
					Name:       workspaceObj.Name,
					Controller: &controller,
				},
			},
		},
		Spec: corev1.ServiceSpec{
			Selector:  selector,
			ClusterIP: "None",
			Ports: []corev1.ServicePort{
				{
					Name:       "torchrun",
					Protocol:   corev1.ProtocolTCP,
					Port:       29500,
					TargetPort: intstr.FromInt(29500),
				},
			},
			PublishNotReadyAddresses: true,
		},
	}
}

func GenerateServiceManifest(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace, serviceType corev1.ServiceType, isStatefulSet bool) *corev1.Service {
	selector := map[string]string{
		kaitov1alpha1.LabelWorkspaceName: workspaceObj.Name,
	}
	// If statefulset, modify the selector to select the pod with index 0 as the endpoint
	if isStatefulSet {
		podNameForIndex0 := fmt.Sprintf("%s-0", workspaceObj.Name)
		selector["statefulset.kubernetes.io/pod-name"] = podNameForIndex0
	}

	return &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      workspaceObj.Name,
			Namespace: workspaceObj.Namespace,
			OwnerReferences: []v1.OwnerReference{
				{
					APIVersion: kaitov1alpha1.GroupVersion.String(),
					Kind:       "Workspace",
					UID:        workspaceObj.UID,
					Name:       workspaceObj.Name,
					Controller: &controller,
				},
			},
		},
		Spec: corev1.ServiceSpec{
			Type: serviceType,
			Ports: []corev1.ServicePort{
				// HTTP API Port
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(5000),
				},
				// Torch NCCL Port
				{
					Name:       "torch",
					Protocol:   corev1.ProtocolTCP,
					Port:       29500,
					TargetPort: intstr.FromInt(29500),
				},
			},
			Selector: selector,
			// Added this to allow pods to discover each other
			// (DNS Resolution) During their initialization phase
			PublishNotReadyAddresses: true,
		},
	}
}

func GenerateStatefulSetManifest(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace, imageName string,
	imagePullSecretRefs []corev1.LocalObjectReference, replicas int, commands []string, containerPorts []corev1.ContainerPort,
	livenessProbe, readinessProbe *corev1.Probe, resourceRequirements corev1.ResourceRequirements,
	tolerations []corev1.Toleration, volumes []corev1.Volume, volumeMount []corev1.VolumeMount) *appsv1.StatefulSet {

	nodeRequirements := make([]corev1.NodeSelectorRequirement, 0, len(workspaceObj.Resource.LabelSelector.MatchLabels))
	for key, value := range workspaceObj.Resource.LabelSelector.MatchLabels {
		nodeRequirements = append(nodeRequirements, corev1.NodeSelectorRequirement{
			Key:      key,
			Operator: corev1.NodeSelectorOpIn,
			Values:   []string{value},
		})
	}

	selector := map[string]string{
		kaitov1alpha1.LabelWorkspaceName: workspaceObj.Name,
	}
	labelselector := &v1.LabelSelector{
		MatchLabels: selector,
	}

	ss := &appsv1.StatefulSet{
		ObjectMeta: v1.ObjectMeta{
			Name:      workspaceObj.Name,
			Namespace: workspaceObj.Namespace,
			OwnerReferences: []v1.OwnerReference{
				{
					APIVersion: kaitov1alpha1.GroupVersion.String(),
					Kind:       "Workspace",
					UID:        workspaceObj.UID,
					Name:       workspaceObj.Name,
					Controller: &controller,
				},
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:            lo.ToPtr(int32(replicas)),
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Selector:            labelselector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: selector,
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: imagePullSecretRefs,
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: nodeRequirements,
									},
								},
							},
						},
					},

					Containers: []corev1.Container{
						{
							Name:           workspaceObj.Name,
							Image:          imageName,
							Command:        commands,
							Resources:      resourceRequirements,
							LivenessProbe:  livenessProbe,
							ReadinessProbe: readinessProbe,
							Ports:          containerPorts,
							VolumeMounts:   volumeMount,
						},
					},
					Tolerations: tolerations,
					Volumes:     volumes,
				},
			},
		},
	}
	ss.Spec.ServiceName = fmt.Sprintf("%s-headless", workspaceObj.Name)
	return ss
}

func dockerSidecarScriptPushToVolume(arg interface{}) string {
	volume, ok := arg.(*corev1.VolumeSource)
	return `
`
}
func dockerSidecarScriptPushImage(arg interface{}) string {
	image, _ := arg.(string)
	return fmt.Sprintf(`
# Start the Docker daemon in the background with specific options for DinD
dockerd &
# Wait for the Docker daemon to be ready
while ! docker info > /dev/null 2>&1; do
  echo "Waiting for Docker daemon to start..."
  sleep 1
done
echo 'Docker daemon started'

while true; do
  FILE_PATH=$(find /workspace/tfs -name 'fine_tuning_completed.txt')
  if [ ! -z "$FILE_PATH" ]; then
    echo "FOUND TRAINING COMPLETED FILE at $FILE_PATH"

    PARENT_DIR=$(dirname "$FILE_PATH")
    echo "Parent directory is $PARENT_DIR"

    TEMP_CONTEXT=$(mktemp -d)
    cp "$PARENT_DIR/adapter_config.json" "$TEMP_CONTEXT/adapter_config.json"
    cp -r "$PARENT_DIR/adapter_model.safetensors" "$TEMP_CONTEXT/adapter_model.safetensors"

    # Create a minimal Dockerfile
    echo 'FROM scratch
    ADD adapter_config.json /
    ADD adapter_model.safetensors /' > "$TEMP_CONTEXT/Dockerfile"

    docker build -t %s "$TEMP_CONTEXT"
    docker push %s

    # Cleanup: Remove the temporary directory
    rm -rf "$TEMP_CONTEXT"

    # Remove the file to prevent repeated builds, or handle as needed
    # rm "$FILE_PATH"
    echo "Upload complete"
    exit 0
  fi
  sleep 10  # Check every 10 seconds
done`, image, image)
}

func determinePushMethod(wObj *kaitov1alpha1.Workspace) (func(arg interface{}) string, interface{}) {
	if wObj.Tuning.Output.Volume != nil {
		return dockerSidecarScriptPushToVolume, wObj.Tuning.Output.Volume
	}
	if wObj.Tuning.Output.Image != "" {
		return dockerSidecarScriptPushImage, wObj.Tuning.Output.Image
	}
	return func(arg interface{}) string { return "" }, ""
}

func GenerateTuningJobManifest(ctx context.Context, wObj *kaitov1alpha1.Workspace, imageName string,
	imagePullSecretRefs []corev1.LocalObjectReference, replicas int, commands []string, containerPorts []corev1.ContainerPort,
	livenessProbe, readinessProbe *corev1.Probe, resourceRequirements corev1.ResourceRequirements, tolerations []corev1.Toleration,
	initContainers []corev1.Container, volumes []corev1.Volume, volumeMounts []corev1.VolumeMount) *batchv1.Job {
	labels := map[string]string{
		kaitov1alpha1.LabelWorkspaceName: wObj.Name,
	}
	pushMethod, pushArg := determinePushMethod(wObj)
	return &batchv1.Job{
		TypeMeta: v1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      wObj.Name,
			Namespace: wObj.Namespace,
			Labels:    labels,
			OwnerReferences: []v1.OwnerReference{
				{
					APIVersion: kaitov1alpha1.GroupVersion.String(),
					Kind:       "Workspace",
					Name:       wObj.Name,
					UID:        wObj.UID,
					Controller: pointer.BoolPtr(true),
				},
			},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					InitContainers: initContainers,
					Containers: []corev1.Container{
						{
							Name:           wObj.Name,
							Image:          imageName,
							Command:        commands,
							Resources:      resourceRequirements,
							LivenessProbe:  livenessProbe,
							ReadinessProbe: readinessProbe,
							Ports:          containerPorts,
							VolumeMounts:   volumeMounts,
						},
						{
							Name:  "docker-sidecar",
							Image: "docker:dind",
							SecurityContext: &corev1.SecurityContext{
								Privileged: pointer.BoolPtr(true),
							},
							VolumeMounts: volumeMounts,
							Command:      []string{"/bin/sh", "-c"},
							Args:         []string{pushMethod(pushArg)},
						},
					},
					RestartPolicy:    corev1.RestartPolicyNever,
					Volumes:          volumes,
					Tolerations:      tolerations,
					ImagePullSecrets: imagePullSecretRefs,
				},
			},
		},
	}
}

func GenerateDeploymentManifest(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace, imageName string,
	imagePullSecretRefs []corev1.LocalObjectReference, replicas int, commands []string, containerPorts []corev1.ContainerPort,
	livenessProbe, readinessProbe *corev1.Probe, resourceRequirements corev1.ResourceRequirements,
	tolerations []corev1.Toleration, volumes []corev1.Volume, volumeMount []corev1.VolumeMount) *appsv1.Deployment {

	nodeRequirements := make([]corev1.NodeSelectorRequirement, 0, len(workspaceObj.Resource.LabelSelector.MatchLabels))
	for key, value := range workspaceObj.Resource.LabelSelector.MatchLabels {
		nodeRequirements = append(nodeRequirements, corev1.NodeSelectorRequirement{
			Key:      key,
			Operator: corev1.NodeSelectorOpIn,
			Values:   []string{value},
		})
	}

	selector := map[string]string{
		kaitov1alpha1.LabelWorkspaceName: workspaceObj.Name,
	}
	labelselector := &v1.LabelSelector{
		MatchLabels: selector,
	}

	return &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      workspaceObj.Name,
			Namespace: workspaceObj.Namespace,
			OwnerReferences: []v1.OwnerReference{
				{
					APIVersion: kaitov1alpha1.GroupVersion.String(),
					Kind:       "Workspace",
					UID:        workspaceObj.UID,
					Name:       workspaceObj.Name,
					Controller: &controller,
				},
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: lo.ToPtr(int32(replicas)),
			Selector: labelselector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: selector,
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: imagePullSecretRefs,
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: nodeRequirements,
									},
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:           workspaceObj.Name,
							Image:          imageName,
							Command:        commands,
							Resources:      resourceRequirements,
							LivenessProbe:  livenessProbe,
							ReadinessProbe: readinessProbe,
							Ports:          containerPorts,
							VolumeMounts:   volumeMount,
						},
					},
					Tolerations: tolerations,
					Volumes:     volumes,
				},
			},
		},
	}
}

func GenerateDeploymentManifestWithPodTemplate(ctx context.Context, workspaceObj *kaitov1alpha1.Workspace, tolerations []corev1.Toleration) *appsv1.Deployment {
	nodeRequirements := make([]corev1.NodeSelectorRequirement, 0, len(workspaceObj.Resource.LabelSelector.MatchLabels))
	for key, value := range workspaceObj.Resource.LabelSelector.MatchLabels {
		nodeRequirements = append(nodeRequirements, corev1.NodeSelectorRequirement{
			Key:      key,
			Operator: corev1.NodeSelectorOpIn,
			Values:   []string{value},
		})
	}

	templateCopy := workspaceObj.Inference.Template.DeepCopy()

	if templateCopy.ObjectMeta.Labels == nil {
		templateCopy.ObjectMeta.Labels = make(map[string]string)
	}
	templateCopy.ObjectMeta.Labels[kaitov1alpha1.LabelWorkspaceName] = workspaceObj.Name
	labelselector := &v1.LabelSelector{
		MatchLabels: map[string]string{
			kaitov1alpha1.LabelWorkspaceName: workspaceObj.Name,
		},
	}
	// Overwrite affinity
	templateCopy.Spec.Affinity = &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: nodeRequirements,
					},
				},
			},
		},
	}

	// append tolerations
	if templateCopy.Spec.Tolerations == nil {
		templateCopy.Spec.Tolerations = tolerations
	} else {
		templateCopy.Spec.Tolerations = append(templateCopy.Spec.Tolerations, tolerations...)
	}

	return &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      workspaceObj.Name,
			Namespace: workspaceObj.Namespace,
			OwnerReferences: []v1.OwnerReference{
				{
					APIVersion: kaitov1alpha1.GroupVersion.String(),
					Kind:       "Workspace",
					UID:        workspaceObj.UID,
					Name:       workspaceObj.Name,
					Controller: &controller,
				},
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: lo.ToPtr(int32(*workspaceObj.Resource.Count)),
			Selector: labelselector,
			Template: *templateCopy,
		},
	}
}

#!/usr/bin/env bash

# Copyright (c) KAITO authors.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

readonly OPERATOR_VERSION="v26.7.0"
readonly OPERATOR_NAMESPACE="gpu-operator"
readonly OPERATOR_RELEASE="gpu-operator"

for command_name in helm kubectl; do
  if ! command -v "${command_name}" >/dev/null 2>&1; then
    echo "required command not found: ${command_name}" >&2
    exit 1
  fi
done

cluster_policy_count="$({ kubectl get clusterpolicies.nvidia.com --no-headers 2>/dev/null || true; } | wc -l | tr -d ' ')"
runtime_class_exists=false
if kubectl get runtimeclass nvidia >/dev/null 2>&1; then
  runtime_class_exists=true
fi

if [[ "${cluster_policy_count}" != "0" || "${runtime_class_exists}" == "true" ]]; then
  if [[ "${cluster_policy_count}" == "0" || "${runtime_class_exists}" != "true" ]]; then
    echo "partial NVIDIA runtime stack detected; refusing to install a second operator" >&2
    exit 1
  fi
  if ! kubectl get clusterpolicies.nvidia.com -o jsonpath='{range .items[*]}{.status.state}{"\n"}{end}' | grep -qi '^ready$'; then
    echo "existing NVIDIA GPU Operator ClusterPolicy is not Ready" >&2
    exit 1
  fi
  echo "reusing existing Ready NVIDIA GPU Operator and RuntimeClass/nvidia"
  exit 0
fi

helm repo add nvidia https://helm.ngc.nvidia.com/nvidia
helm repo update nvidia
helm upgrade --install "${OPERATOR_RELEASE}" nvidia/gpu-operator \
  --version "${OPERATOR_VERSION}" \
  --namespace "${OPERATOR_NAMESPACE}" \
  --create-namespace \
  --set driver.enabled=false \
  --set toolkit.enabled=false \
  --set operator.runtimeClass=nvidia-container-runtime \
  --set cdi.enabled=false \
  --set devicePlugin.enabled=true \
  --set nfd.enabled=true \
  --set-json 'daemonsets.tolerations=[{"key":"nvidia.com/gpu","operator":"Exists","effect":"NoSchedule"},{"key":"sku","operator":"Equal","value":"gpu","effect":"NoSchedule"},{"key":"karpenter.sh/disrupted","operator":"Exists","effect":"NoSchedule"}]' \
  --set-json 'node-feature-discovery.worker.tolerations=[{"key":"node-role.kubernetes.io/control-plane","operator":"Equal","value":"","effect":"NoSchedule"},{"key":"nvidia.com/gpu","operator":"Exists","effect":"NoSchedule"},{"key":"sku","operator":"Equal","value":"gpu","effect":"NoSchedule"},{"key":"karpenter.sh/disrupted","operator":"Exists","effect":"NoSchedule"}]' \
  --wait \
  --timeout 15m

kubectl wait clusterpolicies.nvidia.com/cluster-policy \
  --for=jsonpath='{.status.state}'=ready \
  --timeout=10m
kubectl apply -f - <<'EOF'
apiVersion: node.k8s.io/v1
kind: RuntimeClass
metadata:
  name: nvidia
handler: nvidia-container-runtime
EOF

echo "NVIDIA GPU Operator ${OPERATOR_VERSION} is Ready"
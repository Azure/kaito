
# Image URL to use all building/pushing image targets
REGISTRY ?= YOUR_REGISTRY
IMG_NAME ?= workspace
VERSION ?= v0.3.0
GPU_PROVISIONER_VERSION ?= 0.2.0
IMG_TAG ?= $(subst v,,$(VERSION))

ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
BIN_DIR := $(abspath $(ROOT_DIR)/bin)

TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(abspath $(TOOLS_DIR)/bin)

GOLANGCI_LINT_VER := v1.57.2
GOLANGCI_LINT_BIN := golangci-lint
GOLANGCI_LINT := $(abspath $(TOOLS_BIN_DIR)/$(GOLANGCI_LINT_BIN)-$(GOLANGCI_LINT_VER))

E2E_TEST_BIN := e2e.test
E2E_TEST := $(BIN_DIR)/$(E2E_TEST_BIN)

GINKGO_VER := v2.19.0
GINKGO_BIN := ginkgo
GINKGO := $(TOOLS_BIN_DIR)/$(GINKGO_BIN)-$(GINKGO_VER)

AZURE_SUBSCRIPTION_ID ?= $(AZURE_SUBSCRIPTION_ID)
AZURE_LOCATION ?= eastus
AKS_K8S_VERSION ?= 1.29.2
AZURE_RESOURCE_GROUP ?= demo
AZURE_CLUSTER_NAME ?= kaito-demo
AZURE_RESOURCE_GROUP_MC=MC_$(AZURE_RESOURCE_GROUP)_$(AZURE_CLUSTER_NAME)_$(AZURE_LOCATION)
GPU_NAMESPACE ?= gpu-provisioner
KAITO_NAMESPACE ?= kaito-workspace
GPU_PROVISIONER_MSI_NAME ?= gpuIdentity

RUN_LLAMA_13B ?= false
AI_MODELS_REGISTRY ?= modelregistry.azurecr.io
AI_MODELS_REGISTRY_SECRET ?= modelregistry
SUPPORTED_MODELS_YAML_PATH ?= /home/runner/work/kaito/kaito/presets/models/supported_models.yaml

# Scripts
GO_INSTALL := ./hack/go-install.sh

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

## --------------------------------------
## Tooling Binaries
## --------------------------------------

$(GOLANGCI_LINT):
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) github.com/golangci/golangci-lint/cmd/golangci-lint $(GOLANGCI_LINT_BIN) $(GOLANGCI_LINT_VER)

$(GINKGO):
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) github.com/onsi/ginkgo/v2/ginkgo $(GINKGO_BIN) $(GINKGO_VER)

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	cp config/crd/bases/kaito.sh_workspaces.yaml charts/kaito/workspace/crds/

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

## --------------------------------------
## Tests
## --------------------------------------
.PHONY: unit-test
unit-test: ## Run unit tests.
	go test -v $(shell go list ./pkg/... ./api/... | \
	grep -v -e /vendor -e /api/v1alpha1/zz_generated.deepcopy.go -e /pkg/utils/test/...) \
	-race -coverprofile=coverage.txt -covermode=atomic
	go tool cover -func=coverage.txt

inference-api-e2e: 
	pip install -r presets/inference/text-generation/requirements.txt
	pytest -o log_cli=true -o log_cli_level=INFO .

$(E2E_TEST):
	(cd test/e2e && go test -c . -o $(E2E_TEST))

# Ginkgo configurations
GINKGO_FOCUS ?=
GINKGO_SKIP ?=
GINKGO_NODES ?= 1
GINKGO_NO_COLOR ?= false
GINKGO_TIMEOUT ?= 60m
GINKGO_ARGS ?= -focus="$(GINKGO_FOCUS)" -skip="$(GINKGO_SKIP)" -nodes=$(GINKGO_NODES) -no-color=$(GINKGO_NO_COLOR) -timeout=$(GINKGO_TIMEOUT)

.PHONY: kaito-workspace-e2e-test
kaito-workspace-e2e-test: $(E2E_TEST) $(GINKGO)
	AI_MODELS_REGISTRY_SECRET=$(AI_MODELS_REGISTRY_SECRET) RUN_LLAMA_13B=$(RUN_LLAMA_13B) \
 	AI_MODELS_REGISTRY=$(AI_MODELS_REGISTRY) GPU_NAMESPACE=$(GPU_NAMESPACE) KAITO_NAMESPACE=$(KAITO_NAMESPACE) \
	SUPPORTED_MODELS_YAML_PATH=$(SUPPORTED_MODELS_YAML_PATH) \
 	$(GINKGO) -v -trace $(GINKGO_ARGS) $(E2E_TEST)

.PHONY: create-rg
create-rg: ## Create resource group
	az group create --name $(AZURE_RESOURCE_GROUP) --location $(AZURE_LOCATION) -o none

.PHONY: create-acr
create-acr:  ## Create test ACR
	az acr create --name $(AZURE_ACR_NAME) --resource-group $(AZURE_RESOURCE_GROUP) --sku Standard --admin-enabled -o none
	az acr login  --name $(AZURE_ACR_NAME)

.PHONY: create-aks-cluster
create-aks-cluster: ## Create test AKS cluster (with msi, oidc, and workload identity enabled)
	az aks create  --name $(AZURE_CLUSTER_NAME) --resource-group $(AZURE_RESOURCE_GROUP) --location $(AZURE_LOCATION) \
	--attach-acr $(AZURE_ACR_NAME) 	--kubernetes-version $(AKS_K8S_VERSION) --node-count 1 --generate-ssh-keys  \
	--enable-managed-identity --enable-workload-identity --enable-oidc-issuer --node-vm-size Standard_D2s_v3 -o none

.PHONY: create-aks-cluster-with-kaito
create-aks-cluster-with-kaito: ## Create test AKS cluster (with msi, oidc and kaito enabled)
	az aks create  --name $(AZURE_CLUSTER_NAME) --resource-group $(AZURE_RESOURCE_GROUP) --location $(AZURE_LOCATION) --node-count 1 \
 	--generate-ssh-keys --enable-managed-identity --enable-oidc-issuer --enable-ai-toolchain-operator -o none

	az aks get-credentials --name $(AZURE_CLUSTER_NAME) --resource-group $(AZURE_RESOURCE_GROUP)

.PHONY: prepare-kaito-addon-identity
prepare-kaito-addon-identity:
	IDENTITY_PRINCIPAL_ID=$(shell az identity show --name "ai-toolchain-operator-$(AZURE_CLUSTER_NAME)" -g "$(AZURE_RESOURCE_GROUP_MC)"  --query 'principalId');\
	az role assignment create --assignee $$IDENTITY_PRINCIPAL_ID --scope "/subscriptions/$(AZURE_SUBSCRIPTION_ID)/resourceGroups/$(AZURE_RESOURCE_GROUP_MC)"  --role "Contributor"

	AKS_OIDC_ISSUER=$(shell az aks show -n "$(AZURE_CLUSTER_NAME)" -g "$(AZURE_RESOURCE_GROUP_MC)" --query 'oidcIssuerProfile.issuerUrl');\
	az identity federated-credential create --name gpu-federated-cred --identity-name "ai-toolchain-operator-$(AZURE_CLUSTER_NAME)" \
    -g "$(AZURE_RESOURCE_GROUP)" --issuer $$AKS_OIDC_ISSUER \
    --subject system:serviceaccount:"$(KAITO_NAMESPACE):kaito-gpu-provisioner" --audience api://AzureADTokenExchange

.PHONY: az-patch-install-helm
az-patch-install-helm: ## Update Azure client env vars and settings in helm values.yml
	az aks get-credentials --name $(AZURE_CLUSTER_NAME) --resource-group $(AZURE_RESOURCE_GROUP)

	yq -i '(.image.repository)                                              = "$(REGISTRY)/workspace"'                    ./charts/kaito/workspace/values.yaml
	yq -i '(.image.tag)                                                     = "$(IMG_TAG)"'                               ./charts/kaito/workspace/values.yaml

	helm install kaito-workspace ./charts/kaito/workspace --namespace $(KAITO_NAMESPACE) --create-namespace

##@ Build

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/*.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/main.go

##@ Docker
BUILDX_BUILDER_NAME ?= img-builder
OUTPUT_TYPE ?= type=registry
QEMU_VERSION ?= 5.2.0-2
ARCH ?= amd64,arm64

.PHONY: docker-buildx
docker-buildx: ## Build and push docker image for the manager for cross-platform support
	@if ! docker buildx ls | grep $(BUILDX_BUILDER_NAME); then \
		docker run --rm --privileged multiarch/qemu-user-static:$(QEMU_VERSION) --reset -p yes; \
		docker buildx create --name $(BUILDX_BUILDER_NAME) --use; \
		docker buildx inspect $(BUILDX_BUILDER_NAME) --bootstrap; \
	fi

.PHONY: docker-build-kaito
docker-build-kaito: docker-buildx
	docker buildx build \
		--file ./docker/kaito/Dockerfile \
		--output=$(OUTPUT_TYPE) \
		--platform="linux/$(ARCH)" \
		--pull \
		--tag $(REGISTRY)/$(IMG_NAME):$(IMG_TAG) .

.PHONY: docker-build-adapter
docker-build-adapter: docker-buildx
	docker buildx build \
		--file ./docker/adapter/Dockerfile \
		--output=$(OUTPUT_TYPE) \
		--platform="linux/$(ARCH)" \
		--pull \
		--tag $(REGISTRY)/e2e-adapter:0.0.1 .

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

##@ gpu-provider
gpu-provisioner-identity-perm: ## Create identity for gpu-provisioner
	az identity create --name $(GPU_PROVISIONER_MSI_NAME) --resource-group $(AZURE_RESOURCE_GROUP)

	IDENTITY_PRINCIPAL_ID=$(shell az identity show --name $(GPU_PROVISIONER_MSI_NAME) --resource-group $(AZURE_RESOURCE_GROUP) --subscription $(AZURE_SUBSCRIPTION_ID) --query 'principalId');\
	az role assignment create --assignee $$IDENTITY_PRINCIPAL_ID --scope /subscriptions/$(AZURE_SUBSCRIPTION_ID)/resourceGroups/$(AZURE_RESOURCE_GROUP)  --role "Contributor"

	AKS_OIDC_ISSUER=$(shell az aks show -n "$(AZURE_CLUSTER_NAME)" -g "$(AZURE_RESOURCE_GROUP)" --subscription $(AZURE_SUBSCRIPTION_ID) --query "oidcIssuerProfile.issuerUrl");\
	az identity federated-credential create --name gpu-federatecredential --identity-name $(GPU_PROVISIONER_MSI_NAME) --resource-group "$(AZURE_RESOURCE_GROUP)" --issuer $$AKS_OIDC_ISSUER \
	--subject system:serviceaccount:"$(GPU_NAMESPACE):$(GPU_NAMESPACE)" --audience api://AzureADTokenExchange --subscription $(AZURE_SUBSCRIPTION_ID)

.PHONY: gpu-provisioner-helm
gpu-provisioner-helm:  ## Update Azure client env vars and settings in helm values.yml
	az aks get-credentials --name $(AZURE_CLUSTER_NAME) --resource-group $(AZURE_RESOURCE_GROUP)

	curl -sO https://raw.githubusercontent.com/Azure/gpu-provisioner/main/hack/deploy/configure-helm-values.sh
	chmod +x ./configure-helm-values.sh && ./configure-helm-values.sh $(AZURE_CLUSTER_NAME) $(AZURE_RESOURCE_GROUP) $(GPU_PROVISIONER_MSI_NAME)

	helm install $(GPU_NAMESPACE) --values gpu-provisioner-values.yaml --set settings.azure.clusterName=$(AZURE_CLUSTER_NAME) --wait \
	https://github.com/Azure/gpu-provisioner/raw/gh-pages/charts/gpu-provisioner-$(GPU_PROVISIONER_VERSION).tgz

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUBECTL ?= kubectl
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
CONTROLLER_TOOLS_VERSION ?= v0.15.0

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

## --------------------------------------
## Linting
## --------------------------------------
.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: lint
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run -v

## --------------------------------------
## Release
## To create a release, run `make release VERSION=x.y.z`
## --------------------------------------
.PHONY: release-manifest
release-manifest:
	@sed -i -e 's/^VERSION ?= .*/VERSION ?= ${VERSION}/' ./Makefile
	@sed -i -e "s/version: .*/version: ${IMG_TAG}/" ./charts/kaito/workspace/Chart.yaml
	@sed -i -e "s/appVersion: .*/appVersion: ${IMG_TAG}/" ./charts/kaito/workspace/Chart.yaml
	@sed -i -e "s/tag: .*/tag: ${IMG_TAG}/" ./charts/kaito/workspace/values.yaml
	@sed -i -e 's/IMG_TAG=.*/IMG_TAG=${IMG_TAG}/' ./charts/kaito/workspace/README.md
	git checkout -b release-${VERSION}
	git add ./Makefile ./charts/kaito/workspace/Chart.yaml ./charts/kaito/workspace/values.yaml ./charts/kaito/workspace/README.md
	git commit -s -m "release: update manifest and helm charts for ${VERSION}"

## --------------------------------------
## Cleanup
## --------------------------------------

.PHONY: clean
clean:
	@rm -rf $(BIN_DIR)

# Installation using AWS EKS cluster

Before you begin, ensure you have the following tools installed:
- [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) to provision AWS resources
- [Eksctl](https://eksctl.io/installation/) (>= v0.191.0) to create and manage clusters on EKS
- [Helm](https://helm.sh) to install this operator
- [kubectl](https://kubernetes.io/docs/tasks/tools/) to view Kubernetes resources

## Create EKS Cluster
If you do not already have an EKS cluster, run the following to create one:

```bash
cd ../.. #go back to main directory to use MAKE commands

export CLUSTER_NAME = kaito-aws
export AWS_REGION = us-west-2
export KARPENTER_NAMESPACE = karpenter
export AWS_ACCOUNT_ID="$(aws sts get-caller-identity --query Account --output text)"

make deploy-aws-cloudformation
make create-eks-cluster
```

If you already have an EKS cluster, connect to it using
```bash
aws eks update-kubeconfig --name $CLUSTER_NAME --region $AWS_REGION
```

## Install Karpenter Controller
```bash
make aws-karpenter-helm
```

## Install Workspace Controller
```bash
make aws-patch-install-helm
```

## Verify installation
You can run the following commands to verify the installation of the controllers were successful.

Check status of the Helm chart installations.

```bash
helm list -n default
```

Check status of `workspace`.

```bash
kubectl describe deploy workspace -n kaito-workspace
```

Check status of `karpenter`.

```bash
kubectl describe deploy karpenter -n $KARPENTER_NAMESPACE
```

## Create a Workspace and start an inference service
Once the Kaito and Karpenter controllers are installed, you can follow [these instructions](https://github.com/kaito-project/kaito/tree/main?tab=readme-ov-file#quick-start) to start an inference service. Just update the `resource.InstanceType` to a supported [AWS gpu sku](../../pkg/sku/aws_sku_handler.go).
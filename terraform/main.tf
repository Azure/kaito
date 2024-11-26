terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "=4.5.0"
    }

    random = {
      source  = "hashicorp/random"
      version = "=3.6.3"
    }

    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "=2.33.0"
    }

    helm = {
      source  = "hashicorp/helm"
      version = "=2.16.1"
    }
  }
}

provider "azurerm" {
  features {
    resource_group {
      prevent_deletion_if_contains_resources = false
    }
  }
}

provider "kubernetes" {
  host                   = azurerm_kubernetes_cluster.example.kube_config.0.host
  username               = azurerm_kubernetes_cluster.example.kube_config.0.username
  password               = azurerm_kubernetes_cluster.example.kube_config.0.password
  client_certificate     = base64decode(azurerm_kubernetes_cluster.example.kube_config.0.client_certificate)
  client_key             = base64decode(azurerm_kubernetes_cluster.example.kube_config.0.client_key)
  cluster_ca_certificate = base64decode(azurerm_kubernetes_cluster.example.kube_config.0.cluster_ca_certificate)
}

provider "helm" {
  kubernetes {
    host                   = azurerm_kubernetes_cluster.example.kube_config.0.host
    username               = azurerm_kubernetes_cluster.example.kube_config.0.username
    password               = azurerm_kubernetes_cluster.example.kube_config.0.password
    client_certificate     = base64decode(azurerm_kubernetes_cluster.example.kube_config.0.client_certificate)
    client_key             = base64decode(azurerm_kubernetes_cluster.example.kube_config.0.client_key)
    cluster_ca_certificate = base64decode(azurerm_kubernetes_cluster.example.kube_config.0.cluster_ca_certificate)
  }
}

data "azurerm_client_config" "current" {}

resource "random_integer" "example" {
  min = 10
  max = 99
}

resource "azurerm_resource_group" "example" {
  name     = "rg-${local.random_name}"
  location = var.location
}

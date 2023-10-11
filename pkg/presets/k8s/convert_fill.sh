#!/bin/bash

# Ensure the correct number of arguments are provided
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <acr-name.azurecr.io/repository-name:tag>"
    exit 1
fi

INPUT=$1

# Extract the ACR name, repository name, and tag from the input
ACR_NAME=$(echo "$INPUT" | cut -d. -f1)
REPO_SHORT_NAME=$(echo "$INPUT" | cut -d/ -f2 | cut -d: -f1)
TAG=$(echo "$INPUT" | cut -d: -f2)

# Construct full repository name
FULL_REPO_NAME="$ACR_NAME.azurecr.io/$REPO_SHORT_NAME"

# Fetch the digest using the 'az' CLI
DIGEST=$(az acr repository show-manifests --name $ACR_NAME --repository $REPO_SHORT_NAME --query "[?tags[? @ == '$TAG']].digest" -o tsv)

# Fetch the username and password
USERNAME=$(az acr credential show --name $ACR_NAME --query "username" -o tsv)
PASSWORD=$(az acr credential show --name $ACR_NAME --query "passwords[0].value" -o tsv)

# Combine USERNAME:PASSWORD for the arg
CREDENTIAL="$USERNAME:$PASSWORD"

# Use sed to replace placeholders; we'll use a different delimiter (|) to avoid conflicts with potential special characters in the variables
sed -e "s|<REPO_NAME>|${FULL_REPO_NAME}|g" -e "s|<DIGEST>|${DIGEST}|g" -e "s|<USERNAME>|${CREDENTIAL}|g" convert_template.yaml > convert_filled.yaml

echo "Filled YAML saved as convert_filled.yaml"

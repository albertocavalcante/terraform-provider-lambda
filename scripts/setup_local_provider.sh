#!/bin/bash

# Setup script for local Terraform provider development
set -e

echo "🔧 Setting up local Lambda Cloud Terraform Provider..."

# Build the provider
echo "📦 Building provider..."
go build -o terraform-provider-lambda

# Make it executable
chmod +x terraform-provider-lambda

# Check if we're on macOS (darwin) and determine architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Convert architecture names to Terraform conventions
if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
fi

echo "🖥️  Detected platform: ${OS}_${ARCH}"

# Create the local provider directory structure
PROVIDER_DIR="$HOME/.terraform.d/plugins/github.com/albertocavalcante/lambda/0.1.0/${OS}_${ARCH}"
echo "📁 Creating provider directory: $PROVIDER_DIR"
mkdir -p "$PROVIDER_DIR"

# Copy the built provider to the plugins directory
echo "📋 Installing provider binary..."
cp terraform-provider-lambda "$PROVIDER_DIR/"

# Set up environment for development
echo "🔧 Setting up development environment..."

# Check if TF_CLI_CONFIG_FILE is already set
if [ -z "$TF_CLI_CONFIG_FILE" ]; then
    export TF_CLI_CONFIG_FILE="$PWD/examples/.terraformrc"
    echo "export TF_CLI_CONFIG_FILE=\"$PWD/examples/.terraformrc\"" >> ~/.bashrc 2>/dev/null || true
    echo "export TF_CLI_CONFIG_FILE=\"$PWD/examples/.terraformrc\"" >> ~/.zshrc 2>/dev/null || true
    echo "✅ Set TF_CLI_CONFIG_FILE=$TF_CLI_CONFIG_FILE"
else
    echo "ℹ️  TF_CLI_CONFIG_FILE already set to: $TF_CLI_CONFIG_FILE"
fi

echo ""
echo "🚀 Local provider setup complete!"
echo ""
echo "To test the provider:"
echo "1. Set your API key: export LAMBDA_CLOUD_API_KEY=your-api-key"
echo "2. Navigate to examples: cd examples"
echo "3. Initialize Terraform: terraform init"
echo "4. Test with plan: terraform plan -var-file=terraform.tfvars || terraform plan"
echo ""
echo "For development testing without API key:"
echo "1. cd examples"
echo "2. terraform init -upgrade"
echo "3. terraform plan -target=null_resource.test || terraform validate"
echo ""
echo "Provider installed at: $PROVIDER_DIR/terraform-provider-lambda"

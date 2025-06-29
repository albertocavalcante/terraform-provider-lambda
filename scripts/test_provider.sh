#!/bin/bash

# Test script for Lambda Cloud Terraform Provider
set -e

echo "🔧 Building Terraform Provider..."
go build -o terraform-provider-lambda

echo "📁 Setting up test environment..."
# Auto-detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then ARCH="amd64"; fi

mkdir -p ~/.terraform.d/plugins/github.com/albertocavalcante/lambda/0.1.0/"${OS}"_"${ARCH}"/
cp terraform-provider-lambda ~/.terraform.d/plugins/github.com/albertocavalcante/lambda/0.1.0/"${OS}"_"${ARCH}"/

echo "🔑 Checking for API key..."
if [ -z "$LAMBDA_CLOUD_API_KEY" ]; then
    echo "⚠️  LAMBDA_CLOUD_API_KEY environment variable not set"
    echo "   You can set it with: export LAMBDA_CLOUD_API_KEY=your-api-key"
    echo "   Or configure it in the provider block in your Terraform configuration"
else
    echo "✅ API key found in environment"
fi

echo "📋 Terraform configuration example:"
echo "---"
cat examples/main.tf
echo "---"

echo ""
echo "🚀 Provider is ready! To test:"
echo "1. Get your API key from: https://cloud.lambda.ai/api-keys/cloud-api"
echo "2. Set environment variable: export LAMBDA_CLOUD_API_KEY=your-api-key"
echo "3. Run: cd examples && terraform init && terraform plan"
echo ""
echo "Provider Features Implemented:"
echo "✅ API Key Authentication (Bearer token)"
echo "✅ Environment variable support (LAMBDA_CLOUD_API_KEY)"
echo "✅ Instance resource with full CRUD operations"
echo "✅ SSH key resource (basic structure)"
echo "✅ Instance types data source (basic structure)"
echo ""
echo "Ready for Lambda Cloud GPU instance management! 🚀"

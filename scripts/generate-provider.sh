#!/bin/bash

# Script to generate the complete Lambda Cloud Terraform Provider
# This script handles the full generation workflow from OpenAPI spec to Go code

set -e

OPENAPI_FILE="openapi.json"
CONFIG_FILE="generator_config.yml"
PROVIDER_SPEC_FILE="provider_code_spec.json"
CODEGEN_DIR="codegen"

echo "🚀 Lambda Cloud Terraform Provider Generator"
echo "============================================"

# Check prerequisites
echo "🔍 Checking prerequisites..."

# Install pinned tools using Go 1.24 native tools directive from tools.mod
echo "📦 Installing tools using Go 1.24 native tool management with tools.mod..."
# Tools are already defined in tools.mod, just ensure they're available
go mod download -modfile=tools.mod

# Display tool versions for reproducibility
echo "✅ All prerequisites found"
echo "📋 Tool versions:"
echo "   • tfplugingen-openapi: $(tfplugingen-openapi version 2>/dev/null || echo 'v0.4.0 (pinned)')"
echo "   • tfplugingen-framework: $(tfplugingen-framework version 2>/dev/null || echo 'v0.2.0 (pinned)')"
echo "   • go: $(go version | cut -d' ' -f3)"

# Fetch OpenAPI spec if it doesn't exist
if [ ! -f "$OPENAPI_FILE" ]; then
    echo "📥 OpenAPI spec not found, fetching..."
    ./scripts/fetch-openapi-spec.sh
else
    echo "📋 Using existing OpenAPI spec: $OPENAPI_FILE"
    echo "💡 Run './scripts/fetch-openapi-spec.sh' to update to latest version"
fi

# Verify required files exist
if [ ! -f "$CONFIG_FILE" ]; then
    echo "❌ Generator config not found: $CONFIG_FILE"
    echo "   This file should define the provider configuration"
    exit 1
fi

echo "✅ All required files present"

# Generate provider code specification
echo ""
echo "🔄 Step 1: Generating provider code specification..."
go run -modfile=tools.mod github.com/hashicorp/terraform-plugin-codegen-openapi/cmd/tfplugingen-openapi generate \
    --config "$CONFIG_FILE" \
    --output "$PROVIDER_SPEC_FILE" \
    "$OPENAPI_FILE"

if [ ! -f "$PROVIDER_SPEC_FILE" ]; then
    echo "❌ Failed to generate provider code specification"
    exit 1
fi

echo "✅ Provider code specification generated: $PROVIDER_SPEC_FILE"

# Generate Go framework code
echo ""
echo "🔄 Step 2: Generating Go provider framework code..."

# Clean previous generated code
if [ -d "$CODEGEN_DIR" ]; then
    echo "🧹 Cleaning previous generated code..."
    rm -rf "$CODEGEN_DIR"
fi

go run -modfile=tools.mod github.com/hashicorp/terraform-plugin-codegen-framework/cmd/tfplugingen-framework generate all \
    --input "$PROVIDER_SPEC_FILE" \
    --output "$CODEGEN_DIR/"

if [ ! -d "$CODEGEN_DIR" ]; then
    echo "❌ Failed to generate Go framework code"
    exit 1
fi

echo "✅ Go framework code generated in: $CODEGEN_DIR/"

# Build the provider
echo ""
echo "🔄 Step 3: Building provider binary..."
go build -o terraform-provider-lambda

if [ ! -f "terraform-provider-lambda" ]; then
    echo "❌ Failed to build provider binary"
    exit 1
fi

echo "✅ Provider binary built: terraform-provider-lambda"

# Summary
echo ""
echo "🎉 Provider generation completed successfully!"
echo ""
echo "📊 Generated files:"
echo "   📄 $PROVIDER_SPEC_FILE - Provider specification"
echo "   📁 $CODEGEN_DIR/ - Generated Go framework code" 
echo "   🔧 terraform-provider-lambda - Provider binary"
echo ""
echo "🚀 Next steps:"
echo "   1. Set up local development: ./scripts/setup-local-dev.sh"
echo "   2. Test the provider: cd examples && terraform init && terraform plan"
echo "   3. Or install globally: ./scripts/install-provider.sh"

# Show warnings about ignored API features
echo ""
echo "⚠️  Note: Some complex API features are ignored in the current implementation:"
echo "   • Image specification (defaults to latest Lambda Stack)"
echo "   • User data / cloud-init configuration"  
echo "   • Complex filesystem mount configurations"
echo "   • Instance tags and metadata"
echo ""
echo "💡 These can be added incrementally as the provider matures"
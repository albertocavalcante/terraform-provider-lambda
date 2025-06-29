#!/bin/bash

# Script to set up local development environment for Lambda Cloud Terraform Provider
# This configures Terraform to use the locally built provider binary

set -e

PROVIDER_BINARY="terraform-provider-lambda"
PROVIDER_SOURCE="github.com/albertocavalcante/lambda"
PROVIDER_VERSION="0.1.0"

echo "🔧 Setting up Lambda Cloud Terraform Provider for local development"
echo "=================================================================="

# Check if provider binary exists
if [ ! -f "$PROVIDER_BINARY" ]; then
    echo "❌ Provider binary not found: $PROVIDER_BINARY"
    echo "   Run './scripts/generate-provider.sh' first to build the provider"
    exit 1
fi

# Detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Convert architecture names to Terraform conventions
if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
fi

echo "🖥️  Detected platform: ${OS}_${ARCH}"

# Create provider directory structure
PROVIDER_DIR="$HOME/.terraform.d/plugins/$PROVIDER_SOURCE/$PROVIDER_VERSION/${OS}_${ARCH}"
echo "📁 Creating provider directory: $PROVIDER_DIR"
mkdir -p "$PROVIDER_DIR"

# Copy provider binary with correct naming
PROVIDER_FILENAME="terraform-provider-lambda_v$PROVIDER_VERSION"
echo "📋 Installing provider binary as: $PROVIDER_FILENAME"
cp "$PROVIDER_BINARY" "$PROVIDER_DIR/$PROVIDER_FILENAME"
chmod +x "$PROVIDER_DIR/$PROVIDER_FILENAME"

# Create Terraform CLI configuration for development
TERRAFORMRC_FILE="examples/.terraformrc"
echo "⚙️  Creating Terraform CLI configuration: $TERRAFORMRC_FILE"

mkdir -p examples
cat > "$TERRAFORMRC_FILE" << EOF
provider_installation {
  filesystem_mirror {
    path    = "$HOME/.terraform.d/plugins"
    include = ["$PROVIDER_SOURCE"]
  }
  direct {
    exclude = ["$PROVIDER_SOURCE"]
  }
}
EOF

# Create example .envrc for direnv users (optional)
ENVRC_FILE="examples/.envrc"
if command -v direnv >/dev/null 2>&1; then
    echo "📝 Creating .envrc for direnv users: $ENVRC_FILE"
    cat > "$ENVRC_FILE" << EOF
# Lambda Cloud API Key
# Get your API key from: https://cloud.lambda.ai/api-keys/cloud-api
export LAMBDA_CLOUD_API_KEY=your-api-key-here

# Use local Terraform CLI configuration
export TF_CLI_CONFIG_FILE="\$PWD/.terraformrc"

echo "🔑 Lambda Cloud API key: \${LAMBDA_CLOUD_API_KEY:0:10}..."
echo "⚙️  Using local provider configuration"
EOF
    echo "💡 Add your API key to examples/.envrc and run 'direnv allow' if you use direnv"
else
    echo "💡 Consider installing 'direnv' for automatic environment variable management"
fi

echo ""
echo "✅ Local development environment set up successfully!"
echo ""
echo "📋 Configuration:"
echo "   🔗 Provider source: $PROVIDER_SOURCE"
echo "   📦 Version: $PROVIDER_VERSION"
echo "   📁 Installed at: $PROVIDER_DIR/$PROVIDER_FILENAME"
echo "   ⚙️  CLI config: $TERRAFORMRC_FILE"
echo ""
echo "🚀 To test the provider:"
echo "   1. Set your API key:"
echo "      export LAMBDA_CLOUD_API_KEY=your-api-key"
echo "   2. Configure Terraform CLI:"
echo "      export TF_CLI_CONFIG_FILE=\"\$(pwd)/examples/.terraformrc\""
echo "   3. Test the provider:"
echo "      cd examples"
echo "      terraform init"
echo "      terraform plan"
echo ""
echo "🔄 To update the provider:"
echo "   1. Rebuild: ./scripts/generate-provider.sh"
echo "   2. Reinstall: ./scripts/setup-local-dev.sh"
echo ""
echo "📚 Documentation:"
echo "   • Provider source in main.tf should use: source = \"$PROVIDER_SOURCE\""
echo "   • API key can be set via LAMBDA_CLOUD_API_KEY environment variable"
echo "   • Examples are in the examples/ directory"
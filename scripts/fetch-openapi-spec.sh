#!/bin/bash

# Script to fetch the latest Lambda Cloud OpenAPI specification
# This ensures we always work with the most current API specification

set -e

SPEC_URL="https://cloud.lambda.ai/api/v1/openapi.json"
OUTPUT_FILE="openapi.json"
VERSION_FILE="openapi.version"

echo "🔄 Fetching Lambda Cloud OpenAPI specification..."
echo "📡 URL: $SPEC_URL"

# Download the OpenAPI spec
if command -v curl >/dev/null 2>&1; then
    curl -L -o "$OUTPUT_FILE" "$SPEC_URL"
elif command -v wget >/dev/null 2>&1; then
    wget -O "$OUTPUT_FILE" "$SPEC_URL"
else
    echo "❌ Error: Neither curl nor wget is available"
    echo "   Please install curl or wget to download the OpenAPI specification"
    exit 1
fi

# Verify the downloaded file
if [ ! -f "$OUTPUT_FILE" ]; then
    echo "❌ Error: Failed to download OpenAPI specification"
    exit 1
fi

# Check if it's valid JSON
if command -v jq >/dev/null 2>&1; then
    if ! jq empty "$OUTPUT_FILE" 2>/dev/null; then
        echo "❌ Error: Downloaded file is not valid JSON"
        rm -f "$OUTPUT_FILE"
        exit 1
    fi
    
    # Extract and display API version info
    API_VERSION=$(jq -r '.info.version // "unknown"' "$OUTPUT_FILE" 2>/dev/null)
    API_TITLE=$(jq -r '.info.title // "unknown"' "$OUTPUT_FILE" 2>/dev/null)
    DOWNLOAD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    # Create version metadata file for reproducibility
    cat > "$VERSION_FILE" << EOF
{
  "api_title": "$API_TITLE",
  "api_version": "$API_VERSION", 
  "download_date": "$DOWNLOAD_DATE",
  "source_url": "$SPEC_URL"
}
EOF
    
    echo "✅ Successfully downloaded OpenAPI specification"
    echo "📋 API: $API_TITLE"
    echo "🔖 Version: $API_VERSION"
    echo "📅 Downloaded: $DOWNLOAD_DATE"
    echo "📁 Saved to: $OUTPUT_FILE"
    echo "📄 Version info: $VERSION_FILE"
else
    echo "✅ Successfully downloaded OpenAPI specification"
    echo "📁 Saved to: $OUTPUT_FILE"
    echo "💡 Install 'jq' for validation and version information"
fi

echo ""
echo "🚀 Ready to generate provider code!"
echo "   Next steps:"
echo "   1. Run: ./scripts/generate-provider.sh"
echo "   2. Or manually: tfplugingen-openapi generate --config generator_config.yml --output provider_code_spec.json openapi.json"
#!/bin/bash

# Script to configure GitHub username in all files for a new user
set -e

if [ $# -eq 0 ]; then
    echo "Usage: $0 <github-username>"
    echo ""
    echo "This script replaces albertocavalcante placeholders with your actual GitHub username"
    echo "throughout the repository files."
    echo ""
    echo "Example:"
    echo "  $0 myusername"
    exit 1
fi

USERNAME="$1"

echo "🔧 Configuring repository for GitHub username: $USERNAME"
echo "========================================================="

# Files to update
FILES=(
    "README.md"
    "examples/main.tf"
    "scripts/setup-local-dev.sh"
    "scripts/setup_local_provider.sh"
    "scripts/test_provider.sh"
    "codegen/README.md"
)

# Update each file
for file in "${FILES[@]}"; do
    if [ -f "$file" ]; then
        echo "📝 Updating $file..."
        sed -i.bak "s/albertocavalcante/$USERNAME/g" "$file"
        rm "$file.bak"
        echo "✅ Updated $file"
    else
        echo "⚠️  File not found: $file"
    fi
done

echo ""
echo "🎉 Repository configured successfully!"
echo ""
echo "Summary of changes:"
echo "- Provider source: github.com/$USERNAME/lambda"
echo "- Repository URL: https://github.com/$USERNAME/terraform-provider-lambda"
echo "- Import paths updated for Go modules"
echo ""
echo "🚀 Next steps:"
echo "1. Update go.mod module name if needed: go mod edit -module github.com/$USERNAME/terraform-provider-lambda"
echo "2. Update imports in Go files to match new module name"
echo "3. Run: ./scripts/generate-provider.sh"
echo "4. Run: ./scripts/setup-local-dev.sh"
echo ""
echo "💡 Remember to update the module name in go.mod and Go import statements manually!"

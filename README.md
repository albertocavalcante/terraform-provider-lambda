# Terraform Provider for Lambda GPU Cloud

A Terraform provider for managing [Lambda GPU Cloud](https://cloud.lambda.ai/) resources, generated from their OpenAPI specification.

## Features

- **Instance Management**: Launch, read, and terminate GPU instances
- **Instance Types**: Query available instance types with pricing and specifications
- **SSH Key Management**: Manage SSH keys for instance access
- **API Authentication**: Secure API key-based authentication

## Quick Start

### 1. Get Your API Key

Visit [Lambda Cloud API Keys](https://cloud.lambda.ai/api-keys/cloud-api) to generate your API key.

### 2. Configure the Provider

```hcl
terraform {
  required_providers {
    lambda = {
      source  = "github.com/albertocavalcante/lambda"
      version = "~> 0.1"
    }
  }
}

provider "lambda" {
  api_key = var.lambda_api_key  # or set LAMBDA_CLOUD_API_KEY environment variable
}
```

### 3. Query Available Instance Types

```hcl
data "lambda_instance_types" "available" {}

output "instance_types" {
  value = data.lambda_instance_types.available.instance_types
}
```

### 4. Launch an Instance

```hcl
resource "lambda_instance" "my_gpu" {
  name           = "my-gpu-instance"
  instance_type  = "gpu_1x_a10"
  region         = "us-west-2"
  ssh_key_names  = ["my-ssh-key"]
}
```

## Authentication

The provider supports authentication via:

1. **Environment Variable** (recommended):
   ```bash
   export LAMBDA_CLOUD_API_KEY="your-api-key-here"
   ```

2. **Provider Configuration**:
   ```hcl
   provider "lambda" {
     api_key = "your-api-key-here"
   }
   ```

## Resources

### `lambda_instance`

Manages Lambda GPU Cloud instances.

```hcl
resource "lambda_instance" "example" {
  name           = "my-instance"
  instance_type  = "gpu_1x_a10"
  region         = "us-west-2"
  ssh_key_names  = ["my-key"]
  quantity       = 1
}
```

**Arguments:**
- `name` (Required) - Instance name
- `instance_type` (Required) - Instance type (see data source for available types)
- `region` (Required) - Region to launch in
- `ssh_key_names` (Required) - List of SSH key names
- `quantity` (Optional) - Number of instances to launch (default: 1)

**Attributes:**
- `id` - Instance ID
- `status` - Current instance status
- `ip` - Instance IP address

## Data Sources

### `lambda_instance_types`

Retrieves available instance types with pricing and specifications.

```hcl
data "lambda_instance_types" "all" {}

output "gpu_instances" {
  value = [
    for instance_type in data.lambda_instance_types.all.instance_types :
    instance_type if can(regex("gpu", instance_type.name))
  ]
}
```

**Attributes:**
- `instance_types` - List of available instance types with:
  - `name` - Instance type name
  - `price_cents_per_hour` - Hourly price in cents
  - `description` - Instance description
  - `specs` - Hardware specifications

## Development

### Prerequisites

- [Go](https://golang.org/) 1.24+ (required for latest features)
- [Terraform](https://terraform.io/) 1.6+
- [HashiCorp Plugin Codegen Tools](https://github.com/hashicorp/terraform-plugin-codegen-openapi) (auto-installed via pinned versions)

### Install Codegen Tools

Tools are automatically installed with pinned versions when you run the generation script:

```bash
# Tools are automatically installed from tools.mod with pinned versions
./scripts/generate-provider.sh
```

Or run tools manually using Go 1.24 with separate tools.mod:
```bash
# Download tool dependencies
go mod download -modfile=tools.mod

# Run code generation tools directly with -modfile flag
go run -modfile=tools.mod github.com/hashicorp/terraform-plugin-codegen-openapi/cmd/tfplugingen-openapi --help
go run -modfile=tools.mod github.com/hashicorp/terraform-plugin-codegen-framework/cmd/tfplugingen-framework --help

# Run golangci-lint (install separately due to dependency complexity)
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.2.0
golangci-lint run --timeout=5m
```

### Setup Development Environment

1. **Clone the repository**:
   ```bash
   git clone https://github.com/albertocavalcante/terraform-provider-lambda.git
   cd terraform-provider-lambda
   ```

2. **Fetch the latest OpenAPI specification**:
   ```bash
   ./scripts/fetch-openapi-spec.sh
   ```

3. **Generate and build the provider**:
   ```bash
   ./scripts/generate-provider.sh
   ```

4. **Set up local development**:
   ```bash
   ./scripts/setup-local-dev.sh
   ```

5. **Test the provider**:
   ```bash
   export LAMBDA_CLOUD_API_KEY="your-api-key"
   cd examples
   terraform init
   terraform plan
   ```

### Updating the Provider

The provider is generated from Lambda Cloud's OpenAPI specification. To update:

1. **Fetch the latest API spec**:
   ```bash
   ./scripts/fetch-openapi-spec.sh
   ```

2. **Regenerate the provider**:
   ```bash
   ./scripts/generate-provider.sh
   ```

3. **Test the changes**:
   ```bash
   ./scripts/setup-local-dev.sh
   cd examples && terraform plan
   ```

### Reproducible Builds

This repository uses pinned tool versions for reproducible builds:

- **Tool Versions**: Defined in `tools.mod` (separate from main dependencies)
- **API Versioning**: OpenAPI spec includes version metadata in `openapi.version`
- **Generated Code**: Committed to git for visibility and reproducibility

### Manual OpenAPI Specification Download

If the script doesn't work, download manually:

```bash
curl -L -o openapi.json https://cloud.lambda.ai/api/v1/openapi.json
```

### Generator Configuration

The provider generation is controlled by `generator_config.yml`. Key settings:

- **Resource Mappings**: Maps API endpoints to Terraform resources
- **Field Ignores**: Complex fields that can't be auto-generated
- **Aliases**: Maps API field names to Terraform attribute names

### Directory Structure

```
terraform-provider-lambda/
├── scripts/                    # Development and build scripts
│   ├── fetch-openapi-spec.sh  # Download latest OpenAPI spec
│   ├── generate-provider.sh   # Complete provider generation
│   ├── setup-local-dev.sh     # Local development setup
│   ├── setup_local_provider.sh # Alternative local setup (legacy)
│   └── test_provider.sh       # Quick test script
├── internal/provider/          # Custom provider implementation
├── codegen/                    # Auto-generated framework code (committed)
├── examples/                   # Example Terraform configurations
├── .github/workflows/          # CI/CD workflows
├── tools.mod                   # Go 1.24 native tool management (separate from main deps)
├── generator_config.yml        # OpenAPI generator configuration
└── go.mod                      # Go module definition
```

## Provider Generation Process

This provider is automatically generated from Lambda Cloud's OpenAPI specification:

1. **OpenAPI Spec** → `tfplugingen-openapi` → **Provider Spec JSON**
2. **Provider Spec JSON** → `tfplugingen-framework` → **Go Provider Code**
3. **Go Provider Code** → `go build` → **Provider Binary**

## Known Limitations

Some complex API features are not yet implemented:

- **Image Specification**: Currently uses default Lambda Stack
- **User Data**: Cloud-init configuration not supported
- **Complex Filesystem Mounts**: Advanced storage configurations
- **Instance Tags**: Metadata and tagging not implemented

These features can be added incrementally as the provider matures.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes to the generator configuration or custom provider code
4. Test with `./scripts/generate-provider.sh && ./scripts/setup-local-dev.sh`
5. Submit a pull request

## Resources

- [Lambda Cloud API v1.7.0](https://cloud.lambda.ai/api/v1/docs)
- [HashiCorp OpenAPI Generator](https://developer.hashicorp.com/terraform/plugin/code-generation/openapi-generator)
- [Terraform Provider Development](https://developer.hashicorp.com/terraform/plugin)

## License

This project is licensed under the MPL-2.0 License - see the [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/albertocavalcante/terraform-provider-lambda/issues)
- **Lambda Cloud API**: [API Documentation](https://cloud.lambda.ai/api/v1/docs)
- **Terraform**: [Provider Development Guide](https://developer.hashicorp.com/terraform/plugin)

# CLAUDE.md

This document provides comprehensive guidance for developing the Lambda GPU Cloud Terraform Provider.
It is designed specifically for Claude AI assistance but serves as a complete developer reference.

## Project Overview

**Lambda GPU Cloud Terraform Provider** is a code-generated Terraform provider for managing Lambda GPU Cloud
infrastructure. Lambda GPU Cloud provides on-demand GPU compute instances for machine learning workloads
(completely separate from AWS Lambda).

### Architecture Philosophy
This provider is **entirely code-generated** from Lambda Cloud's OpenAPI specification using HashiCorp's
official toolchain. This approach ensures:
- **API Consistency**: Provider schema automatically matches Lambda Cloud's API
- **Reduced Maintenance**: Updates require only regenerating from new OpenAPI specs
- **Type Safety**: Go types are derived directly from API specifications
- **Terraform Best Practices**: Generated code follows HashiCorp's framework patterns

### Code Generation Pipeline
```
Lambda Cloud OpenAPI 3.1.0 Spec
    ↓ (tfplugingen-openapi)
Provider Code Specification (JSON)
    ↓ (tfplugingen-framework)  
Go Framework Code (codegen/)
    ↓ (custom implementations)
Working Terraform Provider
```

## Essential Commands

### Complete Development Setup
```bash
# Full provider generation and local setup
./scripts/generate-provider.sh && ./scripts/setup-local-dev.sh

# Quick testing
export LAMBDA_CLOUD_API_KEY=your-api-key
cd examples && terraform init && terraform plan
```

### Manual Development Workflow
```bash
# 1. Fetch latest OpenAPI specification
./scripts/fetch-openapi-spec.sh

# 2. Generate provider specification from OpenAPI
tfplugingen-openapi generate \
  --config generator_config.yml \
  --output provider_code_spec.json \
  openapi.json

# 3. Generate Go framework code
tfplugingen-framework generate all \
  --input provider_code_spec.json \
  --output codegen/

# 4. Build provider binary
go build -o terraform-provider-lambda

# 5. Install for local development
./scripts/setup-local-dev.sh
```

### Tool Management (Go 1.24 Native)
```bash
# Install tools using Go 1.24 native tool directives
cd tools
go get -tool github.com/hashicorp/terraform-plugin-codegen-openapi/cmd/tfplugingen-openapi@v0.3.0
go get -tool github.com/hashicorp/terraform-plugin-codegen-framework/cmd/tfplugingen-framework@v0.4.1

# Run tools using go tool command
go tool tfplugingen-openapi --help
go tool tfplugingen-framework --help
```

## Repository Structure

```
terraform-provider-lambda/
├── generator_config.yml           # OpenAPI → Terraform resource mappings
├── codegen/                       # Generated Go framework code (COMMITTED)
│   ├── datasource_instance_types/
│   ├── provider_lambda/
│   ├── resource_instance/
│   └── resource_ssh_key/
├── internal/provider/             # Custom implementations using generated code
│   ├── provider.go               # Main provider with authentication
│   ├── data_source_instance_types.go  # Instance types data source
│   ├── resource_instance.go      # GPU instance resource
│   └── resource_ssh_key.go       # SSH key resource
├── scripts/                      # All automation scripts
│   ├── generate-provider.sh     # Complete generation workflow
│   ├── setup-local-dev.sh       # Local Terraform filesystem mirror setup
│   └── fetch-openapi-spec.sh    # Download latest API specification
├── tools/                         # Go 1.24 native tool management
│   └── go.mod                    # Tool dependencies with native tool directives
└── examples/                     # Working Terraform configurations
```

## Current Implementation Status

### ✅ Implemented Resources
- **`lambda_instance`**: Complete CRUD for GPU instances (launch/terminate/read)
- **`lambda_ssh_key`**: SSH key management for instance access
- **`lambda_instance_types`** (data source): Query available instance types with pricing

### ✅ Working Features
- **Authentication**: API key via environment variable `LAMBDA_CLOUD_API_KEY`
- **Real API Integration**: Successfully tested against Lambda Cloud production API
- **Local Development**: Terraform filesystem mirror for local provider testing
- **Reproducible Builds**: Pinned tool versions with Go 1.24 tool directives

### 🚧 Known Limitations
- **Image Specification**: Complex `anyOf` schemas ignored - uses default Lambda Stack
- **User Data**: Cloud-init configuration not yet supported
- **Instance Tags**: Metadata tagging not implemented
- **File System Mounts**: Advanced storage configurations not supported

## Technical Implementation Details

### Code Generation Challenges Solved

**Non-Standard CRUD Patterns**: Lambda Cloud uses separate endpoints for instance lifecycle:
```yaml
# In generator_config.yml
resources:
  instance:
    create:
      path: /api/v1/instance-operations/launch    # Not /instances
      method: POST
    delete:
      path: /api/v1/instance-operations/terminate # Not DELETE /instances/{id}
      method: POST
```

**Complex Schema Handling**: OpenAPI generator fails on `anyOf`/`oneOf` schemas:
```yaml
# Skip fields that break generation
ignores:
  - image                # Complex multi-type schema
  - user_data           # Nested object variations
  - file_system_mounts  # Array of complex objects
```

**Field Mapping**: API parameter names don't match Terraform conventions:
```yaml
# Map path parameters to attributes
aliases:
  id: instance_id       # Terraform 'id' → API 'instance_id'
```

### Authentication Implementation
```go
// Bearer token authentication
func (c *ProviderConfig) AddAuthHeader(req *http.Request) {
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.ApiKey))
}

// Environment variable support
if apiKey := os.Getenv("LAMBDA_CLOUD_API_KEY"); apiKey != "" {
    config.ApiKey = apiKey
}
```

### Local Development Pattern
The provider uses Terraform's filesystem mirror for local development:
```hcl
# .terraformrc
provider_installation {
  filesystem_mirror {
    path = "~/.terraform.d/plugins"
    include = ["github.com/albertocavalcante/lambda"]
  }
}
```

## Development Guidelines

### Code Generation Best Practices
1. **Always regenerate from OpenAPI spec** - Don't manually edit generated code
2. **Use `ignores` liberally** - Skip fields that break generation rather than fix manually
3. **Test each generation iteration** - Verify provider builds and basic functionality
4. **Commit generated code** - Enables code review and build reproducibility

### API Integration Patterns
1. **Environment-first authentication** - Always support `LAMBDA_CLOUD_API_KEY`
2. **Graceful error handling** - Lambda Cloud returns custom error formats
3. **Resource state management** - Some operations return async status, handle appropriately

### Contributing Guidelines
1. **Keep pull requests focused** - Single resource or feature per PR
2. **Test with real API** - Include integration testing with actual Lambda Cloud API
3. **Update documentation** - Keep README and examples current with new features
4. **Version tool dependencies** - Use `tools.mod` for reproducible builds

## API-Specific Notes

### Lambda Cloud API Characteristics
- **Authentication**: API key with Bearer token format
- **Base URL**: `https://cloud.lambda.ai/api/v1`
- **Rate Limiting**: Reasonable limits for development/testing
- **Regional Model**: Multi-region GPU infrastructure (us-west-2, us-east-1, etc.)

### Critical API Endpoints
```
GET  /api/v1/instance-types          # List available GPU instance types
POST /api/v1/instance-operations/launch    # Launch new instances
POST /api/v1/instance-operations/terminate # Terminate instances
GET  /api/v1/instances/{id}          # Get instance details
GET  /api/v1/ssh-keys               # List SSH keys
POST /api/v1/ssh-keys               # Create SSH key
```

### Testing Strategy
```bash
# Unit tests (when implemented)
go test ./internal/provider/...

# Integration testing with real API
export LAMBDA_CLOUD_API_KEY=your-key
cd examples && terraform plan  # Should show real instance types

# Acceptance testing
TF_ACC=1 go test ./internal/provider/... -v -timeout 120m
```

## Performance and Optimization

### Build Performance
- **Incremental builds**: Only regenerate when OpenAPI spec changes
- **Parallel execution**: Scripts support concurrent tool installation
- **Cached dependencies**: Go modules and tool binaries cached appropriately

### Runtime Performance
- **Minimal API calls**: Data sources cache results appropriately
- **Efficient schema**: Generated schemas use optimal Terraform types
- **Connection reuse**: HTTP client configured for connection pooling

## Critical Notes

⚠️ **DO NOT** manually edit files in `codegen/` - they will be overwritten
⚠️ **DO NOT** commit `openapi.json` - it should be downloaded fresh each time
⚠️ **DO** test with real API keys - the provider is designed for production use
⚠️ **DO** use the automation scripts - they handle complex setup correctly

## Common Issues and Solutions

### Generation Failures
```bash
# Problem: tfplugingen-openapi fails on complex schemas
# Solution: Add problematic fields to 'ignores' in generator_config.yml

# Problem: Module import errors in generated code  
# Solution: Verify go.mod module name matches repository structure
```

### Local Testing Issues
```bash
# Problem: Terraform can't find local provider
# Solution: Run ./scripts/setup-local-dev.sh and use TF_CLI_CONFIG_FILE

# Problem: Provider authentication fails
# Solution: Verify LAMBDA_CLOUD_API_KEY is set and valid
```

This provider represents a successful implementation of HashiCorp's code generation toolchain for a real-world cloud provider with non-standard API patterns.

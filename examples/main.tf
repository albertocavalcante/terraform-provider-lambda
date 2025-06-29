# Example configuration for Lambda Cloud provider

terraform {
  required_providers {
    lambda = {
      source  = "github.com/albertocavalcante/lambda"
      version = "0.1.0"
    }
  }
}

# Configure the Lambda provider
# The API key can be set via LAMBDA_CLOUD_API_KEY environment variable
provider "lambda" {
  # api_key is optional here if you use the environment variable
  # api_key = "your-api-key-here"  # Don't commit real API keys!

  # endpoint defaults to https://cloud.lambda.ai
  # endpoint = "https://cloud.lambda.ai"
}

# Example: Query available instance types
data "lambda_instance_types" "available" {}

# Example: Create an SSH key (commented out to avoid accidental execution)
# resource "lambda_ssh_key" "example" {
#   name = "my-terraform-key"
# }

# Example: Launch a GPU instance (commented out to avoid accidental execution)
# resource "lambda_instance" "gpu_workstation" {
#   name               = "ml-workstation"
#   region_name        = "us-west-2"
#   instance_type_name = "gpu_1x_a100"
#   ssh_key_names      = [lambda_ssh_key.example.name]
#
#   # Optional: Mount filesystems
#   # file_system_names = ["shared-storage"]
# }

# Output the available instance types for reference
output "instance_types" {
  value = data.lambda_instance_types.available
}

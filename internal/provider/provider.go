package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure LambdaProvider satisfies various provider interfaces.
var _ provider.Provider = &LambdaProvider{}
var _ provider.ProviderWithFunctions = &LambdaProvider{}

// LambdaProvider defines the provider implementation.
type LambdaProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// LambdaProviderModel describes the provider data model.
type LambdaProviderModel struct {
	ApiKey   types.String `tfsdk:"api_key"`
	Endpoint types.String `tfsdk:"endpoint"`
}

// ProviderConfig holds the configuration for API requests
type ProviderConfig struct {
	ApiKey     string
	Endpoint   string
	HTTPClient *http.Client
}

func (p *LambdaProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "lambda"
	resp.Version = p.version
}

func (p *LambdaProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Lambda Cloud API key. Can also be set via the LAMBDA_CLOUD_API_KEY environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Lambda Cloud API endpoint. Defaults to https://cloud.lambda.ai",
				Optional:            true,
			},
		},
	}
}

func (p *LambdaProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data LambdaProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override with Terraform configuration.
	apiKey := os.Getenv("LAMBDA_CLOUD_API_KEY")
	endpoint := "https://cloud.lambda.ai"

	if !data.ApiKey.IsNull() {
		apiKey = data.ApiKey.ValueString()
	}

	if !data.Endpoint.IsNull() {
		endpoint = data.Endpoint.ValueString()
	}

	// Validate required configuration
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Unable to find api_key",
			"api_key cannot be an empty string. "+
				"Set the api_key attribute in the provider configuration or use the LAMBDA_CLOUD_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create HTTP client with authentication
	client := &http.Client{}

	// Create provider configuration
	config := &ProviderConfig{
		ApiKey:     apiKey,
		Endpoint:   endpoint,
		HTTPClient: client,
	}

	// Make the configuration available to resources and data sources
	resp.DataSourceData = config
	resp.ResourceData = config
}

func (p *LambdaProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewInstanceResource,
		NewSshKeyResource,
	}
}

func (p *LambdaProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewInstanceTypesDataSource,
	}
}

func (p *LambdaProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &LambdaProvider{
			version: version,
		}
	}
}

// AddAuthHeader adds the Authorization header to HTTP requests
func (c *ProviderConfig) AddAuthHeader(req *http.Request) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.ApiKey))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
}

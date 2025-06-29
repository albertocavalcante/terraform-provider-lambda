package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewInstanceTypesDataSource() datasource.DataSource {
	return &InstanceTypesDataSource{}
}

type InstanceTypesDataSource struct {
	client *ProviderConfig
}

func (d *InstanceTypesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance_types"
}

func (d *InstanceTypesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch available Lambda Cloud instance types.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Data source identifier",
			},
			"instance_types": schema.MapNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Map of available instance types",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the instance type",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "A description of the instance type",
						},
						"gpu_description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Description of the GPUs in this instance type",
						},
						"price_cents_per_hour": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Price in US cents per hour",
						},
						"gpus": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "The number of GPUs",
						},
						"memory_gib": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "The amount of RAM in gibibytes (GiB)",
						},
						"storage_gib": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "The amount of storage in gibibytes (GiB)",
						},
						"vcpus": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "The number of virtual CPUs",
						},
					},
				},
			},
		},
	}
}

func (d *InstanceTypesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*ProviderConfig)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ProviderConfig, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// InstanceTypesDataSourceModel describes the data source data model.
type InstanceTypesDataSourceModel struct {
	Id            types.String `tfsdk:"id"`
	InstanceTypes types.Map    `tfsdk:"instance_types"`
}

// InstanceTypeData represents an instance type
type InstanceTypeData struct {
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	GpuDescription    types.String `tfsdk:"gpu_description"`
	PriceCentsPerHour types.Int64  `tfsdk:"price_cents_per_hour"`
	Gpus              types.Int64  `tfsdk:"gpus"`
	MemoryGib         types.Int64  `tfsdk:"memory_gib"`
	StorageGib        types.Int64  `tfsdk:"storage_gib"`
	Vcpus             types.Int64  `tfsdk:"vcpus"`
}

// API response structures
type InstanceTypesResponse struct {
	Data map[string]InstanceTypeAPIResponse `json:"data"`
}

type InstanceTypeAPIResponse struct {
	InstanceType InstanceTypeDetails `json:"instance_type"`
}

type InstanceTypeDetails struct {
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	GpuDescription    string            `json:"gpu_description"`
	PriceCentsPerHour int64             `json:"price_cents_per_hour"`
	Specs             InstanceTypeSpecs `json:"specs"`
}

type InstanceTypeSpecs struct {
	Gpus       int64 `json:"gpus"`
	MemoryGib  int64 `json:"memory_gib"`
	StorageGib int64 `json:"storage_gib"`
	Vcpus      int64 `json:"vcpus"`
}

func (d *InstanceTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data InstanceTypesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Make API call to get instance types
	httpReq, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/api/v1/instance-types", d.client.Endpoint), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request, got error: %s", err))
		return
	}

	d.client.AddAuthHeader(httpReq)

	httpResp, err := d.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read instance types, got error: %s", err))
		return
	}
	defer func() {
		if err := httpResp.Body.Close(); err != nil {
			tflog.Warn(ctx, "Failed to close response body", map[string]interface{}{"error": err})
		}
	}()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Instance types API returned status %d", httpResp.StatusCode))
		return
	}

	var apiResp InstanceTypesResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&apiResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse instance types response, got error: %s", err))
		return
	}

	// Convert API response to Terraform types
	instanceTypesMap := make(map[string]InstanceTypeData)
	for key, item := range apiResp.Data {
		instanceTypesMap[key] = InstanceTypeData{
			Name:              types.StringValue(item.InstanceType.Name),
			Description:       types.StringValue(item.InstanceType.Description),
			GpuDescription:    types.StringValue(item.InstanceType.GpuDescription),
			PriceCentsPerHour: types.Int64Value(item.InstanceType.PriceCentsPerHour),
			Gpus:              types.Int64Value(item.InstanceType.Specs.Gpus),
			MemoryGib:         types.Int64Value(item.InstanceType.Specs.MemoryGib),
			StorageGib:        types.Int64Value(item.InstanceType.Specs.StorageGib),
			Vcpus:             types.Int64Value(item.InstanceType.Specs.Vcpus),
		}
	}

	// Convert to Terraform Map type
	instanceTypesMapValue, diags := types.MapValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":                 types.StringType,
			"description":          types.StringType,
			"gpu_description":      types.StringType,
			"price_cents_per_hour": types.Int64Type,
			"gpus":                 types.Int64Type,
			"memory_gib":           types.Int64Type,
			"storage_gib":          types.Int64Type,
			"vcpus":                types.Int64Type,
		},
	}, instanceTypesMap)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set computed values
	data.Id = types.StringValue("instance_types")
	data.InstanceTypes = instanceTypesMapValue

	// Write logs using the tflog package
	tflog.Trace(ctx, "read instance types data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

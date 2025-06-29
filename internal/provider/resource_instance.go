package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &InstanceResource{}
var _ resource.ResourceWithImportState = &InstanceResource{}

func NewInstanceResource() resource.Resource {
	return &InstanceResource{}
}

// InstanceResource defines the resource implementation.
type InstanceResource struct {
	client *ProviderConfig
}

func (r *InstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

func (r *InstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	// Use the generated schema from the codegen
	// We'll need to import the generated package and use its schema
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Lambda Cloud GPU instance.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier (ID) of the instance",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "User-provided name for the instance (max 64 chars)",
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 64),
				},
			},
			"region_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Lambda Cloud region code where instance will be launched",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"europe-central-1", "asia-south-1", "australia-east-1",
						"me-west-1", "asia-northeast-1", "asia-northeast-2",
						"us-east-1", "us-west-2", "us-west-1", "us-south-1",
						"us-west-3", "us-midwest-1", "us-east-2", "us-south-2",
						"us-south-3", "us-east-3", "us-midwest-2",
						"test-east-1", "test-west-1",
					),
				},
			},
			"instance_type_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the instance type to launch",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ssh_key_names": schema.ListAttribute{
				ElementType:         types.StringType,
				Required:            true,
				MarkdownDescription: "List of SSH key names to add to the instance",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"file_system_names": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of filesystem names to mount to the instance",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"ip": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The public IP address of the instance",
			},
			"private_ip": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The private IP address of the instance",
			},
			"hostname": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The hostname of the instance",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The current status of the instance",
			},
		},
	}
}

func (r *InstanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*ProviderConfig)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ProviderConfig, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// InstanceModel describes the resource data model.
type InstanceModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	RegionName       types.String `tfsdk:"region_name"`
	InstanceTypeName types.String `tfsdk:"instance_type_name"`
	SshKeyNames      types.List   `tfsdk:"ssh_key_names"`
	FileSystemNames  types.List   `tfsdk:"file_system_names"`
	Ip               types.String `tfsdk:"ip"`
	PrivateIp        types.String `tfsdk:"private_ip"`
	Hostname         types.String `tfsdk:"hostname"`
	Status           types.String `tfsdk:"status"`
}

// LaunchRequest represents the request body for launching instances
type LaunchRequest struct {
	RegionName       string   `json:"region_name"`
	InstanceTypeName string   `json:"instance_type_name"`
	SshKeyNames      []string `json:"ssh_key_names"`
	Name             *string  `json:"name,omitempty"`
	FileSystemNames  []string `json:"file_system_names,omitempty"`
}

// LaunchResponse represents the response from the launch API
type LaunchResponse struct {
	Data struct {
		InstanceIds []string `json:"instance_ids"`
	} `json:"data"`
}

// Instance represents an instance from the read API
type Instance struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Ip        string `json:"ip"`
	PrivateIp string `json:"private_ip"`
	Hostname  string `json:"hostname"`
	Status    string `json:"status"`
}

// InstanceResponse represents the response from the read API
type InstanceResponse struct {
	Data Instance `json:"data"`
}

func (r *InstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InstanceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert ssh_key_names and file_system_names to Go slices
	var sshKeyNames []string
	resp.Diagnostics.Append(data.SshKeyNames.ElementsAs(ctx, &sshKeyNames, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var fileSystemNames []string
	if !data.FileSystemNames.IsNull() {
		resp.Diagnostics.Append(data.FileSystemNames.ElementsAs(ctx, &fileSystemNames, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Create launch request
	launchReq := LaunchRequest{
		RegionName:       data.RegionName.ValueString(),
		InstanceTypeName: data.InstanceTypeName.ValueString(),
		SshKeyNames:      sshKeyNames,
		FileSystemNames:  fileSystemNames,
	}

	if !data.Name.IsNull() {
		name := data.Name.ValueString()
		launchReq.Name = &name
	}

	// Make launch API call
	instanceId, err := r.launchInstance(ctx, launchReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create instance, got error: %s", err))
		return
	}

	// Set the instance ID
	data.Id = types.StringValue(instanceId)

	// Read the created instance to get computed values
	err = r.readInstance(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read instance after creation, got error: %s", err))
		return
	}

	// Write logs using the tflog package
	tflog.Trace(ctx, "created an instance resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InstanceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the instance
	err := r.readInstance(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read instance, got error: %s", err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Lambda instances don't support update operations - all changes require replacement
	resp.Diagnostics.AddError(
		"Update not supported",
		"Instance updates are not supported. All changes require resource replacement.",
	)
}

func (r *InstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data InstanceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Terminate the instance
	err := r.terminateInstance(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete instance, got error: %s", err))
		return
	}
}

func (r *InstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper methods for API calls

func (r *InstanceResource) launchInstance(ctx context.Context, launchReq LaunchRequest) (string, error) {
	jsonData, err := json.Marshal(launchReq)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/api/v1/instance-operations/launch", r.client.Endpoint),
		strings.NewReader(string(jsonData)))
	if err != nil {
		return "", err
	}

	r.client.AddAuthHeader(httpReq)

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := httpResp.Body.Close(); err != nil {
			tflog.Warn(ctx, "Failed to close response body", map[string]interface{}{"error": err})
		}
	}()

	if httpResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("launch API returned status %d", httpResp.StatusCode)
	}

	var launchResp LaunchResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&launchResp); err != nil {
		return "", err
	}

	if len(launchResp.Data.InstanceIds) == 0 {
		return "", fmt.Errorf("no instance IDs returned from launch API")
	}

	return launchResp.Data.InstanceIds[0], nil
}

func (r *InstanceResource) readInstance(ctx context.Context, data *InstanceModel) error {
	httpReq, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/api/v1/instances/%s", r.client.Endpoint, data.Id.ValueString()),
		nil)
	if err != nil {
		return err
	}

	r.client.AddAuthHeader(httpReq)

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer func() {
		if err := httpResp.Body.Close(); err != nil {
			tflog.Warn(ctx, "Failed to close response body", map[string]interface{}{"error": err})
		}
	}()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("read API returned status %d", httpResp.StatusCode)
	}

	var instanceResp InstanceResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&instanceResp); err != nil {
		return err
	}

	// Update computed fields
	data.Ip = types.StringValue(instanceResp.Data.Ip)
	data.PrivateIp = types.StringValue(instanceResp.Data.PrivateIp)
	data.Hostname = types.StringValue(instanceResp.Data.Hostname)
	data.Status = types.StringValue(instanceResp.Data.Status)

	return nil
}

func (r *InstanceResource) terminateInstance(ctx context.Context, instanceId string) error {
	terminateReq := map[string][]string{
		"instance_ids": {instanceId},
	}

	jsonData, err := json.Marshal(terminateReq)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/api/v1/instance-operations/terminate", r.client.Endpoint),
		strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}

	r.client.AddAuthHeader(httpReq)

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer func() {
		if err := httpResp.Body.Close(); err != nil {
			tflog.Warn(ctx, "Failed to close response body", map[string]interface{}{"error": err})
		}
	}()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("terminate API returned status %d", httpResp.StatusCode)
	}

	return nil
}

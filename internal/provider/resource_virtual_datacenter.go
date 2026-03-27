package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
)

var _ resource.Resource = &VirtualDatacenterResource{}

type VirtualDatacenterResource struct {
	client *apiclient.Client
}

type VirtualDatacenterResourceModel struct {
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Customer types.String `tfsdk:"customer"`
}

type virtualDatacenterAPIModel struct {
	Id       string `json:"id,omitempty"`
	Name     string `json:"name"`
	Customer string `json:"customer,omitempty"`
}

func NewVirtualDatacenterResource() resource.Resource {
	return &VirtualDatacenterResource{}
}

func (r *VirtualDatacenterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_datacenter"
}

func (r *VirtualDatacenterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"customer": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *VirtualDatacenterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type",
			fmt.Sprintf("Expected *apiclient.Client, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *VirtualDatacenterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VirtualDatacenterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := virtualDatacenterAPIModel{
		Name:     plan.Name.ValueString(),
		Customer: plan.Customer.ValueString(),
	}

	var apiResp virtualDatacenterAPIModel
	err := r.client.Post(ctx, "/api/virtualization/virtualdatacenter/", apiReq, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating virtual datacenter", err.Error())
		return
	}

	plan.Id = types.StringValue(apiResp.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *VirtualDatacenterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VirtualDatacenterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp virtualDatacenterAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/virtualization/virtualdatacenter/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading virtual datacenter", err.Error())
		return
	}

	state.Id = types.StringValue(apiResp.Id)
	state.Name = types.StringValue(apiResp.Name)
	state.Customer = types.StringValue(apiResp.Customer)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *VirtualDatacenterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan VirtualDatacenterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := virtualDatacenterAPIModel{
		Name:     plan.Name.ValueString(),
		Customer: plan.Customer.ValueString(),
	}

	var apiResp virtualDatacenterAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/virtualization/virtualdatacenter/%s/", plan.Id.ValueString()), apiReq, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating virtual datacenter", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *VirtualDatacenterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VirtualDatacenterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/virtualization/virtualdatacenter/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting virtual datacenter", err.Error())
	}
}

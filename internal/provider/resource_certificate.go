package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
)

var _ resource.Resource = &CertificateResource{}

type CertificateResource struct {
	client *apiclient.Client
}

type CertificateResourceModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Ca               types.Bool   `tfsdk:"ca"`
	ValidToTimestamp types.String `tfsdk:"valid_to_timestamp"`
	Loadbalancer     types.String `tfsdk:"loadbalancer"`
}

type certificateAPIModel struct {
	Id               string  `json:"id,omitempty"`
	Name             *string `json:"name,omitempty"`
	Ca               *bool   `json:"ca,omitempty"`
	ValidToTimestamp *string `json:"valid_to_timestamp,omitempty"`
	Loadbalancer     *string `json:"loadbalancer,omitempty"`
}

func NewCertificateResource() resource.Resource {
	return &CertificateResource{}
}

func (r *CertificateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

func (r *CertificateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"ca": schema.BoolAttribute{
				Optional: true,
			},
			"valid_to_timestamp": schema.StringAttribute{
				Optional: true,
			},
			"loadbalancer": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *CertificateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *apiclient.Client, got: %T", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *CertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CertificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel := certificateAPIModel{}
	if !plan.Name.IsNull() {
		v := plan.Name.ValueString()
		apiModel.Name = &v
	}
	if !plan.Ca.IsNull() {
		v := plan.Ca.ValueBool()
		apiModel.Ca = &v
	}
	if !plan.ValidToTimestamp.IsNull() {
		v := plan.ValidToTimestamp.ValueString()
		apiModel.ValidToTimestamp = &v
	}
	if !plan.Loadbalancer.IsNull() {
		v := plan.Loadbalancer.ValueString()
		apiModel.Loadbalancer = &v
	}

	var apiResp certificateAPIModel
	err := r.client.Post(ctx, "/api/loadbalancing/certificate/", apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating certificate", err.Error())
		return
	}

	plan.Id = types.StringValue(apiResp.Id)
	plan.Name = types.StringPointerValue(apiResp.Name)
	plan.Ca = types.BoolPointerValue(apiResp.Ca)
	plan.ValidToTimestamp = types.StringPointerValue(apiResp.ValidToTimestamp)
	plan.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp certificateAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/certificate/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading certificate", err.Error())
		return
	}

	state.Id = types.StringValue(apiResp.Id)
	state.Name = types.StringPointerValue(apiResp.Name)
	state.Ca = types.BoolPointerValue(apiResp.Ca)
	state.ValidToTimestamp = types.StringPointerValue(apiResp.ValidToTimestamp)
	state.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CertificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state CertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel := certificateAPIModel{}
	if !plan.Name.IsNull() {
		v := plan.Name.ValueString()
		apiModel.Name = &v
	}
	if !plan.Ca.IsNull() {
		v := plan.Ca.ValueBool()
		apiModel.Ca = &v
	}
	if !plan.ValidToTimestamp.IsNull() {
		v := plan.ValidToTimestamp.ValueString()
		apiModel.ValidToTimestamp = &v
	}
	if !plan.Loadbalancer.IsNull() {
		v := plan.Loadbalancer.ValueString()
		apiModel.Loadbalancer = &v
	}

	var apiResp certificateAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/loadbalancing/certificate/%s/", state.Id.ValueString()), apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating certificate", err.Error())
		return
	}

	plan.Id = types.StringValue(apiResp.Id)
	plan.Name = types.StringPointerValue(apiResp.Name)
	plan.Ca = types.BoolPointerValue(apiResp.Ca)
	plan.ValidToTimestamp = types.StringPointerValue(apiResp.ValidToTimestamp)
	plan.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/loadbalancing/certificate/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting certificate", err.Error())
	}
}

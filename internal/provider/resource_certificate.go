package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	Customer         types.String `tfsdk:"customer"`
	Protected        types.Bool   `tfsdk:"protected"`
}

type certificateAPIModel struct {
	Id               string  `json:"id,omitempty"`
	Name             string  `json:"name"`
	Ca               *bool   `json:"ca,omitempty"`
	ValidToTimestamp *string `json:"valid_to_timestamp,omitempty"`
	Loadbalancer     string  `json:"loadbalancer"`
	Customer         *string `json:"customer,omitempty"`
	Protected        *bool   `json:"protected,omitempty"`
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
				Required: true,
			},
			"ca": schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"valid_to_timestamp": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"loadbalancer": schema.StringAttribute{
				Required: true,
			},
			"customer": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"protected": schema.BoolAttribute{
				Computed:      true,
				Description:   "Whether the certificate is protected from changes.",
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
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

func certificateToAPI(plan *CertificateResourceModel) certificateAPIModel {
	return certificateAPIModel{
		Name:             plan.Name.ValueString(),
		Ca:               boolPtr(plan.Ca),
		ValidToTimestamp: stringPtr(plan.ValidToTimestamp),
		Loadbalancer:     plan.Loadbalancer.ValueString(),
		Customer:         stringPtr(plan.Customer),
	}
}

func certificateFromAPI(m *CertificateResourceModel, api *certificateAPIModel) {
	m.Id = types.StringValue(api.Id)
	m.Name = types.StringValue(api.Name)
	m.Ca = types.BoolPointerValue(api.Ca)
	m.ValidToTimestamp = types.StringPointerValue(api.ValidToTimestamp)
	m.Loadbalancer = types.StringValue(api.Loadbalancer)
	m.Customer = types.StringPointerValue(api.Customer)
	m.Protected = types.BoolValue(api.Protected != nil && *api.Protected)
}

func (r *CertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CertificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp certificateAPIModel
	err := r.client.Post(ctx, "/api/loadbalancing/certificate/", certificateToAPI(&plan), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating certificate", err.Error())
		return
	}

	certificateFromAPI(&plan, &apiResp)
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

	certificateFromAPI(&state, &apiResp)
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

	var apiResp certificateAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/loadbalancing/certificate/%s/", state.Id.ValueString()), certificateToAPI(&plan), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating certificate", err.Error())
		return
	}

	certificateFromAPI(&plan, &apiResp)
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

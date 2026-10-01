package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &LbServicegroupResource{}

type LbServicegroupResource struct {
	client *apiclient.Client
}

type LbServicegroupResourceModel struct {
	Id                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Type              types.String `tfsdk:"type"`
	State             types.String `tfsdk:"state"`
	Members           types.List   `tfsdk:"members"`
	Monitors          types.List   `tfsdk:"monitors"`
	Healthmonitor     types.String `tfsdk:"healthmonitor"`
	ClientCertificate types.String `tfsdk:"client_certificate"`
	Cip               types.String `tfsdk:"cip"`
	Cipheader         types.String `tfsdk:"cipheader"`
	Customer          types.String `tfsdk:"customer"`
	Loadbalancer      types.String `tfsdk:"loadbalancer"`
}

type lbServicegroupAPIModel struct {
	Id                string    `json:"id,omitempty"`
	Name              string    `json:"name"`
	Type              string    `json:"type"`
	State             *string   `json:"state,omitempty"`
	Members           *[]string `json:"members,omitempty"`
	Monitors          *[]string `json:"monitors,omitempty"`
	Healthmonitor     *string   `json:"healthmonitor,omitempty"`
	ClientCertificate *string   `json:"client_certificate"`
	Cip               *string   `json:"cip,omitempty"`
	Cipheader         *string   `json:"cipheader,omitempty"`
	Customer          *string   `json:"customer,omitempty"`
	Loadbalancer      *string   `json:"loadbalancer,omitempty"`
}

func NewLbServicegroupResource() resource.Resource {
	return &LbServicegroupResource{}
}

func (r *LbServicegroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lb_servicegroup"
}

func (r *LbServicegroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "One of `HTTP`, `SSL`, `ssl_bridge`, `tcp`, `udp`.",
				Validators:  []validator.String{stringvalidator.OneOf("HTTP", "SSL", "ssl_bridge", "tcp", "udp")},
			},
			"state": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Service group state: `enable` or `disable`.",
				Validators:    []validator.String{stringvalidator.OneOf("enable", "disable")},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"members": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"monitors": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "UUIDs of the load balancing monitors bound to the service group.",
			},
			"healthmonitor": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "`YES` or `NO`.",
				Validators:    []validator.String{stringvalidator.OneOf("YES", "NO")},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"client_certificate": schema.StringAttribute{
				Optional:    true,
				Description: "UUID of the certificate presented to backend servers when they request client authentication.",
			},
			"cip": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Insert the client IP in a header: `ENABLED` or `DISABLED`.",
				Validators:    []validator.String{stringvalidator.OneOf("ENABLED", "DISABLED")},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"cipheader": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Header name for the client IP. The server defaults it to `X-Forwarded-For`.",
				Validators:    []validator.String{stringvalidator.LengthAtMost(255)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"customer": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"loadbalancer": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *LbServicegroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func lbServicegroupToAPI(plan *LbServicegroupResourceModel) lbServicegroupAPIModel {
	return lbServicegroupAPIModel{
		Name:              plan.Name.ValueString(),
		Type:              plan.Type.ValueString(),
		State:             stringPtr(plan.State),
		Healthmonitor:     stringPtr(plan.Healthmonitor),
		ClientCertificate: stringPtr(plan.ClientCertificate),
		Cip:               stringPtr(plan.Cip),
		Cipheader:         stringPtr(plan.Cipheader),
		Customer:          stringPtr(plan.Customer),
		Loadbalancer:      stringPtr(plan.Loadbalancer),
	}
}

func lbServicegroupFromAPI(ctx context.Context, m *LbServicegroupResourceModel, api *lbServicegroupAPIModel, diags *diag.Diagnostics) {
	m.Id = types.StringValue(api.Id)
	m.Name = types.StringValue(api.Name)
	m.Type = types.StringValue(api.Type)
	m.State = types.StringPointerValue(api.State)
	m.Members = listValue(ctx, types.StringType, m.Members, lbStringSlice(api.Members), diags)
	m.Monitors = listValue(ctx, types.StringType, m.Monitors, lbStringSlice(api.Monitors), diags)
	m.Healthmonitor = types.StringPointerValue(api.Healthmonitor)
	m.ClientCertificate = types.StringPointerValue(api.ClientCertificate)
	m.Cip = types.StringPointerValue(api.Cip)
	m.Cipheader = types.StringPointerValue(api.Cipheader)
	m.Customer = types.StringPointerValue(api.Customer)
	m.Loadbalancer = types.StringPointerValue(api.Loadbalancer)
}

func (r *LbServicegroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LbServicegroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Stage 1: create the service group without members
	apiModel := lbServicegroupToAPI(&plan)
	apiModel.Monitors = listElems[string](ctx, plan.Monitors, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var createResp lbServicegroupAPIModel
	err := r.client.Post(ctx, "/api/loadbalancing/lbservicegroup/", apiModel, &createResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating lb_servicegroup", err.Error())
		return
	}

	// Stage 2: set members via PUT once the service group exists
	finalResp := createResp
	if !plan.Members.IsNull() && len(plan.Members.Elements()) > 0 {
		apiModel.Members = listElems[string](ctx, plan.Members, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		var updateResp lbServicegroupAPIModel
		err := r.client.Put(ctx, fmt.Sprintf("/api/loadbalancing/lbservicegroup/%s/", createResp.Id), apiModel, &updateResp)
		if err != nil {
			resp.Diagnostics.AddError("Error setting lb_servicegroup members", err.Error())
			return
		}
		finalResp = updateResp
	}

	lbServicegroupFromAPI(ctx, &plan, &finalResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LbServicegroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LbServicegroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp lbServicegroupAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/lbservicegroup/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading lb_servicegroup", err.Error())
		return
	}

	lbServicegroupFromAPI(ctx, &state, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *LbServicegroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LbServicegroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state LbServicegroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel := lbServicegroupToAPI(&plan)
	apiModel.Members = listElemsForUpdate[string](ctx, plan.Members, state.Members, &resp.Diagnostics)
	apiModel.Monitors = listElemsForUpdate[string](ctx, plan.Monitors, state.Monitors, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp lbServicegroupAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/loadbalancing/lbservicegroup/%s/", state.Id.ValueString()), apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating lb_servicegroup", err.Error())
		return
	}

	lbServicegroupFromAPI(ctx, &plan, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LbServicegroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LbServicegroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/loadbalancing/lbservicegroup/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting lb_servicegroup", err.Error())
	}
}

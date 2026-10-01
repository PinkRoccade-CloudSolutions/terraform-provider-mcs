package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &LbMonitorResource{}

type LbMonitorResource struct {
	client *apiclient.Client
}

type LbMonitorResourceModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	Interval     types.Int64  `tfsdk:"interval"`
	Resptimeout  types.Int64  `tfsdk:"resptimeout"`
	Downtime     types.Int64  `tfsdk:"downtime"`
	Respcode     types.String `tfsdk:"respcode"`
	Secure       types.String `tfsdk:"secure"`
	Httprequest  types.String `tfsdk:"httprequest"`
	Loadbalancer types.String `tfsdk:"loadbalancer"`
	Protected    types.Bool   `tfsdk:"protected"`
	Customer     types.String `tfsdk:"customer"`
}

type lbMonitorAPIModel struct {
	Id           string  `json:"id,omitempty"`
	Name         string  `json:"name"`
	Type         *string `json:"type,omitempty"`
	Interval     *int64  `json:"interval,omitempty"`
	Resptimeout  *int64  `json:"resptimeout,omitempty"`
	Downtime     *int64  `json:"downtime,omitempty"`
	Respcode     *string `json:"respcode,omitempty"`
	Secure       *string `json:"secure,omitempty"`
	Httprequest  *string `json:"httprequest,omitempty"`
	Loadbalancer *string `json:"loadbalancer,omitempty"`
	Protected    *bool   `json:"protected,omitempty"`
	Customer     *string `json:"customer,omitempty"`
}

func NewLbMonitorResource() resource.Resource {
	return &LbMonitorResource{}
}

// lbOptString is an Optional+Computed string attribute for values the server fills in or defaults.
func lbOptString(description string, validators ...validator.String) schema.StringAttribute {
	return schema.StringAttribute{
		Optional:      true,
		Computed:      true,
		Description:   description,
		Validators:    validators,
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
}

// lbOptInt64 is an Optional+Computed int64 attribute for values the server fills in or defaults.
func lbOptInt64(description string, validators ...validator.Int64) schema.Int64Attribute {
	return schema.Int64Attribute{
		Optional:      true,
		Computed:      true,
		Description:   description,
		Validators:    validators,
		PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
	}
}

// lbOptBool is an Optional+Computed bool attribute for values the server fills in or defaults.
func lbOptBool(description string) schema.BoolAttribute {
	return schema.BoolAttribute{
		Optional:      true,
		Computed:      true,
		Description:   description,
		PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
	}
}

func (r *LbMonitorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lb_monitor"
}

func (r *LbMonitorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"type": lbOptString("One of `HTTP`, `TCP`, `ssl_bridge`, `tcp`, `udp`.",
				stringvalidator.OneOf("HTTP", "TCP", "ssl_bridge", "tcp", "udp")),
			"interval":    lbOptInt64("Probe interval."),
			"resptimeout": lbOptInt64("Response timeout."),
			"downtime":    lbOptInt64("Downtime."),
			"respcode": lbOptString("Accepted response codes: `[\"200\"]` or `[\"200\", \"403\", \"500\"]` (or empty).",
				stringvalidator.OneOf(`["200"]`, `["200", "403", "500"]`, "")),
			"secure": lbOptString("`YES`, `NO` or empty.",
				stringvalidator.OneOf("YES", "NO", "")),
			"httprequest":  lbOptString("HTTP request sent by the monitor.", stringvalidator.LengthAtMost(255)),
			"loadbalancer": lbOptString(""),
			"protected":    lbOptBool("Whether the monitor is protected from changes."),
			"customer":     lbOptString(""),
		},
	}
}

func (r *LbMonitorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func lbMonitorToAPI(plan *LbMonitorResourceModel) lbMonitorAPIModel {
	return lbMonitorAPIModel{
		Name:         plan.Name.ValueString(),
		Type:         stringPtr(plan.Type),
		Interval:     int64Ptr(plan.Interval),
		Resptimeout:  int64Ptr(plan.Resptimeout),
		Downtime:     int64Ptr(plan.Downtime),
		Respcode:     stringPtr(plan.Respcode),
		Secure:       stringPtr(plan.Secure),
		Httprequest:  stringPtr(plan.Httprequest),
		Loadbalancer: stringPtr(plan.Loadbalancer),
		Protected:    boolPtr(plan.Protected),
		Customer:     stringPtr(plan.Customer),
	}
}

func lbMonitorFromAPI(m *LbMonitorResourceModel, api *lbMonitorAPIModel) {
	m.Id = types.StringValue(api.Id)
	m.Name = types.StringValue(api.Name)
	m.Type = types.StringPointerValue(api.Type)
	m.Interval = types.Int64PointerValue(api.Interval)
	m.Resptimeout = types.Int64PointerValue(api.Resptimeout)
	m.Downtime = types.Int64PointerValue(api.Downtime)
	m.Respcode = types.StringPointerValue(api.Respcode)
	m.Secure = types.StringPointerValue(api.Secure)
	m.Httprequest = types.StringPointerValue(api.Httprequest)
	m.Loadbalancer = types.StringPointerValue(api.Loadbalancer)
	m.Protected = types.BoolPointerValue(api.Protected)
	m.Customer = types.StringPointerValue(api.Customer)
}

func (r *LbMonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LbMonitorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp lbMonitorAPIModel
	err := r.client.Post(ctx, "/api/loadbalancing/monitor/", lbMonitorToAPI(&plan), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating lb_monitor", err.Error())
		return
	}

	lbMonitorFromAPI(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LbMonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LbMonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp lbMonitorAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/monitor/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading lb_monitor", err.Error())
		return
	}

	lbMonitorFromAPI(&state, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *LbMonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LbMonitorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state LbMonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp lbMonitorAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/loadbalancing/monitor/%s/", state.Id.ValueString()), lbMonitorToAPI(&plan), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating lb_monitor", err.Error())
		return
	}

	lbMonitorFromAPI(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LbMonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LbMonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/loadbalancing/monitor/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting lb_monitor", err.Error())
	}
}

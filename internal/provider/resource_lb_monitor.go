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
			"type": schema.StringAttribute{
				Optional: true,
			},
			"interval": schema.Int64Attribute{
				Optional: true,
			},
			"resptimeout": schema.Int64Attribute{
				Optional: true,
			},
			"downtime": schema.Int64Attribute{
				Optional: true,
			},
			"respcode": schema.StringAttribute{
				Optional: true,
			},
			"secure": schema.StringAttribute{
				Optional: true,
			},
			"httprequest": schema.StringAttribute{
				Optional: true,
			},
			"loadbalancer": schema.StringAttribute{
				Optional: true,
			},
			"protected": schema.BoolAttribute{
				Optional: true,
			},
			"customer": schema.StringAttribute{
				Optional: true,
			},
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

func (r *LbMonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LbMonitorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel := lbMonitorAPIModel{
		Name: plan.Name.ValueString(),
	}
	if !plan.Type.IsNull() {
		v := plan.Type.ValueString()
		apiModel.Type = &v
	}
	if !plan.Interval.IsNull() {
		v := plan.Interval.ValueInt64()
		apiModel.Interval = &v
	}
	if !plan.Resptimeout.IsNull() {
		v := plan.Resptimeout.ValueInt64()
		apiModel.Resptimeout = &v
	}
	if !plan.Downtime.IsNull() {
		v := plan.Downtime.ValueInt64()
		apiModel.Downtime = &v
	}
	if !plan.Respcode.IsNull() {
		v := plan.Respcode.ValueString()
		apiModel.Respcode = &v
	}
	if !plan.Secure.IsNull() {
		v := plan.Secure.ValueString()
		apiModel.Secure = &v
	}
	if !plan.Httprequest.IsNull() {
		v := plan.Httprequest.ValueString()
		apiModel.Httprequest = &v
	}
	if !plan.Loadbalancer.IsNull() {
		v := plan.Loadbalancer.ValueString()
		apiModel.Loadbalancer = &v
	}
	if !plan.Protected.IsNull() {
		v := plan.Protected.ValueBool()
		apiModel.Protected = &v
	}
	if !plan.Customer.IsNull() {
		v := plan.Customer.ValueString()
		apiModel.Customer = &v
	}

	var apiResp lbMonitorAPIModel
	err := r.client.Post(ctx, "/api/loadbalancing/monitor/", apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating lb_monitor", err.Error())
		return
	}

	plan.Id = types.StringValue(apiResp.Id)
	plan.Name = types.StringValue(apiResp.Name)
	plan.Type = types.StringPointerValue(apiResp.Type)
	plan.Interval = types.Int64PointerValue(apiResp.Interval)
	plan.Resptimeout = types.Int64PointerValue(apiResp.Resptimeout)
	plan.Downtime = types.Int64PointerValue(apiResp.Downtime)
	plan.Respcode = types.StringPointerValue(apiResp.Respcode)
	plan.Secure = types.StringPointerValue(apiResp.Secure)
	plan.Httprequest = types.StringPointerValue(apiResp.Httprequest)
	plan.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)
	plan.Protected = types.BoolPointerValue(apiResp.Protected)
	plan.Customer = types.StringPointerValue(apiResp.Customer)

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

	state.Id = types.StringValue(apiResp.Id)
	state.Name = types.StringValue(apiResp.Name)
	state.Type = types.StringPointerValue(apiResp.Type)
	state.Interval = types.Int64PointerValue(apiResp.Interval)
	state.Resptimeout = types.Int64PointerValue(apiResp.Resptimeout)
	state.Downtime = types.Int64PointerValue(apiResp.Downtime)
	state.Respcode = types.StringPointerValue(apiResp.Respcode)
	state.Secure = types.StringPointerValue(apiResp.Secure)
	state.Httprequest = types.StringPointerValue(apiResp.Httprequest)
	state.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)
	state.Protected = types.BoolPointerValue(apiResp.Protected)
	state.Customer = types.StringPointerValue(apiResp.Customer)

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

	apiModel := lbMonitorAPIModel{
		Name: plan.Name.ValueString(),
	}
	if !plan.Type.IsNull() {
		v := plan.Type.ValueString()
		apiModel.Type = &v
	}
	if !plan.Interval.IsNull() {
		v := plan.Interval.ValueInt64()
		apiModel.Interval = &v
	}
	if !plan.Resptimeout.IsNull() {
		v := plan.Resptimeout.ValueInt64()
		apiModel.Resptimeout = &v
	}
	if !plan.Downtime.IsNull() {
		v := plan.Downtime.ValueInt64()
		apiModel.Downtime = &v
	}
	if !plan.Respcode.IsNull() {
		v := plan.Respcode.ValueString()
		apiModel.Respcode = &v
	}
	if !plan.Secure.IsNull() {
		v := plan.Secure.ValueString()
		apiModel.Secure = &v
	}
	if !plan.Httprequest.IsNull() {
		v := plan.Httprequest.ValueString()
		apiModel.Httprequest = &v
	}
	if !plan.Loadbalancer.IsNull() {
		v := plan.Loadbalancer.ValueString()
		apiModel.Loadbalancer = &v
	}
	if !plan.Protected.IsNull() {
		v := plan.Protected.ValueBool()
		apiModel.Protected = &v
	}
	if !plan.Customer.IsNull() {
		v := plan.Customer.ValueString()
		apiModel.Customer = &v
	}

	var apiResp lbMonitorAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/loadbalancing/monitor/%s/", state.Id.ValueString()), apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating lb_monitor", err.Error())
		return
	}

	plan.Id = types.StringValue(apiResp.Id)
	plan.Name = types.StringValue(apiResp.Name)
	plan.Type = types.StringPointerValue(apiResp.Type)
	plan.Interval = types.Int64PointerValue(apiResp.Interval)
	plan.Resptimeout = types.Int64PointerValue(apiResp.Resptimeout)
	plan.Downtime = types.Int64PointerValue(apiResp.Downtime)
	plan.Respcode = types.StringPointerValue(apiResp.Respcode)
	plan.Secure = types.StringPointerValue(apiResp.Secure)
	plan.Httprequest = types.StringPointerValue(apiResp.Httprequest)
	plan.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)
	plan.Protected = types.BoolPointerValue(apiResp.Protected)
	plan.Customer = types.StringPointerValue(apiResp.Customer)

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

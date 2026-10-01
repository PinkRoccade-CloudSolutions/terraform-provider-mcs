package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &LbServicegroupMemberResource{}
	_ resource.ResourceWithImportState = &LbServicegroupMemberResource{}
)

type LbServicegroupMemberResource struct {
	client *apiclient.Client
}

type LbServicegroupMemberResourceModel struct {
	Id           types.String `tfsdk:"id"`
	Address      types.String `tfsdk:"address"`
	Port         types.Int64  `tfsdk:"port"`
	Servername   types.String `tfsdk:"servername"`
	Weight       types.Int64  `tfsdk:"weight"`
	State        types.String `tfsdk:"state"`
	Customer     types.String `tfsdk:"customer"`
	Loadbalancer types.String `tfsdk:"loadbalancer"`
}

type lbServicegroupMemberAPIModel struct {
	Id           string  `json:"id,omitempty"`
	Address      string  `json:"address"`
	Port         *int64  `json:"port,omitempty"`
	Servername   string  `json:"servername"`
	Weight       *int64  `json:"weight,omitempty"`
	State        *string `json:"state,omitempty"`
	Customer     *string `json:"customer,omitempty"`
	Loadbalancer *string `json:"loadbalancer,omitempty"`
}

func NewLbServicegroupMemberResource() resource.Resource {
	return &LbServicegroupMemberResource{}
}

func (r *LbServicegroupMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lb_servicegroup_member"
}

func (r *LbServicegroupMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"address": schema.StringAttribute{
				Required: true,
			},
			"port": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
			},
			"servername": schema.StringAttribute{
				Required: true,
			},
			"weight": schema.Int64Attribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"state": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Member state: `UP` or `DOWN`.",
				Validators:    []validator.String{stringvalidator.OneOf("UP", "DOWN")},
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

func (r *LbServicegroupMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func lbServicegroupMemberToAPI(plan *LbServicegroupMemberResourceModel) lbServicegroupMemberAPIModel {
	return lbServicegroupMemberAPIModel{
		Address:      plan.Address.ValueString(),
		Port:         int64Ptr(plan.Port),
		Servername:   plan.Servername.ValueString(),
		Weight:       int64Ptr(plan.Weight),
		State:        stringPtr(plan.State),
		Customer:     stringPtr(plan.Customer),
		Loadbalancer: stringPtr(plan.Loadbalancer),
	}
}

func lbServicegroupMemberFromAPI(m *LbServicegroupMemberResourceModel, api *lbServicegroupMemberAPIModel) {
	m.Id = types.StringValue(api.Id)
	m.Address = types.StringValue(api.Address)
	if api.Port != nil {
		m.Port = types.Int64Value(*api.Port)
	} else if m.Port.IsUnknown() {
		m.Port = types.Int64Null()
	}
	m.Servername = types.StringValue(api.Servername)
	m.Weight = types.Int64PointerValue(api.Weight)
	m.State = types.StringPointerValue(api.State)
	m.Customer = types.StringPointerValue(api.Customer)
	m.Loadbalancer = types.StringPointerValue(api.Loadbalancer)
}

func (r *LbServicegroupMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LbServicegroupMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp lbServicegroupMemberAPIModel
	err := r.client.Post(ctx, "/api/loadbalancing/lbservicegroupmember/", lbServicegroupMemberToAPI(&plan), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating lb_servicegroup_member", err.Error())
		return
	}

	lbServicegroupMemberFromAPI(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LbServicegroupMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LbServicegroupMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp lbServicegroupMemberAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/lbservicegroupmember/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading lb_servicegroup_member", err.Error())
		return
	}

	lbServicegroupMemberFromAPI(&state, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *LbServicegroupMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LbServicegroupMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state LbServicegroupMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp lbServicegroupMemberAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/loadbalancing/lbservicegroupmember/%s/", state.Id.ValueString()), lbServicegroupMemberToAPI(&plan), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating lb_servicegroup_member", err.Error())
		return
	}

	lbServicegroupMemberFromAPI(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LbServicegroupMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LbServicegroupMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/loadbalancing/lbservicegroupmember/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting lb_servicegroup_member", err.Error())
	}
}

func (r *LbServicegroupMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

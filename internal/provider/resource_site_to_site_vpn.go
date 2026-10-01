package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &SiteToSiteVPNResource{}

type SiteToSiteVPNResource struct {
	client *apiclient.Client
}

type SiteToSiteVPNResourceModel struct {
	Id                 types.String `tfsdk:"id"`
	Uuid               types.String `tfsdk:"uuid"`
	Name               types.String `tfsdk:"name"`
	Domain             types.Int64  `tfsdk:"domain"`
	Tenant             types.Int64  `tfsdk:"tenant"`
	Customer           types.String `tfsdk:"customer"`
	Firewall           types.String `tfsdk:"firewall"`
	Contact            types.List   `tfsdk:"contact"`
	State              types.String `tfsdk:"state"`
	LastStatus         types.String `tfsdk:"last_status"`
	Resets             types.Int64  `tfsdk:"resets"`
	LastCheck          types.String `tfsdk:"last_check"`
	LastReset          types.String `tfsdk:"last_reset"`
	CreatedAtTimestamp types.String `tfsdk:"created_at_timestamp"`
	UpdatedAtTimestamp types.String `tfsdk:"updated_at_timestamp"`
	CreatedByUser      types.Int64  `tfsdk:"created_by_user"`
	UpdatedByUser      types.Int64  `tfsdk:"updated_by_user"`
}

type siteToSiteVPNAPIModel struct {
	Id                 int      `json:"id,omitempty"`
	Uuid               string   `json:"uuid,omitempty"`
	Name               string   `json:"name"`
	Domain             int64    `json:"domain"`
	Tenant             int64    `json:"tenant"`
	Customer           *string  `json:"customer,omitempty"`
	Firewall           *string  `json:"firewall,omitempty"`
	Contact            *[]int64 `json:"contact,omitempty"`
	State              *string  `json:"state,omitempty"`
	LastStatus         *string  `json:"last_status,omitempty"`
	Resets             *int64   `json:"resets,omitempty"`
	LastCheck          *string  `json:"last_check,omitempty"`
	LastReset          *string  `json:"last_reset,omitempty"`
	CreatedAtTimestamp string   `json:"created_at_timestamp,omitempty"`
	UpdatedAtTimestamp string   `json:"updated_at_timestamp,omitempty"`
	CreatedByUser      *int64   `json:"created_by_user,omitempty"`
	UpdatedByUser      *int64   `json:"updated_by_user,omitempty"`
}

func NewSiteToSiteVPNResource() resource.Resource {
	return &SiteToSiteVPNResource{}
}

func (r *SiteToSiteVPNResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_site_to_site_vpn"
}

func (r *SiteToSiteVPNResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	keep := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	optionalComputed := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{Optional: true, Computed: true, Description: desc}
	}

	resp.Schema = schema.Schema{
		Description: "Manages a site-to-site VPN in MCS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: keep,
			},
			"uuid": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: keep,
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Name of the VPN.",
				Validators:  []validator.String{stringvalidator.LengthAtMost(255)},
			},
			"domain": schema.Int64Attribute{
				Required:      true,
				Description:   "ID of the domain the VPN belongs to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"tenant": schema.Int64Attribute{
				Required:      true,
				Description:   "ID of the tenant the VPN belongs to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"customer": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Customer identifier.",
				PlanModifiers: keep,
			},
			"firewall": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "UUID of the firewall terminating the VPN.",
				PlanModifiers: keep,
			},
			"contact": schema.ListAttribute{
				ElementType: types.Int64Type,
				Optional:    true,
				Description: "IDs of the contacts for this VPN.",
			},
			"state":       optionalComputed("Tunnel state."),
			"last_status": optionalComputed("Last reported status."),
			"resets": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Number of tunnel resets.",
			},
			"last_check": optionalComputed("Timestamp of the last check."),
			"last_reset": optionalComputed("Timestamp of the last reset."),
			"created_at_timestamp": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: keep,
			},
			"updated_at_timestamp": schema.StringAttribute{
				Computed: true,
			},
			"created_by_user": schema.Int64Attribute{
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"updated_by_user": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (r *SiteToSiteVPNResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SiteToSiteVPNResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SiteToSiteVPNResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := buildSiteToSiteVPNRequest(&plan)
	apiReq.Contact = listElems[int64](ctx, plan.Contact, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp siteToSiteVPNAPIModel
	err := r.client.Post(ctx, "/api/vpn/site_to_site/", apiReq, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating site-to-site VPN", err.Error())
		return
	}

	mapSiteToSiteVPNToState(ctx, &plan, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SiteToSiteVPNResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SiteToSiteVPNResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp siteToSiteVPNAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/vpn/site_to_site/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading site-to-site VPN", err.Error())
		return
	}

	mapSiteToSiteVPNToState(ctx, &state, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SiteToSiteVPNResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state SiteToSiteVPNResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := buildSiteToSiteVPNRequest(&plan)
	apiReq.Contact = listElemsForUpdate[int64](ctx, plan.Contact, state.Contact, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp siteToSiteVPNAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/vpn/site_to_site/%s/", state.Id.ValueString()), apiReq, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating site-to-site VPN", err.Error())
		return
	}

	mapSiteToSiteVPNToState(ctx, &plan, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SiteToSiteVPNResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SiteToSiteVPNResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/vpn/site_to_site/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting site-to-site VPN", err.Error())
	}
}

func buildSiteToSiteVPNRequest(plan *SiteToSiteVPNResourceModel) siteToSiteVPNAPIModel {
	return siteToSiteVPNAPIModel{
		Name:       plan.Name.ValueString(),
		Domain:     plan.Domain.ValueInt64(),
		Tenant:     plan.Tenant.ValueInt64(),
		Customer:   stringPtr(plan.Customer),
		Firewall:   stringPtr(plan.Firewall),
		State:      stringPtr(plan.State),
		LastStatus: stringPtr(plan.LastStatus),
		Resets:     int64Ptr(plan.Resets),
		LastCheck:  stringPtr(plan.LastCheck),
		LastReset:  stringPtr(plan.LastReset),
	}
}

func mapSiteToSiteVPNToState(ctx context.Context, state *SiteToSiteVPNResourceModel, api *siteToSiteVPNAPIModel, diags *diag.Diagnostics) {
	str := func(s *string) types.String {
		if s == nil {
			return types.StringValue("")
		}
		return types.StringValue(*s)
	}
	state.Id = types.StringValue(strconv.Itoa(api.Id))
	state.Uuid = types.StringValue(api.Uuid)
	state.Name = types.StringValue(api.Name)
	state.Domain = types.Int64Value(api.Domain)
	state.Tenant = types.Int64Value(api.Tenant)
	state.Customer = types.StringPointerValue(api.Customer)
	state.Firewall = types.StringPointerValue(api.Firewall)
	var contact []int64
	if api.Contact != nil {
		contact = *api.Contact
	}
	state.Contact = listValue(ctx, types.Int64Type, state.Contact, contact, diags)
	state.State = str(api.State)
	state.LastStatus = str(api.LastStatus)
	if api.Resets != nil {
		state.Resets = types.Int64Value(*api.Resets)
	} else {
		state.Resets = types.Int64Value(0)
	}
	state.LastCheck = types.StringPointerValue(api.LastCheck)
	state.LastReset = types.StringPointerValue(api.LastReset)
	state.CreatedAtTimestamp = types.StringValue(api.CreatedAtTimestamp)
	state.UpdatedAtTimestamp = types.StringValue(api.UpdatedAtTimestamp)
	state.CreatedByUser = types.Int64PointerValue(api.CreatedByUser)
	state.UpdatedByUser = types.Int64PointerValue(api.UpdatedByUser)
}

package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &DnsDomainResource{}
	_ resource.ResourceWithImportState = &DnsDomainResource{}
)

type DnsDomainResource struct {
	client *apiclient.Client
}

type DnsDomainResourceModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Comment      types.String `tfsdk:"comment"`
	Enddate      types.String `tfsdk:"enddate"`
	Customer     types.String `tfsdk:"customer"`
	Type         types.String `tfsdk:"type"`
	ZoneType     types.String `tfsdk:"zone_type"`
	ProviderID   types.Int64  `tfsdk:"provider_id"`
	ProviderName types.String `tfsdk:"provider_name"`
}

type dnsDomainUpdateAPIModel struct {
	Name     string  `json:"name"`
	Comment  *string `json:"comment,omitempty"`
	Enddate  *string `json:"enddate,omitempty"`
	Customer *string `json:"customer,omitempty"`
	Type     *string `json:"type,omitempty"`
	ZoneType *string `json:"zone_type,omitempty"`
}

func NewDnsDomainResource() resource.Resource {
	return &DnsDomainResource{}
}

func (r *DnsDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_domain"
}

func (r *DnsDomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the settings of an existing MCS DNS domain (zone). Zones are synchronised from the DNS " +
			"provider and cannot be created through the API: creating this resource adopts the existing zone with the " +
			"given name, and destroying it only removes it from Terraform state (the zone is not deleted in MCS).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "UUID of the DNS domain.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:      true,
				Description:   "Full zone name as known by the DNS provider (e.g. domain.nl). Must match an existing zone.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"comment":  lbOptString("Comment for the domain.", stringvalidator.LengthAtMost(255)),
			"enddate":  lbOptString("End date for the domain (YYYY-MM-DD), if known."),
			"customer": lbOptString("Customer associated with the domain."),
			"type": lbOptString("Domain use type: `external` or `internal`.",
				stringvalidator.OneOf("external", "internal", "")),
			"zone_type": lbOptString("Zone type: `forward` or `reverse`.",
				stringvalidator.OneOf("forward", "reverse")),
			"provider_id": schema.Int64Attribute{
				Computed:      true,
				Description:   "ID of the DNS provider integration.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"provider_name": schema.StringAttribute{
				Computed:      true,
				Description:   "Name of the DNS provider integration.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *DnsDomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *apiclient.Client, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func dnsDomainToAPI(plan *DnsDomainResourceModel) dnsDomainUpdateAPIModel {
	return dnsDomainUpdateAPIModel{
		Name:     plan.Name.ValueString(),
		Comment:  stringPtr(plan.Comment),
		Enddate:  stringPtr(plan.Enddate),
		Customer: stringPtr(plan.Customer),
		Type:     stringPtr(plan.Type),
		ZoneType: stringPtr(plan.ZoneType),
	}
}

func dnsDomainFromAPI(m *DnsDomainResourceModel, api *dnsDomainAPIModel) {
	m.Id = types.StringValue(api.UUID)
	m.Name = types.StringValue(api.Name)
	m.Comment = types.StringValue(api.Comment)
	m.Enddate = types.StringPointerValue(api.Enddate)
	m.Customer = types.StringPointerValue(api.Customer)
	m.Type = types.StringValue(api.Type)
	m.ZoneType = types.StringValue(api.ZoneType)
	m.ProviderID = types.Int64Value(int64(api.Provider.ID))
	m.ProviderName = types.StringValue(api.Provider.Name)
}

func (r *DnsDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DnsDomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	domains, err := listAll[dnsDomainAPIModel](ctx, r.client, "/api/dns/domains/?name__icontains="+url.QueryEscape(name))
	if err != nil {
		resp.Diagnostics.AddError("Error looking up DNS domain", err.Error())
		return
	}
	var existing *dnsDomainAPIModel
	for i := range domains {
		if domains[i].Name == name {
			existing = &domains[i]
			break
		}
	}
	if existing == nil {
		resp.Diagnostics.AddError("DNS domain not found",
			fmt.Sprintf("No DNS domain named %q exists in MCS. DNS domains cannot be created through the API; "+
				"they are synchronised from the DNS provider. Make sure the zone exists at the provider and has been synced.", name))
		return
	}

	var apiResp dnsDomainAPIModel
	err = r.client.Put(ctx, fmt.Sprintf("/api/dns/domains/%s/", existing.UUID), dnsDomainToAPI(&plan), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating DNS domain", err.Error())
		return
	}

	dnsDomainFromAPI(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DnsDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DnsDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp dnsDomainAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/dns/domains/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading DNS domain", err.Error())
		return
	}

	dnsDomainFromAPI(&state, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DnsDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DnsDomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state DnsDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp dnsDomainAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/dns/domains/%s/", state.Id.ValueString()), dnsDomainToAPI(&plan), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating DNS domain", err.Error())
		return
	}

	dnsDomainFromAPI(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DnsDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DnsDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.AddWarning("DNS domain not deleted",
		fmt.Sprintf("DNS domain %q was removed from Terraform state only. The zone still exists in MCS and at the DNS provider.",
			state.Name.ValueString()))
}

func (r *DnsDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

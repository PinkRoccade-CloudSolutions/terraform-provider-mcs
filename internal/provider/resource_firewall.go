package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ resource.Resource                = &FirewallResource{}
	_ resource.ResourceWithImportState = &FirewallResource{}
)

type FirewallResource struct {
	client *apiclient.Client
}

type firewallRequest struct {
	Name                  *string `json:"name,omitempty"`
	Description           *string `json:"description,omitempty"`
	Customer              string  `json:"customer"`
	Type                  *string `json:"type,omitempty"`
	Device                string  `json:"device"`
	Context               *string `json:"context,omitempty"`
	ExternalInterface     *string `json:"external_interface,omitempty"`
	InternalInterface     *string `json:"internal_interface,omitempty"`
	DefaultLogProfile     *string `json:"default_log_profile,omitempty"`
	DefaultProtectProfile *string `json:"default_protect_profile,omitempty"`
	MultiTenant           *bool   `json:"multi_tenant,omitempty"`
	TagName               *string `json:"tag_name,omitempty"`
	NatIpSyncEnabled      *bool   `json:"nat_ip_sync_enabled,omitempty"`
}

func NewFirewallResource() resource.Resource {
	return &FirewallResource{}
}

func (r *FirewallResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall"
}

func (r *FirewallResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	optString := func(desc string, maxLen int) schema.StringAttribute {
		return schema.StringAttribute{
			Optional:      true,
			Computed:      true,
			Description:   desc,
			Validators:    []validator.String{stringvalidator.LengthAtMost(maxLen)},
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	}
	optBool := func(desc string) schema.BoolAttribute {
		return schema.BoolAttribute{
			Optional:      true,
			Computed:      true,
			Description:   desc,
			PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
		}
	}
	computedString := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{
			Computed:      true,
			Description:   desc,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	}

	resp.Schema = schema.Schema{
		Description: "Manages a firewall in the MCS API.",
		Attributes: map[string]schema.Attribute{
			"id": computedString("UUID of the firewall."),
			"customer": schema.StringAttribute{
				Required:      true,
				Description:   "Customer the firewall belongs to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"device": schema.StringAttribute{
				Required:      true,
				Description:   "Firewall device (see /api/networking/firewalls/device-options/). Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name":        optString("Name of the firewall.", 255),
			"description": optString("Description of the firewall.", 4096),
			"type": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Firewall type: `internet` or `wan`.",
				Validators:    []validator.String{stringvalidator.OneOf("internet", "wan")},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"context":                    optString("VDOM on Fortimanager or Device Group on Panorama.", 255),
			"external_interface":         optString("Name of the external (internet or WAN facing) interface.", 255),
			"internal_interface":         optString("Name of the internal (VDOM or transit facing) interface.", 255),
			"default_log_profile":        optString("Default log profile used when creating firewall rules.", 255),
			"default_protect_profile":    optString("Default protect group profile used when creating firewall rules.", 255),
			"multi_tenant":               optBool("Whether the device is multi tenant; `tag_name` is then used for rule separation."),
			"tag_name":                   optString("TAG name for object lookups on a multi tenant firewall (PaloAlto only).", 255),
			"nat_ip_sync_enabled":        optBool("Opt in to the daily Panorama NAT public IP import job for this device group."),
			"customer_name":              computedString("Name of the customer."),
			"device_name":                computedString("Name of the device."),
			"platform":                   computedString("Platform of the device."),
			"supports_threat_protection": schema.BoolAttribute{Computed: true, Description: "Whether the device supports threat protection.", PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		},
	}
}

func (r *FirewallResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("Expected *apiclient.Client, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func firewallBody(plan *FirewallModel) firewallRequest {
	return firewallRequest{
		Name:                  stringPtr(plan.Name),
		Description:           stringPtr(plan.Description),
		Customer:              plan.Customer.ValueString(),
		Type:                  stringPtr(plan.Type),
		Device:                plan.Device.ValueString(),
		Context:               stringPtr(plan.Context),
		ExternalInterface:     stringPtr(plan.ExternalInterface),
		InternalInterface:     stringPtr(plan.InternalInterface),
		DefaultLogProfile:     stringPtr(plan.DefaultLogProfile),
		DefaultProtectProfile: stringPtr(plan.DefaultProtectProfile),
		MultiTenant:           boolPtr(plan.MultiTenant),
		TagName:               stringPtr(plan.TagName),
		NatIpSyncEnabled:      boolPtr(plan.NatIpSyncEnabled),
	}
}

func (r *FirewallResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FirewallModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp firewallAPIModel
	if err := r.client.Post(ctx, "/api/networking/firewalls/", firewallBody(&plan), &apiResp); err != nil {
		resp.Diagnostics.AddError("Error creating firewall", err.Error())
		return
	}

	state := firewallModelFromAPI(&apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FirewallResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FirewallModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp firewallAPIModel
	if err := r.client.Get(ctx, fmt.Sprintf("/api/networking/firewalls/%s/", state.Id.ValueString()), &apiResp); err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading firewall", err.Error())
		return
	}

	state = firewallModelFromAPI(&apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FirewallResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state FirewallModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp firewallAPIModel
	if err := r.client.Patch(ctx, fmt.Sprintf("/api/networking/firewalls/%s/", state.Id.ValueString()), firewallBody(&plan), &apiResp); err != nil {
		resp.Diagnostics.AddError("Error updating firewall", err.Error())
		return
	}

	newState := firewallModelFromAPI(&apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *FirewallResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FirewallModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Delete(ctx, fmt.Sprintf("/api/networking/firewalls/%s/", state.Id.ValueString())); err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting firewall", err.Error())
	}
}

func (r *FirewallResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

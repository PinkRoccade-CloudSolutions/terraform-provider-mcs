package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &NetworkDhcpLeaseResource{}
	_ resource.ResourceWithImportState = &NetworkDhcpLeaseResource{}
)

const dhcpLeasesPath = "/api/v3/networking/dhcp-leases/"

type NetworkDhcpLeaseResource struct {
	client *apiclient.Client
}

type NetworkDhcpLeaseResourceModel struct {
	Id          types.String   `tfsdk:"id"`
	Network     types.String   `tfsdk:"network"`
	Ip          types.String   `tfsdk:"ip"`
	Mac         types.String   `tfsdk:"mac"`
	Description types.String   `tfsdk:"description"`
	CreatedAt   types.String   `tfsdk:"created_at"`
	UpdatedAt   types.String   `tfsdk:"updated_at"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}

// dhcpLeaseAPIModel is DhcpLeaseV3: a static lease as the FortiGate runs it.
type dhcpLeaseAPIModel struct {
	Id          string `json:"id"`
	Network     string `json:"network"`
	Ip          string `json:"ip"`
	Mac         string `json:"mac"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at_timestamp"`
	UpdatedAt   string `json:"updated_at_timestamp"`
}

// dhcpLeaseWriteAPIModel is DhcpLeaseWriteV3; network is sent on create only.
type dhcpLeaseWriteAPIModel struct {
	Network     string `json:"network,omitempty"`
	Ip          string `json:"ip"`
	Mac         string `json:"mac"`
	Description string `json:"description"`
}

func NewNetworkDhcpLeaseResource() resource.Resource {
	return &NetworkDhcpLeaseResource{}
}

func dhcpLeasePath(id string) string {
	return dhcpLeasesPath + id + "/"
}

func (r *NetworkDhcpLeaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_dhcp_lease"
}

func (r *NetworkDhcpLeaseResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a static DHCP lease (a FortiOS reserved address) on an MCS network with a DHCP server. " +
			"Every change is queued as a job that configures the FortiGate; Terraform waits for it to finish.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "UUID of the lease.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"network": schema.StringAttribute{
				Required: true,
				Description: "UUID of the network, which must have a DHCP server. Refer to `mcs_network_dhcp_server.x.network` " +
					"so the lease is created after the server and destroyed before it. Changing it recreates the lease.",
				Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"ip": schema.StringAttribute{
				Required: true,
				Description: "IPv4 address to reserve. Must be inside the network's prefix, must not be its network, broadcast " +
					"or gateway address, and must be unique on the network.",
				Validators: []validator.String{ipv4Validator{}},
			},
			"mac": schema.StringAttribute{
				Required: true,
				Description: "MAC address of the client, in any common notation (`00:50:56:aa:bb:01`, `00-50-56-AA-BB-01`, " +
					"`0050.56aa.bb01`). MCS stores it lower case and colon separated; state keeps the notation you " +
					"configured, and rewriting the same address in another notation sends nothing to MCS. Must be unique on the network.",
				Validators: []validator.String{macValidator{}},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "What the reservation is for (max 255 characters).",
				Validators:  []validator.String{stringvalidator.LengthAtMost(255)},
			},
			"created_at": schema.StringAttribute{
				Computed:      true,
				Description:   "Time the lease was created.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Time the lease was last updated.",
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{Create: true, Update: true, Delete: true}),
		},
	}
}

func (r *NetworkDhcpLeaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// dhcpLeaseFromAPI copies the API record into the model. A MAC in state or configuration that
// only differs in notation from the stored one is kept, so it does not show as a change.
func dhcpLeaseFromAPI(m *NetworkDhcpLeaseResourceModel, api *dhcpLeaseAPIModel) {
	m.Id = types.StringValue(api.Id)
	m.Network = types.StringValue(api.Network)
	m.Ip = types.StringValue(api.Ip)
	if !known(m.Mac) || !sameMAC(m.Mac.ValueString(), api.Mac) {
		m.Mac = types.StringValue(api.Mac)
	}
	m.Description = types.StringValue(api.Description)
	m.CreatedAt = types.StringValue(api.CreatedAt)
	m.UpdatedAt = types.StringValue(api.UpdatedAt)
}

func dhcpLeaseToAPI(m *NetworkDhcpLeaseResourceModel) dhcpLeaseWriteAPIModel {
	return dhcpLeaseWriteAPIModel{
		Ip:          m.Ip.ValueString(),
		Mac:         m.Mac.ValueString(),
		Description: m.Description.ValueString(),
	}
}

func (r *NetworkDhcpLeaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NetworkDhcpLeaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	createTimeout, diags := plan.Timeouts.Create(ctx, dhcpDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	waitCtx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	body := dhcpLeaseToAPI(&plan)
	body.Network = plan.Network.ValueString()

	// Sent once: a retried POST could queue a second change, which would then be refused as a
	// duplicate IP or MAC only after the first one had been applied.
	var op networkOperationAPIModel
	if err := r.client.PostOnce(waitCtx, dhcpLeasesPath, body, &op); err != nil {
		resp.Diagnostics.AddError("Error creating network DHCP lease",
			fmt.Sprintf("The request failed and was not retried. If the error was a timeout or server error, check in "+
				"MCS whether a lease change was queued before trying again.\n\n%s", err))
		return
	}

	final, rolledBack, err := applyDhcpChange(waitCtx, r.client, &op, &resp.Diagnostics)
	if err != nil {
		if !rolledBack && final.DhcpLease != nil && final.DhcpLease.Id != "" {
			// The lease row exists: keep it in state so Terraform marks it tainted instead of losing it.
			dhcpLeaseFromAPI(&plan, final.DhcpLease)
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		}
		resp.Diagnostics.AddError("Error creating network DHCP lease", dhcpChangeError(err, rolledBack))
		return
	}
	if final.DhcpLease == nil || final.DhcpLease.Id == "" {
		resp.Diagnostics.AddError("Error creating network DHCP lease",
			fmt.Sprintf("The change finished without returning a lease: %s", final.describe()))
		return
	}

	r.readInto(ctx, final.DhcpLease.Id, &plan, "Error reading network DHCP lease after create", &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		// The lease exists; save what the operation returned so it is not lost.
		dhcpLeaseFromAPI(&plan, final.DhcpLease)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NetworkDhcpLeaseResource) readInto(ctx context.Context, id string, m *NetworkDhcpLeaseResourceModel, summary string, diags *diag.Diagnostics) {
	var fresh dhcpLeaseAPIModel
	if err := r.client.Get(ctx, dhcpLeasePath(id), &fresh); err != nil {
		diags.AddError(summary, err.Error())
		return
	}
	dhcpLeaseFromAPI(m, &fresh)
}

func (r *NetworkDhcpLeaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NetworkDhcpLeaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var api dhcpLeaseAPIModel
	if err := r.client.Get(ctx, dhcpLeasePath(state.Id.ValueString()), &api); err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading network DHCP lease", err.Error())
		return
	}
	dhcpLeaseFromAPI(&state, &api)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NetworkDhcpLeaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state NetworkDhcpLeaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.Id.ValueString()

	// Only timeouts changed, or the MAC is written differently: nothing to send to MCS.
	if plan.Ip.Equal(state.Ip) && plan.Description.Equal(state.Description) && sameMAC(plan.Mac.ValueString(), state.Mac.ValueString()) {
		plan.Id, plan.CreatedAt, plan.UpdatedAt = state.Id, state.CreatedAt, state.UpdatedAt
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, dhcpDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	waitCtx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	var op networkOperationAPIModel
	if err := r.client.Patch(waitCtx, dhcpLeasePath(id), dhcpLeaseToAPI(&plan), &op); err != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		resp.Diagnostics.AddError("Error changing network DHCP lease", err.Error())
		return
	}
	if _, rolledBack, err := applyDhcpChange(waitCtx, r.client, &op, &resp.Diagnostics); err != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		resp.Diagnostics.AddError("Error changing network DHCP lease", dhcpChangeError(err, rolledBack))
		return
	}

	r.readInto(ctx, id, &plan, "Error reading network DHCP lease after update", &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		// The change was applied; save it without the unknown timestamp.
		plan.UpdatedAt = state.UpdatedAt
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NetworkDhcpLeaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NetworkDhcpLeaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	deleteTimeout, diags := state.Timeouts.Delete(ctx, dhcpDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	waitCtx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	var op networkOperationAPIModel
	err := r.client.DeleteWithResult(waitCtx, dhcpLeasePath(state.Id.ValueString()), &op)
	switch {
	case apiclient.IsNotFound(err):
		return
	case err != nil:
		resp.Diagnostics.AddError("Error deleting network DHCP lease", err.Error())
		return
	}
	if _, rolledBack, err := applyDhcpChange(waitCtx, r.client, &op, &resp.Diagnostics); err != nil {
		resp.Diagnostics.AddError("Error deleting network DHCP lease", dhcpChangeError(err, rolledBack))
	}
}

func (r *NetworkDhcpLeaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

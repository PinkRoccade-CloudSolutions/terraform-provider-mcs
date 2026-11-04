package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &NetworkDhcpServerResource{}
	_ resource.ResourceWithImportState = &NetworkDhcpServerResource{}
)

// dhcpDefaultTimeout bounds a DHCP change: one job that configures the FortiGate and has
// FortiManager retrieve it.
const dhcpDefaultTimeout = 20 * time.Minute

type NetworkDhcpServerResource struct {
	client *apiclient.Client
}

type NetworkDhcpServerResourceModel struct {
	Id             types.String   `tfsdk:"id"`
	Network        types.String   `tfsdk:"network"`
	PoolStart      types.String   `tfsdk:"pool_start"`
	PoolEnd        types.String   `tfsdk:"pool_end"`
	DnsServers     types.List     `tfsdk:"dns_servers"`
	LeaseTime      types.Int64    `tfsdk:"lease_time"`
	Netmask        types.String   `tfsdk:"netmask"`
	DefaultGateway types.String   `tfsdk:"default_gateway"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
}

// dhcpServerAPIModel is DhcpServerV3: a Network's DHCP server, `enabled: false` when there is none.
type dhcpServerAPIModel struct {
	Network        string   `json:"network"`
	Enabled        bool     `json:"enabled"`
	PoolStart      *string  `json:"pool_start"`
	PoolEnd        *string  `json:"pool_end"`
	Netmask        *string  `json:"netmask"`
	DefaultGateway *string  `json:"default_gateway"`
	DnsServers     []string `json:"dns_servers"`
	LeaseTime      *int64   `json:"lease_time"`
}

// dhcpServerWriteAPIModel is DhcpServerWriteV3.
type dhcpServerWriteAPIModel struct {
	PoolStart  string   `json:"pool_start"`
	PoolEnd    string   `json:"pool_end"`
	DnsServers []string `json:"dns_servers"`
	LeaseTime  int64    `json:"lease_time"`
}

func NewNetworkDhcpServerResource() resource.Resource {
	return &NetworkDhcpServerResource{}
}

func dhcpServerPath(networkID string) string {
	return networkPath(networkID) + "dhcp-server/"
}

func (r *NetworkDhcpServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_dhcp_server"
}

func (r *NetworkDhcpServerResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the DHCP server of an MCS network (one per network) through the v3 networking API. " +
			"Every change is queued as a job that configures the FortiGate; Terraform waits for it to finish. " +
			"Destroying the resource disables the DHCP server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "UUID of the network (the DHCP server has no ID of its own).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"network": schema.StringAttribute{
				Required:      true,
				Description:   "UUID of the network to run the DHCP server on, e.g. `mcs_network.x.id`. Changing it moves the server to another network.",
				Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"pool_start": schema.StringAttribute{
				Required:    true,
				Description: "First address of the dynamic pool. Must be inside the network's prefix.",
				Validators:  []validator.String{ipv4Validator{}},
			},
			"pool_end": schema.StringAttribute{
				Required:    true,
				Description: "Last address of the dynamic pool. Must be inside the network's prefix, not before `pool_start`, and the pool must not contain the gateway.",
				Validators:  []validator.String{ipv4Validator{}},
			},
			"dns_servers": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "DNS servers handed out to clients: one to four IPv4 addresses.",
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 4),
					listvalidator.ValueStringsAre(ipv4Validator{}),
				},
			},
			"lease_time": schema.Int64Attribute{
				Required:    true,
				Description: "Lease time in seconds (300-8604800).",
				Validators:  []validator.Int64{int64validator.Between(300, 8604800)},
			},
			"netmask": schema.StringAttribute{
				Computed:      true,
				Description:   "Netmask handed out to clients; follows the network's prefix.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"default_gateway": schema.StringAttribute{
				Computed:      true,
				Description:   "Default gateway handed out to clients; the network's gateway.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{Create: true, Update: true, Delete: true}),
		},
	}
}

func (r *NetworkDhcpServerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func dhcpServerToAPI(ctx context.Context, m *NetworkDhcpServerResourceModel, diags *diag.Diagnostics) dhcpServerWriteAPIModel {
	var dns []string
	diags.Append(m.DnsServers.ElementsAs(ctx, &dns, false)...)
	return dhcpServerWriteAPIModel{
		PoolStart:  m.PoolStart.ValueString(),
		PoolEnd:    m.PoolEnd.ValueString(),
		DnsServers: dns,
		LeaseTime:  m.LeaseTime.ValueInt64(),
	}
}

// dhcpServerFromAPI copies an enabled server into the model.
func dhcpServerFromAPI(m *NetworkDhcpServerResourceModel, api *dhcpServerAPIModel) {
	m.Id = types.StringValue(api.Network)
	m.Network = types.StringValue(api.Network)
	m.PoolStart = strOrNull(api.PoolStart)
	m.PoolEnd = strOrNull(api.PoolEnd)
	m.DnsServers = stringList(api.DnsServers)
	m.LeaseTime = types.Int64PointerValue(api.LeaseTime)
	m.Netmask = strOrNull(api.Netmask)
	m.DefaultGateway = strOrNull(api.DefaultGateway)
}

// applyDhcpChange waits for a queued DHCP change. A job that finished unsuccessfully has been
// rolled back, so rolledBack tells the caller that nothing changed; otherwise (a timeout or a
// polling error) the outcome is unknown.
func applyDhcpChange(ctx context.Context, c *apiclient.Client, op *networkOperationAPIModel, diags *diag.Diagnostics) (final *networkOperationAPIModel, rolledBack bool, err error) {
	final, err = waitForNetworkOperation(ctx, c, op)
	if err != nil {
		return final, final.state() == networkOperationFailed, err
	}
	final.addWarnings("DHCP change applied with a warning", diags)
	return final, false, nil
}

func (r *NetworkDhcpServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NetworkDhcpServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	networkID := plan.Network.ValueString()

	// The server is a singleton of its network: refuse to take over one Terraform does not manage.
	var current dhcpServerAPIModel
	if err := r.client.Get(ctx, dhcpServerPath(networkID), &current); err != nil {
		resp.Diagnostics.AddError("Error reading network DHCP server", err.Error())
		return
	}
	if current.Enabled {
		resp.Diagnostics.AddError("Network already has a DHCP server",
			fmt.Sprintf("Network %s already runs a DHCP server. Import it with "+
				"`terraform import <address> %s` to manage it with Terraform.", networkID, networkID))
		return
	}

	body := dhcpServerToAPI(ctx, &plan, &resp.Diagnostics)
	createTimeout, diags := plan.Timeouts.Create(ctx, dhcpDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	waitCtx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	var op networkOperationAPIModel
	if err := r.client.Put(waitCtx, dhcpServerPath(networkID), body, &op); err != nil {
		resp.Diagnostics.AddError("Error enabling network DHCP server", err.Error())
		return
	}
	if _, rolledBack, err := applyDhcpChange(waitCtx, r.client, &op, &resp.Diagnostics); err != nil {
		if !rolledBack {
			// The change may still be applied: keep the server in state so Terraform marks it
			// tainted and the next apply puts it right.
			plan.Id = types.StringValue(networkID)
			plan.Netmask = strOrNull(current.Netmask)
			plan.DefaultGateway = strOrNull(current.DefaultGateway)
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		}
		resp.Diagnostics.AddError("Error enabling network DHCP server", dhcpChangeError(err, rolledBack))
		return
	}

	r.readInto(ctx, networkID, &plan, "Error reading network DHCP server after create", &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// dhcpChangeError explains a failed DHCP change.
func dhcpChangeError(err error, rolledBack bool) string {
	if rolledBack {
		return fmt.Sprintf("%s\n\nThe change was rolled back: the FortiGate and MCS are as they were before.", err)
	}
	return fmt.Sprintf("%s\n\nThe change may still be applied; check the job in MCS.", err)
}

// readInto reads the server after a successful change. The configured dns_servers, pool and
// lease time are what the API stores, so the whole model is taken from the API.
func (r *NetworkDhcpServerResource) readInto(ctx context.Context, networkID string, m *NetworkDhcpServerResourceModel, summary string, diags *diag.Diagnostics) {
	var fresh dhcpServerAPIModel
	if err := r.client.Get(ctx, dhcpServerPath(networkID), &fresh); err != nil {
		diags.AddError(summary, err.Error())
		return
	}
	if !fresh.Enabled {
		diags.AddError(summary, fmt.Sprintf("The change finished, but network %s reads without a DHCP server.", networkID))
		return
	}
	if fresh.Network == "" {
		fresh.Network = networkID
	}
	dhcpServerFromAPI(m, &fresh)
}

func (r *NetworkDhcpServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NetworkDhcpServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	networkID := state.Id.ValueString()

	var api dhcpServerAPIModel
	if err := r.client.Get(ctx, dhcpServerPath(networkID), &api); err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading network DHCP server", err.Error())
		return
	}
	if !api.Enabled {
		resp.State.RemoveResource(ctx)
		return
	}
	if api.Network == "" {
		api.Network = networkID
	}
	dhcpServerFromAPI(&state, &api)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NetworkDhcpServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state NetworkDhcpServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	networkID := state.Id.ValueString()
	body := dhcpServerToAPI(ctx, &plan, &resp.Diagnostics)
	updateTimeout, diags := plan.Timeouts.Update(ctx, dhcpDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	waitCtx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	var op networkOperationAPIModel
	if err := r.client.Put(waitCtx, dhcpServerPath(networkID), body, &op); err != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		resp.Diagnostics.AddError("Error changing network DHCP server", err.Error())
		return
	}
	if _, rolledBack, err := applyDhcpChange(waitCtx, r.client, &op, &resp.Diagnostics); err != nil {
		// Keep the prior state; the next refresh shows what the FortiGate runs.
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		resp.Diagnostics.AddError("Error changing network DHCP server", dhcpChangeError(err, rolledBack))
		return
	}

	r.readInto(ctx, networkID, &plan, "Error reading network DHCP server after update", &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NetworkDhcpServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NetworkDhcpServerResourceModel
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
	err := r.client.DeleteWithResult(waitCtx, dhcpServerPath(state.Id.ValueString()), &op)
	switch {
	case apiclient.IsNotFound(err):
		return
	case apiclient.IsConflict(err):
		resp.Diagnostics.AddError("Error disabling network DHCP server",
			fmt.Sprintf("%s\n\nMCS refuses to disable a DHCP server while the network has static leases, or while the "+
				"network is not active. Remove the leases first; an `mcs_network_dhcp_lease` whose `network` refers to "+
				"`mcs_network_dhcp_server.<name>.network` is destroyed before its server.", err))
		return
	case err != nil:
		resp.Diagnostics.AddError("Error disabling network DHCP server", err.Error())
		return
	}
	if _, rolledBack, err := applyDhcpChange(waitCtx, r.client, &op, &resp.Diagnostics); err != nil {
		resp.Diagnostics.AddError("Error disabling network DHCP server", dhcpChangeError(err, rolledBack))
	}
}

func (r *NetworkDhcpServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		resp.Diagnostics.AddError("Invalid import ID", errors.New("expected the network UUID").Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("network"), req.ID)...)
}

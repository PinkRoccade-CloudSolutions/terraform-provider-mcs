package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &NetworkBgpPeeringResource{}
	_ resource.ResourceWithImportState = &NetworkBgpPeeringResource{}
)

// bgpPeeringDefaultTimeout bounds a BGP peering change: one job that configures the Domain's
// FortiGate VDOM.
const bgpPeeringDefaultTimeout = 20 * time.Minute

type NetworkBgpPeeringResource struct {
	client *apiclient.Client
}

type NetworkBgpPeeringResourceModel struct {
	Id               types.String   `tfsdk:"id"`
	Network          types.String   `tfsdk:"network"`
	NeighborGroup    types.String   `tfsdk:"neighbor_group"`
	AcceptedPrefixes types.List     `tfsdk:"accepted_prefixes"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}

// bgpPeeringAPIModel is BgpPeeringV3: a Network's BGP peering, `enabled: false` while it is off.
type bgpPeeringAPIModel struct {
	Enabled          bool     `json:"enabled"`
	NeighborGroup    *string  `json:"neighbor_group"`
	AcceptedPrefixes []string `json:"accepted_prefixes"`
}

func NewNetworkBgpPeeringResource() resource.Resource {
	return &NetworkBgpPeeringResource{}
}

func bgpPeeringPath(networkID string) string {
	return networkPath(networkID) + "bgp-peering/"
}

func (r *NetworkBgpPeeringResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_bgp_peering"
}

func (r *NetworkBgpPeeringResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Switches BGP peering on for an MCS network (one per network) through the v3 networking API, so " +
			"hosts on the network, such as Kubernetes nodes, can peer with the domain's FortiGate VDOM. Every change is " +
			"queued as a job that configures the FortiGate; Terraform waits for it to finish. Destroying the resource " +
			"switches BGP peering off.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "UUID of the network (the BGP peering has no ID of its own).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"network": schema.StringAttribute{
				Required:      true,
				Description:   "UUID of the network to switch BGP peering on for, e.g. `mcs_network.x.id`. Changing it switches peering off on the old network and on for the new one.",
				Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"neighbor_group": schema.StringAttribute{
				Computed:      true,
				Description:   "FortiGate neighbor group of the network (`MCS-K8S-<vlan id>`).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"accepted_prefixes": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Prefixes the VDOM accepts from peers on the network: the load balancer pools of the " +
					"Kubernetes clusters using it. Empty: it accepts nothing. MCS keeps it up to date as clusters change.",
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{Create: true, Update: true, Delete: true}),
		},
	}
}

func (r *NetworkBgpPeeringResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// bgpPeeringFromAPI copies switched-on peering into the model.
func bgpPeeringFromAPI(m *NetworkBgpPeeringResourceModel, networkID string, api *bgpPeeringAPIModel) {
	m.Id = types.StringValue(networkID)
	m.Network = types.StringValue(networkID)
	m.NeighborGroup = strOrNull(api.NeighborGroup)
	m.AcceptedPrefixes = stringList(api.AcceptedPrefixes)
}

// bgpPeeringChangeWarning heads the warnings of an applied BGP peering change.
const bgpPeeringChangeWarning = "BGP peering change applied with a warning"

// switchOn queues a PUT and waits for it. PUT is idempotent: it also (re)applies peering that is
// already on, e.g. after `bgp_peering = true` on the network's create. On failure it returns the
// error message and whether the change may still be applied.
func (r *NetworkBgpPeeringResource) switchOn(ctx context.Context, networkID string, diags *diag.Diagnostics) (msg string, mayBeApplied bool) {
	var op networkOperationAPIModel
	if err := r.client.Put(ctx, bgpPeeringPath(networkID), nil, &op); err != nil {
		if d := apiclient.Detail(err); d != "" && apiclient.IsConflict(err) {
			return fmt.Sprintf("MCS refused to switch BGP peering on for network %s: %s", networkID, d), false
		}
		return err.Error(), false
	}
	if _, rolledBack, err := applyNetworkChange(ctx, r.client, &op, bgpPeeringChangeWarning, diags); err != nil {
		return networkChangeError(err, rolledBack), !rolledBack
	}
	return "", false
}

func (r *NetworkBgpPeeringResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NetworkBgpPeeringResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	networkID := plan.Network.ValueString()

	createTimeout, diags := plan.Timeouts.Create(ctx, bgpPeeringDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	waitCtx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	if msg, mayBeApplied := r.switchOn(waitCtx, networkID, &resp.Diagnostics); msg != "" {
		if mayBeApplied {
			// The change may still be applied: keep the peering in state so Terraform marks it
			// tainted and the next apply puts it right.
			plan.Id = types.StringValue(networkID)
			plan.NeighborGroup = types.StringNull()
			plan.AcceptedPrefixes = stringList(nil)
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		}
		resp.Diagnostics.AddError("Error switching on network BGP peering", msg)
		return
	}

	r.readInto(ctx, networkID, &plan, "Error reading network BGP peering after create", &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// readInto reads the peering after a successful change.
func (r *NetworkBgpPeeringResource) readInto(ctx context.Context, networkID string, m *NetworkBgpPeeringResourceModel, summary string, diags *diag.Diagnostics) {
	var fresh bgpPeeringAPIModel
	if err := r.client.Get(ctx, bgpPeeringPath(networkID), &fresh); err != nil {
		diags.AddError(summary, err.Error())
		return
	}
	if !fresh.Enabled {
		diags.AddError(summary, fmt.Sprintf("The change finished, but network %s reads with BGP peering off.", networkID))
		return
	}
	bgpPeeringFromAPI(m, networkID, &fresh)
}

func (r *NetworkBgpPeeringResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NetworkBgpPeeringResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	networkID := state.Id.ValueString()

	var api bgpPeeringAPIModel
	if err := r.client.Get(ctx, bgpPeeringPath(networkID), &api); err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading network BGP peering", err.Error())
		return
	}
	if !api.Enabled {
		resp.State.RemoveResource(ctx)
		return
	}
	bgpPeeringFromAPI(&state, networkID, &api)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update only runs when the timeouts change (network forces replacement). It switches peering on
// again, which reapplies the VDOM configuration and waits for it.
func (r *NetworkBgpPeeringResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state NetworkBgpPeeringResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	networkID := state.Id.ValueString()
	updateTimeout, diags := plan.Timeouts.Update(ctx, bgpPeeringDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	waitCtx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	if msg, _ := r.switchOn(waitCtx, networkID, &resp.Diagnostics); msg != "" {
		// Keep the prior state; the next refresh shows what the FortiGate runs.
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		resp.Diagnostics.AddError("Error changing network BGP peering", msg)
		return
	}

	r.readInto(ctx, networkID, &plan, "Error reading network BGP peering after update", &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NetworkBgpPeeringResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NetworkBgpPeeringResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	deleteTimeout, diags := state.Timeouts.Delete(ctx, bgpPeeringDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	waitCtx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	networkID := state.Id.ValueString()
	var op networkOperationAPIModel
	err := r.client.DeleteWithResult(waitCtx, bgpPeeringPath(networkID), &op)
	switch {
	case apiclient.IsNotFound(err):
		return
	case apiclient.IsConflict(err):
		resp.Diagnostics.AddError("Cannot switch off network BGP peering", bgpPeeringConflictMessage(networkID, err))
		return
	case err != nil:
		resp.Diagnostics.AddError("Error switching off network BGP peering", err.Error())
		return
	}
	if _, rolledBack, err := applyNetworkChange(waitCtx, r.client, &op, bgpPeeringChangeWarning, &resp.Diagnostics); err != nil {
		resp.Diagnostics.AddError("Error switching off network BGP peering", networkChangeError(err, rolledBack))
	}
}

// bgpPeeringConflictMessage explains a 409 on switching peering off with MCS's own reason, e.g.
// the Kubernetes clusters that still use the network.
func bgpPeeringConflictMessage(networkID string, err error) string {
	reason := apiclient.Detail(err)
	if reason == "" {
		reason = err.Error()
	}
	return fmt.Sprintf("MCS refused to switch BGP peering off for network %s: %s\n\n"+
		"MCS keeps BGP peering on while a Kubernetes cluster that has not been destroyed uses the network, "+
		"and refuses changes while the network is not active or a deploy or teardown of it is queued. "+
		"Destroy the clusters first (make the cluster depend on this resource so Terraform destroys it "+
		"before the peering), or wait for the running job, then apply again.", networkID, reason)
}

func (r *NetworkBgpPeeringResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		resp.Diagnostics.AddError("Invalid import ID", errors.New("expected the network UUID").Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("network"), req.ID)...)
}

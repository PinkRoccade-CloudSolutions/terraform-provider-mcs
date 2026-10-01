package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &NetworkResource{}
	_ resource.ResourceWithImportState = &NetworkResource{}
)

const (
	networkDefaultCreateTimeout = 30 * time.Minute
	networkDefaultDeleteTimeout = 30 * time.Minute
)

// networkProducts are the products a network can be created under (ProductEnum in the v3 spec).
var networkProducts = []string{
	"Algemeen", "CareCTRL", "Caress", "CreAim", "DMZ", "Geniq - Planywhere", "MyHealthOnline",
	"NPQ", "PQConnect", "PRLG", "Quarant", "ValueCare", "Windex",
}

type NetworkResource struct {
	client *apiclient.Client
}

type NetworkResourceModel struct {
	Id                    types.String   `tfsdk:"id"`
	Name                  types.String   `tfsdk:"name"`
	DomainUuid            types.String   `tfsdk:"domain_uuid"`
	Product               types.String   `tfsdk:"product"`
	Bitmask               types.Int64    `tfsdk:"bitmask"`
	Description           types.String   `tfsdk:"description"`
	NetworkPoolId         types.String   `tfsdk:"network_pool_id"`
	Prefix                types.String   `tfsdk:"prefix"`
	DomainId              types.Int64    `tfsdk:"domain_id"`
	DomainName            types.String   `tfsdk:"domain_name"`
	DomainVdomId          types.String   `tfsdk:"domain_vdom_id"`
	Customer              types.String   `tfsdk:"customer"`
	CustomerName          types.String   `tfsdk:"customer_name"`
	Type                  types.String   `tfsdk:"type"`
	Parent                types.String   `tfsdk:"parent"`
	VlanId                types.Int64    `tfsdk:"vlan_id"`
	Ipv4Prefix            types.String   `tfsdk:"ipv4_prefix"`
	Ipv4Address           types.String   `tfsdk:"ipv4_address"`
	Ipv6Prefix            types.String   `tfsdk:"ipv6_prefix"`
	Ipv6Address           types.String   `tfsdk:"ipv6_address"`
	DhcpServer            types.Bool     `tfsdk:"dhcp_server"`
	DhcpServerAddressPool types.String   `tfsdk:"dhcp_server_address_pool"`
	DhcpServerNetmask     types.String   `tfsdk:"dhcp_server_netmask"`
	DhcpServerDnsServers  types.String   `tfsdk:"dhcp_server_dns_servers"`
	DhcpServerLeaseTime   types.Int64    `tfsdk:"dhcp_server_lease_time"`
	Visible               types.Bool     `tfsdk:"visible"`
	CreatedAt             types.String   `tfsdk:"created_at"`
	UpdatedAt             types.String   `tfsdk:"updated_at"`
	Timeouts              timeouts.Value `tfsdk:"timeouts"`
}

type networkCreateAPIModel struct {
	Domain      string  `json:"domain"`
	NetworkPool *string `json:"network_pool,omitempty"`
	Prefix      *string `json:"prefix,omitempty"`
	Bitmask     int64   `json:"bitmask"`
	Description string  `json:"description"`
	Product     string  `json:"product"`
}

func NewNetworkResource() resource.Resource {
	return &NetworkResource{}
}

func (r *NetworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

// The v3 API cannot change a network in place, so every argument forces replacement. The
// exception is a null prior state value: product, network_pool_id and prefix are never returned
// by the API, so after an import they are null in state, and filling them in from the
// configuration must not recreate the network.
const replaceIfStateKnownDesc = "Changing this value recreates the network, unless it is not yet known in state (e.g. after an import)."

func replaceIfStateKnownString() planmodifier.String {
	return stringplanmodifier.RequiresReplaceIf(
		func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
			resp.RequiresReplace = !req.StateValue.IsNull()
		}, replaceIfStateKnownDesc, replaceIfStateKnownDesc)
}

func replaceIfStateKnownInt64() planmodifier.Int64 {
	return int64planmodifier.RequiresReplaceIf(
		func(_ context.Context, req planmodifier.Int64Request, resp *int64planmodifier.RequiresReplaceIfFuncResponse) {
			resp.RequiresReplace = !req.StateValue.IsNull()
		}, replaceIfStateKnownDesc, replaceIfStateKnownDesc)
}

func networkComputedString(desc string) schema.StringAttribute {
	return schema.StringAttribute{
		Computed:      true,
		Description:   desc,
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
}

func networkComputedInt64(desc string) schema.Int64Attribute {
	return schema.Int64Attribute{
		Computed:      true,
		Description:   desc,
		PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
	}
}

func networkComputedBool(desc string) schema.BoolAttribute {
	return schema.BoolAttribute{
		Computed:      true,
		Description:   desc,
		PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
	}
}

func (r *NetworkResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an MCS network through the v3 networking API. Creating the resource queues a deploy " +
			"(VLAN, prefix and interface) and waits for it to finish; destroying it queues a teardown and waits " +
			"for the network to be gone. Networks cannot be changed in place, so changing any argument recreates the network.",
		Attributes: map[string]schema.Attribute{
			"id":   networkComputedString("UUID of the network."),
			"name": networkComputedString("Name of the network, assigned by MCS."),
			"domain_uuid": schema.StringAttribute{
				Required:      true,
				Description:   "UUID of the domain (VDOM) to attach the network to, e.g. `data.mcs_domain.x.uuid`.",
				Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
				PlanModifiers: []planmodifier.String{replaceIfStateKnownString()},
			},
			"product": schema.StringAttribute{
				Required: true,
				Description: "Product that owns the network; picks the PRODID band. One of: " +
					"`" + strings.Join(networkProducts, "`, `") + "`. See `data.mcs_network_products`.",
				Validators:    []validator.String{stringvalidator.OneOf(networkProducts...)},
				PlanModifiers: []planmodifier.String{replaceIfStateKnownString()},
			},
			"bitmask": schema.Int64Attribute{
				Required:      true,
				Description:   "Prefix length of the network (24-29).",
				Validators:    []validator.Int64{int64validator.Between(24, 29)},
				PlanModifiers: []planmodifier.Int64{replaceIfStateKnownInt64()},
			},
			"description": schema.StringAttribute{
				Required:      true,
				Description:   "What the network is for.",
				Validators:    []validator.String{stringvalidator.LengthBetween(1, 4096)},
				PlanModifiers: []planmodifier.String{replaceIfStateKnownString()},
			},
			"network_pool_id": schema.StringAttribute{
				Optional:    true,
				Description: "UUID of the root network pool to allocate the prefix from. Conflicts with `prefix`.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.ConflictsWith(path.MatchRoot("prefix")),
				},
				PlanModifiers: []planmodifier.String{replaceIfStateKnownString()},
			},
			"prefix": schema.StringAttribute{
				Optional:    true,
				Description: "Explicit network address to use instead of allocating from a pool, e.g. `10.0.0.0`. Conflicts with `network_pool_id`.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.ConflictsWith(path.MatchRoot("network_pool_id")),
				},
				PlanModifiers: []planmodifier.String{replaceIfStateKnownString()},
			},
			"domain_id":                networkComputedInt64("Integer ID of the network's domain."),
			"domain_name":              networkComputedString("Name of the network's domain."),
			"domain_vdom_id":           networkComputedString("VDOM ID of the network's domain."),
			"customer":                 networkComputedString("Customer that owns the network."),
			"customer_name":            networkComputedString("Name of the customer that owns the network."),
			"type":                     networkComputedString("Interface type: `vlan`, `tunnel`, `lacp` or `physical`."),
			"parent":                   networkComputedString("Parent LACP interface."),
			"vlan_id":                  networkComputedInt64("VLAN ID assigned to the network."),
			"ipv4_prefix":              networkComputedString("IPv4 prefix of the network."),
			"ipv4_address":             networkComputedString("IPv4 address of the network."),
			"ipv6_prefix":              networkComputedString("IPv6 prefix of the network."),
			"ipv6_address":             networkComputedString("IPv6 address of the network."),
			"dhcp_server":              networkComputedBool("Whether a DHCP server is enabled on the interface."),
			"dhcp_server_address_pool": networkComputedString("DHCP server address pool."),
			"dhcp_server_netmask":      networkComputedString("DHCP server netmask."),
			"dhcp_server_dns_servers":  networkComputedString("Comma-separated DNS servers handed out by the DHCP server."),
			"dhcp_server_lease_time":   networkComputedInt64("DHCP lease time."),
			"visible":                  networkComputedBool("Whether the interface is visible to end users."),
			"created_at":               networkComputedString("Time the network was created."),
			"updated_at":               networkComputedString("Time the network was last updated."),
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{Create: true, Delete: true}),
		},
	}
}

func (r *NetworkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func networkToAPI(plan *NetworkResourceModel) networkCreateAPIModel {
	return networkCreateAPIModel{
		Domain:      plan.DomainUuid.ValueString(),
		NetworkPool: stringPtr(plan.NetworkPoolId),
		Prefix:      stringPtr(plan.Prefix),
		Bitmask:     plan.Bitmask.ValueInt64(),
		Description: plan.Description.ValueString(),
		Product:     plan.Product.ValueString(),
	}
}

// networkFromAPI copies the API record into the model. The create-only arguments the API never
// returns (domain_uuid, product, network_pool_id, prefix) are left untouched; bitmask is only
// derived from ipv4_prefix when it is not known yet (after an import).
func networkFromAPI(m *NetworkResourceModel, api *networkV3APIModel) {
	n := api.toModel()
	m.Id = n.Id
	m.Name = n.Name
	m.Description = n.Description
	m.DomainId = n.Domain
	m.DomainName = n.DomainName
	m.DomainVdomId = n.DomainVdomId
	m.Customer = n.Customer
	m.CustomerName = n.CustomerName
	m.Type = n.Type
	m.Parent = n.Parent
	m.VlanId = n.VlanId
	m.Ipv4Prefix = n.Ipv4Prefix
	m.Ipv4Address = n.Ipv4Address
	m.Ipv6Prefix = n.Ipv6Prefix
	m.Ipv6Address = n.Ipv6Address
	m.DhcpServer = n.DhcpServer
	m.DhcpServerAddressPool = n.DhcpServerAddressPool
	m.DhcpServerNetmask = n.DhcpServerNetmask
	m.DhcpServerDnsServers = n.DhcpServerDnsServers
	m.DhcpServerLeaseTime = n.DhcpServerLeaseTime
	m.Visible = n.Visible
	m.CreatedAt = n.CreatedAt
	m.UpdatedAt = n.UpdatedAt

	if m.Bitmask.IsNull() || m.Bitmask.IsUnknown() {
		m.Bitmask = types.Int64Null()
		if i := strings.LastIndexByte(api.Ipv4Prefix, '/'); i != -1 {
			if bits, err := strconv.ParseInt(api.Ipv4Prefix[i+1:], 10, 64); err == nil {
				m.Bitmask = types.Int64Value(bits)
			}
		}
	}
}

func networkPath(id string) string {
	return fmt.Sprintf("/api/v3/networking/networks/%s/", id)
}

func (r *NetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NetworkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, networkDefaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	waitCtx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	// The deploy is sent once: a retried POST could queue a second network and VLAN.
	var op networkOperationAPIModel
	if err := r.client.PostOnce(waitCtx, "/api/v3/networking/networks/", networkToAPI(&plan), &op); err != nil {
		resp.Diagnostics.AddError("Error creating network",
			fmt.Sprintf("The deploy request failed and was not retried. If the error was a timeout or server error, "+
				"check in MCS whether a network was queued before trying again.\n\n%s", err))
		return
	}

	final, waitErr := waitForNetworkOperation(waitCtx, r.client, &op)
	if final.Network == nil || final.Network.Id == "" {
		if waitErr == nil {
			waitErr = fmt.Errorf("the deploy finished without returning a network: %s", final.describe())
		}
		resp.Diagnostics.AddError("Error creating network", waitErr.Error())
		return
	}

	description := plan.Description
	networkFromAPI(&plan, final.Network)
	if waitErr != nil {
		// The network row exists: keep it in state so Terraform marks it tainted instead of losing it.
		plan.Description = description
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		resp.Diagnostics.AddError("Error creating network",
			fmt.Sprintf("Network %s was created but its deploy did not complete. It has been saved as tainted "+
				"and will be replaced on the next apply.\n\n%s", plan.Id.ValueString(), waitErr))
		return
	}

	var fresh networkV3APIModel
	if err := r.client.Get(ctx, networkPath(plan.Id.ValueString()), &fresh); err != nil {
		plan.Description = description
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		resp.Diagnostics.AddError("Error reading network after create", err.Error())
		return
	}
	networkFromAPI(&plan, &fresh)
	// Keep the configured description so the apply result matches the plan; a normalised value
	// returned by the API shows up on the next refresh.
	plan.Description = description
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NetworkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp networkV3APIModel
	if err := r.client.Get(ctx, networkPath(state.Id.ValueString()), &apiResp); err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading network", err.Error())
		return
	}
	networkFromAPI(&state, &apiResp)

	if state.DomainUuid.IsNull() || state.DomainUuid.IsUnknown() {
		state.DomainUuid = r.lookupDomainUuid(ctx, apiResp.Domain, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// lookupDomainUuid maps the integer domain ID the v3 API returns to the domain UUID the create
// endpoint takes. It is only needed after an import, when domain_uuid is not in state yet.
func (r *NetworkResource) lookupDomainUuid(ctx context.Context, domainID *int64, diags *diag.Diagnostics) types.String {
	if domainID == nil {
		return types.StringNull()
	}
	domains, err := listAll[domainAPIModel](ctx, r.client, "/api/tenant/domains/")
	if err != nil {
		diags.AddError("Error looking up network domain", err.Error())
		return types.StringNull()
	}
	for _, d := range domains {
		if d.Id == *domainID {
			return types.StringValue(d.Uuid)
		}
	}
	diags.AddWarning("Network domain not found",
		fmt.Sprintf("Domain %d of the network was not found, so domain_uuid is left empty and will be taken from the configuration.", *domainID))
	return types.StringNull()
}

// Update is only reached when nothing has to change in MCS: the timeouts changed, or create-only
// arguments the API does not return (e.g. after an import) are filled in from the configuration.
// Every other change forces replacement.
func (r *NetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NetworkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NetworkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, networkDefaultDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	waitCtx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	id := state.Id.ValueString()
	var op networkOperationAPIModel
	err := r.client.DeleteWithResult(waitCtx, networkPath(id), &op)
	switch {
	case apiclient.IsNotFound(err):
		return
	case apiclient.IsConflict(err):
		// A teardown is already running (or there was nothing left to tear down): wait for the
		// network to disappear.
		tflog.Debug(ctx, "network teardown conflict, waiting for network to be removed", map[string]interface{}{"id": id, "error": err.Error()})
	case err != nil:
		resp.Diagnostics.AddError("Error deleting network", err.Error())
		return
	default:
		if _, err := waitForNetworkOperation(waitCtx, r.client, &op); err != nil {
			resp.Diagnostics.AddError("Error deleting network", err.Error())
			return
		}
	}

	if err := waitForNetworkGone(waitCtx, r.client, networkPath(id)); err != nil {
		resp.Diagnostics.AddError("Error deleting network", err.Error())
	}
}

func (r *NetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

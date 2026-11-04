package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &FirewallCommitResource{}

const firewallCommitDefaultCreateTimeout = 20 * time.Minute

// FirewallCommitResource commits a domain's pending firewall changes so they go live. It is an
// action modelled as a resource: creating it commits, changing `triggers` replaces it (and so
// commits again), and destroying it only forgets it.
type FirewallCommitResource struct {
	client *apiclient.Client
}

type FirewallCommitModel struct {
	Id          types.String   `tfsdk:"id"`
	Domain      types.String   `tfsdk:"domain"`
	ChangeId    types.String   `tfsdk:"changeid"`
	Triggers    types.Map      `tfsdk:"triggers"`
	DeployFlow  types.String   `tfsdk:"deploy_flow"`
	CommittedAt types.String   `tfsdk:"committed_at"`
	JobId       types.Int64    `tfsdk:"job_id"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}

type firewallCommitAPI struct {
	ChangeId string `json:"changeid"`
}

type fortiManagerInstallAPI struct {
	Domain   string `json:"domain"`
	ChangeId string `json:"changeid"`
}

type fortiManagerInstallQueuedAPI struct {
	JobId    int64 `json:"job_id"`
	Networks int64 `json:"networks"`
}

func NewFirewallCommitResource() resource.Resource {
	return &FirewallCommitResource{}
}

func (r *FirewallCommitResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_commit"
}

func (r *FirewallCommitResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Commits a domain's pending firewall changes so they go live. V1 domains are committed " +
			"through the firewall commit API; V2 domains get a FortiManager install of their policy package. " +
			"Make it depend on the firewall resources of the domain and put their values in `triggers` so " +
			"every change is committed. Requires the `firewall.direct_management` permission.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Identifier of the last commit: `<domain>/<committed_at>`.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"domain": schema.StringAttribute{
				Required:      true,
				Description:   "Name of the firewall domain (VDOM) to commit, as used by `mcs_firewall_rule`. Changing this commits the new domain.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"changeid": schema.StringAttribute{
				Required: true,
				Description: "Free-text change ID recorded with the commit, e.g. a ticket from your ITSM tool or a pipeline run ID. " +
					"Changing only this does not commit again.",
			},
			"triggers": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Arbitrary values that cause a new commit whenever any of them change, " +
					"e.g. a hash of the firewall rules of the domain.",
				PlanModifiers: []planmodifier.Map{mapplanmodifier.RequiresReplace()},
			},
			"deploy_flow": schema.StringAttribute{
				Computed:      true,
				Description:   "Deploy flow of the domain at commit time: `v1` (firewall commit) or `v2` (FortiManager install).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"committed_at": schema.StringAttribute{
				Computed:      true,
				Description:   "RFC 3339 time of the last commit.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"job_id": schema.Int64Attribute{
				Computed:      true,
				Description:   "V2 only: the FortiManager install job. Null for V1 domains or when nothing was pending.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{Create: true}),
		},
	}
}

func (r *FirewallCommitResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FirewallCommitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FirewallCommitModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, firewallCommitDefaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	name := plan.Domain.ValueString()
	domain, err := findDomainByName(ctx, r.client, name)
	if err != nil {
		resp.Diagnostics.AddError("Error looking up firewall domain", err.Error())
		return
	}

	plan.JobId = types.Int64Null()
	plan.DeployFlow = types.StringValue(domain.DeployFlow)
	switch domain.DeployFlow {
	case "v1":
		r.commitV1(ctx, name, plan.ChangeId.ValueString(), &resp.Diagnostics)
	case "v2":
		plan.JobId = r.installV2(ctx, domain, plan.ChangeId.ValueString(), &resp.Diagnostics)
	default:
		resp.Diagnostics.AddError("Unsupported deploy flow",
			fmt.Sprintf("Domain %q has deploy flow %q; only v1 and v2 can be committed.", name, domain.DeployFlow))
	}
	if resp.Diagnostics.HasError() {
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	plan.CommittedAt = types.StringValue(now)
	plan.Id = types.StringValue(name + "/" + now)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FirewallCommitResource) commitV1(ctx context.Context, domain, changeID string, diags *diag.Diagnostics) {
	path := fmt.Sprintf("/api/networking/domain/%s/commit/", domain)
	err := r.client.PostOnce(ctx, path, firewallCommitAPI{ChangeId: changeID}, nil)
	addCommitError(err, "Error committing firewall domain "+domain, diags)
}

// installV2 queues a FortiManager install and waits for it. It returns the job ID, or null when
// nothing was pending.
func (r *FirewallCommitResource) installV2(ctx context.Context, domain *domainAPIModel, changeID string, diags *diag.Diagnostics) types.Int64 {
	var queued fortiManagerInstallQueuedAPI
	err := r.client.PostOnce(ctx, "/api/v3/networking/fortimanager-installs/",
		fortiManagerInstallAPI{Domain: domain.Uuid, ChangeId: changeID}, &queued)
	if err != nil || queued.JobId == 0 {
		addCommitError(err, "Error installing FortiManager policy package for domain "+domain.Name, diags)
		return types.Int64Null()
	}

	op, err := waitForNetworkOperation(ctx, r.client, &networkOperationAPIModel{JobID: queued.JobId})
	op.addWarnings("FortiManager install warning", diags)
	if err != nil {
		diags.AddError("FortiManager install for domain "+domain.Name+" failed", err.Error())
	}
	return types.Int64Value(queued.JobId)
}

// addCommitError reports a commit or install error. A 409 means nothing was pending (or an
// install is already queued), which is not a reason to fail the apply.
func addCommitError(err error, summary string, diags *diag.Diagnostics) {
	if err == nil {
		return
	}
	if apiclient.IsConflict(err) {
		diags.AddWarning("Nothing to commit",
			"The API reported nothing to commit, or a commit/install is already queued for this domain.\n\n"+err.Error())
		return
	}
	var apiErr *apiclient.APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == 403 {
		diags.AddError(summary, "Committing requires the firewall.direct_management permission; "+
			"Change-Request-only users cannot commit directly.\n\n"+err.Error())
		return
	}
	diags.AddError(summary, err.Error())
}

// Read keeps the state as is: a commit leaves nothing on the server to refresh.
func (r *FirewallCommitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FirewallCommitModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update only happens for changeid or timeouts; it stores them without committing again.
func (r *FirewallCommitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FirewallCommitModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete only removes the resource from state; a commit cannot be undone.
func (r *FirewallCommitResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Firewall commit removed from state only",
		"Destroying mcs_firewall_commit does not roll back the firewall; committed changes stay live.")
}

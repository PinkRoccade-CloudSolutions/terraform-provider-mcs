package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &DnsEntryResource{}
	_ resource.ResourceWithImportState = &DnsEntryResource{}
)

type DnsEntryResource struct {
	client *apiclient.Client
}

type DnsEntryResourceModel struct {
	Id         types.String `tfsdk:"id"`
	DomainUUID types.String `tfsdk:"domain_uuid"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	Content    types.String `tfsdk:"content"`
	Expire     types.Int64  `tfsdk:"expire"`
}

type dnsEntryAPIModel struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	Expire  int    `json:"expire"`
}

func NewDnsEntryResource() resource.Resource {
	return &DnsEntryResource{}
}

func (r *DnsEntryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_entry"
}

func (r *DnsEntryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DNS entry within an MCS DNS domain. Updates are performed by deleting the old entry and creating a new one.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Composite identifier: domain_uuid/name/type/content.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"domain_uuid": schema.StringAttribute{
				Required:      true,
				Description:   "UUID of the DNS domain this entry belongs to.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:      true,
				Description:   "DNS record name (e.g. www, mail).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"type": schema.StringAttribute{
				Required:      true,
				Description:   "DNS record type (e.g. A, AAAA, CNAME, MX, TXT).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"content": schema.StringAttribute{
				Required:      true,
				Description:   "DNS record content (e.g. IP address, hostname).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"expire": schema.Int64Attribute{
				Required:      true,
				Description:   "TTL in seconds (minimum 60, maximum 604800).",
				Validators:    []validator.Int64{int64validator.Between(60, 604800)},
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *DnsEntryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DnsEntryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DnsEntryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := dnsEntryAPIModel{
		Name:    plan.Name.ValueString(),
		Type:    plan.Type.ValueString(),
		Content: plan.Content.ValueString(),
		Expire:  int(plan.Expire.ValueInt64()),
	}

	domainUUID := plan.DomainUUID.ValueString()
	var apiResp dnsEntryAPIModel
	err := r.client.Post(ctx, fmt.Sprintf("/api/dns/domains/%s/entries/", domainUUID), apiReq, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating DNS entry", err.Error())
		return
	}

	plan.Id = types.StringValue(buildDnsEntryID(domainUUID, apiResp.Name, apiResp.Type, apiResp.Content))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DnsEntryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DnsEntryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainUUID := state.DomainUUID.ValueString()
	entries, err := r.listEntries(ctx, domainUUID)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading DNS entries", err.Error())
		return
	}

	wantName := state.Name.ValueString()
	wantType := state.Type.ValueString()
	wantContent := state.Content.ValueString()

	var found *dnsEntryAPIModel
	for i := range entries {
		if entries[i].Name == wantName && entries[i].Type == wantType && entries[i].Content == wantContent {
			found = &entries[i]
			break
		}
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(found.Name)
	state.Type = types.StringValue(found.Type)
	state.Content = types.StringValue(found.Content)
	state.Expire = types.Int64Value(int64(found.Expire))
	state.Id = types.StringValue(buildDnsEntryID(domainUUID, found.Name, found.Type, found.Content))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DnsEntryResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"DNS entries cannot be updated in-place. All attributes use RequiresReplace, so Terraform should destroy and recreate instead.",
	)
}

func (r *DnsEntryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DnsEntryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := dnsEntryAPIModel{
		Name:    state.Name.ValueString(),
		Type:    state.Type.ValueString(),
		Content: state.Content.ValueString(),
		Expire:  int(state.Expire.ValueInt64()),
	}

	domainUUID := state.DomainUUID.ValueString()
	entryName := state.Name.ValueString()
	err := r.client.DeleteWithBody(ctx, fmt.Sprintf("/api/dns/domains/%s/entries/%s/", domainUUID, entryName), apiReq)
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting DNS entry", err.Error())
	}
}

func buildDnsEntryID(domainUUID, name, entryType, content string) string {
	return strings.Join([]string{domainUUID, name, entryType, content}, "/")
}

// listEntries fetches the entries of a domain. The spec documents a bare JSON array, but the
// paginated {results} envelope is accepted as well.
func (r *DnsEntryResource) listEntries(ctx context.Context, domainUUID string) ([]dnsEntryAPIModel, error) {
	path := fmt.Sprintf("/api/dns/domains/%s/entries/", domainUUID)
	var raw json.RawMessage
	if err := r.client.Get(ctx, path, &raw); err != nil {
		return nil, err
	}
	if trimmed := bytes.TrimSpace(raw); len(trimmed) > 0 && trimmed[0] == '[' {
		var entries []dnsEntryAPIModel
		if err := json.Unmarshal(trimmed, &entries); err != nil {
			return nil, fmt.Errorf("parsing DNS entries: %w", err)
		}
		return entries, nil
	}
	return listAll[dnsEntryAPIModel](ctx, r.client, path)
}

// ImportState takes the composite id "<domain_uuid>/<name>/<type>/<content>". There is no
// single-entry GET, so Read matches name, type and content exactly against the zone's entries;
// content is the last part and may itself contain slashes.
func (r *DnsEntryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, ok := parseImportID(req.ID, 4, "<domain_uuid>/<name>/<type>/<content>", &resp.Diagnostics)
	if !ok {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain_uuid"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("content"), parts[3])...)
}

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &DblResource{}
	_ resource.ResourceWithImportState = &DblResource{}
)

type DblResource struct {
	client *apiclient.Client
}

type dblModel struct {
	Id         types.String `tfsdk:"id"`
	IpAddress  types.String `tfsdk:"ipaddress"`
	Timestamp  types.String `tfsdk:"timestamp"`
	Source     types.String `tfsdk:"source"`
	Occurrence types.Int64  `tfsdk:"occurrence"`
	Persistent types.Bool   `tfsdk:"persistent"`
	Blackholed types.Bool   `tfsdk:"blackholed"`
	Hostname   types.String `tfsdk:"hostname"`
	Meta       types.String `tfsdk:"meta"`
	Itsm       types.String `tfsdk:"itsm"`
	Reason     types.String `tfsdk:"reason"`
}

type dblAPIModel struct {
	Id         int     `json:"id,omitempty"`
	IpAddress  string  `json:"ipaddress"`
	Timestamp  string  `json:"timestamp,omitempty"`
	Source     *string `json:"source,omitempty"`
	Occurrence int     `json:"occurrence,omitempty"`
	Persistent *bool   `json:"persistent,omitempty"`
	Blackholed *bool   `json:"blackholed,omitempty"`
	Hostname   string  `json:"hostname,omitempty"`
	Meta       *string `json:"meta,omitempty"`
	Itsm       *string `json:"itsm,omitempty"`
	Reason     *string `json:"reason,omitempty"`
}

func NewDblResource() resource.Resource {
	return &DblResource{}
}

func (r *DblResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dbl"
}

func (r *DblResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ipaddress": schema.StringAttribute{
				Required:    true,
				Description: "IP address to block. The API addresses entries by IP, so changing it replaces the entry.",
				Validators:  []validator.String{stringvalidator.LengthBetween(1, 255)},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"timestamp": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source": schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{stringvalidator.LengthAtMost(255)},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"occurrence": schema.Int64Attribute{
				Computed: true,
			},
			"persistent": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"blackholed": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Drop traffic from the IP address early in the defense-in-depth model.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"hostname": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"meta": schema.StringAttribute{
				Computed: true,
			},
			"itsm": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "ITSM registration in which the block was requested. If none exists, set `reason`.",
				Validators:  []validator.String{stringvalidator.LengthAtMost(255)},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"reason": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Why the IP address is added when no ITSM ticket exists.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *DblResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func dblBodyFromPlan(plan *dblModel) dblAPIModel {
	return dblAPIModel{
		IpAddress:  plan.IpAddress.ValueString(),
		Source:     stringPtr(plan.Source),
		Persistent: boolPtr(plan.Persistent),
		Blackholed: boolPtr(plan.Blackholed),
		Itsm:       stringPtr(plan.Itsm),
		Reason:     stringPtr(plan.Reason),
	}
}

func dblStateFromAPI(result *dblAPIModel, state *dblModel) {
	state.Id = types.StringValue(strconv.Itoa(result.Id))
	state.IpAddress = types.StringValue(result.IpAddress)
	state.Timestamp = types.StringValue(result.Timestamp)
	state.Source = types.StringPointerValue(result.Source)
	state.Occurrence = types.Int64Value(int64(result.Occurrence))
	state.Persistent = types.BoolPointerValue(result.Persistent)
	state.Blackholed = types.BoolPointerValue(result.Blackholed)
	state.Hostname = types.StringValue(result.Hostname)
	state.Meta = types.StringPointerValue(result.Meta)
	state.Itsm = types.StringPointerValue(result.Itsm)
	state.Reason = types.StringPointerValue(result.Reason)
}

func (r *DblResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dblModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := dblBodyFromPlan(&plan)

	var result dblAPIModel
	err := r.client.Post(ctx, "/api/dbl/", body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error creating dbl entry", err.Error())
		return
	}

	dblStateFromAPI(&result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DblResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dblModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result dblAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/dbl/%s/", state.IpAddress.ValueString()), &result)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading dbl entry", err.Error())
		return
	}

	dblStateFromAPI(&result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DblResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dblModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := dblBodyFromPlan(&plan)

	var result dblAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/dbl/%s/", plan.IpAddress.ValueString()), body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error updating dbl entry", err.Error())
		return
	}

	dblStateFromAPI(&result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DblResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dblModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/dbl/%s/", state.IpAddress.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting dbl entry", err.Error())
	}
}

// ImportState takes the IP address, because the API addresses DBL entries by IP rather than by id.
func (r *DblResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("ipaddress"), req, resp)
}

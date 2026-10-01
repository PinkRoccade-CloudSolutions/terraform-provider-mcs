package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &DomainDblResource{}
	_ resource.ResourceWithImportState = &DomainDblResource{}
)

type DomainDblResource struct {
	client *apiclient.Client
}

type domainDblModel struct {
	Id         types.String `tfsdk:"id"`
	DomainName types.String `tfsdk:"domainname"`
	Timestamp  types.String `tfsdk:"timestamp"`
	Source     types.String `tfsdk:"source"`
	Persistent types.Bool   `tfsdk:"persistent"`
	Occurrence types.Int64  `tfsdk:"occurrence"`
	Labels     types.List   `tfsdk:"labels"`
}

// dblLabelAPIModel is the nested Label object ({id, name}) used by the DBL endpoints.
type dblLabelAPIModel struct {
	Id   int    `json:"id,omitempty"`
	Name string `json:"name"`
}

type domainDblAPIModel struct {
	Id         int                `json:"id,omitempty"`
	DomainName string             `json:"domainname"`
	Timestamp  string             `json:"timestamp,omitempty"`
	Source     *string            `json:"source,omitempty"`
	Persistent *bool              `json:"persistent,omitempty"`
	Occurrence int                `json:"occurrence,omitempty"`
	Labels     []dblLabelAPIModel `json:"labels"`
}

func NewDomainDblResource() resource.Resource {
	return &DomainDblResource{}
}

func (r *DomainDblResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_dbl"
}

func (r *DomainDblResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domainname": schema.StringAttribute{
				Required:   true,
				Validators: []validator.String{stringvalidator.LengthBetween(5, 255)},
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
				Validators: []validator.String{stringvalidator.LengthBetween(5, 255)},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"persistent": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"occurrence": schema.Int64Attribute{
				Computed: true,
			},
			"labels": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Names of the DBL labels attached to the entry.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *DomainDblResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func domainDblBodyFromPlan(ctx context.Context, plan *domainDblModel, diags *diag.Diagnostics) domainDblAPIModel {
	body := domainDblAPIModel{
		DomainName: plan.DomainName.ValueString(),
		Source:     stringPtr(plan.Source),
		Persistent: boolPtr(plan.Persistent),
		Labels:     []dblLabelAPIModel{},
	}
	if names := listElems[string](ctx, plan.Labels, diags); names != nil {
		for _, n := range *names {
			body.Labels = append(body.Labels, dblLabelAPIModel{Name: n})
		}
	}
	return body
}

func dblLabelNames(labels []dblLabelAPIModel) []string {
	names := make([]string, 0, len(labels))
	for _, l := range labels {
		names = append(names, l.Name)
	}
	return names
}

func domainDblStateFromAPI(ctx context.Context, result *domainDblAPIModel, state *domainDblModel, diags *diag.Diagnostics) {
	state.Id = types.StringValue(strconv.Itoa(result.Id))
	state.DomainName = types.StringValue(result.DomainName)
	state.Timestamp = types.StringValue(result.Timestamp)
	state.Source = types.StringPointerValue(result.Source)
	state.Persistent = types.BoolPointerValue(result.Persistent)
	state.Occurrence = types.Int64Value(int64(result.Occurrence))
	state.Labels = computedListValue(ctx, types.StringType, dblLabelNames(result.Labels), diags)
}

func (r *DomainDblResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan domainDblModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := domainDblBodyFromPlan(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var result domainDblAPIModel
	err := r.client.Post(ctx, "/api/dbl/domaindbl/", body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error creating domain dbl entry", err.Error())
		return
	}

	domainDblStateFromAPI(ctx, &result, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DomainDblResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state domainDblModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result domainDblAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/dbl/domaindbl/%s/", state.Id.ValueString()), &result)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading domain dbl entry", err.Error())
		return
	}

	domainDblStateFromAPI(ctx, &result, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DomainDblResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan domainDblModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := domainDblBodyFromPlan(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var result domainDblAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/dbl/domaindbl/%s/", plan.Id.ValueString()), body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error updating domain dbl entry", err.Error())
		return
	}

	domainDblStateFromAPI(ctx, &result, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DomainDblResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state domainDblModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/dbl/domaindbl/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting domain dbl entry", err.Error())
	}
}

func (r *DomainDblResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

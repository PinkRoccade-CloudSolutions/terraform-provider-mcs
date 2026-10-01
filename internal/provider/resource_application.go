package provider

import (
	"context"
	"fmt"
	"regexp"

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

const applicationBasePath = "/api/loadbalancing/application/"

var (
	_ resource.Resource                = &ApplicationResource{}
	_ resource.ResourceWithImportState = &ApplicationResource{}
)

// applicationPentestTypes are the values of the spec's PentestTypeEnum.
var applicationPentestTypes = []string{"blackbox", "graybox", "whitebox", "unknown"}

var applicationDateRegexp = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

type ApplicationResource struct {
	client *apiclient.Client
}

type ApplicationResourceModel struct {
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Pentested       types.Bool   `tfsdk:"pentested"`
	PentestType     types.String `tfsdk:"pentest_type"`
	PentestDate     types.String `tfsdk:"pentest_date"`
	PentestFindings types.Int64  `tfsdk:"pentest_findings"`
}

type applicationAPIModel struct {
	Id              string  `json:"id,omitempty"`
	Name            string  `json:"name"`
	Pentested       *bool   `json:"pentested,omitempty"`
	PentestType     *string `json:"pentest_type,omitempty"`
	PentestDate     *string `json:"pentest_date"`
	PentestFindings *int64  `json:"pentest_findings"`
}

func NewApplicationResource() resource.Resource {
	return &ApplicationResource{}
}

func (r *ApplicationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (r *ApplicationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an application in the tenant's application catalogue. Applications are referenced by `mcs_cs_policy.application`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "UUID of the application.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the application (max 255 characters).",
				Validators:  []validator.String{stringvalidator.LengthBetween(1, 255)},
			},
			"pentested": schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Whether the application has been pentested.",
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"pentest_type": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Type of pentest: `blackbox`, `graybox`, `whitebox` or `unknown`.",
				Validators:    []validator.String{stringvalidator.OneOf(applicationPentestTypes...)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"pentest_date": schema.StringAttribute{
				Optional:    true,
				Description: "Date of the pentest (`YYYY-MM-DD`).",
				Validators: []validator.String{
					stringvalidator.RegexMatches(applicationDateRegexp, "must be a date in YYYY-MM-DD format"),
				},
			},
			"pentest_findings": schema.Int64Attribute{
				Optional:    true,
				Description: "Number of findings in the pentest.",
			},
		},
	}
}

func (r *ApplicationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildApplicationAPIRequest(plan *ApplicationResourceModel) applicationAPIModel {
	return applicationAPIModel{
		Name:            plan.Name.ValueString(),
		Pentested:       boolPtr(plan.Pentested),
		PentestType:     stringPtr(plan.PentestType),
		PentestDate:     stringPtr(plan.PentestDate),
		PentestFindings: int64Ptr(plan.PentestFindings),
	}
}

func mapApplicationToState(m *ApplicationResourceModel, a *applicationAPIModel) {
	m.Id = types.StringValue(a.Id)
	m.Name = types.StringValue(a.Name)
	m.Pentested = types.BoolValue(a.Pentested != nil && *a.Pentested)
	if a.PentestType != nil {
		m.PentestType = types.StringValue(*a.PentestType)
	} else {
		m.PentestType = types.StringNull()
	}
	m.PentestDate = types.StringPointerValue(a.PentestDate)
	m.PentestFindings = types.Int64PointerValue(a.PentestFindings)
}

func (r *ApplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApplicationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp applicationAPIModel
	if err := r.client.Post(ctx, applicationBasePath, buildApplicationAPIRequest(&plan), &apiResp); err != nil {
		resp.Diagnostics.AddError("Error creating application", err.Error())
		return
	}

	mapApplicationToState(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp applicationAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("%s%s/", applicationBasePath, state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading application", err.Error())
		return
	}

	mapApplicationToState(&state, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ApplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ApplicationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp applicationAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("%s%s/", applicationBasePath, state.Id.ValueString()), buildApplicationAPIRequest(&plan), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating application", err.Error())
		return
	}

	mapApplicationToState(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApplicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("%s%s/", applicationBasePath, state.Id.ValueString()))
	if err != nil && !apiclient.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting application", err.Error())
	}
}

func (r *ApplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

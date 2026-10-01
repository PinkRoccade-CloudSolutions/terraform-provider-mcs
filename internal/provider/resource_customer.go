package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CustomerResource{}

type CustomerResource struct {
	client *apiclient.Client
}

type CustomerResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	ContractId    types.String `tfsdk:"contractid"`
	Sdm           types.Int64  `tfsdk:"sdm"`
	TechContacts  types.List   `tfsdk:"tech_contacts"`
	AdminContacts types.List   `tfsdk:"admin_contacts"`
}

type customerAPIModel struct {
	Id            string   `json:"id,omitempty"`
	Name          string   `json:"name"`
	ContractId    *string  `json:"contractid,omitempty"`
	Sdm           *int64   `json:"sdm,omitempty"`
	TechContacts  *[]int64 `json:"tech_contacts,omitempty"`
	AdminContacts *[]int64 `json:"admin_contacts,omitempty"`
}

func NewCustomerResource() resource.Resource {
	return &CustomerResource{}
}

func (r *CustomerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_customer"
}

func (r *CustomerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:   true,
				Validators: []validator.String{stringvalidator.LengthBetween(1, 255)},
			},
			"contractid": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Validators:    []validator.String{stringvalidator.LengthAtMost(255)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"sdm": schema.Int64Attribute{
				Optional:      true,
				Computed:      true,
				Description:   "Service Delivery Manager (user id) for the customer; null when not set.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"tech_contacts": schema.ListAttribute{
				Optional:      true,
				Computed:      true,
				ElementType:   types.Int64Type,
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
			"admin_contacts": schema.ListAttribute{
				Optional:      true,
				Computed:      true,
				ElementType:   types.Int64Type,
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *CustomerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func customerBodyFromPlan(ctx context.Context, plan *CustomerResourceModel, diags *diag.Diagnostics) customerAPIModel {
	return customerAPIModel{
		Name:          plan.Name.ValueString(),
		ContractId:    stringPtr(plan.ContractId),
		Sdm:           int64Ptr(plan.Sdm),
		TechContacts:  listElems[int64](ctx, plan.TechContacts, diags),
		AdminContacts: listElems[int64](ctx, plan.AdminContacts, diags),
	}
}

func customerStateFromAPI(ctx context.Context, apiResp *customerAPIModel, state *CustomerResourceModel, diags *diag.Diagnostics) {
	state.Id = types.StringValue(apiResp.Id)
	state.Name = types.StringValue(apiResp.Name)
	if apiResp.ContractId != nil {
		state.ContractId = types.StringValue(*apiResp.ContractId)
	} else {
		state.ContractId = types.StringValue("")
	}
	if apiResp.Sdm != nil {
		state.Sdm = types.Int64Value(*apiResp.Sdm)
	} else {
		state.Sdm = types.Int64Null()
	}
	var tech, admin []int64
	if apiResp.TechContacts != nil {
		tech = *apiResp.TechContacts
	}
	if apiResp.AdminContacts != nil {
		admin = *apiResp.AdminContacts
	}
	state.TechContacts = computedListValue(ctx, types.Int64Type, tech, diags)
	state.AdminContacts = computedListValue(ctx, types.Int64Type, admin, diags)
}

func (r *CustomerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CustomerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := customerBodyFromPlan(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp customerAPIModel
	err := r.client.Post(ctx, "/api/tenant/customers/", apiReq, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating customer", err.Error())
		return
	}

	customerStateFromAPI(ctx, &apiResp, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CustomerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CustomerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp customerAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/tenant/customers/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading customer", err.Error())
		return
	}

	customerStateFromAPI(ctx, &apiResp, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CustomerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CustomerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := customerBodyFromPlan(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp customerAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/tenant/customers/%s/", plan.Id.ValueString()), apiReq, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating customer", err.Error())
		return
	}

	customerStateFromAPI(ctx, &apiResp, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CustomerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CustomerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/tenant/customers/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting customer", err.Error())
	}
}

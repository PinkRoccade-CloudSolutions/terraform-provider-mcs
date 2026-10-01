package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ContactResource{}

type ContactResource struct {
	client *apiclient.Client
}

type ContactResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Company   types.String `tfsdk:"company"`
	Firstname types.String `tfsdk:"firstname"`
	Lastname  types.String `tfsdk:"lastname"`
	Email     types.String `tfsdk:"email"`
	Phone     types.String `tfsdk:"phone"`
	Address   types.String `tfsdk:"address"`
}

type contactAPIModel struct {
	Id        int     `json:"id,omitempty"`
	Company   string  `json:"company"`
	Firstname *string `json:"firstname,omitempty"`
	Lastname  *string `json:"lastname,omitempty"`
	Email     *string `json:"email,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Address   *string `json:"address,omitempty"`
}

func NewContactResource() resource.Resource {
	return &ContactResource{}
}

func (r *ContactResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_contact"
}

func contactOptionalString() schema.StringAttribute {
	return schema.StringAttribute{
		Optional:      true,
		Computed:      true,
		Validators:    []validator.String{stringvalidator.LengthAtMost(255)},
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
}

func (r *ContactResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"company": schema.StringAttribute{
				Required:   true,
				Validators: []validator.String{stringvalidator.LengthBetween(1, 255)},
			},
			"firstname": contactOptionalString(),
			"lastname":  contactOptionalString(),
			"email":     contactOptionalString(),
			"phone":     contactOptionalString(),
			"address":   contactOptionalString(),
		},
	}
}

func (r *ContactResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func contactBodyFromPlan(plan *ContactResourceModel) contactAPIModel {
	return contactAPIModel{
		Company:   plan.Company.ValueString(),
		Firstname: stringPtr(plan.Firstname),
		Lastname:  stringPtr(plan.Lastname),
		Email:     stringPtr(plan.Email),
		Phone:     stringPtr(plan.Phone),
		Address:   stringPtr(plan.Address),
	}
}

func contactStringValue(v *string) types.String {
	if v == nil {
		return types.StringValue("")
	}
	return types.StringValue(*v)
}

func contactStateFromAPI(apiResp *contactAPIModel, state *ContactResourceModel) {
	state.Id = types.StringValue(strconv.Itoa(apiResp.Id))
	state.Company = types.StringValue(apiResp.Company)
	state.Firstname = contactStringValue(apiResp.Firstname)
	state.Lastname = contactStringValue(apiResp.Lastname)
	state.Email = contactStringValue(apiResp.Email)
	state.Phone = contactStringValue(apiResp.Phone)
	state.Address = contactStringValue(apiResp.Address)
}

func (r *ContactResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ContactResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp contactAPIModel
	err := r.client.Post(ctx, "/api/tenant/contacts/", contactBodyFromPlan(&plan), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating contact", err.Error())
		return
	}

	contactStateFromAPI(&apiResp, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ContactResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ContactResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp contactAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/tenant/contacts/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading contact", err.Error())
		return
	}

	contactStateFromAPI(&apiResp, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ContactResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ContactResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp contactAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/tenant/contacts/%s/", plan.Id.ValueString()), contactBodyFromPlan(&plan), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating contact", err.Error())
		return
	}

	contactStateFromAPI(&apiResp, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ContactResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ContactResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/tenant/contacts/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting contact", err.Error())
	}
}

package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CsvServerResource{}

type CsvServerResource struct {
	client *apiclient.Client
}

type CsvServerResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Ufname        types.String `tfsdk:"ufname"`
	Ipaddress     types.String `tfsdk:"ipaddress"`
	Port          types.Int64  `tfsdk:"port"`
	Type          types.String `tfsdk:"type"`
	Policies      types.List   `tfsdk:"policies"`
	Customer      types.String `tfsdk:"customer"`
	Certificate   types.List   `tfsdk:"certificate"`
	CaCertificate types.List   `tfsdk:"ca_certificate"`
	Loadbalancer  types.String `tfsdk:"loadbalancer"`
	Clientauth    types.Bool   `tfsdk:"clientauth"`
	Clientcert    types.String `tfsdk:"clientcert"`
}

type csvServerAPIModel struct {
	Id            string    `json:"id,omitempty"`
	Name          string    `json:"name"`
	Ufname        string    `json:"ufname"`
	Ipaddress     *string   `json:"ipaddress"`
	Port          *int64    `json:"port,omitempty"`
	Type          string    `json:"type"`
	Policies      *[]string `json:"policies,omitempty"`
	Customer      *string   `json:"customer,omitempty"`
	Certificate   *[]string `json:"certificate,omitempty"`
	CaCertificate *[]string `json:"ca_certificate,omitempty"`
	Loadbalancer  *string   `json:"loadbalancer,omitempty"`
	Clientauth    *bool     `json:"clientauth,omitempty"`
	Clientcert    *string   `json:"clientcert,omitempty"`
}

func NewCsvServerResource() resource.Resource {
	return &CsvServerResource{}
}

func (r *CsvServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_csv_server"
}

func (r *CsvServerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"ufname": schema.StringAttribute{
				Required: true,
			},
			"ipaddress": schema.StringAttribute{
				Optional:    true,
				Description: "UUID of the associated PublicIPAddress.",
			},
			"port": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(443),
			},
			"type": schema.StringAttribute{
				Required:   true,
				Validators: []validator.String{stringvalidator.OneOf("http", "ssl")},
			},
			"policies": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"customer": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"certificate": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"ca_certificate": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "UUIDs of CA certificates used to verify client certificates.",
			},
			"loadbalancer": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"clientauth": schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Enable client certificate authentication.",
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"clientcert": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Client certificate requirement: `Mandatory` or `Optional`.",
				Validators:    []validator.String{stringvalidator.OneOf("Mandatory", "Optional")},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *CsvServerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// lbStringSlice dereferences an optional API list.
func lbStringSlice(p *[]string) []string {
	if p == nil {
		return nil
	}
	return *p
}

func csvServerToAPI(ctx context.Context, plan, state *CsvServerResourceModel, diags *diag.Diagnostics) csvServerAPIModel {
	m := csvServerAPIModel{
		Name:         plan.Name.ValueString(),
		Ufname:       plan.Ufname.ValueString(),
		Ipaddress:    stringPtr(plan.Ipaddress),
		Port:         int64Ptr(plan.Port),
		Type:         plan.Type.ValueString(),
		Customer:     stringPtr(plan.Customer),
		Loadbalancer: stringPtr(plan.Loadbalancer),
		Clientauth:   boolPtr(plan.Clientauth),
		Clientcert:   stringPtr(plan.Clientcert),
	}
	if state == nil {
		m.Policies = listElems[string](ctx, plan.Policies, diags)
		m.Certificate = listElems[string](ctx, plan.Certificate, diags)
		m.CaCertificate = listElems[string](ctx, plan.CaCertificate, diags)
	} else {
		m.Policies = listElemsForUpdate[string](ctx, plan.Policies, state.Policies, diags)
		m.Certificate = listElemsForUpdate[string](ctx, plan.Certificate, state.Certificate, diags)
		m.CaCertificate = listElemsForUpdate[string](ctx, plan.CaCertificate, state.CaCertificate, diags)
	}
	return m
}

func csvServerFromAPI(ctx context.Context, m *CsvServerResourceModel, api *csvServerAPIModel, diags *diag.Diagnostics) {
	m.Id = types.StringValue(api.Id)
	m.Name = types.StringValue(api.Name)
	m.Ufname = types.StringValue(api.Ufname)
	m.Ipaddress = types.StringPointerValue(api.Ipaddress)
	if api.Port != nil {
		m.Port = types.Int64Value(*api.Port)
	} else if m.Port.IsUnknown() {
		m.Port = types.Int64Null()
	}
	m.Type = types.StringValue(api.Type)
	m.Policies = listValue(ctx, types.StringType, m.Policies, lbStringSlice(api.Policies), diags)
	m.Customer = types.StringPointerValue(api.Customer)
	m.Certificate = listValue(ctx, types.StringType, m.Certificate, lbStringSlice(api.Certificate), diags)
	m.CaCertificate = listValue(ctx, types.StringType, m.CaCertificate, lbStringSlice(api.CaCertificate), diags)
	m.Loadbalancer = types.StringPointerValue(api.Loadbalancer)
	m.Clientauth = types.BoolPointerValue(api.Clientauth)
	m.Clientcert = types.StringPointerValue(api.Clientcert)
}

func (r *CsvServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CsvServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel := csvServerToAPI(ctx, &plan, nil, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp csvServerAPIModel
	err := r.client.Post(ctx, "/api/loadbalancing/csvserver/", apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating csv_server", err.Error())
		return
	}

	csvServerFromAPI(ctx, &plan, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CsvServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CsvServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp csvServerAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/csvserver/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading csv_server", err.Error())
		return
	}

	csvServerFromAPI(ctx, &state, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CsvServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CsvServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state CsvServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel := csvServerToAPI(ctx, &plan, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp csvServerAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/loadbalancing/csvserver/%s/", state.Id.ValueString()), apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating csv_server", err.Error())
		return
	}

	csvServerFromAPI(ctx, &plan, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CsvServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CsvServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/loadbalancing/csvserver/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting csv_server", err.Error())
	}
}

package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CustomerDataSource{}

type CustomerDataSource struct {
	client *apiclient.Client
}

type CustomerDataSourceModel struct {
	Id            types.String        `tfsdk:"id"`
	Name          types.String        `tfsdk:"name"`
	ContractId    types.String        `tfsdk:"contractid"`
	Sdm           types.Int64         `tfsdk:"sdm"`
	AdminContacts types.List          `tfsdk:"admin_contacts"`
	TechContacts  types.List          `tfsdk:"tech_contacts"`
	Customers     []CustomerListModel `tfsdk:"customers"`
}

type CustomerListModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	ContractId    types.String `tfsdk:"contractid"`
	Sdm           types.Int64  `tfsdk:"sdm"`
	AdminContacts types.List   `tfsdk:"admin_contacts"`
	TechContacts  types.List   `tfsdk:"tech_contacts"`
}

type customerDSAPIModel struct {
	Id            string  `json:"id"`
	Name          string  `json:"name"`
	ContractId    string  `json:"contractid"`
	Sdm           *int64  `json:"sdm"`
	AdminContacts []int64 `json:"admin_contacts"`
	TechContacts  []int64 `json:"tech_contacts"`
}

func NewCustomerDataSource() datasource.DataSource {
	return &CustomerDataSource{}
}

func (d *CustomerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_customer"
}

func (d *CustomerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	itemAttrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true},
		"name": schema.StringAttribute{
			Computed: true,
		},
		"contractid": schema.StringAttribute{Computed: true},
		"sdm":        schema.Int64Attribute{Computed: true},
		"admin_contacts": schema.ListAttribute{
			Computed:    true,
			ElementType: types.Int64Type,
		},
		"tech_contacts": schema.ListAttribute{
			Computed:    true,
			ElementType: types.Int64Type,
		},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS customers by id, by name, or list all.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Description: "Customer id to fetch.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by name (uses name__icontains then exact match).",
			},
			"contractid": schema.StringAttribute{Computed: true},
			"sdm": schema.Int64Attribute{
				Computed:    true,
				Description: "Service Delivery Manager (user id); null when not set.",
			},
			"admin_contacts": schema.ListAttribute{
				Computed:    true,
				ElementType: types.Int64Type,
			},
			"tech_contacts": schema.ListAttribute{
				Computed:    true,
				ElementType: types.Int64Type,
			},
			"customers": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: itemAttrs,
				},
			},
		},
	}
}

func (d *CustomerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type",
			fmt.Sprintf("Expected *apiclient.Client, got %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *CustomerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CustomerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var item customerDSAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/tenant/customers/%s/", config.Id.ValueString()), &item)
		if err != nil {
			resp.Diagnostics.AddError("Error reading customer", err.Error())
			return
		}
		setSingleCustomer(ctx, &config, &item, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	path := "/api/tenant/customers/"
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		path += "?name__icontains=" + url.QueryEscape(config.Name.ValueString())
	}

	items, err := listAll[customerDSAPIModel](ctx, d.client, path)
	if err != nil {
		resp.Diagnostics.AddError("Error reading customers", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		want := config.Name.ValueString()
		var match *customerDSAPIModel
		for i := range items {
			if items[i].Name == want {
				match = &items[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("Customer not found",
				fmt.Sprintf("No customer with exact name %q was found.", want))
			return
		}
		setSingleCustomer(ctx, &config, match, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := CustomerDataSourceModel{
		Id:            types.StringNull(),
		Name:          types.StringNull(),
		ContractId:    types.StringNull(),
		Sdm:           types.Int64Null(),
		AdminContacts: types.ListNull(types.Int64Type),
		TechContacts:  types.ListNull(types.Int64Type),
		Customers:     make([]CustomerListModel, 0, len(items)),
	}
	for i := range items {
		state.Customers = append(state.Customers, customerToListModel(ctx, &items[i], &resp.Diagnostics))
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleCustomer(ctx context.Context, state *CustomerDataSourceModel, item *customerDSAPIModel, diags *diag.Diagnostics) {
	lm := customerToListModel(ctx, item, diags)
	state.Id = lm.Id
	state.Name = lm.Name
	state.ContractId = lm.ContractId
	state.Sdm = lm.Sdm
	state.AdminContacts = lm.AdminContacts
	state.TechContacts = lm.TechContacts
	state.Customers = []CustomerListModel{}
}

func customerToListModel(ctx context.Context, item *customerDSAPIModel, diags *diag.Diagnostics) CustomerListModel {
	sdm := types.Int64Null()
	if item.Sdm != nil {
		sdm = types.Int64Value(*item.Sdm)
	}
	return CustomerListModel{
		Id:            types.StringValue(item.Id),
		Name:          types.StringValue(item.Name),
		ContractId:    types.StringValue(item.ContractId),
		Sdm:           sdm,
		AdminContacts: computedListValue(ctx, types.Int64Type, item.AdminContacts, diags),
		TechContacts:  computedListValue(ctx, types.Int64Type, item.TechContacts, diags),
	}
}

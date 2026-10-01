package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &NATTranslationDataSource{}

type NATTranslationDataSource struct {
	client *apiclient.Client
}

type NATTranslationDataSourceModel struct {
	Id                      types.String          `tfsdk:"id"`
	PublicIP                types.String          `tfsdk:"public_ip"`
	InterfaceField          types.String          `tfsdk:"interface"`
	Firewall                types.String          `tfsdk:"firewall"`
	Translation             types.String          `tfsdk:"translation"`
	PrivateIP               types.String          `tfsdk:"private_ip"`
	TranslationType         types.String          `tfsdk:"translation_type"`
	PublicPort              types.Int64           `tfsdk:"public_port"`
	PrivatePort             types.Int64           `tfsdk:"private_port"`
	Protocol                types.String          `tfsdk:"protocol"`
	Customer                types.String          `tfsdk:"customer"`
	Description             types.String          `tfsdk:"description"`
	State                   types.String          `tfsdk:"state"`
	Enabled                 types.Bool            `tfsdk:"enabled"`
	SnatSourceAddresses     types.List            `tfsdk:"snat_source_addresses"`
	SnatTranslatedAddresses types.List            `tfsdk:"snat_translated_addresses"`
	NatTranslations         []NATTranslationModel `tfsdk:"nat_translations"`
}

type NATTranslationModel struct {
	Id                      types.String `tfsdk:"id"`
	PublicIP                types.String `tfsdk:"public_ip"`
	InterfaceField          types.String `tfsdk:"interface"`
	Firewall                types.String `tfsdk:"firewall"`
	Translation             types.String `tfsdk:"translation"`
	PrivateIP               types.String `tfsdk:"private_ip"`
	TranslationType         types.String `tfsdk:"translation_type"`
	PublicPort              types.Int64  `tfsdk:"public_port"`
	PrivatePort             types.Int64  `tfsdk:"private_port"`
	Protocol                types.String `tfsdk:"protocol"`
	Customer                types.String `tfsdk:"customer"`
	Description             types.String `tfsdk:"description"`
	State                   types.String `tfsdk:"state"`
	Enabled                 types.Bool   `tfsdk:"enabled"`
	SnatSourceAddresses     types.List   `tfsdk:"snat_source_addresses"`
	SnatTranslatedAddresses types.List   `tfsdk:"snat_translated_addresses"`
}

type natTranslationDSAPIModel struct {
	Id                      string         `json:"id"`
	PublicIP                *string        `json:"public_ip"`
	InterfaceField          *string        `json:"interface"`
	Firewall                string         `json:"firewall"`
	Translation             string         `json:"translation"`
	PrivateIP               string         `json:"private_ip"`
	TranslationType         string         `json:"translation_type"`
	PublicPort              *int64         `json:"public_port"`
	PrivatePort             *int64         `json:"private_port"`
	Protocol                string         `json:"protocol"`
	Customer                string         `json:"customer"`
	Description             string         `json:"description"`
	State                   string         `json:"state"`
	Enabled                 bool           `json:"enabled"`
	SnatSourceAddresses     natAddressList `json:"snat_source_addresses"`
	SnatTranslatedAddresses natAddressList `json:"snat_translated_addresses"`
}

func NewNATTranslationDataSource() datasource.DataSource {
	return &NATTranslationDataSource{}
}

func (d *NATTranslationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nat_translation"
}

func (d *NATTranslationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	natAttrs := map[string]schema.Attribute{
		"id":                        schema.StringAttribute{Computed: true},
		"public_ip":                 schema.StringAttribute{Computed: true},
		"interface":                 schema.StringAttribute{Computed: true},
		"firewall":                  schema.StringAttribute{Computed: true},
		"translation":               schema.StringAttribute{Computed: true},
		"private_ip":                schema.StringAttribute{Computed: true},
		"translation_type":          schema.StringAttribute{Computed: true},
		"public_port":               schema.Int64Attribute{Computed: true},
		"private_port":              schema.Int64Attribute{Computed: true},
		"protocol":                  schema.StringAttribute{Computed: true},
		"customer":                  schema.StringAttribute{Computed: true},
		"description":               schema.StringAttribute{Computed: true},
		"state":                     schema.StringAttribute{Computed: true},
		"enabled":                   schema.BoolAttribute{Computed: true},
		"snat_source_addresses":     schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"snat_translated_addresses": schema.ListAttribute{Computed: true, ElementType: types.StringType},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS NAT translations. Set `id` to fetch a single translation, or omit it to list all.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific NAT translation to look up.",
			},
			"public_ip":                 schema.StringAttribute{Computed: true},
			"interface":                 schema.StringAttribute{Computed: true},
			"firewall":                  schema.StringAttribute{Computed: true},
			"translation":               schema.StringAttribute{Computed: true},
			"private_ip":                schema.StringAttribute{Computed: true},
			"translation_type":          schema.StringAttribute{Computed: true},
			"public_port":               schema.Int64Attribute{Computed: true},
			"private_port":              schema.Int64Attribute{Computed: true},
			"protocol":                  schema.StringAttribute{Computed: true},
			"customer":                  schema.StringAttribute{Computed: true},
			"description":               schema.StringAttribute{Computed: true},
			"state":                     schema.StringAttribute{Computed: true},
			"enabled":                   schema.BoolAttribute{Computed: true},
			"snat_source_addresses":     schema.ListAttribute{Computed: true, ElementType: types.StringType},
			"snat_translated_addresses": schema.ListAttribute{Computed: true, ElementType: types.StringType},
			"nat_translations": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All NAT translations (populated when `id` is not set).",
				NestedObject: schema.NestedAttributeObject{Attributes: natAttrs},
			},
		},
	}
}

func (d *NATTranslationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NATTranslationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config NATTranslationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var nat natTranslationDSAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/networking/nattranslations/%s/", config.Id.ValueString()), &nat)
		if err != nil {
			resp.Diagnostics.AddError("Error reading NAT translation", err.Error())
			return
		}
		m := toNATTranslationListModel(ctx, &nat, &resp.Diagnostics)
		config = NATTranslationDataSourceModel{
			Id:                      m.Id,
			PublicIP:                m.PublicIP,
			InterfaceField:          m.InterfaceField,
			Firewall:                m.Firewall,
			Translation:             m.Translation,
			PrivateIP:               m.PrivateIP,
			TranslationType:         m.TranslationType,
			PublicPort:              m.PublicPort,
			PrivatePort:             m.PrivatePort,
			Protocol:                m.Protocol,
			Customer:                m.Customer,
			Description:             m.Description,
			State:                   m.State,
			Enabled:                 m.Enabled,
			SnatSourceAddresses:     m.SnatSourceAddresses,
			SnatTranslatedAddresses: m.SnatTranslatedAddresses,
			NatTranslations:         []NATTranslationModel{},
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	items, err := listAll[natTranslationDSAPIModel](ctx, d.client, "/api/networking/nattranslations/")
	if err != nil {
		resp.Diagnostics.AddError("Error reading NAT translations", err.Error())
		return
	}

	state := NATTranslationDataSourceModel{
		Id:                      types.StringNull(),
		PublicIP:                types.StringNull(),
		InterfaceField:          types.StringNull(),
		Firewall:                types.StringNull(),
		Translation:             types.StringNull(),
		PrivateIP:               types.StringNull(),
		TranslationType:         types.StringNull(),
		PublicPort:              types.Int64Null(),
		PrivatePort:             types.Int64Null(),
		Protocol:                types.StringNull(),
		Customer:                types.StringNull(),
		Description:             types.StringNull(),
		State:                   types.StringNull(),
		Enabled:                 types.BoolNull(),
		SnatSourceAddresses:     types.ListNull(types.StringType),
		SnatTranslatedAddresses: types.ListNull(types.StringType),
		NatTranslations:         make([]NATTranslationModel, 0, len(items)),
	}
	for i := range items {
		state.NatTranslations = append(state.NatTranslations, toNATTranslationListModel(ctx, &items[i], &resp.Diagnostics))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func toNATTranslationListModel(ctx context.Context, nat *natTranslationDSAPIModel, diags *diag.Diagnostics) NATTranslationModel {
	return NATTranslationModel{
		Id:                      types.StringValue(nat.Id),
		PublicIP:                types.StringPointerValue(nat.PublicIP),
		InterfaceField:          types.StringPointerValue(nat.InterfaceField),
		Firewall:                types.StringValue(nat.Firewall),
		Translation:             types.StringValue(nat.Translation),
		PrivateIP:               types.StringValue(nat.PrivateIP),
		TranslationType:         types.StringValue(nat.TranslationType),
		PublicPort:              types.Int64PointerValue(nat.PublicPort),
		PrivatePort:             types.Int64PointerValue(nat.PrivatePort),
		Protocol:                types.StringValue(nat.Protocol),
		Customer:                types.StringValue(nat.Customer),
		Description:             types.StringValue(nat.Description),
		State:                   types.StringValue(nat.State),
		Enabled:                 types.BoolValue(nat.Enabled),
		SnatSourceAddresses:     computedListValue(ctx, types.StringType, []string(nat.SnatSourceAddresses), diags),
		SnatTranslatedAddresses: computedListValue(ctx, types.StringType, []string(nat.SnatTranslatedAddresses), diags),
	}
}

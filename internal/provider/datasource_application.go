package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ApplicationDataSource{}

type ApplicationDataSource struct {
	client *apiclient.Client
}

type ApplicationDataSourceModel struct {
	Id              types.String       `tfsdk:"id"`
	Name            types.String       `tfsdk:"name"`
	Pentested       types.Bool         `tfsdk:"pentested"`
	PentestType     types.String       `tfsdk:"pentest_type"`
	PentestDate     types.String       `tfsdk:"pentest_date"`
	PentestFindings types.Int64        `tfsdk:"pentest_findings"`
	Applications    []ApplicationModel `tfsdk:"applications"`
}

type ApplicationModel struct {
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Pentested       types.Bool   `tfsdk:"pentested"`
	PentestType     types.String `tfsdk:"pentest_type"`
	PentestDate     types.String `tfsdk:"pentest_date"`
	PentestFindings types.Int64  `tfsdk:"pentest_findings"`
}

func NewApplicationDataSource() datasource.DataSource {
	return &ApplicationDataSource{}
}

func (d *ApplicationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (d *ApplicationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	itemAttrs := map[string]schema.Attribute{
		"id":               schema.StringAttribute{Computed: true},
		"name":             schema.StringAttribute{Computed: true},
		"pentested":        schema.BoolAttribute{Computed: true},
		"pentest_type":     schema.StringAttribute{Computed: true},
		"pentest_date":     schema.StringAttribute{Computed: true},
		"pentest_findings": schema.Int64Attribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up applications in the tenant's application catalogue. Set `name` or `id` to fetch a single application, or omit both to list all.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific application to look up.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Exact application name to look up.",
			},
			"pentested": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the application has been pentested.",
			},
			"pentest_type": schema.StringAttribute{
				Computed:    true,
				Description: "Type of pentest: blackbox, graybox, whitebox or unknown.",
			},
			"pentest_date": schema.StringAttribute{
				Computed:    true,
				Description: "Date of the pentest (YYYY-MM-DD).",
			},
			"pentest_findings": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of findings in the pentest.",
			},
			"applications": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All applications (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: itemAttrs},
			},
		},
	}
}

func (d *ApplicationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ApplicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ApplicationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var item applicationAPIModel
		if err := d.client.Get(ctx, fmt.Sprintf("%s%s/", applicationBasePath, config.Id.ValueString()), &item); err != nil {
			resp.Diagnostics.AddError("Error reading application", err.Error())
			return
		}
		state := applicationSingleState(&item)
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	// The list endpoint documents no filters, so name matching happens client-side.
	items, err := listAll[applicationAPIModel](ctx, d.client, applicationBasePath)
	if err != nil {
		resp.Diagnostics.AddError("Error listing applications", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		for i := range items {
			if items[i].Name == config.Name.ValueString() {
				state := applicationSingleState(&items[i])
				resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
				return
			}
		}
		resp.Diagnostics.AddError("Application not found",
			fmt.Sprintf("No application with exact name %q was found.", config.Name.ValueString()))
		return
	}

	state := ApplicationDataSourceModel{
		Id:              types.StringNull(),
		Name:            types.StringNull(),
		Pentested:       types.BoolNull(),
		PentestType:     types.StringNull(),
		PentestDate:     types.StringNull(),
		PentestFindings: types.Int64Null(),
		Applications:    make([]ApplicationModel, 0, len(items)),
	}
	for i := range items {
		state.Applications = append(state.Applications, applicationItemToModel(&items[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func applicationItemToModel(a *applicationAPIModel) ApplicationModel {
	var r ApplicationResourceModel
	mapApplicationToState(&r, a)
	return ApplicationModel(r)
}

func applicationSingleState(a *applicationAPIModel) ApplicationDataSourceModel {
	m := applicationItemToModel(a)
	return ApplicationDataSourceModel{
		Id:              m.Id,
		Name:            m.Name,
		Pentested:       m.Pentested,
		PentestType:     m.PentestType,
		PentestDate:     m.PentestDate,
		PentestFindings: m.PentestFindings,
		Applications:    []ApplicationModel{},
	}
}

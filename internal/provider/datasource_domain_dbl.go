package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DomainDblDataSource{}

type DomainDblDataSource struct {
	client *apiclient.Client
}

type DomainDblDataSourceModel struct {
	Id         types.String         `tfsdk:"id"`
	DomainName types.String         `tfsdk:"domainname"`
	Timestamp  types.String         `tfsdk:"timestamp"`
	Source     types.String         `tfsdk:"source"`
	Persistent types.Bool           `tfsdk:"persistent"`
	Occurrence types.Int64          `tfsdk:"occurrence"`
	Labels     types.List           `tfsdk:"labels"`
	DomainDbls []DomainDblListModel `tfsdk:"domain_dbls"`
}

type DomainDblListModel struct {
	Id         types.String `tfsdk:"id"`
	DomainName types.String `tfsdk:"domainname"`
	Timestamp  types.String `tfsdk:"timestamp"`
	Source     types.String `tfsdk:"source"`
	Persistent types.Bool   `tfsdk:"persistent"`
	Occurrence types.Int64  `tfsdk:"occurrence"`
	Labels     types.List   `tfsdk:"labels"`
}

func NewDomainDblDataSource() datasource.DataSource {
	return &DomainDblDataSource{}
}

func (d *DomainDblDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_dbl"
}

func (d *DomainDblDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	itemAttrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true},
		"domainname": schema.StringAttribute{
			Computed: true,
		},
		"timestamp":  schema.StringAttribute{Computed: true},
		"source":     schema.StringAttribute{Computed: true},
		"persistent": schema.BoolAttribute{Computed: true},
		"occurrence": schema.Int64Attribute{Computed: true},
		"labels":     schema.ListAttribute{Computed: true, ElementType: types.StringType},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS domain DBL entries by id or list all.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Description: "Numeric id to fetch a single entry; omit to list all.",
			},
			"domainname": schema.StringAttribute{Computed: true},
			"timestamp":  schema.StringAttribute{Computed: true},
			"source":     schema.StringAttribute{Computed: true},
			"persistent": schema.BoolAttribute{Computed: true},
			"occurrence": schema.Int64Attribute{Computed: true},
			"labels": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Names of the DBL labels attached to the entry.",
			},
			"domain_dbls": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: itemAttrs,
				},
			},
		},
	}
}

func (d *DomainDblDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DomainDblDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DomainDblDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var item domainDblAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/dbl/domaindbl/%s/", config.Id.ValueString()), &item)
		if err != nil {
			resp.Diagnostics.AddError("Error reading domain dbl entry", err.Error())
			return
		}
		setSingleDomainDbl(ctx, &config, &item, &resp.Diagnostics)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	items, err := listAll[domainDblAPIModel](ctx, d.client, "/api/dbl/domaindbl/")
	if err != nil {
		resp.Diagnostics.AddError("Error reading domain dbl entries", err.Error())
		return
	}

	state := DomainDblDataSourceModel{
		Id:         types.StringNull(),
		DomainName: types.StringNull(),
		Timestamp:  types.StringNull(),
		Source:     types.StringNull(),
		Persistent: types.BoolNull(),
		Occurrence: types.Int64Null(),
		Labels:     types.ListNull(types.StringType),
		DomainDbls: make([]DomainDblListModel, 0, len(items)),
	}
	for i := range items {
		state.DomainDbls = append(state.DomainDbls, domainDblToListModel(ctx, &items[i], &resp.Diagnostics))
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleDomainDbl(ctx context.Context, state *DomainDblDataSourceModel, item *domainDblAPIModel, diags *diag.Diagnostics) {
	lm := domainDblToListModel(ctx, item, diags)
	state.Id = lm.Id
	state.DomainName = lm.DomainName
	state.Timestamp = lm.Timestamp
	state.Source = lm.Source
	state.Persistent = lm.Persistent
	state.Occurrence = lm.Occurrence
	state.Labels = lm.Labels
	state.DomainDbls = []DomainDblListModel{}
}

func domainDblToListModel(ctx context.Context, item *domainDblAPIModel, diags *diag.Diagnostics) DomainDblListModel {
	return DomainDblListModel{
		Id:         types.StringValue(strconv.Itoa(item.Id)),
		DomainName: types.StringValue(item.DomainName),
		Timestamp:  types.StringValue(item.Timestamp),
		Source:     types.StringPointerValue(item.Source),
		Persistent: types.BoolPointerValue(item.Persistent),
		Occurrence: types.Int64Value(int64(item.Occurrence)),
		Labels:     computedListValue(ctx, types.StringType, dblLabelNames(item.Labels), diags),
	}
}

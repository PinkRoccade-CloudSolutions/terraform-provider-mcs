package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DblDataSource{}

type DblDataSource struct {
	client *apiclient.Client
}

type DblDataSourceModel struct {
	IpAddress  types.String   `tfsdk:"ipaddress"`
	Id         types.String   `tfsdk:"id"`
	Timestamp  types.String   `tfsdk:"timestamp"`
	Source     types.String   `tfsdk:"source"`
	Occurrence types.Int64    `tfsdk:"occurrence"`
	Persistent types.Bool     `tfsdk:"persistent"`
	Blackholed types.Bool     `tfsdk:"blackholed"`
	Hostname   types.String   `tfsdk:"hostname"`
	Meta       types.String   `tfsdk:"meta"`
	Itsm       types.String   `tfsdk:"itsm"`
	Reason     types.String   `tfsdk:"reason"`
	Dbls       []DblListModel `tfsdk:"dbls"`
}

type DblListModel struct {
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

func NewDblDataSource() datasource.DataSource {
	return &DblDataSource{}
}

func (d *DblDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dbl"
}

func (d *DblDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	dblAttrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true},
		"ipaddress": schema.StringAttribute{
			Computed: true,
		},
		"timestamp":  schema.StringAttribute{Computed: true},
		"source":     schema.StringAttribute{Computed: true},
		"occurrence": schema.Int64Attribute{Computed: true},
		"persistent": schema.BoolAttribute{Computed: true},
		"blackholed": schema.BoolAttribute{Computed: true},
		"hostname":   schema.StringAttribute{Computed: true},
		"meta":       schema.StringAttribute{Computed: true},
		"itsm":       schema.StringAttribute{Computed: true},
		"reason":     schema.StringAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS DBL entries by ipaddress or list all.",
		Attributes: map[string]schema.Attribute{
			"ipaddress": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "IP address to look up; omit to list all.",
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"timestamp":  schema.StringAttribute{Computed: true},
			"source":     schema.StringAttribute{Computed: true},
			"occurrence": schema.Int64Attribute{Computed: true},
			"persistent": schema.BoolAttribute{Computed: true},
			"blackholed": schema.BoolAttribute{Computed: true},
			"hostname":   schema.StringAttribute{Computed: true},
			"meta":       schema.StringAttribute{Computed: true},
			"itsm":       schema.StringAttribute{Computed: true},
			"reason":     schema.StringAttribute{Computed: true},
			"dbls": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: dblAttrs,
				},
			},
		},
	}
}

func (d *DblDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DblDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DblDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.IpAddress.IsNull() && config.IpAddress.ValueString() != "" {
		var item dblAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/dbl/%s/", config.IpAddress.ValueString()), &item)
		if err != nil {
			resp.Diagnostics.AddError("Error reading dbl entry", err.Error())
			return
		}
		setSingleDbl(&config, &item)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	items, err := listAll[dblAPIModel](ctx, d.client, "/api/dbl/")
	if err != nil {
		resp.Diagnostics.AddError("Error reading dbl entries", err.Error())
		return
	}

	state := DblDataSourceModel{
		IpAddress:  types.StringNull(),
		Id:         types.StringNull(),
		Timestamp:  types.StringNull(),
		Source:     types.StringNull(),
		Occurrence: types.Int64Null(),
		Persistent: types.BoolNull(),
		Blackholed: types.BoolNull(),
		Hostname:   types.StringNull(),
		Meta:       types.StringNull(),
		Itsm:       types.StringNull(),
		Reason:     types.StringNull(),
		Dbls:       make([]DblListModel, 0, len(items)),
	}
	for i := range items {
		state.Dbls = append(state.Dbls, dblToListModel(&items[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleDbl(state *DblDataSourceModel, item *dblAPIModel) {
	lm := dblToListModel(item)
	state.IpAddress = lm.IpAddress
	state.Id = lm.Id
	state.Timestamp = lm.Timestamp
	state.Source = lm.Source
	state.Occurrence = lm.Occurrence
	state.Persistent = lm.Persistent
	state.Blackholed = lm.Blackholed
	state.Hostname = lm.Hostname
	state.Meta = lm.Meta
	state.Itsm = lm.Itsm
	state.Reason = lm.Reason
	state.Dbls = []DblListModel{}
}

func dblToListModel(item *dblAPIModel) DblListModel {
	return DblListModel{
		Id:         types.StringValue(strconv.Itoa(item.Id)),
		IpAddress:  types.StringValue(item.IpAddress),
		Timestamp:  types.StringValue(item.Timestamp),
		Source:     types.StringPointerValue(item.Source),
		Occurrence: types.Int64Value(int64(item.Occurrence)),
		Persistent: types.BoolPointerValue(item.Persistent),
		Blackholed: types.BoolPointerValue(item.Blackholed),
		Hostname:   types.StringValue(item.Hostname),
		Meta:       types.StringPointerValue(item.Meta),
		Itsm:       types.StringPointerValue(item.Itsm),
		Reason:     types.StringPointerValue(item.Reason),
	}
}

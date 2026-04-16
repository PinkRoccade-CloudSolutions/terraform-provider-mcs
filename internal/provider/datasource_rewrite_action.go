package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &RewriteActionDataSource{}

type RewriteActionDataSource struct {
	client *apiclient.Client
}

type RewriteActionDataSourceModel struct {
	Name              types.String         `tfsdk:"name"`
	Id                types.String         `tfsdk:"id"`
	Type              types.String         `tfsdk:"type"`
	Target            types.String         `tfsdk:"target"`
	Stringbuilderexpr types.String         `tfsdk:"stringbuilderexpr"`
	Search            types.String         `tfsdk:"search"`
	Comment           types.String         `tfsdk:"comment"`
	Customer          types.String         `tfsdk:"customer"`
	Loadbalancer      types.String         `tfsdk:"loadbalancer"`
	RewriteActions    []RewriteActionModel `tfsdk:"rewrite_actions"`
}

type RewriteActionModel struct {
	Id                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Type              types.String `tfsdk:"type"`
	Target            types.String `tfsdk:"target"`
	Stringbuilderexpr types.String `tfsdk:"stringbuilderexpr"`
	Search            types.String `tfsdk:"search"`
	Comment           types.String `tfsdk:"comment"`
	Customer          types.String `tfsdk:"customer"`
	Loadbalancer      types.String `tfsdk:"loadbalancer"`
}

type rewriteActionDSAPIModel struct {
	Id                string  `json:"id"`
	Name              string  `json:"name"`
	Type              *string `json:"type,omitempty"`
	Target            *string `json:"target,omitempty"`
	Stringbuilderexpr *string `json:"stringbuilderexpr,omitempty"`
	Search            *string `json:"search,omitempty"`
	Comment           *string `json:"comment,omitempty"`
	Customer          *string `json:"customer,omitempty"`
	Loadbalancer      *string `json:"loadbalancer,omitempty"`
}

func NewRewriteActionDataSource() datasource.DataSource {
	return &RewriteActionDataSource{}
}

func (d *RewriteActionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rewrite_action"
}

func (d *RewriteActionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	actionAttrs := map[string]schema.Attribute{
		"id":                schema.StringAttribute{Computed: true},
		"name":              schema.StringAttribute{Computed: true},
		"type":              schema.StringAttribute{Computed: true},
		"target":            schema.StringAttribute{Computed: true},
		"stringbuilderexpr": schema.StringAttribute{Computed: true},
		"search":            schema.StringAttribute{Computed: true},
		"comment":           schema.StringAttribute{Computed: true},
		"customer":          schema.StringAttribute{Computed: true},
		"loadbalancer":      schema.StringAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS rewrite actions. Set `name` or `id` to fetch a single rewrite action, or omit both to list all.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact rewrite action name to look up.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific rewrite action to look up.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "Rewrite action type.",
			},
			"target": schema.StringAttribute{
				Computed:    true,
				Description: "Rewrite target expression.",
			},
			"stringbuilderexpr": schema.StringAttribute{
				Computed:    true,
				Description: "String builder expression.",
			},
			"search": schema.StringAttribute{
				Computed:    true,
				Description: "Search expression.",
			},
			"comment": schema.StringAttribute{
				Computed:    true,
				Description: "Rewrite action comment.",
			},
			"customer": schema.StringAttribute{
				Computed:    true,
				Description: "Customer identifier.",
			},
			"loadbalancer": schema.StringAttribute{
				Computed:    true,
				Description: "Associated load balancer.",
			},
			"rewrite_actions": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All rewrite actions (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: actionAttrs},
			},
		},
	}
}

func (d *RewriteActionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RewriteActionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config RewriteActionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var item rewriteActionDSAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/rewriteaction/%s/", config.Id.ValueString()), &item)
		if err != nil {
			resp.Diagnostics.AddError("Error reading rewrite action", err.Error())
			return
		}
		setSingleRewriteAction(&config, &item)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	path := "/api/loadbalancing/rewriteaction/?page_size=1000"
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		path += "&name__icontains=" + url.QueryEscape(config.Name.ValueString())
	}

	var page struct {
		Results []rewriteActionDSAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, path, &page); err != nil {
		resp.Diagnostics.AddError("Error reading rewrite actions", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *rewriteActionDSAPIModel
		for i := range page.Results {
			if page.Results[i].Name == config.Name.ValueString() {
				match = &page.Results[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("Rewrite action not found",
				fmt.Sprintf("No rewrite action with exact name %q was found.", config.Name.ValueString()))
			return
		}
		setSingleRewriteAction(&config, match)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := RewriteActionDataSourceModel{
		Name:              types.StringNull(),
		Id:                types.StringNull(),
		Type:              types.StringNull(),
		Target:            types.StringNull(),
		Stringbuilderexpr: types.StringNull(),
		Search:            types.StringNull(),
		Comment:           types.StringNull(),
		Customer:          types.StringNull(),
		Loadbalancer:      types.StringNull(),
		RewriteActions:    make([]RewriteActionModel, 0, len(page.Results)),
	}
	for i := range page.Results {
		state.RewriteActions = append(state.RewriteActions, rewriteActionItemToModel(&page.Results[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleRewriteAction(state *RewriteActionDataSourceModel, item *rewriteActionDSAPIModel) {
	state.Id = types.StringValue(item.Id)
	state.Name = types.StringValue(item.Name)
	if item.Type != nil {
		state.Type = types.StringValue(*item.Type)
	} else {
		state.Type = types.StringNull()
	}
	if item.Target != nil {
		state.Target = types.StringValue(*item.Target)
	} else {
		state.Target = types.StringNull()
	}
	if item.Stringbuilderexpr != nil {
		state.Stringbuilderexpr = types.StringValue(*item.Stringbuilderexpr)
	} else {
		state.Stringbuilderexpr = types.StringNull()
	}
	if item.Search != nil {
		state.Search = types.StringValue(*item.Search)
	} else {
		state.Search = types.StringNull()
	}
	if item.Comment != nil {
		state.Comment = types.StringValue(*item.Comment)
	} else {
		state.Comment = types.StringNull()
	}
	if item.Customer != nil {
		state.Customer = types.StringValue(*item.Customer)
	} else {
		state.Customer = types.StringNull()
	}
	if item.Loadbalancer != nil {
		state.Loadbalancer = types.StringValue(*item.Loadbalancer)
	} else {
		state.Loadbalancer = types.StringNull()
	}
	state.RewriteActions = []RewriteActionModel{}
}

func rewriteActionItemToModel(item *rewriteActionDSAPIModel) RewriteActionModel {
	m := RewriteActionModel{
		Id:   types.StringValue(item.Id),
		Name: types.StringValue(item.Name),
	}
	if item.Type != nil {
		m.Type = types.StringValue(*item.Type)
	} else {
		m.Type = types.StringNull()
	}
	if item.Target != nil {
		m.Target = types.StringValue(*item.Target)
	} else {
		m.Target = types.StringNull()
	}
	if item.Stringbuilderexpr != nil {
		m.Stringbuilderexpr = types.StringValue(*item.Stringbuilderexpr)
	} else {
		m.Stringbuilderexpr = types.StringNull()
	}
	if item.Search != nil {
		m.Search = types.StringValue(*item.Search)
	} else {
		m.Search = types.StringNull()
	}
	if item.Comment != nil {
		m.Comment = types.StringValue(*item.Comment)
	} else {
		m.Comment = types.StringNull()
	}
	if item.Customer != nil {
		m.Customer = types.StringValue(*item.Customer)
	} else {
		m.Customer = types.StringNull()
	}
	if item.Loadbalancer != nil {
		m.Loadbalancer = types.StringValue(*item.Loadbalancer)
	} else {
		m.Loadbalancer = types.StringNull()
	}
	return m
}

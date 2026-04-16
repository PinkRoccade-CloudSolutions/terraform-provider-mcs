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

var _ datasource.DataSource = &RewritePolicyDataSource{}

type RewritePolicyDataSource struct {
	client *apiclient.Client
}

type RewritePolicyDataSourceModel struct {
	Name                   types.String         `tfsdk:"name"`
	Id                     types.String         `tfsdk:"id"`
	Rule                   types.String         `tfsdk:"rule"`
	Action                 types.String         `tfsdk:"action"`
	Undefaction            types.String         `tfsdk:"undefaction"`
	Comment                types.String         `tfsdk:"comment"`
	Priority               types.Float64        `tfsdk:"priority"`
	Bindpoint              types.String         `tfsdk:"bindpoint"`
	Gotopriorityexpression types.String         `tfsdk:"gotopriorityexpression"`
	Customer               types.String         `tfsdk:"customer"`
	Loadbalancer           types.String         `tfsdk:"loadbalancer"`
	RewritePolicies        []RewritePolicyModel `tfsdk:"rewrite_policies"`
}

type RewritePolicyModel struct {
	Id                     types.String  `tfsdk:"id"`
	Name                   types.String  `tfsdk:"name"`
	Rule                   types.String  `tfsdk:"rule"`
	Action                 types.String  `tfsdk:"action"`
	Undefaction            types.String  `tfsdk:"undefaction"`
	Comment                types.String  `tfsdk:"comment"`
	Priority               types.Float64 `tfsdk:"priority"`
	Bindpoint              types.String  `tfsdk:"bindpoint"`
	Gotopriorityexpression types.String  `tfsdk:"gotopriorityexpression"`
	Customer               types.String  `tfsdk:"customer"`
	Loadbalancer           types.String  `tfsdk:"loadbalancer"`
}

type rewritePolicyDSAPIModel struct {
	Id                     string   `json:"id"`
	Name                   string   `json:"name"`
	Rule                   *string  `json:"rule,omitempty"`
	Action                 *string  `json:"action,omitempty"`
	Undefaction            *string  `json:"undefaction,omitempty"`
	Comment                *string  `json:"comment,omitempty"`
	Priority               *float64 `json:"priority,omitempty"`
	Bindpoint              *string  `json:"bindpoint,omitempty"`
	Gotopriorityexpression *string  `json:"gotopriorityexpression,omitempty"`
	Customer               *string  `json:"customer,omitempty"`
	Loadbalancer           *string  `json:"loadbalancer,omitempty"`
}

func NewRewritePolicyDataSource() datasource.DataSource {
	return &RewritePolicyDataSource{}
}

func (d *RewritePolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rewrite_policy"
}

func (d *RewritePolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	policyAttrs := map[string]schema.Attribute{
		"id":                     schema.StringAttribute{Computed: true},
		"name":                   schema.StringAttribute{Computed: true},
		"rule":                   schema.StringAttribute{Computed: true},
		"action":                 schema.StringAttribute{Computed: true},
		"undefaction":            schema.StringAttribute{Computed: true},
		"comment":                schema.StringAttribute{Computed: true},
		"priority":               schema.Float64Attribute{Computed: true},
		"bindpoint":              schema.StringAttribute{Computed: true},
		"gotopriorityexpression": schema.StringAttribute{Computed: true},
		"customer":               schema.StringAttribute{Computed: true},
		"loadbalancer":           schema.StringAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS rewrite policies. Set `name` or `id` to fetch a single rewrite policy, or omit both to list all.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact rewrite policy name to look up.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific rewrite policy to look up.",
			},
			"rule": schema.StringAttribute{
				Computed:    true,
				Description: "Rewrite rule expression.",
			},
			"action": schema.StringAttribute{
				Computed:    true,
				Description: "Associated rewrite action UUID.",
			},
			"undefaction": schema.StringAttribute{
				Computed:    true,
				Description: "Action taken when the rule result is undefined.",
			},
			"comment": schema.StringAttribute{
				Computed:    true,
				Description: "Rewrite policy comment.",
			},
			"priority": schema.Float64Attribute{
				Computed:    true,
				Description: "Rewrite policy priority.",
			},
			"bindpoint": schema.StringAttribute{
				Computed:    true,
				Description: "Rewrite bind point.",
			},
			"gotopriorityexpression": schema.StringAttribute{
				Computed:    true,
				Description: "Priority expression after policy evaluation.",
			},
			"customer": schema.StringAttribute{
				Computed:    true,
				Description: "Customer identifier.",
			},
			"loadbalancer": schema.StringAttribute{
				Computed:    true,
				Description: "Associated load balancer.",
			},
			"rewrite_policies": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All rewrite policies (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: policyAttrs},
			},
		},
	}
}

func (d *RewritePolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RewritePolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config RewritePolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var item rewritePolicyDSAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/rewritepolicy/%s/", config.Id.ValueString()), &item)
		if err != nil {
			resp.Diagnostics.AddError("Error reading rewrite policy", err.Error())
			return
		}
		setSingleRewritePolicy(&config, &item)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	path := "/api/loadbalancing/rewritepolicy/?page_size=1000"
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		path += "&name__icontains=" + url.QueryEscape(config.Name.ValueString())
	}

	var page struct {
		Results []rewritePolicyDSAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, path, &page); err != nil {
		resp.Diagnostics.AddError("Error reading rewrite policies", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *rewritePolicyDSAPIModel
		for i := range page.Results {
			if page.Results[i].Name == config.Name.ValueString() {
				match = &page.Results[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("Rewrite policy not found",
				fmt.Sprintf("No rewrite policy with exact name %q was found.", config.Name.ValueString()))
			return
		}
		setSingleRewritePolicy(&config, match)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := RewritePolicyDataSourceModel{
		Name:                   types.StringNull(),
		Id:                     types.StringNull(),
		Rule:                   types.StringNull(),
		Action:                 types.StringNull(),
		Undefaction:            types.StringNull(),
		Comment:                types.StringNull(),
		Priority:               types.Float64Null(),
		Bindpoint:              types.StringNull(),
		Gotopriorityexpression: types.StringNull(),
		Customer:               types.StringNull(),
		Loadbalancer:           types.StringNull(),
		RewritePolicies:        make([]RewritePolicyModel, 0, len(page.Results)),
	}
	for i := range page.Results {
		state.RewritePolicies = append(state.RewritePolicies, rewritePolicyItemToModel(&page.Results[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleRewritePolicy(state *RewritePolicyDataSourceModel, item *rewritePolicyDSAPIModel) {
	state.Id = types.StringValue(item.Id)
	state.Name = types.StringValue(item.Name)
	if item.Rule != nil {
		state.Rule = types.StringValue(*item.Rule)
	} else {
		state.Rule = types.StringNull()
	}
	if item.Action != nil {
		state.Action = types.StringValue(*item.Action)
	} else {
		state.Action = types.StringNull()
	}
	if item.Undefaction != nil {
		state.Undefaction = types.StringValue(*item.Undefaction)
	} else {
		state.Undefaction = types.StringNull()
	}
	if item.Comment != nil {
		state.Comment = types.StringValue(*item.Comment)
	} else {
		state.Comment = types.StringNull()
	}
	if item.Priority != nil {
		state.Priority = types.Float64Value(*item.Priority)
	} else {
		state.Priority = types.Float64Null()
	}
	if item.Bindpoint != nil {
		state.Bindpoint = types.StringValue(*item.Bindpoint)
	} else {
		state.Bindpoint = types.StringNull()
	}
	if item.Gotopriorityexpression != nil {
		state.Gotopriorityexpression = types.StringValue(*item.Gotopriorityexpression)
	} else {
		state.Gotopriorityexpression = types.StringNull()
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
	state.RewritePolicies = []RewritePolicyModel{}
}

func rewritePolicyItemToModel(item *rewritePolicyDSAPIModel) RewritePolicyModel {
	m := RewritePolicyModel{
		Id:   types.StringValue(item.Id),
		Name: types.StringValue(item.Name),
	}
	if item.Rule != nil {
		m.Rule = types.StringValue(*item.Rule)
	} else {
		m.Rule = types.StringNull()
	}
	if item.Action != nil {
		m.Action = types.StringValue(*item.Action)
	} else {
		m.Action = types.StringNull()
	}
	if item.Undefaction != nil {
		m.Undefaction = types.StringValue(*item.Undefaction)
	} else {
		m.Undefaction = types.StringNull()
	}
	if item.Comment != nil {
		m.Comment = types.StringValue(*item.Comment)
	} else {
		m.Comment = types.StringNull()
	}
	if item.Priority != nil {
		m.Priority = types.Float64Value(*item.Priority)
	} else {
		m.Priority = types.Float64Null()
	}
	if item.Bindpoint != nil {
		m.Bindpoint = types.StringValue(*item.Bindpoint)
	} else {
		m.Bindpoint = types.StringNull()
	}
	if item.Gotopriorityexpression != nil {
		m.Gotopriorityexpression = types.StringValue(*item.Gotopriorityexpression)
	} else {
		m.Gotopriorityexpression = types.StringNull()
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

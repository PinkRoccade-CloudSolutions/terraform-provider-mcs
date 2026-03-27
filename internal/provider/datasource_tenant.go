package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinkroccade/terraform-provider-mcs/internal/apiclient"
)

var _ datasource.DataSource = &TenantDataSource{}

// TenantDataSource reads the current tenant for the configured API token (GET /api/tenant/).
type TenantDataSource struct {
	client *apiclient.Client
}

// TenantDataSourceModel holds Terraform state for data.mcs_tenant.
type TenantDataSourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	RawJSON     types.String `tfsdk:"raw_json"`
}

func NewTenantDataSource() datasource.DataSource {
	return &TenantDataSource{}
}

func (d *TenantDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant"
}

func (d *TenantDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads tenant information for the authenticated token from GET /api/tenant/.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "Tenant primary key when the API returns a numeric id.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Tenant display name when present in the API response.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Tenant description when present in the API response.",
			},
			"raw_json": schema.StringAttribute{
				Computed:    true,
				Description: "Full JSON response body from /api/tenant/ for any fields not mapped above (use with jsondecode).",
			},
		},
	}
}

func (d *TenantDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TenantDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var raw json.RawMessage
	err := d.client.Get(ctx, "/api/tenant/", &raw)
	if err != nil {
		resp.Diagnostics.AddError("Error reading tenant", err.Error())
		return
	}

	state := TenantDataSourceModel{
		ID:          types.Int64Null(),
		Name:        types.StringNull(),
		Description: types.StringNull(),
		RawJSON:     types.StringValue(string(raw)),
	}

	if len(raw) == 0 {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var obj map[string]interface{}
	if err := dec.Decode(&obj); err != nil {
		resp.Diagnostics.AddError("Error decoding tenant JSON", err.Error())
		return
	}

	state.ID = int64FromTenantMap(obj, "id", "pk", "tenant")
	state.Name = stringFromTenantMap(obj, "name")
	state.Description = stringFromTenantMap(obj, "description")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func int64FromTenantMap(m map[string]interface{}, keys ...string) types.Int64 {
	for _, k := range keys {
		v, ok := m[k]
		if !ok || v == nil {
			continue
		}
		switch t := v.(type) {
		case json.Number:
			i, err := t.Int64()
			if err == nil {
				return types.Int64Value(i)
			}
		case float64:
			return types.Int64Value(int64(t))
		}
	}
	return types.Int64Null()
}

func stringFromTenantMap(m map[string]interface{}, keys ...string) types.String {
	for _, k := range keys {
		v, ok := m[k]
		if !ok || v == nil {
			continue
		}
		switch t := v.(type) {
		case string:
			return types.StringValue(t)
		default:
			return types.StringValue(fmt.Sprint(t))
		}
	}
	return types.StringNull()
}

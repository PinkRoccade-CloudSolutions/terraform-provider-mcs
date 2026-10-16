package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// listDataSource looks up objects in a REST collection. Setting `id` fetches
// one object, setting `name` finds one by exact name, and omitting both lists
// every object into the plural attribute. A is the API representation and I
// the Terraform item model, whose tfsdk tags must match itemAttrs.
type listDataSource[A any, I any] struct {
	client *apiclient.Client

	typeSuffix  string
	label       string
	description string
	collection  string
	pluralAttr  string
	nameFilter  string
	idIsInt     bool

	// itemAttrs describes every item field, including id and name. All
	// attributes must be Computed.
	itemAttrs map[string]schema.Attribute
	nameOf    func(*A) string
	toItem    func(*A) I
}

func (d *listDataSource[A, I]) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.typeSuffix
}

func (d *listDataSource[A, I]) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := make(map[string]schema.Attribute, len(d.itemAttrs)+3)
	for k, v := range d.itemAttrs {
		attrs[k] = v
	}
	if d.idIsInt {
		attrs["id"] = schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Description: fmt.Sprintf("ID of a specific %s to look up.", d.label),
		}
	} else {
		attrs["id"] = schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: fmt.Sprintf("ID of a specific %s to look up.", d.label),
		}
	}
	if d.nameOf != nil {
		attrs["name"] = schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: fmt.Sprintf("Exact %s name to look up.", d.label),
		}
	}
	attrs[d.pluralAttr] = schema.ListNestedAttribute{
		Computed:     true,
		Description:  fmt.Sprintf("All %s objects (populated when no lookup argument is set).", d.label),
		NestedObject: schema.NestedAttributeObject{Attributes: d.itemAttrs},
	}

	resp.Schema = schema.Schema{
		Description: d.description,
		Attributes:  attrs,
	}
}

func (d *listDataSource[A, I]) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *listDataSource[A, I]) configID(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) string {
	if d.idIsInt {
		var id types.Int64
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("id"), &id)...)
		if known(id) {
			return strconv.FormatInt(id.ValueInt64(), 10)
		}
		return ""
	}
	var id types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("id"), &id)...)
	if known(id) {
		return id.ValueString()
	}
	return ""
}

func (d *listDataSource[A, I]) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	id := d.configID(ctx, req, resp)
	var name types.String
	if d.nameOf != nil {
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("name"), &name)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	if id != "" {
		var item A
		if err := d.client.Get(ctx, d.collection+url.PathEscape(id)+"/", &item); err != nil {
			resp.Diagnostics.AddError("Error reading "+d.label, err.Error())
			return
		}
		d.setSingle(ctx, &item, resp)
		return
	}

	listPath := d.collection
	wantName := ""
	if known(name) && name.ValueString() != "" {
		wantName = name.ValueString()
		if d.nameFilter != "" {
			listPath += "?" + d.nameFilter + "=" + url.QueryEscape(wantName)
		}
	}

	raw, err := d.client.ListAll(ctx, listPath)
	if err != nil {
		resp.Diagnostics.AddError("Error listing "+d.label+" objects", err.Error())
		return
	}
	items := make([]A, 0, len(raw))
	for _, r := range raw {
		var item A
		if err := json.Unmarshal(r, &item); err != nil {
			resp.Diagnostics.AddError("Error parsing "+d.label, err.Error())
			return
		}
		items = append(items, item)
	}

	if wantName != "" {
		for i := range items {
			if d.nameOf(&items[i]) == wantName {
				d.setSingle(ctx, &items[i], resp)
				return
			}
		}
		resp.Diagnostics.AddError(d.label+" not found",
			fmt.Sprintf("No %s with exact name %q was found.", d.label, wantName))
		return
	}

	out := make([]I, 0, len(items))
	for i := range items {
		out = append(out, d.toItem(&items[i]))
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(d.pluralAttr), out)...)
}

func (d *listDataSource[A, I]) setSingle(ctx context.Context, item *A, resp *datasource.ReadResponse) {
	attrTypes := make(map[string]attr.Type, len(d.itemAttrs))
	for k, v := range d.itemAttrs {
		attrTypes[k] = v.GetType()
	}
	obj, diags := types.ObjectValueFrom(ctx, attrTypes, d.toItem(item))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	for k, v := range obj.Attributes() {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(k), v)...)
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(d.pluralAttr), []I{})...)
}

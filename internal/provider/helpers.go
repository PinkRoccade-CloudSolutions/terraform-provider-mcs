package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// listAll fetches every page of a paginated list endpoint and decodes each item into T.
func listAll[T any](ctx context.Context, c *apiclient.Client, path string) ([]T, error) {
	raw, err := c.ListAll(ctx, path)
	if err != nil {
		return nil, err
	}
	items := make([]T, 0, len(raw))
	for _, r := range raw {
		var item T
		if err := json.Unmarshal(r, &item); err != nil {
			return nil, fmt.Errorf("parsing list item: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

// listElems converts a configured list into an API value. A null or unknown list yields nil
// (omitted from the request body via omitempty); a known list, even an empty one, yields a
// non-nil pointer so that `[]` is sent and the API clears the field.
func listElems[T any](ctx context.Context, l types.List, diags *diag.Diagnostics) *[]T {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	out := make([]T, 0, len(l.Elements()))
	diags.Append(l.ElementsAs(ctx, &out, false)...)
	return &out
}

// listElemsForUpdate behaves like listElems, but when the attribute was removed from the
// configuration (plan null, state non-empty) it returns an empty slice so the PUT/PATCH clears it.
func listElemsForUpdate[T any](ctx context.Context, plan, state types.List, diags *diag.Diagnostics) *[]T {
	if plan.IsNull() && !state.IsNull() && !state.IsUnknown() && len(state.Elements()) > 0 {
		out := []T{}
		return &out
	}
	return listElems[T](ctx, plan, diags)
}

// listValue converts API values into a list for state. When the attribute was not configured
// (configured is null) and the API returned nothing, the list stays null to avoid a diff.
func listValue[T any](ctx context.Context, elemType attr.Type, configured types.List, values []T, diags *diag.Diagnostics) types.List {
	if configured.IsNull() && len(values) == 0 {
		return types.ListNull(elemType)
	}
	if values == nil {
		values = []T{}
	}
	v, d := types.ListValueFrom(ctx, elemType, values)
	diags.Append(d...)
	return v
}

// computedListValue converts API values into a list for an Optional+Computed or Computed-only
// attribute: the API value is always authoritative, an absent value becomes an empty list.
func computedListValue[T any](ctx context.Context, elemType attr.Type, values []T, diags *diag.Diagnostics) types.List {
	if values == nil {
		values = []T{}
	}
	v, d := types.ListValueFrom(ctx, elemType, values)
	diags.Append(d...)
	return v
}

// stringPtr returns a pointer to the value of a known, non-null string attribute, or nil.
func stringPtr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

// int64Ptr returns a pointer to the value of a known, non-null int64 attribute, or nil.
func int64Ptr(v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := v.ValueInt64()
	return &i
}

// boolPtr returns a pointer to the value of a known, non-null bool attribute, or nil.
func boolPtr(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	b := v.ValueBool()
	return &b
}

// parseImportID splits a composite import ID on "/" into exactly n non-empty parts. The last
// part keeps any further slashes, so it may hold free text such as TXT record content. On a
// malformed ID it adds an error naming the expected format and returns false.
func parseImportID(id string, n int, format string, diags *diag.Diagnostics) ([]string, bool) {
	parts := strings.SplitN(id, "/", n)
	if len(parts) != n {
		diags.AddError("Invalid import ID", fmt.Sprintf("Expected an import ID of the form %q, got %q.", format, id))
		return nil, false
	}
	for _, p := range parts {
		if p == "" {
			diags.AddError("Invalid import ID", fmt.Sprintf("Expected an import ID of the form %q, got %q.", format, id))
			return nil, false
		}
	}
	return parts, true
}

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func stringList(xs []string) types.List {
	elems := make([]attr.Value, 0, len(xs))
	for _, x := range xs {
		elems = append(elems, types.StringValue(x))
	}
	return types.ListValueMust(types.StringType, elems)
}

func stringMap(m map[string]string) types.Map {
	elems := make(map[string]attr.Value, len(m))
	for k, v := range m {
		elems[k] = types.StringValue(v)
	}
	return types.MapValueMust(types.StringType, elems)
}

func mapToStrings(m types.Map) map[string]string {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	out := make(map[string]string, len(m.Elements()))
	for k, e := range m.Elements() {
		if s, ok := e.(types.String); ok && !s.IsNull() && !s.IsUnknown() {
			out[k] = s.ValueString()
		}
	}
	return out
}

func known(v attr.Value) bool {
	return !v.IsNull() && !v.IsUnknown()
}

func strOrNull(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

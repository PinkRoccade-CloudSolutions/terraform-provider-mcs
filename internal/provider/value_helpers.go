package provider

import (
	"encoding/json"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// decodeStringList accepts fields the API spec leaves untyped: a JSON array of
// strings, a comma-separated string, or null.
func decodeStringList(raw json.RawMessage) []string {
	if len(raw) == 0 || string(raw) == "null" {
		return []string{}
	}
	var list []string
	if err := json.Unmarshal(raw, &list); err == nil {
		return list
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		out := []string{}
		for _, part := range strings.Split(s, ",") {
			if p := strings.TrimSpace(part); p != "" {
				out = append(out, p)
			}
		}
		return out
	}
	return []string{}
}

func stringList(xs []string) types.List {
	elems := make([]attr.Value, 0, len(xs))
	for _, x := range xs {
		elems = append(elems, types.StringValue(x))
	}
	return types.ListValueMust(types.StringType, elems)
}

// optionalStringList keeps an unset optional list null when the API returns no
// elements, so an omitted attribute does not show a perpetual diff.
func optionalStringList(prior types.List, values []string) types.List {
	if prior.IsNull() && len(values) == 0 {
		return types.ListNull(types.StringType)
	}
	return stringList(values)
}

// optionalString keeps an unset optional string null when the API echoes it
// back as an empty string or omits it.
func optionalString(prior types.String, value *string) types.String {
	if value == nil || (*value == "" && prior.IsNull()) {
		return types.StringNull()
	}
	return types.StringValue(*value)
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

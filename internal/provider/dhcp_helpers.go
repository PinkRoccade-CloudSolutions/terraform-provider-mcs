package provider

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// ipv4Validator accepts a dotted-quad IPv4 address.
type ipv4Validator struct{}

func (ipv4Validator) Description(context.Context) string { return "value must be an IPv4 address" }

func (v ipv4Validator) MarkdownDescription(ctx context.Context) string { return v.Description(ctx) }

func (ipv4Validator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	s := req.ConfigValue.ValueString()
	if ip := net.ParseIP(s); ip == nil || ip.To4() == nil || strings.Contains(s, ":") {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid IPv4 address",
			fmt.Sprintf("%q is not an IPv4 address.", s))
	}
}

// macValidator accepts a 48-bit MAC address in any notation net.ParseMAC understands
// (00:50:56:aa:bb:01, 00-50-56-AA-BB-01, 0050.56aa.bb01).
type macValidator struct{}

func (macValidator) Description(context.Context) string { return "value must be a 48-bit MAC address" }

func (v macValidator) MarkdownDescription(ctx context.Context) string { return v.Description(ctx) }

func (macValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if normalizeMAC(req.ConfigValue.ValueString()) == "" {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid MAC address",
			fmt.Sprintf("%q is not a 48-bit MAC address.", req.ConfigValue.ValueString()))
	}
}

// normalizeMAC returns the notation MCS stores (lower case, colon separated), or "" when s is
// not a 48-bit MAC address.
func normalizeMAC(s string) string {
	hw, err := net.ParseMAC(strings.TrimSpace(s))
	if err != nil || len(hw) != 6 {
		return ""
	}
	return hw.String()
}

// sameMAC reports whether a and b are the same MAC address written differently.
func sameMAC(a, b string) bool {
	na := normalizeMAC(a)
	return na != "" && na == normalizeMAC(b)
}

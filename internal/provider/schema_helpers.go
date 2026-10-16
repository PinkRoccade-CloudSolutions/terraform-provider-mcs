package provider

import (
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func rsID() schema.StringAttribute {
	return schema.StringAttribute{
		Computed:      true,
		Description:   "Identifier assigned by the API.",
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
}

func rsIntID() schema.Int64Attribute {
	return schema.Int64Attribute{
		Computed:      true,
		Description:   "Identifier assigned by the API.",
		PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
	}
}

func rsRequiredString(desc string) schema.StringAttribute {
	return schema.StringAttribute{Required: true, Description: desc}
}

func rsOptionalString(desc string) schema.StringAttribute {
	return schema.StringAttribute{Optional: true, Description: desc}
}

// rsOptComputedString is an optional string the API fills with a default
// when omitted.
func rsOptComputedString(desc string) schema.StringAttribute {
	return schema.StringAttribute{
		Optional:      true,
		Computed:      true,
		Description:   desc,
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
}

func rsOptComputedInt64(desc string) schema.Int64Attribute {
	return schema.Int64Attribute{
		Optional:      true,
		Computed:      true,
		Description:   desc,
		PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
	}
}

// rsStableString is a read-only string that does not change after creation.
func rsStableString(desc string) schema.StringAttribute {
	return schema.StringAttribute{
		Computed:      true,
		Description:   desc,
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
}

// rsVolatileString is a read-only string that may change on every update.
func rsVolatileString(desc string) schema.StringAttribute {
	return schema.StringAttribute{Computed: true, Description: desc}
}

func rsVolatileBool(desc string) schema.BoolAttribute {
	return schema.BoolAttribute{Computed: true, Description: desc}
}

func dsString(desc string) dschema.StringAttribute {
	return dschema.StringAttribute{Computed: true, Description: desc}
}

func dsBool(desc string) dschema.BoolAttribute {
	return dschema.BoolAttribute{Computed: true, Description: desc}
}

func dsInt64(desc string) dschema.Int64Attribute {
	return dschema.Int64Attribute{Computed: true, Description: desc}
}

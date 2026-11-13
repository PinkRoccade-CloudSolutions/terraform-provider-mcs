package provider

import (
	"fmt"
	"net/netip"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type LoadBalancerPoolResourceModel struct {
	Id                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	IpPool              types.String `tfsdk:"ip_pool"`
	IpPoolName          types.String `tfsdk:"ip_pool_name"`
	Prefix              types.String `tfsdk:"prefix"`
	PrefixLength        types.Int64  `tfsdk:"prefix_length"`
	Status              types.String `tfsdk:"status"`
	Customer            types.String `tfsdk:"customer"`
	AssignedObjectType  types.String `tfsdk:"assigned_object_type"`
	AssignedObjectId    types.String `tfsdk:"assigned_object_id"`
	AssignedObjectLabel types.String `tfsdk:"assigned_object_label"`
}

type loadBalancerPoolAPIModel struct {
	ID                  string  `json:"id"`
	Name                string  `json:"name"`
	Description         string  `json:"description"`
	IpPool              string  `json:"ip_pool"`
	IpPoolName          string  `json:"ip_pool_name"`
	Prefix              string  `json:"prefix"`
	Status              string  `json:"status"`
	Customer            string  `json:"customer"`
	AssignedObjectType  string  `json:"assigned_object_type"`
	AssignedObjectId    *string `json:"assigned_object_id"`
	AssignedObjectLabel string  `json:"assigned_object_label"`
}

// loadBalancerPoolCreateBody is the LoadBalancerPoolCreate schema.
type loadBalancerPoolCreateBody struct {
	IpPool       string  `json:"ip_pool"`
	Customer     string  `json:"customer"`
	Prefix       *string `json:"prefix,omitempty"`
	PrefixLength *int64  `json:"prefix_length,omitempty"`
	Name         string  `json:"name"`
	Description  *string `json:"description,omitempty"`
}

// loadBalancerPoolBody is the edit shape: only name and description can change.
type loadBalancerPoolBody struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// prefixLengthOf returns the length of a CIDR such as 203.0.113.16/28, or null.
func prefixLengthOf(prefix string) types.Int64 {
	if p, err := netip.ParsePrefix(prefix); err == nil {
		return types.Int64Value(int64(p.Bits()))
	}
	return types.Int64Null()
}

func NewLoadBalancerPoolResource() resource.Resource {
	return &crudResource[LoadBalancerPoolResourceModel, loadBalancerPoolAPIModel]{
		typeSuffix: "loadbalancer_pool",
		label:      "load balancer pool",
		collection: "/api/networking/loadbalancerpools/",
		schema: schema.Schema{
			Description: "Manages a load balancer pool: a subnet cut from an `lbpools` IP pool for a customer. Only `name` and `description` can be changed in place; a pool can only be deleted while it is free or assigned to a Kubernetes cluster.",
			Attributes: map[string]schema.Attribute{
				"id":          rsID(),
				"name":        rsRequiredString("Name of the pool."),
				"description": rsOptComputedString("Description of the pool."),
				"ip_pool": schema.StringAttribute{
					Required:      true,
					Description:   "UUID of the `lbpools` IP pool to allocate the subnet from. Changing it forces a new pool.",
					PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				},
				"customer": schema.StringAttribute{
					Required:      true,
					Description:   "Customer the pool belongs to. Changing it forces a new pool.",
					PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				},
				"prefix": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Subnet in CIDR notation, e.g. `203.0.113.16/28`. Set it to request a specific subnet; allocated by the API when omitted. Changing it forces a new pool.",
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
						stringplanmodifier.RequiresReplaceIfConfigured(),
					},
				},
				"prefix_length": schema.Int64Attribute{
					Optional:    true,
					Computed:    true,
					Description: "Length of the subnet to allocate, e.g. `28`. Chosen by the API when omitted; read from `prefix`. Changing it forces a new pool.",
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.UseStateForUnknown(),
						int64planmodifier.RequiresReplaceIfConfigured(),
					},
				},
				"ip_pool_name":          rsStableString("Name of the IP pool."),
				"status":                rsVolatileString("Pool status: `free` or `in_use`."),
				"assigned_object_type":  rsVolatileString("Type of the object the pool is assigned to."),
				"assigned_object_id":    rsVolatileString("ID of the object the pool is assigned to."),
				"assigned_object_label": rsVolatileString("Label of the object the pool is assigned to."),
			},
		},
		idOf: func(m *LoadBalancerPoolResourceModel) string { return m.Id.ValueString() },
		toAPI: func(m *LoadBalancerPoolResourceModel) any {
			return loadBalancerPoolCreateBody{
				IpPool:       m.IpPool.ValueString(),
				Customer:     m.Customer.ValueString(),
				Prefix:       stringPtr(m.Prefix),
				PrefixLength: int64Ptr(m.PrefixLength),
				Name:         m.Name.ValueString(),
				Description:  stringPtr(m.Description),
			}
		},
		updateAPI: func(m *LoadBalancerPoolResourceModel) any {
			return loadBalancerPoolBody{
				Name:        m.Name.ValueString(),
				Description: stringPtr(m.Description),
			}
		},
		fromAPI: func(a *loadBalancerPoolAPIModel, m *LoadBalancerPoolResourceModel) {
			m.Id = types.StringValue(a.ID)
			m.Name = types.StringValue(a.Name)
			m.Description = types.StringValue(a.Description)
			m.IpPool = types.StringValue(a.IpPool)
			m.IpPoolName = types.StringValue(a.IpPoolName)
			m.Prefix = types.StringValue(a.Prefix)
			m.PrefixLength = prefixLengthOf(a.Prefix)
			m.Status = types.StringValue(a.Status)
			m.Customer = types.StringValue(a.Customer)
			m.AssignedObjectType = types.StringValue(a.AssignedObjectType)
			m.AssignedObjectId = strOrNull(a.AssignedObjectId)
			m.AssignedObjectLabel = types.StringValue(a.AssignedObjectLabel)
		},
		deleteErr: func(err error) error {
			if apiclient.IsConflict(err) {
				return fmt.Errorf("the pool is still in use; unassign it from its load balancer before destroying it (pools of a Kubernetes cluster can be destroyed directly): %w", err)
			}
			return err
		},
	}
}

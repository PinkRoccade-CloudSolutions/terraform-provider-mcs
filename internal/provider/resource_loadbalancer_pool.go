package provider

import (
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

type loadBalancerPoolBody struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	IpPool      *string `json:"ip_pool,omitempty"`
}

func NewLoadBalancerPoolResource() resource.Resource {
	return &crudResource[LoadBalancerPoolResourceModel, loadBalancerPoolAPIModel]{
		typeSuffix: "loadbalancer_pool",
		label:      "load balancer pool",
		collection: "/api/networking/loadbalancerpools/",
		schema: schema.Schema{
			Description: "Manages a load balancer pool: a subnet cut from an `lbpools` IP pool. Only `name` and `description` can be changed in place; a pool can only be deleted while it is free.",
			Attributes: map[string]schema.Attribute{
				"id":          rsID(),
				"name":        rsRequiredString("Name of the pool."),
				"description": rsOptComputedString("Description of the pool."),
				"ip_pool": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "UUID of the `lbpools` IP pool to allocate the subnet from. Selected by the API when omitted.",
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
						stringplanmodifier.RequiresReplaceIfConfigured(),
					},
				},
				"ip_pool_name":          rsStableString("Name of the IP pool."),
				"prefix":                rsStableString("Allocated subnet in CIDR notation."),
				"status":                rsVolatileString("Pool status: `free` or `in_use`."),
				"customer":              rsStableString("Customer the pool belongs to."),
				"assigned_object_type":  rsVolatileString("Type of the object the pool is assigned to."),
				"assigned_object_id":    rsVolatileString("ID of the object the pool is assigned to."),
				"assigned_object_label": rsVolatileString("Label of the object the pool is assigned to."),
			},
		},
		idOf: func(m *LoadBalancerPoolResourceModel) string { return m.Id.ValueString() },
		toAPI: func(m *LoadBalancerPoolResourceModel) any {
			return loadBalancerPoolBody{
				Name:        m.Name.ValueString(),
				Description: stringPtr(m.Description),
				IpPool:      stringPtr(m.IpPool),
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
			m.Status = types.StringValue(a.Status)
			m.Customer = types.StringValue(a.Customer)
			m.AssignedObjectType = types.StringValue(a.AssignedObjectType)
			m.AssignedObjectId = strOrNull(a.AssignedObjectId)
			m.AssignedObjectLabel = types.StringValue(a.AssignedObjectLabel)
		},
		deleteErr: func(err error) error {
			if apiclient.IsConflict(err) {
				return fmt.Errorf("the pool is still in use; unassign it from its load balancer before destroying it: %w", err)
			}
			return err
		},
	}
}

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type loadBalancerPoolItem struct {
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

func NewLoadBalancerPoolDataSource() datasource.DataSource {
	return &listDataSource[loadBalancerPoolAPIModel, loadBalancerPoolItem]{
		typeSuffix:  "loadbalancer_pool",
		label:       "load balancer pool",
		description: "Looks up load balancer pools by ID or name, or lists all of them.",
		collection:  "/api/networking/loadbalancerpools/",
		pluralAttr:  "pools",
		nameFilter:  "name__icontains",
		itemAttrs: map[string]dschema.Attribute{
			"id":                    dsString("Pool ID."),
			"name":                  dsString("Name of the pool."),
			"description":           dsString("Description of the pool."),
			"ip_pool":               dsString("UUID of the IP pool."),
			"ip_pool_name":          dsString("Name of the IP pool."),
			"prefix":                dsString("Allocated subnet in CIDR notation."),
			"status":                dsString("Pool status: `free` or `in_use`."),
			"customer":              dsString("Customer the pool belongs to."),
			"assigned_object_type":  dsString("Type of the object the pool is assigned to."),
			"assigned_object_id":    dsString("ID of the object the pool is assigned to."),
			"assigned_object_label": dsString("Label of the object the pool is assigned to."),
		},
		nameOf: func(a *loadBalancerPoolAPIModel) string { return a.Name },
		toItem: func(a *loadBalancerPoolAPIModel) loadBalancerPoolItem {
			return loadBalancerPoolItem{
				Id:                  types.StringValue(a.ID),
				Name:                types.StringValue(a.Name),
				Description:         types.StringValue(a.Description),
				IpPool:              types.StringValue(a.IpPool),
				IpPoolName:          types.StringValue(a.IpPoolName),
				Prefix:              types.StringValue(a.Prefix),
				Status:              types.StringValue(a.Status),
				Customer:            types.StringValue(a.Customer),
				AssignedObjectType:  types.StringValue(a.AssignedObjectType),
				AssignedObjectId:    strOrNull(a.AssignedObjectId),
				AssignedObjectLabel: types.StringValue(a.AssignedObjectLabel),
			}
		},
	}
}

package provider

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &VirtualMachineDataSource{}

type VirtualMachineDataSource struct {
	client *apiclient.Client
}

type VirtualMachineDataSourceModel struct {
	Name            types.String              `tfsdk:"name"`
	Id              types.String              `tfsdk:"id"`
	DomainId        types.Int64               `tfsdk:"domain_id"`
	CPU             types.Int64               `tfsdk:"cpu"`
	Memory          types.Int64               `tfsdk:"memory"`
	OS              types.String              `tfsdk:"os"`
	State           types.String              `tfsdk:"state"`
	Disks           []VMDiskModel             `tfsdk:"disks"`
	Interfaces      []VMInterfaceModel        `tfsdk:"interfaces"`
	VirtualMachines []VirtualMachineListModel `tfsdk:"virtual_machines"`
}

type VirtualMachineListModel struct {
	Id         types.String       `tfsdk:"id"`
	Name       types.String       `tfsdk:"name"`
	CPU        types.Int64        `tfsdk:"cpu"`
	Memory     types.Int64        `tfsdk:"memory"`
	OS         types.String       `tfsdk:"os"`
	State      types.String       `tfsdk:"state"`
	Disks      []VMDiskModel      `tfsdk:"disks"`
	Interfaces []VMInterfaceModel `tfsdk:"interfaces"`
}

type VMDiskModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Size types.Int64  `tfsdk:"size"`
	Path types.String `tfsdk:"path"`
	Type types.String `tfsdk:"type"`
}

type VMInterfaceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	IPAddress   types.String `tfsdk:"ipaddress"`
	IPv6Address types.String `tfsdk:"ipv6address"`
	Network     types.String `tfsdk:"network"`
	MACAddress  types.String `tfsdk:"mac_address"`
	VMName      types.String `tfsdk:"vm_name"`
}

type vmAPIModel struct {
	Id         string            `json:"id"`
	Name       string            `json:"name"`
	CPU        int64             `json:"cpu"`
	Memory     int64             `json:"memory"`
	OS         string            `json:"os"`
	State      string            `json:"state"`
	Disks      []vmDiskAPIModel  `json:"disks"`
	Interfaces []vmIfaceAPIModel `json:"interfaces"`
}

type vmDiskAPIModel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
	Path string `json:"path"`
	Type string `json:"type"`
}

type vmIfaceAPIModel struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	IPAddress   string  `json:"ipaddress"`
	IPv6Address string  `json:"ipv6address"`
	Network     *string `json:"network"`
	MACAddress  string  `json:"macAddress"`
	VMName      string  `json:"vm_name"`
}

func NewVirtualMachineDataSource() datasource.DataSource {
	return &VirtualMachineDataSource{}
}

func (d *VirtualMachineDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtualmachine"
}

func (d *VirtualMachineDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	diskAttrs := map[string]schema.Attribute{
		"id":   schema.StringAttribute{Computed: true},
		"name": schema.StringAttribute{Computed: true},
		"size": schema.Int64Attribute{Computed: true, Description: "Size in GB."},
		"path": schema.StringAttribute{Computed: true},
		"type": schema.StringAttribute{Computed: true},
	}
	ifaceAttrs := map[string]schema.Attribute{
		"id":          schema.StringAttribute{Computed: true},
		"name":        schema.StringAttribute{Computed: true},
		"ipaddress":   schema.StringAttribute{Computed: true},
		"ipv6address": schema.StringAttribute{Computed: true},
		"network":     schema.StringAttribute{Computed: true},
		"mac_address": schema.StringAttribute{Computed: true},
		"vm_name":     schema.StringAttribute{Computed: true},
	}

	vmAttrs := map[string]schema.Attribute{
		"id":     schema.StringAttribute{Computed: true},
		"name":   schema.StringAttribute{Computed: true},
		"cpu":    schema.Int64Attribute{Computed: true},
		"memory": schema.Int64Attribute{Computed: true, Description: "Memory in MB."},
		"os":     schema.StringAttribute{Computed: true},
		"state":  schema.StringAttribute{Computed: true},
		"disks": schema.ListNestedAttribute{
			Computed:     true,
			NestedObject: schema.NestedAttributeObject{Attributes: diskAttrs},
		},
		"interfaces": schema.ListNestedAttribute{
			Computed:     true,
			NestedObject: schema.NestedAttributeObject{Attributes: ifaceAttrs},
		},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS virtual machines. Set `name` or `id` to fetch a single VM, or omit both to list all.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact VM name to look up.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific VM to look up, or the UUID of the matched VM when filtering by name.",
			},
			"domain_id": schema.Int64Attribute{
				Optional:    true,
				Description: "Only consider VMs with an interface in a network of this domain (interfaces__network__domain__id).",
			},
			"cpu": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of vCPUs (set when a single VM is matched).",
			},
			"memory": schema.Int64Attribute{
				Computed:    true,
				Description: "Memory in MB (set when a single VM is matched).",
			},
			"os": schema.StringAttribute{
				Computed:    true,
				Description: "Operating system (set when a single VM is matched).",
			},
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "Power state (set when a single VM is matched).",
			},
			"disks": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "Disks attached to the matched VM.",
				NestedObject: schema.NestedAttributeObject{Attributes: diskAttrs},
			},
			"interfaces": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "Network interfaces of the matched VM.",
				NestedObject: schema.NestedAttributeObject{Attributes: ifaceAttrs},
			},
			"virtual_machines": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All virtual machines (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: vmAttrs},
			},
		},
	}
}

func (d *VirtualMachineDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VirtualMachineDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config VirtualMachineDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var vm vmAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/virtualization/virtualmachine/%s/", config.Id.ValueString()), &vm)
		if err != nil {
			resp.Diagnostics.AddError("Error reading virtual machine", err.Error())
			return
		}
		setSingleVM(&config, &vm)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	query := url.Values{}
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		query.Set("name__icontains", config.Name.ValueString())
	}
	if !config.DomainId.IsNull() && !config.DomainId.IsUnknown() {
		query.Set("interfaces__network__domain__id", strconv.FormatInt(config.DomainId.ValueInt64(), 10))
	}
	path := "/api/virtualization/virtualmachine/"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	items, err := listAll[vmAPIModel](ctx, d.client, path)
	if err != nil {
		resp.Diagnostics.AddError("Error reading virtual machines", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *vmAPIModel
		for i := range items {
			if items[i].Name == config.Name.ValueString() {
				match = &items[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("Virtual machine not found",
				fmt.Sprintf("No virtual machine with exact name %q was found.", config.Name.ValueString()))
			return
		}
		setSingleVM(&config, match)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := VirtualMachineDataSourceModel{
		Name:            types.StringNull(),
		Id:              types.StringNull(),
		DomainId:        config.DomainId,
		CPU:             types.Int64Null(),
		Memory:          types.Int64Null(),
		OS:              types.StringNull(),
		State:           types.StringNull(),
		Disks:           []VMDiskModel{},
		Interfaces:      []VMInterfaceModel{},
		VirtualMachines: make([]VirtualMachineListModel, 0, len(items)),
	}
	for i := range items {
		state.VirtualMachines = append(state.VirtualMachines, toVMListModel(&items[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleVM(state *VirtualMachineDataSourceModel, vm *vmAPIModel) {
	state.Id = types.StringValue(vm.Id)
	state.Name = types.StringValue(vm.Name)
	state.CPU = types.Int64Value(vm.CPU)
	state.Memory = types.Int64Value(vm.Memory)
	state.OS = types.StringValue(vm.OS)
	state.State = types.StringValue(vm.State)
	state.Disks = mapDisks(vm.Disks)
	state.Interfaces = mapIfaces(vm.Interfaces)
	state.VirtualMachines = []VirtualMachineListModel{}
}

func toVMListModel(vm *vmAPIModel) VirtualMachineListModel {
	return VirtualMachineListModel{
		Id:         types.StringValue(vm.Id),
		Name:       types.StringValue(vm.Name),
		CPU:        types.Int64Value(vm.CPU),
		Memory:     types.Int64Value(vm.Memory),
		OS:         types.StringValue(vm.OS),
		State:      types.StringValue(vm.State),
		Disks:      mapDisks(vm.Disks),
		Interfaces: mapIfaces(vm.Interfaces),
	}
}

func mapDisks(disks []vmDiskAPIModel) []VMDiskModel {
	out := make([]VMDiskModel, 0, len(disks))
	for _, d := range disks {
		out = append(out, VMDiskModel{
			Id:   types.StringValue(d.Id),
			Name: types.StringValue(d.Name),
			Size: types.Int64Value(d.Size),
			Path: types.StringValue(d.Path),
			Type: types.StringValue(d.Type),
		})
	}
	return out
}

func mapIfaces(ifaces []vmIfaceAPIModel) []VMInterfaceModel {
	out := make([]VMInterfaceModel, 0, len(ifaces))
	for _, i := range ifaces {
		out = append(out, VMInterfaceModel{
			Id:          types.StringValue(i.Id),
			Name:        types.StringValue(i.Name),
			IPAddress:   types.StringValue(i.IPAddress),
			IPv6Address: types.StringValue(i.IPv6Address),
			Network:     types.StringPointerValue(i.Network),
			MACAddress:  types.StringValue(i.MACAddress),
			VMName:      types.StringValue(i.VMName),
		})
	}
	return out
}

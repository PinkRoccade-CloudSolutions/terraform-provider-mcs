package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &JobDataSource{}

type JobDataSource struct {
	client *apiclient.Client
}

type JobDataSourceModel struct {
	Id                      types.Int64  `tfsdk:"id"`
	JobName                 types.String `tfsdk:"jobname"`
	Status                  types.String `tfsdk:"status"`
	Result                  types.String `tfsdk:"result"`
	Timestamp               types.String `tfsdk:"timestamp"`
	EndTime                 types.String `tfsdk:"endtime"`
	Message                 types.String `tfsdk:"message"`
	DryRun                  types.Bool   `tfsdk:"dryrun"`
	ContinueOnFailure       types.Bool   `tfsdk:"continue_on_failure"`
	CanForceContinueTransit types.String `tfsdk:"can_force_continue_transit"`
	Parent                  types.Int64  `tfsdk:"parent"`
	SubJobs                 types.List   `tfsdk:"sub_jobs"`
	CreatedAtTimestamp      types.String `tfsdk:"created_at_timestamp"`
	UpdatedAtTimestamp      types.String `tfsdk:"updated_at_timestamp"`
	CreatedByUser           types.Int64  `tfsdk:"created_by_user"`
	UpdatedByUser           types.Int64  `tfsdk:"updated_by_user"`
}

// jobNestedString decodes the nested Status ({"status": "..."}) and Result ({"result": "..."})
// objects into their single string value; a bare string or number is accepted as well.
type jobNestedString struct {
	Value *string
}

func (j *jobNestedString) UnmarshalJSON(b []byte) error {
	trimmed := strings.TrimSpace(string(b))
	if trimmed == "null" {
		j.Value = nil
		return nil
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(b, &obj); err == nil {
		for _, k := range []string{"status", "result", "name"} {
			if v, ok := obj[k]; ok && v != nil {
				s := fmt.Sprint(v)
				j.Value = &s
				return nil
			}
		}
		j.Value = nil
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		j.Value = &s
		return nil
	}
	j.Value = &trimmed
	return nil
}

type jobAPIModel struct {
	Id                      int             `json:"id"`
	JobName                 string          `json:"jobname"`
	Status                  jobNestedString `json:"status"`
	Result                  jobNestedString `json:"result"`
	Timestamp               string          `json:"timestamp"`
	EndTime                 *string         `json:"endtime"`
	Message                 string          `json:"message"`
	DryRun                  bool            `json:"dryrun"`
	ContinueOnFailure       bool            `json:"continue_on_failure"`
	CanForceContinueTransit *string         `json:"can_force_continue_transit"`
	Parent                  *int64          `json:"parent"`
	SubJobs                 []int64         `json:"sub_jobs"`
	CreatedAtTimestamp      *string         `json:"created_at_timestamp"`
	UpdatedAtTimestamp      *string         `json:"updated_at_timestamp"`
	CreatedByUser           *int64          `json:"created_by_user"`
	UpdatedByUser           *int64          `json:"updated_by_user"`
}

func NewJobDataSource() datasource.DataSource {
	return &JobDataSource{}
}

func (d *JobDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job"
}

func (d *JobDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Required: true,
			},
			"jobname": schema.StringAttribute{
				Computed: true,
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Job status name.",
			},
			"result": schema.StringAttribute{
				Computed:    true,
				Description: "Job result name.",
			},
			"timestamp": schema.StringAttribute{
				Computed: true,
			},
			"endtime": schema.StringAttribute{
				Computed:    true,
				Description: "End time of the job; null while the job is still running.",
			},
			"message": schema.StringAttribute{
				Computed: true,
			},
			"dryrun": schema.BoolAttribute{
				Computed: true,
			},
			"continue_on_failure": schema.BoolAttribute{
				Computed: true,
			},
			"can_force_continue_transit": schema.StringAttribute{
				Computed: true,
			},
			"parent": schema.Int64Attribute{
				Computed:    true,
				Description: "ID of the parent job.",
			},
			"sub_jobs": schema.ListAttribute{
				Computed:    true,
				ElementType: types.Int64Type,
			},
			"created_at_timestamp": schema.StringAttribute{
				Computed: true,
			},
			"updated_at_timestamp": schema.StringAttribute{
				Computed: true,
			},
			"created_by_user": schema.Int64Attribute{
				Computed: true,
			},
			"updated_by_user": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (d *JobDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *JobDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state JobDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp jobAPIModel
	err := d.client.Get(ctx, fmt.Sprintf("/api/jobs/job/%d/", state.Id.ValueInt64()), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error reading job", err.Error())
		return
	}

	state.Id = types.Int64Value(int64(apiResp.Id))
	state.JobName = types.StringValue(apiResp.JobName)
	state.Status = types.StringPointerValue(apiResp.Status.Value)
	state.Result = types.StringPointerValue(apiResp.Result.Value)
	state.Timestamp = types.StringValue(apiResp.Timestamp)
	state.EndTime = types.StringPointerValue(apiResp.EndTime)
	state.Message = types.StringValue(apiResp.Message)
	state.DryRun = types.BoolValue(apiResp.DryRun)
	state.ContinueOnFailure = types.BoolValue(apiResp.ContinueOnFailure)
	state.CanForceContinueTransit = types.StringPointerValue(apiResp.CanForceContinueTransit)
	state.Parent = types.Int64PointerValue(apiResp.Parent)
	state.SubJobs = computedListValue(ctx, types.Int64Type, apiResp.SubJobs, &resp.Diagnostics)
	state.CreatedAtTimestamp = types.StringPointerValue(apiResp.CreatedAtTimestamp)
	state.UpdatedAtTimestamp = types.StringPointerValue(apiResp.UpdatedAtTimestamp)
	state.CreatedByUser = types.Int64PointerValue(apiResp.CreatedByUser)
	state.UpdatedByUser = types.Int64PointerValue(apiResp.UpdatedByUser)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

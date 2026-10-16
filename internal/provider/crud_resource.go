package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// crudResource implements a resource backed by a plain REST collection:
// POST to collection, GET/PUT|PATCH/DELETE on collection/{id}/.
// M is the Terraform model, A the API representation.
type crudResource[M any, A any] struct {
	client *apiclient.Client

	typeSuffix string
	label      string
	collection string
	schema     schema.Schema
	usePatch   bool

	idOf    func(*M) string
	toAPI   func(*M) any
	fromAPI func(*A, *M)

	// updateAPI optionally builds a different body for updates than for create.
	updateAPI func(*M) any
	// deletePath optionally overrides the path used for DELETE.
	deletePath func(*M) string
	// deleteErr optionally rewrites a delete error into a friendlier one.
	deleteErr func(error) error
}

func (r *crudResource[M, A]) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.typeSuffix
}

func (r *crudResource[M, A]) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.schema
}

func (r *crudResource[M, A]) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *apiclient.Client, got: %T", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *crudResource[M, A]) itemPath(m *M) string {
	return r.collection + r.idOf(m) + "/"
}

func (r *crudResource[M, A]) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan M
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result A
	if err := r.client.Post(ctx, r.collection, r.toAPI(&plan), &result); err != nil {
		resp.Diagnostics.AddError("Error creating "+r.label, err.Error())
		return
	}

	r.fromAPI(&result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *crudResource[M, A]) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state M
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result A
	if err := r.client.Get(ctx, r.itemPath(&state), &result); err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading "+r.label, err.Error())
		return
	}

	r.fromAPI(&result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *crudResource[M, A]) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state M
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := r.toAPI
	if r.updateAPI != nil {
		body = r.updateAPI
	}

	var result A
	var err error
	if r.usePatch {
		err = r.client.Patch(ctx, r.itemPath(&state), body(&plan), &result)
	} else {
		err = r.client.Put(ctx, r.itemPath(&state), body(&plan), &result)
	}
	if err != nil {
		resp.Diagnostics.AddError("Error updating "+r.label, err.Error())
		return
	}

	r.fromAPI(&result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *crudResource[M, A]) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state M
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := r.itemPath(&state)
	if r.deletePath != nil {
		p = r.deletePath(&state)
	}
	err := r.client.Delete(ctx, p)
	if err == nil || apiclient.IsNotFound(err) {
		return
	}
	if r.deleteErr != nil {
		err = r.deleteErr(err)
	}
	resp.Diagnostics.AddError("Error deleting "+r.label, err.Error())
}

func (r *crudResource[M, A]) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if _, isInt := r.schema.Attributes["id"].(schema.Int64Attribute); !isInt {
		resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
		return
	}
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("Expected the numeric ID of the %s, got %q.", r.label, req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

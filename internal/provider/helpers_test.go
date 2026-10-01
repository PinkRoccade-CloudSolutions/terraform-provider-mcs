package provider

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
)

func TestListAll_FollowsPages(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	mock.On("/api/things/", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("page") == "2" {
			_, _ = fmt.Fprint(w, `{"count":2,"next":null,"results":[{"id":"b"}]}`)
			return
		}
		_, _ = fmt.Fprintf(w, `{"count":2,"next":"%s/api/things/?page=2","results":[{"id":"a"}]}`, mock.URL())
	})

	c := apiclient.NewClient(mock.URL(), "t", true)
	items, err := listAll[struct {
		Id string `json:"id"`
	}](context.Background(), c, "/api/things/")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].Id != "a" || items[1].Id != "b" {
		t.Fatalf("unexpected items: %+v", items)
	}
}

func TestListElems(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	if got := listElems[string](ctx, types.ListNull(types.StringType), &diags); got != nil {
		t.Errorf("null list: want nil, got %v", *got)
	}
	empty, _ := types.ListValueFrom(ctx, types.StringType, []string{})
	if got := listElems[string](ctx, empty, &diags); got == nil || len(*got) != 0 {
		t.Errorf("empty list: want non-nil empty slice, got %v", got)
	}
	full, _ := types.ListValueFrom(ctx, types.StringType, []string{"x"})
	if got := listElemsForUpdate[string](ctx, types.ListNull(types.StringType), full, &diags); got == nil || len(*got) != 0 {
		t.Errorf("removed list: want non-nil empty slice, got %v", got)
	}
	if diags.HasError() {
		t.Fatal(diags)
	}
}

package provider

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func TestParseImportID(t *testing.T) {
	cases := []struct {
		id   string
		n    int
		want []string
	}{
		{"dom/obj", 2, []string{"dom", "obj"}},
		{"uuid/_dmarc/TXT/v=DMARC1; rua=mailto:a/b", 4, []string{"uuid", "_dmarc", "TXT", "v=DMARC1; rua=mailto:a/b"}},
		{"dom", 2, nil},
		{"dom/", 2, nil},
		{"/obj", 2, nil},
		{"uuid/www/A", 4, nil},
	}
	for _, c := range cases {
		var diags diag.Diagnostics
		got, ok := parseImportID(c.id, c.n, "<format>", &diags)
		if c.want == nil {
			if ok || !diags.HasError() {
				t.Errorf("%q: expected an error, got %v", c.id, got)
			}
			continue
		}
		if !ok || diags.HasError() || fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("%q: got %q (diags %v), want %q", c.id, got, diags, c.want)
		}
	}
}

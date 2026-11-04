package provider

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// The Terraform Registry only renders per-type pages from docs/resources/<name>.md and
// docs/data-sources/<name>.md. These tests keep those pages, the examples/ tree and the
// provider schema in sync.

const repoRoot = "../.."

var hclBlockRe = regexp.MustCompile("(?s)```hcl\n(.*?)```")

type docTarget struct {
	typeName   string   // e.g. mcs_network
	page       string   // docs/resources/network.md
	example    string   // examples/resources/mcs_network/resource.tf
	importSh   string   // examples/resources/mcs_network/import.sh, empty when not importable
	attributes []string // top-level schema attributes and blocks
}

func docTargets(t *testing.T) []docTarget {
	t.Helper()
	ctx := context.Background()
	p := New("test")()
	meta := &provider.MetadataResponse{}
	p.Metadata(ctx, provider.MetadataRequest{}, meta)

	var targets []docTarget
	for _, f := range p.Resources(ctx) {
		r := f()
		mr := &resource.MetadataResponse{}
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: meta.TypeName}, mr)
		sr := &resource.SchemaResponse{}
		r.Schema(ctx, resource.SchemaRequest{}, sr)
		tgt := docTarget{
			typeName: mr.TypeName,
			page:     filepath.Join("docs", "resources", strings.TrimPrefix(mr.TypeName, meta.TypeName+"_")+".md"),
			example:  filepath.Join("examples", "resources", mr.TypeName, "resource.tf"),
		}
		if _, ok := r.(resource.ResourceWithImportState); ok {
			tgt.importSh = filepath.Join("examples", "resources", mr.TypeName, "import.sh")
		}
		for name := range sr.Schema.Attributes {
			tgt.attributes = append(tgt.attributes, name)
		}
		for name := range sr.Schema.Blocks {
			tgt.attributes = append(tgt.attributes, name)
		}
		targets = append(targets, tgt)
	}
	for _, f := range p.DataSources(ctx) {
		d := f()
		mr := &datasource.MetadataResponse{}
		d.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: meta.TypeName}, mr)
		sr := &datasource.SchemaResponse{}
		d.Schema(ctx, datasource.SchemaRequest{}, sr)
		tgt := docTarget{
			typeName: mr.TypeName,
			page:     filepath.Join("docs", "data-sources", strings.TrimPrefix(mr.TypeName, meta.TypeName+"_")+".md"),
			example:  filepath.Join("examples", "data-sources", mr.TypeName, "data-source.tf"),
		}
		for name := range sr.Schema.Attributes {
			tgt.attributes = append(tgt.attributes, name)
		}
		for name := range sr.Schema.Blocks {
			tgt.attributes = append(tgt.attributes, name)
		}
		targets = append(targets, tgt)
	}
	return targets
}

func readRepoFile(t *testing.T, rel string) (string, bool) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot, rel))
	if err != nil {
		return "", false
	}
	return string(b), true
}

func TestDocs_EveryTypeHasPageAndExample(t *testing.T) {
	for _, tgt := range docTargets(t) {
		t.Run(tgt.typeName, func(t *testing.T) {
			page, ok := readRepoFile(t, tgt.page)
			if !ok {
				t.Fatalf("missing documentation page %s", tgt.page)
			}
			if !strings.HasPrefix(page, "---\n") || !strings.Contains(page, "\nsubcategory: ") {
				t.Errorf("%s: missing Registry front matter with a subcategory", tgt.page)
			}

			example, ok := readRepoFile(t, tgt.example)
			if !ok {
				t.Fatalf("missing example %s", tgt.example)
			}
			found := false
			for _, m := range hclBlockRe.FindAllStringSubmatch(page, -1) {
				if m[1] == example {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s: no ```hcl block matches %s exactly; update one from the other", tgt.page, tgt.example)
			}

			if tgt.importSh != "" {
				imp, ok := readRepoFile(t, tgt.importSh)
				if !ok {
					t.Errorf("missing import example %s", tgt.importSh)
				} else if !strings.Contains(page, "## Import") || !strings.Contains(page, imp) {
					t.Errorf("%s: Import section missing or does not match %s", tgt.page, tgt.importSh)
				}
			}

			var missing []string
			for _, a := range tgt.attributes {
				if !strings.Contains(page, "`"+a+"`") {
					missing = append(missing, a)
				}
			}
			sort.Strings(missing)
			if len(missing) > 0 {
				t.Errorf("%s does not document schema attributes: %s", tgt.page, strings.Join(missing, ", "))
			}
		})
	}
}

func TestDocs_NoOrphanPages(t *testing.T) {
	known := map[string]bool{}
	for _, tgt := range docTargets(t) {
		known[filepath.ToSlash(tgt.page)] = true
	}
	for _, dir := range []string{"docs/resources", "docs/data-sources"} {
		entries, err := os.ReadDir(filepath.Join(repoRoot, dir))
		if err != nil {
			t.Fatalf("reading %s: %v", dir, err)
		}
		for _, e := range entries {
			rel := dir + "/" + e.Name()
			if !known[rel] {
				t.Errorf("%s does not belong to any resource or data source of the provider", rel)
			}
		}
	}
}

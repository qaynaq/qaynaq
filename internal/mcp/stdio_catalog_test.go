package mcp

import "testing"

func TestCatalog_KnownEntries(t *testing.T) {
	required := []string{"filesystem", "git", "github", "slack", "postgres"}
	for _, id := range required {
		if _, ok := LookupCatalogEntry(id); !ok {
			t.Errorf("expected catalog entry %q to exist", id)
		}
	}
}

func TestCatalog_UnknownEntry(t *testing.T) {
	if _, ok := LookupCatalogEntry("definitely-not-a-real-mcp"); ok {
		t.Fatal("expected unknown lookup to fail")
	}
}

func TestCatalog_AllEntriesUseNpx(t *testing.T) {
	for _, e := range ListCatalogEntries() {
		if e.Command != "npx" {
			t.Errorf("entry %q uses non-npx command %q", e.ID, e.Command)
		}
	}
}

func TestCatalog_ListSorted(t *testing.T) {
	prev := ""
	for _, e := range ListCatalogEntries() {
		if prev != "" && e.ID < prev {
			t.Errorf("catalog list not sorted: %q before %q", prev, e.ID)
		}
		prev = e.ID
	}
}

func TestCatalog_RequiredEnvSpecsResolved(t *testing.T) {
	gh, _ := LookupCatalogEntry("github")
	hasRequired := false
	for _, s := range gh.EnvSpec {
		if s.Required {
			hasRequired = true
			break
		}
	}
	if !hasRequired {
		t.Error("github catalog entry has no required env spec")
	}
}

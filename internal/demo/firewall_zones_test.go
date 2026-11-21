package demo

import (
	"context"
	"net/url"
	"testing"
)

func TestFirewallZonePoliciesFilterActualRecords(t *testing.T) {
	page, err := (Fixtures{}).Page(context.Background(), "/firewall", "24h", "warning")
	if err != nil {
		t.Fatal(err)
	}
	if len(page.ZoneNames) != 6 || len(page.ZonePolicies) != 25 {
		t.Fatalf("unexpected matrix dimensions: %d zones, %d policies", len(page.ZoneNames), len(page.ZonePolicies))
	}
	for _, policy := range page.ZonePolicies {
		id := policy.From + "--" + policy.To
		view, err := pageView(page, shell("/firewall", page.Title, "warning", "24h", ""), url.Values{"zone": {id}})
		if err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		if len(view.ZoneRows) != 5 || len(view.Records) != len(policy.RuleIDs) {
			t.Fatalf("%s: got %d matrix rows and %d matching records for %d IDs", id, len(view.ZoneRows), len(view.Records), len(policy.RuleIDs))
		}
	}
	if _, err := pageView(page, shell("/firewall", page.Title, "warning", "24h", ""), url.Values{"zone": {"invalid"}}); err != ErrNotFound {
		t.Fatalf("unknown zone should fail: %v", err)
	}
	missing, err := (Fixtures{}).Page(context.Background(), "/firewall", "24h", "unavailable")
	if err != nil || len(missing.ZonePolicies) != 0 {
		t.Fatalf("unavailable source contains zone policies: %v", err)
	}
}

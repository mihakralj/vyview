package demo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestEvidenceLinksAndUnavailableSnapshots(t *testing.T) {
	handler, err := Handler()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/clients", "/firewall", "/conntrack", "/routing", "/vpn/wireguard", "/qos", "/dataplane", "/system", "/config", "/logs"} {
		t.Run(path, func(t *testing.T) {
			page, err := (Fixtures{}).Page(context.Background(), path, "7d", "normal")
			if err != nil {
				t.Fatal(err)
			}
			missing, err := (Fixtures{}).Page(context.Background(), path, "7d", "unavailable")
			if err != nil {
				t.Fatal(err)
			}
			if len(missing.Summaries) != 0 || len(missing.Evidence) != 0 || len(missing.Series) != 0 {
				t.Fatal("unavailable source contains illustrative data")
			}
			for tab, panels := range page.Evidence {
				view, err := pageView(page, shell(path, page.Title, "normal", "7d", "light"), url.Values{"tab": {tab}})
				if err != nil {
					t.Fatal(err)
				}
				for i, panel := range panels {
					for j, row := range panel.Rows {
						if len(row.Values) != len(panel.Columns) {
							t.Fatalf("%s: row %s has %d values for %d columns", tab, row.Label, len(row.Values), len(panel.Columns))
						}
						href := view.EvidencePanels[i].Rows[j].Href
						if href == "" {
							continue
						}
						target, err := url.Parse(href)
						if err != nil {
							t.Fatal(err)
						}
						if target.Query().Get("scenario") != "normal" || target.Query().Get("range") != "7d" || target.Query().Get("scoutTheme") != "light" {
							t.Fatalf("preferences lost in %s", href)
						}
						response := httptest.NewRecorder()
						handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, href, nil))
						if response.Code != http.StatusOK || (target.Query().Has("sel") && !strings.Contains(response.Body.String(), `id="details"`)) {
							t.Fatalf("evidence %s has broken target: %d", href, response.Code)
						}
					}
				}
			}
		})
	}
}

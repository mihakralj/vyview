package demo

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testProvider struct{ Fixtures }

func (testProvider) Page(_ context.Context, path, rangeID, scenario string) (Page, error) {
	if path != "/system" {
		return Page{}, ErrNotFound
	}
	return Page{Title: "Injected system", Sections: []Section{
		{ID: "resources", Title: "Resources", Columns: []string{"Reading"},
			Records: []Record{row("reading", "Sample", "OK", "Injected data", "", rangeID+" / "+scenario)}},
	}}, nil
}

func TestInjectedProviderAndRouteCoverage(t *testing.T) {
	handler, err := HandlerWithProvider(testProvider{})
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", "/system?range=7d&scenario=normal", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "7d / normal") {
		t.Fatalf("provider parameters or snapshot not reflected in response: %d", rec.Code)
	}
	handler, err = Handler()
	if err != nil {
		t.Fatal(err)
	}
	for _, destination := range append(append([]Destination{}, destinations...),
		Destination{Path: "/interfaces/eth0", Title: "eth0"}, Destination{Path: "/search", Title: "Search & page index"}) {
		t.Run(destination.Path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest("GET", destination.Path, nil))
			if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "<h1>"+template.HTMLEscapeString(destination.Title)+"</h1>") {
				t.Fatalf("unexpected page response: %d", rec.Code)
			}
		})
	}
}

func TestFixtureRangesAndScenarios(t *testing.T) {
	first, err := Fixture("1h", "warning")
	if err != nil {
		t.Fatal(err)
	}
	last, err := Fixture("7d", "normal")
	if err != nil {
		t.Fatal(err)
	}
	if first.RXPath == last.RXPath || first.PeakRX != 630 || last.PeakRX != 720 {
		t.Fatalf("ranges do not contain their expected distinct samples: %q, %q", first.RXPath, last.RXPath)
	}
	if first.Alert == "" || last.Alert != "" {
		t.Fatal("warning and normal scenarios have incorrect alerts")
	}
	missing, err := Fixture("24h", "unavailable")
	if err != nil {
		t.Fatal(err)
	}
	if len(missing.Points) != 0 || missing.WANAddress != "Unavailable" || missing.KPIs[0].Status != "unknown" {
		t.Fatalf("unavailable scenario should not invent values: %+v", missing)
	}
	for _, args := range [][2]string{{"invalid", "normal"}, {"1h", "invalid"}} {
		if _, err := Fixture(args[0], args[1]); err == nil {
			t.Fatalf("expected invalid fixture input to fail: %v", args)
		}
	}
}

func TestHTTPDemo(t *testing.T) {
	handler, err := Handler()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, method, path string
		status             int
		contains           string
	}{
		{"default overview", "GET", "/", http.StatusOK, "Demo data"},
		{"selected range", "GET", "/?range=1h&scenario=normal", http.StatusOK, "Peak receive 630"},
		{"unavailable", "GET", "/?scenario=unavailable", http.StatusOK, "Throughput data unavailable"},
		{"bad range", "GET", "/?range=unknown", http.StatusBadRequest, "unknown range"},
		{"bad scenario", "GET", "/?scenario=unknown", http.StatusBadRequest, "unknown demo scenario"},
		{"missing heatmap sample", "GET", "/system?sel=cpu&sample=core0-99", http.StatusNotFound, ""},
		{"valid heatmap sample", "GET", "/system?sel=cpu&sample=core1-2", http.StatusOK, "Core 1, 20h ago: 81%"},
		{"method rejected", "POST", "/", http.StatusMethodNotAllowed, ""},
		{"missing page", "GET", "/missing", http.StatusNotFound, ""},
		{"asset", "GET", "/assets/app.css", http.StatusOK, "--bg"},
		{"font asset", "GET", "/assets/fonts/IBMPlexSans-Regular.woff2", http.StatusOK, "wOF2"},
		{"head", "HEAD", "/", http.StatusOK, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))
			if rec.Code != tt.status || !strings.Contains(rec.Body.String(), tt.contains) {
				t.Fatalf("got status %d and body %q, want %d containing %q", rec.Code, rec.Body.String(), tt.status, tt.contains)
			}
			if rec.Header().Get("Content-Security-Policy") == "" {
				t.Error("missing content security policy")
			}
		})
	}
}

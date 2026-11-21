package demo

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"errors"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"strings"
)

//go:embed templates/*.html assets/*
var files embed.FS

func Handler() (http.Handler, error) {
	return HandlerWithProvider(Fixtures{})
}

func HandlerWithProvider(provider Provider) (http.Handler, error) {
	tmpl, err := template.ParseFS(files, "templates/*.html")
	if err != nil {
		return nil, err
	}
	if err := parseProTemplates(tmpl); err != nil {
		return nil, err
	}
	assets, err := fs.Sub(files, "assets")
	if err != nil {
		return nil, err
	}
	var themeScript bytes.Buffer
	if err := tmpl.ExecuteTemplate(&themeScript, "theme-script", nil); err != nil {
		return nil, err
	}
	hash := sha256.Sum256(themeScript.Bytes())
	csp := "default-src 'none'; script-src 'self' 'sha256-" + base64.StdEncoding.EncodeToString(hash[:]) +
		"'; style-src 'self'; font-src 'self'; img-src 'self'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'"
	render := func(w http.ResponseWriter, name string, data any) {
		var body bytes.Buffer
		if err := tmpl.ExecuteTemplate(&body, name, data); err != nil {
			log.Printf("render %s: %v", name, err)
			http.Error(w, "Unable to render demo view", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = body.WriteTo(w)
	}
	providerError := func(w http.ResponseWriter, r *http.Request, err error) {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		log.Printf("read snapshot for %s: %v", r.URL.Path, err)
		http.Error(w, "Unable to read snapshot", http.StatusServiceUnavailable)
	}
	mux := http.NewServeMux()
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		rangeID := q.Get("range")
		if rangeID == "" {
			rangeID = "24h"
		}
		if rangeID != "1h" && rangeID != "6h" && rangeID != "24h" && rangeID != "7d" {
			http.Error(w, "unknown range", http.StatusBadRequest)
			return
		}
		scenario := q.Get("scenario")
		if scenario == "" {
			scenario = "warning"
		}
		if scenario != "normal" && scenario != "warning" && scenario != "unavailable" {
			http.Error(w, "unknown demo scenario", http.StatusBadRequest)
			return
		}
		theme := q.Get("scoutTheme")
		if theme != "" && theme != "dark" && theme != "light" && theme != "system" {
			http.Error(w, "unknown theme", http.StatusBadRequest)
			return
		}
		if r.URL.Path == "/" {
			data, err := provider.Overview(r.Context(), rangeID, scenario)
			if err != nil {
				providerError(w, r, err)
				return
			}
			view := OverviewView{Overview: data, Shell: shell("/", "Overview", scenario, rangeID, theme), EvidenceLinks: make(map[string]string)}
			view.Ranges = append([]RangeOption(nil), data.Ranges...)
			for i := range data.Ranges {
				view.Ranges[i].Href = viewURL("/", scenario, data.Ranges[i].ID, "scoutTheme", theme)
			}
			for _, target := range []string{"/interfaces", "/health", "/system", "/clients", "/logs", "/interfaces/eth0"} {
				view.EvidenceLinks[target] = viewURL(target, scenario, rangeID, "scoutTheme", theme)
			}
			for _, link := range data.Links {
				view.EvidenceLinks[link.Name] = viewURL("/interfaces/"+url.PathEscape(link.Name), scenario, rangeID, "scoutTheme", theme)
			}
			render(w, "overview.html", view)
			return
		}
		if r.URL.Path == "/search" {
			query := strings.TrimSpace(q.Get("q"))
			search := SearchView{Shell: shell("/search", "Search & page index", scenario, rangeID, theme), Query: query}
			for _, destination := range destinations {
				if matches(query, destination.Title, destination.Path) {
					search.Results = append(search.Results, SearchHit{destination.Title, "", viewURL(destination.Path, scenario, rangeID, "scoutTheme", theme)})
				}
				if destination.Path == "/" || query == "" {
					continue
				}
				page, err := provider.Page(r.Context(), destination.Path, rangeID, scenario)
				if errors.Is(err, ErrNotFound) {
					continue
				}
				if err != nil {
					providerError(w, r, err)
					return
				}
				for _, section := range page.Sections {
					for _, record := range section.Records {
						if record.matches(query) {
							search.Results = append(search.Results, SearchHit{record.Name, page.Title + " / " + section.Title,
								viewURL(destination.Path, scenario, rangeID, "tab", section.ID, "sel", record.ID, "scoutTheme", theme) + "#details"})
						}
					}
				}
			}
			render(w, "search.html", search)
			return
		}
		page, err := provider.Page(r.Context(), r.URL.Path, rangeID, scenario)
		if err != nil {
			providerError(w, r, err)
			return
		}
		page.Path = r.URL.Path
		if sort := q.Get("sort"); sort != "" && sort != "name" && sort != "status" {
			http.Error(w, "unknown sort", http.StatusBadRequest)
			return
		}
		data, err := pageView(page, shell(page.Path, page.Title, scenario, rangeID, theme), q)
		if err != nil {
			providerError(w, r, err)
			return
		}
		render(w, "page.html", data)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", csp)
		w.Header().Set("Referrer-Policy", "no-referrer")
		mux.ServeHTTP(w, r)
	}), nil
}

type SearchHit struct{ Title, Context, Href string }
type SearchView struct {
	Shell
	Query   string
	Results []SearchHit
}

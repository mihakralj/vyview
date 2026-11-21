package demo

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"slices"
	"strings"
)

var ErrNotFound = errors.New("demo view not found")

// Provider supplies read-only snapshots. Filtering, linking and rendering are HTTP concerns.
type Provider interface {
	Overview(context.Context, string, string) (Overview, error)
	Page(context.Context, string, string, string) (Page, error)
}

type Fixtures struct{}

func (Fixtures) Overview(_ context.Context, rangeID, scenario string) (Overview, error) {
	return Fixture(rangeID, scenario)
}

type Page struct {
	Path, Title, Description, Source string
	Sections                         []Section
	Series                           map[string]*Series
	Meters                           map[string][]Meter
	Heatmaps                         map[string]*Heatmap
	InterfaceDetail                  *InterfaceDetail
	Summaries                        map[string][]KPI
	Evidence                         map[string][]EvidencePanel
	ZoneNames                        []string
	ZonePolicies                     []ZonePolicy
}

type ZonePolicy struct {
	From, To, Action string
	RuleIDs          []string
}

type EvidencePanel struct {
	Title, Source string
	Columns       []string
	Rows          []EvidenceRow
}

type EvidenceRow struct {
	Label, Href string
	Values      []string
}

type EvidencePanelView struct {
	EvidencePanel
	Rows []EvidenceRow
}

type Meter struct {
	ID, Label, Reading string
	Value, Capacity    int
}

type MeterView struct {
	Meter
	Href string
}

type Section struct {
	ID, Title string
	Columns   []string
	Records   []Record
}

type Series struct {
	Title, Unit string
	Max         int
	Points      []int
	Bars        bool
	TX          []int
	Annotation  string
}

type SeriesView struct {
	Series
	Path, Range string
	Options     []NavItem
	Columns     []ChartBar
	TXPath      string
	MarkerX     float64
}

type ChartBar struct{ X, Y, Width, Height float64 }

type Record struct {
	ID, Name, Status, Detail, Related string
	Values                            []string
	Category                          string
	Spark                             []int
	Color                             int
	Errors, Drops                     int
}

type Destination struct{ Path, Title, Icon string }

type NavItem struct {
	Title, Href, Icon string
	Current           bool
}

type CategoryItem struct {
	NavItem
	Count int
}

type Shell struct {
	Scenario, RangeID, Theme, Heading string
	Nav, Mobile                       []NavItem
	SearchURL, Home                   string
}

func shell(path, heading, scenario, rangeID, theme string) Shell {
	s := Shell{Scenario: scenario, RangeID: rangeID, Theme: theme, Heading: heading,
		SearchURL: viewURL("/search", scenario, rangeID, "scoutTheme", theme),
		Home:      viewURL("/", scenario, rangeID, "scoutTheme", theme)}
	for _, dest := range destinations {
		item := NavItem{dest.Title, viewURL(dest.Path, scenario, rangeID, "scoutTheme", theme), dest.Icon,
			path == dest.Path || (dest.Path != "/" && strings.HasPrefix(path, dest.Path+"/"))}
		s.Nav = append(s.Nav, item)
		switch dest.Path {
		case "/", "/interfaces", "/clients", "/health":
			s.Mobile = append(s.Mobile, item)
		}
	}
	s.Mobile = append(s.Mobile, NavItem{"More", s.SearchURL, "fa-list-timeline", path == "/search"})
	return s
}

var destinations = []Destination{
	{"/", "Overview", "fa-gauge-max"}, {"/health", "Health & Alerts", "fa-wave-pulse"}, {"/interfaces", "Interfaces", "fa-ethernet"},
	{"/clients", "Traffic & Clients", "fa-chart-network"}, {"/routing", "Routing", "fa-route"}, {"/firewall", "Firewall", "fa-block-brick-fire"},
	{"/conntrack", "Conntrack & NAT", "fa-arrows-left-right"}, {"/vpn/wireguard", "WireGuard", "fa-lock-keyhole"}, {"/qos", "QoS", "fa-sliders"},
	{"/topology", "Topology", "fa-diagram-project"}, {"/services", "Services", "fa-server"}, {"/system", "System", "fa-router"},
	{"/dataplane", "Dataplane (ASK2)", "fa-microchip"}, {"/config", "Configuration", "fa-code-compare"}, {"/logs", "Logs & Events", "fa-list-timeline"},
	{"/diagnostics", "Diagnostics", "fa-radar"}, {"/settings", "Settings", "fa-gear"},
}

func matches(query string, values ...string) bool {
	haystack := strings.ToLower(strings.Join(values, " "))
	for _, word := range strings.Fields(strings.ToLower(query)) {
		if !strings.Contains(haystack, word) {
			return false
		}
	}
	return true
}

func (r Record) matches(query string) bool {
	return matches(query, append([]string{r.ID, r.Name, r.Status, r.Detail}, r.Values...)...)
}

type RecordView struct {
	Record
	Href, SparkPath string
}

type Tab struct {
	Title, Href string
	Current     bool
}

type PageView struct {
	Shell
	Page
	TabID, Query, Sort, SelectedID string
	Tabs                           []Tab
	Records                        []RecordView
	Columns                        []string
	Selected                       *Record
	Related, Close                 string
	Options                        []NavItem
	Chart                          *SeriesView
	Heatmap                        *HeatmapView
	Visual                         []VisualMark
	Meters                         []MeterView
	Diagram                        bool
	LookupIP                       string
	LookupResult                   *RecordView
	LookupError                    string
	Categories                     []CategoryItem
	SelectedCategories             []string
	ZoneNames                      []string
	ZoneRows                       []ZoneRow
	Summary                        string
	PacketPath, ErrorPath          string
	SummaryMetrics                 []KPI
	EvidencePanels                 []EvidencePanelView
	ZoneFilter, ZoneFilterLabel    string
	ClearZone                      string
}

type VisualMark struct{ Name, Value, Href string }
type ZoneCell struct {
	Label, Value, Action, Href string
	Selected                   bool
}
type ZoneRow struct {
	Name  string
	Cells []ZoneCell
}

func viewURL(path, scenario, rangeID string, pairs ...string) string {
	q := url.Values{}
	if scenario != "warning" {
		q.Set("scenario", scenario)
	}
	if rangeID != "24h" {
		q.Set("range", rangeID)
	}
	for i := 0; i < len(pairs); i += 2 {
		if pairs[i+1] != "" {
			q.Set(pairs[i], pairs[i+1])
		}
	}
	if len(q) == 0 {
		return path
	}
	return path + "?" + q.Encode()
}

func pageView(page Page, shared Shell, query url.Values) (PageView, error) {
	scenario, rangeID, theme := shared.Scenario, shared.RangeID, shared.Theme
	link := func(path string, pairs ...string) string {
		return viewURL(path, scenario, rangeID, append(pairs, "scoutTheme", theme)...)
	}
	v := PageView{Shell: shared, Page: page, TabID: query.Get("tab"),
		Query: strings.TrimSpace(query.Get("q")), Sort: query.Get("sort"), SelectedID: query.Get("sel")}
	if len(page.Sections) == 0 {
		return v, errors.New("page has no sections")
	}
	if v.Sort != "" && v.Sort != "name" && v.Sort != "status" {
		return v, errors.New("unknown sort")
	}
	if v.TabID == "" {
		v.TabID = page.Sections[0].ID
	}
	var selectedTab *Section
	for i := range page.Sections {
		tab := &page.Sections[i]
		if v.TabID == tab.ID {
			selectedTab = tab
		}
		v.Tabs = append(v.Tabs, Tab{tab.Title, link(page.Path, "tab", tab.ID), v.TabID == tab.ID})
	}
	if selectedTab == nil {
		return v, ErrNotFound
	}
	v.Columns = selectedTab.Columns
	var filteredRules []string
	if page.Path == "/firewall" && v.TabID == "chains" && query.Get("zone") != "" {
		v.ZoneFilter = query.Get("zone")
		for _, policy := range page.ZonePolicies {
			if v.ZoneFilter == policy.From+"--"+policy.To {
				filteredRules = policy.RuleIDs
				v.ZoneFilterLabel = policy.From + " → " + policy.To
				break
			}
		}
		if v.ZoneFilterLabel == "" {
			return v, ErrNotFound
		}
		v.ClearZone = link(page.Path, "tab", v.TabID, "q", v.Query, "sort", v.Sort)
	}
	v.SummaryMetrics = page.Summaries[v.TabID]
	for _, panel := range page.Evidence[v.TabID] {
		pv := EvidencePanelView{EvidencePanel: panel}
		for _, record := range panel.Rows {
			copy := record
			if copy.Href != "" {
				target, err := url.Parse(copy.Href)
				if err != nil || target.IsAbs() || target.Host != "" || !strings.HasPrefix(target.Path, "/") {
					return v, fmt.Errorf("invalid evidence link %q", copy.Href)
				}
				pairs := []string{}
				for key, values := range target.Query() {
					for _, value := range values {
						pairs = append(pairs, key, value)
					}
				}
				copy.Href = link(target.Path, pairs...)
				if target.Fragment != "" {
					copy.Href += "#" + url.PathEscape(target.Fragment)
				}
			}
			pv.Rows = append(pv.Rows, copy)
		}
		v.EvidencePanels = append(v.EvidencePanels, pv)
	}
	if page.InterfaceDetail != nil && v.TabID == "overview" {
		detail := *page.InterfaceDetail
		detail.Counters = append([]InterfaceCounter(nil), detail.Counters...)
		for i := range detail.Counters {
			detail.Counters[i].Path = sparkline(detail.Counters[i].Trend)
		}
		v.Page.InterfaceDetail = &detail
		v.PacketPath = sparkline(detail.Packets)
		v.ErrorPath = sparkline(detail.Errors)
	}
	if page.Path == "/interfaces" {
		for _, record := range selectedTab.Records {
			if record.Category != "" && !slices.Contains(v.SelectedCategories, record.Category) &&
				slices.Contains(strings.Split(query.Get("types"), ","), record.Category) {
				v.SelectedCategories = append(v.SelectedCategories, record.Category)
			}
		}
		var added []string
		for _, record := range selectedTab.Records {
			category := record.Category
			if category == "" || slices.Contains(added, category) {
				continue
			}
			added = append(added, category)
			current := slices.Contains(v.SelectedCategories, category)
			next := append([]string(nil), v.SelectedCategories...)
			if current {
				next = slices.DeleteFunc(next, func(item string) bool { return item == category })
			} else {
				next = append(next, category)
			}
			count := 0
			for _, candidate := range selectedTab.Records {
				if candidate.Category == category {
					count++
				}
			}
			v.Categories = append(v.Categories, CategoryItem{NavItem: NavItem{Title: category, Href: link(page.Path, "tab", v.TabID, "q", v.Query,
				"sort", v.Sort, "types", strings.Join(next, ",")), Current: current}, Count: count})
		}
		var up, down, other int
		for _, record := range selectedTab.Records {
			switch record.Status {
			case "Up":
				up++
			case "Down":
				down++
			default:
				other++
			}
		}
		v.Summary = fmt.Sprintf("%d interfaces · %d up · %d down · %d standby or admin down", len(selectedTab.Records), up, down, other)
	}
	types := strings.Join(v.SelectedCategories, ",")
	v.Close = link(page.Path, "tab", v.TabID, "q", v.Query, "sort", v.Sort, "types", types, "ip", query.Get("ip"), "zone", v.ZoneFilter)
	for _, record := range selectedTab.Records {
		if record.ID == v.SelectedID {
			copy := record
			v.Selected = &copy
			if record.Related != "" {
				v.Related = link(record.Related)
			}
		}
		if record.matches(v.Query) && (types == "" || slices.Contains(v.SelectedCategories, record.Category)) &&
			(v.ZoneFilter == "" || slices.Contains(filteredRules, record.ID)) &&
			(page.Path != "/firewall" || v.TabID != "chains" || v.ZoneFilter != "" || record.Category != "zone-policy") {
			href := link(page.Path,
				"tab", v.TabID, "q", v.Query, "sort", v.Sort, "types", types, "zone", v.ZoneFilter, "sel", record.ID) + "#details"
			if page.Path == "/interfaces" && record.Related != "" {
				href = link(record.Related)
			}
			v.Records = append(v.Records, RecordView{Record: record, Href: href, SparkPath: sparkline(record.Spark)})
		}
	}
	if v.SelectedID != "" && v.Selected == nil {
		return v, ErrNotFound
	}
	if page.Path == "/routing" && v.TabID == "routes" {
		v.LookupIP = strings.TrimSpace(query.Get("ip"))
		if v.LookupIP != "" {
			addr, err := netip.ParseAddr(v.LookupIP)
			if err != nil {
				v.LookupError = "Enter a valid IPv4 or IPv6 address."
			} else {
				best := -1
				for _, record := range selectedTab.Records {
					prefix, err := netip.ParsePrefix(record.Name)
					if err == nil && prefix.Bits() > best && prefix.Contains(addr) {
						best = prefix.Bits()
						v.LookupResult = &RecordView{Record: record, Href: link(page.Path, "tab", v.TabID, "sel", record.ID, "ip", v.LookupIP) + "#details"}
					}
				}
			}
		}
	}
	slices.SortStableFunc(v.Records, func(a, b RecordView) int {
		switch v.Sort {
		case "name":
			return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
		case "status":
			return strings.Compare(a.Status, b.Status)
		}
		return 0
	})
	if page.Path == "/settings" {
		for _, option := range []struct{ label, value string }{{"Dark theme", "dark"}, {"Light theme", "light"}, {"System theme", "system"}} {
			v.Options = append(v.Options, NavItem{Title: option.label,
				Href: viewURL(page.Path, scenario, rangeID, "scoutTheme", option.value), Current: theme == option.value || (theme == "" && option.value == "dark")})
		}
		for _, candidate := range []string{"normal", "warning", "unavailable"} {
			v.Options = append(v.Options, NavItem{Title: candidate + " scenario",
				Href: viewURL(page.Path, candidate, rangeID, "scoutTheme", theme), Current: scenario == candidate})
		}
	}
	if series := page.Series[v.TabID]; series != nil && series.Max > 0 && len(series.Points) > 1 {
		chart := &SeriesView{Series: *series, Range: rangeID}
		if series.Bars {
			width := 720.0 / float64(len(series.Points))
			for i, value := range series.Points {
				height := 180 * float64(value) / float64(series.Max)
				chart.Columns = append(chart.Columns, ChartBar{40 + float64(i)*width + width/4, 200 - height, width / 2, height})
			}
		} else {
			if len(series.TX) == len(series.Points) {
				chart.Path = mirroredPath(series.Points, series.Max, false)
				chart.TXPath = mirroredPath(series.TX, series.Max, true)
				chart.MarkerX = 400
			} else {
				chart.Path = pathForValues(series.Points, series.Max)
			}
		}
		for _, option := range []string{"1h", "6h", "24h", "7d"} {
			chart.Options = append(chart.Options, NavItem{Title: option,
				Href: viewURL(page.Path, scenario, option, "tab", v.TabID, "q", v.Query, "sort", v.Sort, "sel", v.SelectedID, "scoutTheme", theme), Current: option == rangeID})
		}
		v.Chart = chart
	}
	visible := make(map[string]bool, len(v.Records))
	for _, record := range v.Records {
		visible[record.ID] = true
	}
	heatmap, err := renderHeatmap(page.Heatmaps[v.TabID], visible, v.SelectedID, query.Get("sample"), func(id, sample string) string {
		return link(page.Path, "tab", v.TabID, "q", v.Query, "sort", v.Sort, "sel", id, "sample", sample) + "#details"
	})
	if err != nil {
		return v, err
	}
	if heatmap != nil && v.Chart == nil {
		for _, option := range []string{"1h", "6h", "24h", "7d"} {
			heatmap.Options = append(heatmap.Options, NavItem{Title: option,
				Href: viewURL(page.Path, scenario, option, "tab", v.TabID, "q", v.Query, "sort", v.Sort, "scoutTheme", theme), Current: option == rangeID})
		}
	}
	v.Heatmap = heatmap
	for _, record := range v.Records {
		if len(record.Values) == 0 {
			continue
		}
		if page.Path == "/topology" || (page.Path == "/clients" && v.TabID == "paths") {
			v.Visual = append(v.Visual, VisualMark{record.Name, strings.Join(record.Values, " · "), record.Href})
		}
		for _, meter := range page.Meters[v.TabID] {
			if meter.ID == record.ID && meter.Capacity > 0 && meter.Value >= 0 && meter.Value <= meter.Capacity {
				v.Meters = append(v.Meters, MeterView{meter, record.Href})
			}
		}
	}
	v.Diagram = page.Path == "/topology" && len(v.Visual) > 0
	if page.Path == "/firewall" && (v.TabID == "chains" || v.TabID == "zones") && len(page.ZonePolicies) > 0 {
		v.ZoneNames = page.ZoneNames
		for _, source := range page.ZoneNames {
			hasSource := false
			row := ZoneRow{Name: source}
			for _, target := range page.ZoneNames {
				cell := ZoneCell{Label: source + " → " + target, Value: "—"}
				for _, policy := range page.ZonePolicies {
					if policy.From == source && policy.To == target {
						hasSource = true
						cell.Action = policy.Action
						cell.Value = policy.Action
						if len(policy.RuleIDs) > 1 {
							cell.Value += fmt.Sprintf(" +%d", len(policy.RuleIDs)-1)
						}
						id := source + "--" + target
						cell.Label += " · " + cell.Value
						cell.Href = link(page.Path, "tab", "chains", "zone", id) + "#firewall-rules"
						cell.Selected = v.ZoneFilter == id
						break
					}
				}
				row.Cells = append(row.Cells, cell)
			}
			if hasSource {
				v.ZoneRows = append(v.ZoneRows, row)
			}
		}
	}
	return v, nil
}

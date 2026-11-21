package demo

import (
	"fmt"
	"strings"
)

type Point struct {
	Label string
	RX    int
	TX    int
}

type Range struct {
	ID     string
	Label  string
	Points []Point
}

type Link struct {
	Name, Description, State, Speed, RX, TX string
}

type KPI struct {
	Name, Value, Detail, Status string
}

type Talker struct {
	Name, Address, Received, Sent string
}

type Event struct {
	Time, Kind, Description, Status string
}

type Overview struct {
	Scenario, RangeID, RangeLabel string
	Alert                         string
	WANState, WANAddress          string
	WANPrefix, WANGateway         string
	WANLatency, WANLoss, WANLease string
	Links                         []Link
	KPIs                          []KPI
	Talkers                       []Talker
	Events                        []Event
	Points                        []Point
	Ranges                        []RangeOption
	RXPath, TXPath                string
	PeakRX, PeakTX                int
}

type RangeOption struct {
	ID, Label string
	Selected  bool
	Href      string
}

type OverviewView struct {
	Overview
	Shell         Shell
	EvidenceLinks map[string]string
}

// Fixture returns a fresh, deterministic snapshot; it never consults the host.
func Fixture(rangeID, scenario string) (Overview, error) {
	ranges := []Range{
		{ID: "1h", Label: "Last hour", Points: []Point{{"60 min ago", 220, 48}, {"50 min ago", 310, 55}, {"40 min ago", 280, 62}, {"30 min ago", 510, 85}, {"20 min ago", 630, 110}, {"10 min ago", 480, 90}, {"Now", 420, 72}}},
		{ID: "6h", Label: "Last 6 hours", Points: []Point{{"6 h ago", 180, 35}, {"5 h ago", 250, 45}, {"4 h ago", 310, 58}, {"3 h ago", 420, 78}, {"2 h ago", 650, 105}, {"1 h ago", 500, 90}, {"Now", 420, 72}}},
		{ID: "24h", Label: "Last 24 hours", Points: []Point{{"24 h ago", 150, 30}, {"20 h ago", 180, 45}, {"16 h ago", 200, 50}, {"12 h ago", 360, 70}, {"8 h ago", 690, 130}, {"4 h ago", 520, 95}, {"Now", 420, 72}}},
		{ID: "7d", Label: "Last 7 days", Points: []Point{{"7 days ago", 290, 40}, {"6 days ago", 440, 60}, {"5 days ago", 320, 80}, {"4 days ago", 610, 120}, {"3 days ago", 410, 75}, {"2 days ago", 720, 140}, {"Now", 420, 72}}},
	}
	if scenario != "normal" && scenario != "warning" && scenario != "unavailable" {
		return Overview{}, fmt.Errorf("unknown demo scenario %q", scenario)
	}
	var selected *Range
	var options []RangeOption
	for i := range ranges {
		r := &ranges[i]
		options = append(options, RangeOption{ID: r.ID, Label: r.Label, Selected: r.ID == rangeID})
		if r.ID == rangeID {
			selected = r
		}
	}
	if selected == nil {
		return Overview{}, fmt.Errorf("unknown range %q", rangeID)
	}
	o := Overview{
		Scenario: scenario, RangeID: rangeID, RangeLabel: selected.Label, Points: selected.Points, Ranges: options,
		WANState: "Up", WANAddress: "203.0.113.42", WANPrefix: "2001:db8:4f2a::/56", WANGateway: "203.0.113.1",
		WANLatency: "8.4 ms", WANLoss: "0.0%", WANLease: "21 h",
		Links: []Link{
			{"eth0", "WAN", "Up", "10G", "420 Mb/s", "72 Mb/s"},
			{"br0", "LAN", "Up", "—", "58 Mb/s", "690 Mb/s"},
			{"eth3", "Lab SFP+", "Up", "10G", "1.2 Mb/s", "0.4 Mb/s"},
			{"wg0", "VPN", "Up", "—", "22 Mb/s", "9 Mb/s"},
			{"pppoe0", "WAN2", "Standby", "—", "—", "—"},
		},
		KPIs: []KPI{
			{"CPU", "14%", "4 cores · load 0.42", "ok"},
			{"Memory", "1.9 GiB", "of 8 GiB · 24%", "ok"},
			{"Temperature", "58 °C", "critical 95 °C", "ok"},
			{"Conntrack", "12.4k", "of 262k · 4.7%", "ok"},
			{"Clients", "47", "31 wired · 16 Wi-Fi", "ok"},
			{"VPN peers", "5 / 6", "1 stale handshake", "warning"},
			{"Offload hit", "93.8%", "ASK2 · 9.9k flows", "ok"},
		},
		Talkers: []Talker{
			{"nas-01", "192.0.2.20", "41.2 GB", "3.1 GB"},
			{"tv-livingroom", "192.0.2.41", "18.7 GB", "0.2 GB"},
			{"workstation", "192.0.2.23", "9.8 GB", "1.6 GB"},
			{"game-console", "192.0.2.35", "6.8 GB", "1.0 GB"},
			{"phone", "192.0.2.12", "2.4 GB", "0.4 GB"},
		},
		Events: []Event{
			{"18:41", "WireGuard", "Example peer handshake is stale", "warning"},
			{"17:02", "DHCP", "Example LAN pool reached 91%", "info"},
			{"15:12", "Commit r41", "Example configuration change", "info"},
			{"14:58", "Link", "Example WAN link recovered", "ok"},
		},
	}
	switch scenario {
	case "warning":
		o.Alert = "2 example warnings: VPN peer handshake is stale; LAN DHCP pool is 91% allocated."
	case "normal":
		o.KPIs[5] = KPI{"VPN peers", "6 / 6", "All peers healthy", "ok"}
		o.Events = []Event{{"14:58", "Link", "Example WAN link recovered", "ok"}}
	case "unavailable":
		o.Alert = "Example data source unavailable. Some values cannot be displayed."
		o.WANState, o.WANAddress = "Unknown", "Unavailable"
		o.WANPrefix, o.WANGateway, o.WANLatency, o.WANLoss, o.WANLease = "", "", "", "", ""
		o.Links = nil
		o.KPIs = []KPI{{"CPU", "Unavailable", "No demo sample", "unknown"}, {"Memory", "Unavailable", "No demo sample", "unknown"}, {"Temperature", "Unavailable", "No demo sample", "unknown"}, {"Conntrack", "Unavailable", "No demo sample", "unknown"}, {"Clients", "Unavailable", "No demo sample", "unknown"}, {"VPN peers", "Unavailable", "No demo sample", "unknown"}, {"Offload hit", "Unavailable", "No demo sample", "unknown"}}
		o.Points, o.Talkers, o.Events = nil, nil, nil
	}
	if len(o.Points) > 0 {
		o.RXPath = chartPath(o.Points, true)
		o.TXPath = chartPath(o.Points, false)
		for _, point := range o.Points {
			o.PeakRX = max(o.PeakRX, point.RX)
			o.PeakTX = max(o.PeakTX, point.TX)
		}
	}
	return o, nil
}

func chartPath(points []Point, rx bool) string {
	values := make([]int, len(points))
	for i, point := range points {
		values[i] = point.TX
		if rx {
			values[i] = point.RX
		}
	}
	return pathForValues(values, 1000)
}

func pathForValues(values []int, ceiling int) string {
	var path strings.Builder
	for i, value := range values {
		x := 40 + float64(i)*720/float64(len(values)-1)
		y := 200 - float64(value)*180/float64(ceiling)
		if i == 0 {
			fmt.Fprintf(&path, "M %.1f %.1f", x, y)
		} else {
			fmt.Fprintf(&path, " L %.1f %.1f", x, y)
		}
	}
	return path.String()
}

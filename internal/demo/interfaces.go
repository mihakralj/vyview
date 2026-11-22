package demo

import (
	"fmt"
	"strings"
)

type InterfaceAddress struct {
	CIDR  string
	Scope string // "IPv4", "IPv6 GUA", "IPv6 ULA", "IPv6 LLA", "IPv6"
	Peer  string // optional P2P remote peer, e.g. "10.99.50.1/32"
	Meta  string // optional quiet tag, e.g. "DHCP · 22h"
}

type PhysicalPort struct {
	Name, Connector, Speed, Status, SL, Description, Href string
	Children                                              []string
}

type InterfaceDetail struct {
	State, Mode, Driver, MAC string
	HasOptics                bool
	Metrics                  []KPI
	Optics                   []OpticalReading
	Counters                 []InterfaceCounter
	Packets, Errors          []int
}

type OpticalReading struct {
	Name, Value, Unit, Status       string
	Reading, AlarmLow, WarningLow   float64
	WarningHigh, AlarmHigh, Optimum float64
}

type InterfaceCounter struct {
	Name, Value, Delta, Rate string
	Trend                    []int
	Path                     string
}

func mirroredPath(values []int, ceiling int, lower bool) string {
	var path strings.Builder
	for i, value := range values {
		x := 40 + float64(i)*720/float64(len(values)-1)
		y := 110 - float64(value)*90/float64(ceiling)
		if lower {
			y = 110 + float64(value)*90/float64(ceiling)
		}
		cmd := "L"
		if i == 0 {
			cmd = "M"
		}
		fmt.Fprintf(&path, "%s %.1f %.1f ", cmd, x, y)
	}
	return strings.TrimSpace(path.String())
}

func sparkline(values []int) string {
	if len(values) < 2 {
		return ""
	}
	ceiling := 1
	for _, value := range values {
		ceiling = max(ceiling, value)
	}
	return pathForValues(values, ceiling)
}

// btopBars computes mirrored RX (rising above centerline) and TX (falling below
// centerline) vertical spike bars on a shared scale inside viewBox="0 0 120 40".
func btopBars(rx, tx []int) ([]SparkBar, []SparkBar) {
	n := len(rx)
	if n == 0 || len(tx) != n {
		return nil, nil
	}
	ceiling := 1
	for i := 0; i < n; i++ {
		ceiling = max(ceiling, rx[i], tx[i])
	}
	step := 120.0 / float64(n)
	barW := step - 1.5
	if barW < 1.5 {
		barW = 1.5
	}
	rxBars := make([]SparkBar, 0, n)
	txBars := make([]SparkBar, 0, n)
	for i := 0; i < n; i++ {
		x := float64(i) * step
		if rx[i] > 0 {
			h := 18.0 * float64(rx[i]) / float64(ceiling)
			if h < 1.5 {
				h = 1.5
			}
			rxBars = append(rxBars, SparkBar{X: x, Y: 19.5 - h, Width: barW, Height: h})
		}
		if tx[i] > 0 {
			h := 18.0 * float64(tx[i]) / float64(ceiling)
			if h < 1.5 {
				h = 1.5
			}
			txBars = append(txBars, SparkBar{X: x, Y: 20.5, Width: barW, Height: h})
		}
	}
	return rxBars, txBars
}

func physicalPorts() []PhysicalPort {
	return []PhysicalPort{
		{Name: "eth0", Connector: "RJ45", Speed: "1G", Status: "Up", SL: "u/u", Description: "MGMT (left RJ45)", Href: "/interfaces/eth0"},
		{Name: "eth1", Connector: "RJ45", Speed: "—", Status: "Down", SL: "u/D", Description: "RJ45 center", Href: "/interfaces/eth1"},
		{Name: "eth2", Connector: "RJ45", Speed: "—", Status: "Down", SL: "u/D", Description: "RJ45 right", Href: "/interfaces/eth2"},
		{Name: "eth3", Connector: "SFP+", Speed: "10G", Status: "Up", SL: "u/u", Description: "SFP+ left", Href: "/interfaces/eth3", Children: []string{"eth3.10", "pppoe10"}},
		{Name: "eth4", Connector: "SFP+", Speed: "10G", Status: "Up", SL: "u/u", Description: "SFP+ right", Href: "/interfaces/eth4", Children: []string{"eth4.20"}},
	}
}

func interfaceRecords() []Record {
	type entry struct {
		name, kind, description string
		group, parent, prefix   string
		depth                   int
		underlay                string
		mac, vrf, mtu, sl, link string
		altName                 string
		rx, tx, state           string
		rxSamples, txSamples    []int
		addresses               []InterfaceAddress
		errors, drops           int
		errorNote               string
	}
	entries := []entry{
		// Group 1: Physical & L2 Hierarchy (Mono Gateway hardware layout)
		{
			name: "eth0", kind: "Ethernet", description: "MGMT (left RJ45)",
			group: "Physical & L2", depth: 0, altName: "end1",
			mac: "e8:f6:d7:00:16:00", vrf: "default", mtu: "1500", sl: "u/u", link: "1G full · RJ45",
			rx: "18.4 Mb/s", tx: "12.1 Mb/s", state: "Up",
			rxSamples: []int{9, 12, 14, 11, 16, 22, 19, 15, 21, 24, 18, 13, 17, 20, 16, 14, 23, 19, 17, 18},
			txSamples: []int{6, 8, 7, 10, 9, 14, 11, 8, 12, 15, 10, 7, 9, 13, 11, 9, 14, 12, 11, 12},
			addresses: []InterfaceAddress{
				{CIDR: "192.168.1.185/16", Scope: "IPv4", Meta: "DHCP · 22h"},
				{CIDR: "2001:db8:1::185/64", Scope: "IPv6 GUA"},
				{CIDR: "fd00:1::185/64", Scope: "IPv6 ULA"},
				{CIDR: "fe80::eaf6:d7ff:fe00:1600/64", Scope: "IPv6 LLA"},
			},
		},
		{
			name: "eth1", kind: "Ethernet", description: "RJ45 center",
			group: "Physical & L2", depth: 0, altName: "end5",
			mac: "e8:f6:d7:00:16:01", vrf: "default", mtu: "1500", sl: "u/D", link: "No carrier · RJ45",
			rx: "—", tx: "—", state: "Down",
		},
		{
			name: "eth2", kind: "Ethernet", description: "RJ45 right",
			group: "Physical & L2", depth: 0, altName: "end0",
			mac: "e8:f6:d7:00:15:ff", vrf: "default", mtu: "1500", sl: "u/D", link: "No carrier · RJ45",
			rx: "—", tx: "—", state: "Down",
		},
		{
			name: "eth3", kind: "Ethernet", description: "SFP+ left",
			group: "Physical & L2", depth: 0, altName: "end3",
			mac: "e8:f6:d7:00:16:02", vrf: "default", mtu: "1500", sl: "u/u", link: "10G full · SFP+",
			rx: "420 Mb/s", tx: "72 Mb/s", state: "Up",
			rxSamples: []int{180, 220, 260, 310, 280, 390, 480, 520, 440, 360, 410, 590, 630, 510, 460, 380, 430, 490, 450, 420},
			txSamples: []int{38, 44, 52, 60, 48, 70, 88, 95, 76, 62, 74, 110, 122, 92, 80, 64, 78, 86, 75, 72},
			addresses: []InterfaceAddress{
				{CIDR: "10.99.1.185/24", Scope: "IPv4"},
				{CIDR: "fd99:1::185/64", Scope: "IPv6 ULA"},
				{CIDR: "fe80::185/64", Scope: "IPv6 LLA"},
			},
			errors: 2, drops: 0, errorNote: "2 rx err",
		},
		{
			name: "eth3.10", kind: "VLAN", description: "ISP trunk on eth3",
			group: "Physical & L2", parent: "eth3", depth: 1, prefix: "└─",
			mac: "e8:f6:d7:00:16:02", vrf: "default", mtu: "1500", sl: "u/u", link: "802.1Q · ID 10",
			rx: "312 Mb/s", tx: "54 Mb/s", state: "Up",
			rxSamples: []int{140, 165, 195, 240, 210, 290, 360, 390, 330, 270, 310, 440, 470, 380, 340, 285, 320, 365, 335, 312},
			txSamples: []int{28, 32, 39, 46, 36, 52, 66, 72, 58, 46, 55, 82, 90, 68, 60, 48, 58, 64, 56, 54},
			addresses: []InterfaceAddress{
				{CIDR: "10.99.10.185/24", Scope: "IPv4"},
				{CIDR: "fd99:10::185/64", Scope: "IPv6 ULA"},
				{CIDR: "fe80::185/64", Scope: "IPv6 LLA"},
			},
		},
		{
			name: "pppoe10", kind: "PPPoE", description: "Session over eth3.10",
			group: "Physical & L2", parent: "eth3.10", depth: 2, prefix: "└─",
			mac: "n/a", vrf: "default", mtu: "1492", sl: "u/u", link: "PPPoE · eth3.10",
			rx: "298 Mb/s", tx: "49 Mb/s", state: "Up",
			rxSamples: []int{132, 158, 184, 228, 198, 276, 342, 372, 315, 255, 294, 420, 448, 362, 324, 270, 305, 348, 320, 298},
			txSamples: []int{25, 29, 36, 42, 32, 48, 60, 66, 52, 42, 50, 75, 82, 62, 54, 44, 52, 58, 51, 49},
			addresses: []InterfaceAddress{
				{CIDR: "10.99.50.16", Scope: "IPv4", Peer: "10.99.50.1/32"},
				{CIDR: "fe80::6124:fc3d:c9ac:75a3", Scope: "IPv6 LLA", Peer: "fe80::bcf3:88a:cbe4:5b37/128"},
			},
		},
		{
			name: "eth4", kind: "Ethernet", description: "SFP+ right",
			group: "Physical & L2", depth: 0, altName: "end4",
			mac: "e8:f6:d7:00:16:03", vrf: "default", mtu: "1500", sl: "u/u", link: "10G full · SFP+",
			rx: "184 Mb/s", tx: "640 Mb/s", state: "Up",
			rxSamples: []int{90, 110, 125, 160, 140, 195, 230, 210, 175, 150, 180, 250, 270, 220, 195, 165, 190, 215, 196, 184},
			txSamples: []int{310, 380, 420, 520, 470, 610, 720, 680, 590, 510, 600, 790, 840, 710, 660, 580, 650, 700, 670, 640},
			addresses: []InterfaceAddress{
				{CIDR: "10.99.2.185/24", Scope: "IPv4"},
				{CIDR: "fd99:2::185/64", Scope: "IPv6 ULA"},
				{CIDR: "fe80::eaf6:d7ff:fe00:1603/64", Scope: "IPv6 LLA"},
				{CIDR: "fe80::185/64", Scope: "IPv6 LLA"},
			},
			errors: 0, drops: 12, errorNote: "12 rx drop",
		},
		{
			name: "eth4.20", kind: "VLAN", description: "LAN / Lab on eth4",
			group: "Physical & L2", parent: "eth4", depth: 1, prefix: "└─",
			mac: "e8:f6:d7:00:16:03", vrf: "default", mtu: "1500", sl: "u/u", link: "802.1Q · ID 20",
			rx: "142 Mb/s", tx: "510 Mb/s", state: "Up",
			rxSamples: []int{70, 86, 98, 124, 110, 152, 180, 162, 136, 118, 140, 194, 210, 172, 152, 128, 148, 168, 150, 142},
			txSamples: []int{245, 300, 335, 415, 375, 490, 575, 540, 470, 405, 480, 630, 670, 565, 525, 460, 520, 560, 535, 510},
			addresses: []InterfaceAddress{
				{CIDR: "10.99.20.185/24", Scope: "IPv4"},
				{CIDR: "fd99:20::185/64", Scope: "IPv6 ULA"},
				{CIDR: "fe80::eaf6:d7ff:fe00:1603/64", Scope: "IPv6 LLA"},
			},
		},
		// Group 2: L3 Tunnels & VPN Overlays (routed — shows live underlay egress chip)
		{
			name: "wg0", kind: "WireGuard", description: "Remote access · 2 peers",
			group: "Tunnels & VPN Overlays", depth: 0, underlay: "pppoe10",
			mac: "n/a", vrf: "default", mtu: "1420", sl: "u/u", link: "L3 overlay · 2 peers",
			rx: "22 Mb/s", tx: "9 Mb/s", state: "Up",
			rxSamples: []int{12, 18, 15, 24, 28, 19, 14, 26, 32, 25, 18, 21, 29, 34, 27, 20, 24, 28, 25, 22},
			txSamples: []int{5, 8, 6, 11, 12, 9, 6, 10, 14, 11, 8, 9, 12, 15, 11, 8, 10, 12, 10, 9},
			addresses: []InterfaceAddress{
				{CIDR: "10.200.0.1/24", Scope: "IPv4"},
				{CIDR: "fd00:200::1/64", Scope: "IPv6 ULA"},
			},
		},
		{
			name: "xfrm0", kind: "IPsec", description: "Site-to-site · CAAM HW",
			group: "Tunnels & VPN Overlays", depth: 0, underlay: "eth3",
			mac: "n/a", vrf: "default", mtu: "1400", sl: "u/u", link: "L3 overlay · 1 SA",
			rx: "14.2 Mb/s", tx: "11.4 Mb/s", state: "Up",
			rxSamples: []int{8, 10, 12, 11, 15, 18, 14, 13, 17, 19, 15, 12, 14, 16, 18, 15, 13, 16, 15, 14},
			txSamples: []int{7, 9, 10, 9, 12, 14, 12, 11, 14, 16, 12, 10, 11, 13, 15, 12, 11, 13, 12, 11},
			addresses: []InterfaceAddress{
				{CIDR: "169.254.10.1/30", Scope: "IPv4"},
				{CIDR: "fe80::10:1/64", Scope: "IPv6 LLA"},
			},
		},
		// Group 3: Loopback & Control
		{
			name: "lo", kind: "Loopback", description: "Local host",
			group: "Loopback & Control", depth: 0,
			mac: "00:00:00:00:00:00", vrf: "default", mtu: "65536", sl: "u/u", link: "Virtual",
			rx: "—", tx: "—", state: "Up",
			addresses: []InterfaceAddress{
				{CIDR: "127.0.0.1/8", Scope: "IPv4"},
				{CIDR: "::1/128", Scope: "IPv6"},
				{CIDR: "fe80::200:ff:fe00:0/64", Scope: "IPv6 LLA"},
			},
		},
	}

	typeIcons := map[string]string{
		"Ethernet":  "fa-ethernet",
		"VLAN":      "fa-layer-group",
		"PPPoE":     "fa-plug",
		"Bridge":    "fa-bridge",
		"Bond":      "fa-link",
		"WireGuard": "fa-lock-keyhole",
		"IPsec":     "fa-shield-halved",
		"Loopback":  "fa-arrows-rotate",
	}

	records := make([]Record, 0, len(entries))
	for i, e := range entries {
		addrSummary := "—"
		if len(e.addresses) > 0 {
			parts := make([]string, len(e.addresses))
			for idx, a := range e.addresses {
				if a.Peer != "" {
					parts[idx] = a.CIDR + " ↔ " + a.Peer
				} else {
					parts[idx] = a.CIDR
				}
			}
			addrSummary = strings.Join(parts, " · ")
		}
		tpSummary := "—"
		if e.rx != "—" {
			tpSummary = e.rx + " / " + e.tx
		}
		r := row(e.name, e.name, e.state, e.description, "/interfaces/"+e.name,
			addrSummary, e.vrf, e.link, tpSummary)
		r.Category = e.kind
		r.Icon = typeIcons[e.kind]
		if r.Icon == "" {
			r.Icon = "fa-ethernet"
		}
		r.Color = i%16 + 1
		r.Errors = e.errors
		r.Drops = e.drops
		r.ErrorNote = e.errorNote
		r.AltName = e.altName
		r.Group = e.group
		r.Parent = e.parent
		r.Depth = e.depth
		r.TreePrefix = e.prefix
		r.SL = e.sl
		r.MAC = e.mac
		r.VRF = e.vrf
		r.MTU = e.mtu
		r.Underlay = e.underlay
		if e.underlay != "" {
			r.UnderlayHref = "/interfaces/" + e.underlay
		}
		r.Addresses = e.addresses
		r.RXLabel = e.rx
		r.TXLabel = e.tx
		r.Spark = e.rxSamples
		r.SparkTX = e.txSamples
		records = append(records, r)
	}
	return records
}

func interfaceDetail(name, rangeID string) (Page, bool) {
	var item Record
	found := false
	for _, candidate := range interfaceRecords() {
		if candidate.Name == name {
			item, found = candidate, true
			break
		}
	}
	if !found {
		return Page{}, false
	}

	page := interfacePage(name)
	page.Description = item.Detail + " · " + item.Category + " · VRF " + item.VRF
	if item.AltName != "" {
		page.Description += " · altname " + item.AltName
	}
	page.Source = "Link statistics, interface addresses and driver diagnostics"

	// Build full address records from the interface's multi-IPv4/IPv6 slice.
	addrRecords := make([]Record, 0, len(item.Addresses)+1)
	for i, a := range item.Addresses {
		val := a.CIDR
		if a.Peer != "" {
			val += " ↔ " + a.Peer
		}
		scopeLabel := a.Scope
		if a.Meta != "" {
			scopeLabel += " · " + a.Meta
		}
		addrRecords = append(addrRecords, row(
			fmt.Sprintf("addr-%d", i),
			a.Scope+" address",
			item.Status,
			fmt.Sprintf("Configured %s address on %s.", scopeLabel, item.Name),
			"/routing",
			val,
			scopeLabel,
		))
	}
	if len(addrRecords) == 0 {
		addrRecords = append(addrRecords, row("addr-none", "No addresses assigned", item.Status, "Interface currently has no IPv4 or IPv6 addresses.", "/interfaces", "—", "Unassigned"))
	}

	// Rich detail for eth0 (1G RJ45 MGMT), eth3 (10G SFP+ left), and eth4 (10G SFP+ right).
	isSFP := name == "eth3" || name == "eth4"
	if name == "eth0" || isSFP {
		mode := "1G full · RJ45"
		rxVal, txVal, rxPkts, p95RX, p95TX := "18.4 Mb/s", "12.1 Mb/s", "3.4 k/s", "42 Mb/s", "28 Mb/s"
		errVal, dropVal := "0", "0"
		errDetail, dropDetail := "CRC 0 · frame 0", "RX FIFO 0"
		errStatus, dropStatus := "ok", "ok"
		if name == "eth3" {
			mode = "10G full · SFP+"
			rxVal, txVal, rxPkts, p95RX, p95TX = "420 Mb/s", "72 Mb/s", "68.2 k/s", "690 Mb/s", "130 Mb/s"
			errVal, errDetail, errStatus = "2", "rx_fcs_err 2 (1 h)", "warning"
		} else if name == "eth4" {
			mode = "10G full · SFP+"
			rxVal, txVal, rxPkts, p95RX, p95TX = "184 Mb/s", "640 Mb/s", "94.1 k/s", "290 Mb/s", "880 Mb/s"
			dropVal, dropDetail, dropStatus = "12", "rx_dropped 12 (1 h)", "warning"
		}

		detail := &InterfaceDetail{
			State: item.Status + " (" + item.SL + ")", Mode: mode, Driver: "fsl_dpaa_eth", MAC: item.MAC,
			HasOptics: isSFP,
			Metrics: []KPI{
				{"RX", rxVal, "p95 " + p95RX, "ok"},
				{"TX", txVal, "p95 " + p95TX, "ok"},
				{"RX packets", rxPkts, "VRF " + item.VRF + " · MTU " + item.MTU, "ok"},
				{"Errors", errVal, errDetail, errStatus},
				{"Drops", dropVal, dropDetail, dropStatus},
				{"Link flaps", "1", "last 24 hours", "warning"},
			},
			Counters: []InterfaceCounter{
				{"rx_packets", "9,811,204,551", "+1.47 G", rxPkts, []int{3, 5, 6, 4, 8, 4, 5}, ""},
				{"rx_bytes", "13.36 TB", "+2.01 GB", rxVal, []int{2, 3, 4, 7, 8, 5, 4}, ""},
				{"rx_fcs_err", errVal, "+" + errVal, "—", []int{0, 0, 0, 0, 2, 0, 0}, ""},
				{"tx_packets", "4,102,667,980", "+0.66 G", "31.1 k/s", []int{1, 2, 3, 4, 3, 2, 3}, ""},
				{"rx_dropped", dropVal, "+" + dropVal, "0", []int{0, 0, 0, 0, 0, 0, 0}, ""},
				{"bp_count[0]", "498", "—", "—", []int{4, 4, 5, 5, 5, 5, 5}, ""},
			},
			Packets: fixtureSamples(rangeID, []int{18, 30, 25, 45, 70, 50, 36, 82, 22}),
			Errors:  fixtureSamples(rangeID, []int{0, 0, 0, 2, 0, 0, 0, 0, 2}),
		}
		if isSFP {
			detail.Optics = []OpticalReading{
				{"TX power", "-2.41 dBm", "dBm", "OK", -2.41, -10, -7.3, 0, 2, -3},
				{"RX power", "-5.86 dBm", "dBm", "OK", -5.86, -14, -9.9, 0, 2, -5},
				{"Temperature", "47.2 °C", "°C", "OK", 47.2, -10, -5, 75, 85, 35},
				{"Supply", "3.29 V", "V", "OK", 3.29, 3, 3.13, 3.47, 3.6, 3.3},
				{"Laser bias", "6.8 mA", "mA", "OK", 6.8, 1, 2, 12, 14, 7},
			}
		}
		page.InterfaceDetail = detail

		sections := []Section{
			{"overview", "Overview", []string{"RX", "TX", "Speed", "VRF"}, []Record{
				row("rate", "Throughput snapshot", item.Status, item.Detail, "/interfaces", rxVal, txVal, mode, item.VRF),
			}},
			{"counters", "Counters", []string{"Value", "Delta", "Rate"}, []Record{
				row("rx-packets", "rx_packets", "OK", "Received packets since boot.", "", "9,811,204,551", "+1.47 G", rxPkts),
				row("fcs", "rx_fcs_err", "Warning", "Example receive frame check errors.", "", errVal, "+"+errVal, "—"),
				row("bman", "bp_count[0]", "OK", "Example driver-specific buffer pool count.", "/dataplane", "498", "—", "—"),
			}},
		}
		if isSFP {
			sections = append(sections, Section{"optics", "Optics", []string{"Reading", "Operating range"}, []Record{
				row("rx-power", "Receive power", "OK", "Optical module diagnostic sample.", "", "-5.86 dBm", "-9.9 to 0 dBm"),
				row("tx-power", "Transmit power", "OK", "Optical transmit power.", "", "-2.41 dBm", "-7.3 to 0 dBm"),
				row("temp", "Module temperature", "OK", "Example module temperature.", "", "47.2 °C", "-5 to 75 °C"),
				row("supply", "Supply voltage", "OK", "Example module supply voltage.", "", "3.29 V", "3.13 to 3.47 V"),
				row("bias", "Laser bias", "OK", "Example laser bias current.", "", "6.8 mA", "2.0 to 12.0 mA"),
			}})
		}
		sections = append(sections,
			Section{"addresses", "Addresses", []string{"Address", "Scope"}, addrRecords},
			Section{"queues", "Queues", []string{"Qdisc", "Backlog", "Drops"}, []Record{
				row("queue", "Egress queue", "Healthy", "Example interface queue.", "/qos", "CAKE", "24 KiB", "15"),
			}},
			Section{"config", "Config", []string{"Value"}, []Record{
				row("description", "Description", "Configured", "Read-only example configuration.", "/config", item.Detail),
				row("mtu", "MTU", "Configured", "Read-only example configuration.", "/config", item.MTU),
			}},
		)
		page.Sections = sections
		rx := fixtureSamples(rangeID, []int{220, 310, 280, 510, 630, 480, 420, 690, 150})
		tx := fixtureSamples(rangeID, []int{48, 55, 62, 85, 110, 90, 72, 130, 30})
		page.Series = map[string]*Series{"overview": {Title: "Throughput · RX above / TX below", Unit: "Mb/s", Max: 1000, Points: rx, TX: tx, Annotation: "Example link flap"}}
		return page, true
	}

	// Standard detail for VLANs, PPPoE, down RJ45 ports, WireGuard, IPsec, and Loopback.
	parentDetail := item.Detail
	if item.Parent != "" {
		parentDetail += " · Parent: " + item.Parent
	} else if item.Underlay != "" {
		parentDetail += " · Active underlay: " + item.Underlay
	}
	page.Sections = []Section{
		{"overview", "Overview", []string{"RX", "TX", "Link", "MTU", "VRF"}, []Record{
			row("rate", "Throughput snapshot", item.Status+" ("+item.SL+")", parentDetail, "/interfaces", item.RXLabel, item.TXLabel, item.Values[2], item.MTU, item.VRF),
		}},
		{"counters", "Counters", []string{"Errors", "Drops"}, []Record{
			row("errors", "Errors and drops", item.Status, "Example link counter deltas.", "", fmt.Sprint(item.Errors), fmt.Sprint(item.Drops)),
		}},
		{"addresses", "Addresses", []string{"Address", "Scope"}, addrRecords},
	}
	return page, true
}

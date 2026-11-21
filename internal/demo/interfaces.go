package demo

import (
	"fmt"
	"strings"
)

type InterfaceDetail struct {
	State, Mode, Driver, MAC string
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

func interfaceRecords() []Record {
	type entry struct {
		name, kind, description, address, mtu, link, rx, tx, state string
		errors, drops                                              int
	}
	entries := []entry{
		{"eth0", "Ethernet", "WAN · SFP+ to ONT", "203.0.113.42/24", "1500", "10G full", "420 Mb/s", "72 Mb/s", "Up", 0, 0},
		{"eth1", "Ethernet", "LAN trunk", "—", "1500", "1G full", "58 Mb/s", "690 Mb/s", "Up", 0, 12},
		{"eth2", "Ethernet", "AP uplink", "—", "1500", "1G full", "12.9 Mb/s", "84 Mb/s", "Up", 0, 0},
		{"eth3", "Ethernet", "Lab SFP+", "198.51.100.17/24", "9000", "10G full", "1.2 Mb/s", "0.4 Mb/s", "Up", 2, 0},
		{"eth4", "Ethernet", "Spare SFP+", "—", "1500", "No carrier", "—", "—", "Down", 0, 0},
		{"br0", "Bridge", "LAN bridge", "192.0.2.1/24", "1500", "Virtual", "58 Mb/s", "690 Mb/s", "Up", 0, 0},
		{"br0.20", "VLAN", "Lab", "192.0.2.65/26", "1500", "Virtual", "3.3 Mb/s", "2 Mb/s", "Up", 0, 0},
		{"br0.30", "VLAN", "IoT", "192.0.2.129/26", "1500", "Virtual", "4.6 Mb/s", "38.8 Mb/s", "Up", 0, 0},
		{"br0.40", "VLAN", "Guest", "192.0.2.193/26", "1500", "Virtual", "0.8 Mb/s", "6.2 Mb/s", "Up", 0, 3},
		{"wg0", "WireGuard", "Remote access", "192.0.2.129/27", "1420", "2 peers", "22 Mb/s", "9 Mb/s", "Up", 0, 0},
		{"wg1", "WireGuard", "Site-to-site", "198.51.100.65/30", "1420", "1 peer", "5.2 Mb/s", "4.8 Mb/s", "Up", 0, 0},
		{"pppoe0", "PPPoE", "Backup WAN", "—", "1492", "Standby", "—", "—", "Standby", 0, 0},
		{"dum0", "Dummy", "Router ID", "192.0.2.254/32", "1500", "Admin down", "—", "—", "Admin down", 0, 0},
		{"lo", "Loopback", "Local host", "127.0.0.1/8 · ::1/128", "65536", "Virtual", "—", "—", "Up", 0, 0},
	}
	records := make([]Record, 0, len(entries))
	for i, e := range entries {
		r := row(e.name, e.name, e.state, e.description, "/interfaces/"+e.name,
			e.kind, e.description, e.address, e.mtu, e.link, e.rx, e.tx, fmt.Sprintf("%d / %d", e.errors, e.drops))
		r.Category, r.Color, r.Errors, r.Drops = e.kind, i%8+1, e.errors, e.drops
		if e.rx != "—" {
			r.Spark = []int{20 + i, 28 + i, 24 + i, 38 + i, 31 + i, 46 + i, 35 + i, 40 + i}
		}
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
	page.Description = item.Values[1] + " · " + item.Values[0]
	page.Source = "Link statistics, interface addresses and driver diagnostics"
	if name != "eth0" {
		var optics *Section
		for i := range page.Sections {
			if page.Sections[i].ID == "optics" {
				copy := page.Sections[i]
				optics = &copy
			}
		}
		page.Sections = []Section{
			{"overview", "Overview", []string{"RX", "TX", "Link", "MTU"}, []Record{row("rate", "Throughput snapshot", item.Status, item.Detail, "/interfaces", item.Values[5], item.Values[6], item.Values[4], item.Values[3])}},
			{"counters", "Counters", []string{"Errors", "Drops"}, []Record{row("errors", "Errors and drops", item.Status, "Example link counter deltas.", "", fmt.Sprint(item.Errors), fmt.Sprint(item.Drops))}},
			{"addresses", "Addresses", []string{"Address"}, []Record{row("address", "Address", item.Status, "Example address for this interface.", "/routing", item.Values[2])}},
		}
		if optics != nil {
			page.Sections = append(page.Sections, *optics)
		}
		return page, true
	}
	page.InterfaceDetail = &InterfaceDetail{
		State: "Up", Mode: "10G full", Driver: "fsl_dpaa_eth", MAC: "02:00:5e:10:01:10",
		Metrics: []KPI{{"RX", "420 Mb/s", "p95 690 Mb/s", "ok"}, {"TX", "72 Mb/s", "p95 130 Mb/s", "ok"},
			{"RX packets", "68.2 k/s", "average packet 1.36 kB", "ok"}, {"Errors", "0", "CRC 0 · frame 0", "ok"},
			{"Drops", "0", "RX FIFO 0", "ok"}, {"Link flaps", "1", "last 24 hours", "warning"}},
		Optics: []OpticalReading{
			{"TX power", "-2.41 dBm", "dBm", "OK", -2.41, -10, -7.3, 0, 2, -3},
			{"RX power", "-5.86 dBm", "dBm", "OK", -5.86, -14, -9.9, 0, 2, -5},
			{"Temperature", "47.2 °C", "°C", "OK", 47.2, -10, -5, 75, 85, 35},
			{"Supply", "3.29 V", "V", "OK", 3.29, 3, 3.13, 3.47, 3.6, 3.3},
			{"Laser bias", "6.8 mA", "mA", "OK", 6.8, 1, 2, 12, 14, 7},
		},
		Counters: []InterfaceCounter{
			{"rx_packets", "9,811,204,551", "+1.47 G", "68.2 k/s", []int{3, 5, 6, 4, 8, 4, 5}, ""},
			{"rx_bytes", "13.36 TB", "+2.01 GB", "420 Mb/s", []int{2, 3, 4, 7, 8, 5, 4}, ""},
			{"rx_fcs_err", "54", "+54", "—", []int{0, 0, 0, 0, 54, 0, 0}, ""},
			{"tx_packets", "4,102,667,980", "+0.66 G", "31.1 k/s", []int{1, 2, 3, 4, 3, 2, 3}, ""},
			{"rx_dropped", "0", "0", "0", []int{0, 0, 0, 0, 0, 0, 0}, ""},
			{"bp_count[0]", "498", "—", "—", []int{4, 4, 5, 5, 5, 5, 5}, ""},
		},
		Packets: fixtureSamples(rangeID, []int{18, 30, 25, 45, 70, 50, 36, 82, 22}),
		Errors:  fixtureSamples(rangeID, []int{0, 0, 0, 54, 0, 0, 0, 0, 54}),
	}
	page.Sections = []Section{
		{"overview", "Overview", []string{"RX", "TX", "Speed"}, []Record{row("rate", "Throughput snapshot", "Up", "Example receive and transmit rates.", "/interfaces", "420 Mb/s", "72 Mb/s", "10G full")}},
		{"counters", "Counters", []string{"Value", "Delta", "Rate"}, []Record{
			row("rx-packets", "rx_packets", "OK", "Received packets since boot.", "", "9,811,204,551", "+1.47 G", "68.2 k/s"),
			row("fcs", "rx_fcs_err", "Warning", "Example receive frame check errors.", "", "54", "+54", "—"),
			row("bman", "bp_count[0]", "OK", "Example driver-specific buffer pool count.", "/dataplane", "498", "—", "—"),
		}},
		{"optics", "Optics", []string{"Reading", "Operating range"}, []Record{
			row("rx-power", "Receive power", "OK", "Optical module diagnostic sample.", "", "-5.86 dBm", "-9.9 to 0 dBm"),
			row("tx-power", "Transmit power", "OK", "Optical transmit power.", "", "-2.41 dBm", "-7.3 to 0 dBm"),
			row("temp", "Module temperature", "OK", "Example module temperature.", "", "47.2 °C", "-5 to 75 °C"),
			row("supply", "Supply voltage", "OK", "Example module supply voltage.", "", "3.29 V", "3.13 to 3.47 V"),
			row("bias", "Laser bias", "OK", "Example laser bias current.", "", "6.8 mA", "2.0 to 12.0 mA"),
		}},
		{"addresses", "Addresses", []string{"Address", "Source"}, []Record{
			row("ipv4", "IPv4 address", "Reachable", "Example WAN address.", "/routing", "203.0.113.42/24", "DHCP"),
			row("ipv6", "IPv6 address", "Reachable", "Example IPv6 address.", "/routing", "2001:db8:4f2a:ff::2/64", "SLAAC"),
			row("gateway", "Gateway", "Reachable", "Example upstream gateway.", "/routing", "203.0.113.1", "Neighbor"),
		}},
		{"queues", "Queues", []string{"Qdisc", "Backlog", "Drops"}, []Record{row("queue", "Egress queue", "Healthy", "Example WAN queue.", "/qos", "CAKE", "24 KiB", "15")}},
		{"config", "Config", []string{"Value"}, []Record{
			row("description", "Description", "Configured", "Read-only example configuration.", "/config", "WAN · SFP+ to ONT"),
			row("mtu", "MTU", "Configured", "Read-only example configuration.", "/config", "1500"),
		}},
	}
	rx := fixtureSamples(rangeID, []int{220, 310, 280, 510, 630, 480, 420, 690, 150})
	tx := fixtureSamples(rangeID, []int{48, 55, 62, 85, 110, 90, 72, 130, 30})
	page.Series = map[string]*Series{"overview": {Title: "Throughput · RX above / TX below", Unit: "Mb/s", Max: 1000, Points: rx, TX: tx, Annotation: "Example link flap"}}
	return page, true
}

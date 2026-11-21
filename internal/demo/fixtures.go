package demo

import (
	"context"
	"strings"
)

func row(id, name, status, detail, related string, values ...string) Record {
	return Record{ID: id, Name: name, Status: status, Detail: detail, Related: related, Values: values}
}

func (Fixtures) Page(_ context.Context, path, rangeID, scenario string) (Page, error) {
	var page Page
	if strings.HasPrefix(path, "/interfaces/") {
		var found bool
		page, found = interfaceDetail(strings.TrimPrefix(path, "/interfaces/"), rangeID)
		if !found {
			return Page{}, ErrNotFound
		}
	} else {
		page = fixturePages()[path]
	}
	if page.Title == "" {
		return Page{}, ErrNotFound
	}
	page.Path = path
	if path == "/interfaces" {
		page.Sections[0].Columns = []string{"Type", "Description", "Address", "MTU", "Link", "RX", "TX", "Errors / drops"}
		page.Sections[0].Records = interfaceRecords()
	}
	if scenario == "unavailable" && path != "/settings" {
		for i := range page.Sections {
			page.Sections[i].Records = nil
		}
		page.InterfaceDetail = nil
		page.Series = nil
	}
	if scenario == "normal" && path == "/health" {
		page.Sections[0].Records = nil
	}
	if scenario != "unavailable" {
		enhancePage(&page, path, rangeID)
		switch {
		case strings.HasPrefix(path, "/interfaces/") && path != "/interfaces/pppoe0" && path != "/interfaces/eth0":
			samples := map[string][]int{
				"eth0": {220, 310, 280, 510, 630, 480, 420, 690, 150},
				"eth3": {0, 1, 0, 1, 2, 1, 1, 2, 1},
				"br0":  {42, 48, 65, 60, 75, 55, 58, 80, 31},
				"wg0":  {11, 18, 19, 30, 25, 16, 22, 38, 12},
			}[strings.TrimPrefix(path, "/interfaces/")]
			if len(samples) > 0 {
				page.Series = map[string]*Series{"overview": newSeries("Receive throughput", "Mb/s", rangeID, samples)}
			}
		case path == "/system":
			page.Series = map[string]*Series{"resources": newSeries("CPU utilization", "%", rangeID, []int{9, 22, 30, 13, 25, 16, 14, 53, 32})}
			page.Heatmaps = map[string]*Heatmap{"resources": fixtureHeatmap("Per-core CPU utilization", "%", 100, rangeID, []HeatmapRow{
				{"core0", "cpu", "Core 0", []int{7, 18, 30, 9, 28, 16, 12, 72, 39}},
				{"core1", "cpu", "Core 1", []int{9, 26, 34, 13, 32, 13, 17, 81, 33}},
				{"core2", "cpu", "Core 2", []int{12, 20, 33, 17, 20, 18, 15, 44, 31}},
				{"core3", "cpu", "Core 3", []int{8, 24, 23, 13, 20, 17, 12, 15, 25}},
			})}
			page.Meters = map[string][]Meter{"resources": {
				{"cpu", "CPU utilization", "14%", 14, 100},
				{"memory", "Memory used", "24%", 24, 100},
			}}
		case path == "/logs":
			page.Series = map[string]*Series{"journal": newSeries("Journal entries per bucket", "entries", rangeID, []int{1, 4, 3, 5, 2, 3, 2, 12, 9})}
			page.Series["journal"].Bars = true
		case path == "/vpn/wireguard":
			page.Series = map[string]*Series{"peers": newSeries("Phone peer handshake age", "min", rangeID, []int{1, 2, 3, 4, 5, 6, 7, 9, 12})}
		case path == "/qos":
			page.Series = map[string]*Series{"queues": newSeries("Bulk queue delay", "ms", rangeID, []int{4, 7, 8, 12, 19, 12, 10, 25, 21})}
			page.Meters = map[string][]Meter{"queues": {
				{"best-effort", "Best effort peak delay", "8 ms", 8, 25},
				{"bulk", "Bulk peak delay", "19 ms", 19, 25},
			}}
		case path == "/dataplane":
			page.Heatmaps = map[string]*Heatmap{"queues": fixtureHeatmap("CEETM occupancy", "descriptors", 128, rangeID, []HeatmapRow{
				{"queue0", "ceetm", "Queue 0", []int{4, 12, 33, 14, 25, 20, 18, 66, 35}},
				{"queue1", "ceetm-1", "Queue 1", []int{12, 39, 61, 110, 82, 43, 36, 124, 92}},
			})}
			page.Meters = map[string][]Meter{
				"offload": {{"hardware", "Hardware share", "93.8%", 938, 1000}, {"software", "Software share", "6.2%", 62, 1000}},
				"queues":  {{"ceetm", "CEETM occupancy", "18 of 128", 18, 128}, {"bman", "BMan occupancy", "112 of 512", 112, 512}, {"shunt", "Shunt occupancy", "160 of 1,024", 160, 1024}},
			}
		case path == "/clients":
			page.Meters = map[string][]Meter{"talkers": {
				{"nas", "nas-01 received", "41.2 GB", 412, 412},
				{"tv", "tv-livingroom received", "18.7 GB", 187, 412},
				{"phone", "phone received", "2.4 GB", 24, 412},
			}}
		}
	}
	return page, nil
}

func fixtureSamples(rangeID string, samples []int) []int {
	points := append([]int(nil), samples[:7]...)
	switch rangeID {
	case "6h":
		points[1], points[2] = samples[8], samples[7]
	case "24h":
		points[1], points[3] = samples[7], samples[8]
	case "7d":
		points[0], points[2], points[4] = samples[8], samples[7], samples[8]
	}
	return points
}

func fixtureHeatmap(title, unit string, capacity int, rangeID string, rows []HeatmapRow) *Heatmap {
	buckets := map[string][]string{
		"1h":  {"60m ago", "50m ago", "40m ago", "30m ago", "20m ago", "10m ago", "Latest"},
		"6h":  {"6h ago", "5h ago", "4h ago", "3h ago", "2h ago", "1h ago", "Latest"},
		"24h": {"24h ago", "20h ago", "16h ago", "12h ago", "8h ago", "4h ago", "Latest"},
		"7d":  {"6d ago", "5d ago", "4d ago", "3d ago", "2d ago", "1d ago", "Latest"},
	}[rangeID]
	for i := range rows {
		rows[i].Samples = fixtureSamples(rangeID, rows[i].Samples)
	}
	return &Heatmap{Title: title, Unit: unit, Capacity: capacity, Buckets: buckets, Rows: rows}
}

func newSeries(title, unit, rangeID string, samples []int) *Series {
	ceiling := 1
	for _, point := range samples {
		ceiling = max(ceiling, point)
	}
	return &Series{Title: title, Unit: unit, Max: ceiling, Points: fixtureSamples(rangeID, samples)}
}
func interfacePage(name string) Page {
	type sample struct{ kind, address, state, rx, tx, speed string }
	data := map[string]sample{
		"eth0":   {"WAN ethernet", "203.0.113.42/24", "Up", "420 Mb/s", "72 Mb/s", "10 Gbps"},
		"eth3":   {"Lab SFP+", "198.51.100.17/24", "Up", "1.2 Mb/s", "0.4 Mb/s", "10 Gbps"},
		"br0":    {"LAN bridge", "192.0.2.1/24", "Up", "58 Mb/s", "690 Mb/s", "Virtual"},
		"wg0":    {"WireGuard", "192.0.2.129/27", "Up", "22 Mb/s", "9 Mb/s", "Virtual"},
		"pppoe0": {"Backup WAN", "Unavailable", "Standby", "Unavailable", "Unavailable", "Virtual"},
	}[name]
	sections := []Section{
		{"overview", "Overview", []string{"RX", "TX", "Speed"}, []Record{row("rate", "Throughput snapshot", data.state, "Receive and transmit rates from the latest fixture sample.", "/interfaces", data.rx, data.tx, data.speed)}},
		{"counters", "Counters", []string{"Errors", "Drops"}, []Record{
			row("rx", "Receive counters", "OK", "Example receive errors and drops since boot.", "", "0", "0"),
			row("tx", "Transmit counters", "OK", "Example transmit errors and drops since boot.", "", "0", "0"),
		}},
		{"addresses", "Addresses", []string{"Prefix"}, []Record{row("ipv4", "IPv4 address", data.state, "Interface addressing uses documentation ranges.", "/routing", data.address)}},
	}
	if name == "eth0" || name == "eth3" {
		sections = append(sections, Section{"optics", "Optics", []string{"Reading", "Threshold"}, []Record{
			row("rx-power", "Receive power", "OK", "Optical module diagnostic sample.", "", "-5.86 dBm", "-14 dBm"),
			row("temp", "Module temperature", "OK", "Example module temperature.", "", "47.2 °C", "85 °C"),
		}})
	}
	return Page{Title: name, Description: data.kind + " · interface detail", Source: "Link statistics and optical diagnostics", Sections: sections}
}

func fixturePages() map[string]Page {
	return map[string]Page{
		"/health": {Title: "Health & Alerts", Description: "Active checks and recent health changes.", Source: "Health checks", Sections: []Section{
			{"active", "Active", []string{"Observed", "Threshold"}, []Record{
				row("vpn", "VPN handshake stale", "Warning", "The phone peer has not exchanged a handshake in 7 minutes.", "/vpn/wireguard", "7 min", "3 min"),
				row("dhcp", "LAN DHCP pool", "Warning", "The LAN pool is close to exhaustion.", "/services", "91%", "90%"),
			}},
			{"history", "History", []string{"Time"}, []Record{row("link-flap", "WAN link recovered", "OK", "eth0 was down for 4 seconds.", "/interfaces/eth0", "14:58")}},
		}},
		"/interfaces": {Title: "Interfaces", Description: "Link inventory, rates, and errors.", Source: "Link statistics", Sections: []Section{
			{"links", "All interfaces", []string{"Type", "Address", "Speed", "RX", "TX", "Errors"}, []Record{
				row("eth0", "eth0", "Up", "Primary WAN on a 10 Gbps optical link.", "/interfaces/eth0", "Ethernet · WAN", "203.0.113.42/24", "10G", "420 Mb/s", "72 Mb/s", "0"),
				row("br0", "br0", "Up", "Bridge for the example LAN.", "/interfaces/br0", "Bridge · LAN", "192.0.2.1/24", "—", "58 Mb/s", "690 Mb/s", "0"),
				row("eth3", "eth3", "Up", "Lab SFP+ uplink.", "/interfaces/eth3", "Ethernet · Lab", "198.51.100.17/24", "10G", "1.2 Mb/s", "0.4 Mb/s", "2"),
				row("wg0", "wg0", "Up", "Remote access tunnel.", "/interfaces/wg0", "WireGuard", "192.0.2.129/27", "—", "22 Mb/s", "9 Mb/s", "0"),
				row("pppoe0", "pppoe0", "Standby", "Backup WAN is not carrying traffic.", "/interfaces/pppoe0", "PPPoE", "—", "—", "—", "—", "—"),
			}},
		}},
		"/clients": {Title: "Traffic & Clients", Description: "Bandwidth consumers and identity evidence.", Source: "Flow accounting and DHCP leases", Sections: []Section{
			{"talkers", "Top talkers", []string{"Address", "Segment", "Identity", "Received", "Sent", "Last seen"}, []Record{
				row("nas", "nas-01", "Wired", "Identity from a static DHCP lease; traffic is aggregated over the last hour.", "/conntrack", "192.0.2.20", "LAN", "Static lease", "41.2 GB", "3.1 GB", "Now"),
				row("tv", "tv-livingroom", "Wired", "Identity from a DHCP hostname.", "/conntrack", "192.0.2.41", "IoT", "DHCP hostname", "18.7 GB", "0.2 GB", "Now"),
				row("phone", "phone", "Wi-Fi", "Randomized MAC address hidden; identity from a mock lease.", "", "192.0.2.12", "LAN", "Private MAC", "2.4 GB", "0.4 GB", "2 min ago"),
			}},
			{"paths", "Traffic paths", []string{"From", "To", "Volume"}, []Record{
				row("wan-lan", "Internet to LAN", "Observed", "Aggregate downstream path through the primary WAN.", "/interfaces/eth0", "eth0", "br0", "72 GB"),
				row("lan-wan", "LAN to Internet", "Observed", "Aggregate upstream path through the primary WAN.", "/topology", "br0", "eth0", "9 GB"),
			}},
		}},
		"/firewall": {Title: "Firewall", Description: "Rule counters and zone policy.", Source: "Firewall counters", Sections: []Section{
			{"chains", "Chains & rules", []string{"Chain", "Match", "Verdict", "Packets", "Bytes"}, []Record{
				row("wan-10", "Rule 10", "Accept", "Accept established and related packets.", "/conntrack", "WAN-IN", "state established, related", "accept", "1.28M", "1.3 GB"),
				row("wan-90", "Rule 90", "Drop", "Default drop for unmatched input.", "/logs", "WAN-IN", "any", "drop", "407", "38 KiB"),
				row("lan-20", "Rule 20", "Accept", "Permit LAN forwarding to the WAN.", "/topology", "LAN-OUT", "source LAN", "accept", "813k", "900 MiB"),
			}},
			{"zones", "Zone matrix", []string{"From", "To", "Verdict"}, []Record{
				row("lan-wan", "LAN → WAN", "Accept", "LAN traffic may leave through the WAN.", "", "LAN", "WAN", "allow"),
				row("wan-lan", "WAN → LAN", "Drop", "Unsolicited ingress is dropped.", "", "WAN", "LAN", "drop"),
			}},
		}},
		"/conntrack": {Title: "Conntrack & NAT", Description: "Example connection state, hardware path, and translations.", Source: "Conntrack and NAT snapshots", Sections: []Section{
			{"flows", "Flows", []string{"Protocol", "Source", "Destination", "Path"}, []Record{
				row("flow-41", "Flow 41", "Established", "Original 192.0.2.20:52646 → 198.51.100.80:443; reply 198.51.100.80:443 → 203.0.113.42:52646. Source NAT translates 192.0.2.20 to 203.0.113.42. Example hardware path.", "/dataplane", "TCP", "192.0.2.20:52646", "198.51.100.80:443", "Hardware"),
				row("flow-42", "Flow 42", "Replied", "Original 192.0.2.12:50253 → 198.51.100.53:53; reply 198.51.100.53:53 → 203.0.113.42:50253. Example DNS query tracked in software.", "/services", "UDP", "192.0.2.12:50253", "198.51.100.53:53", "Software"),
			}},
			{"nat", "NAT rules", []string{"Type", "Translation"}, []Record{row("src-100", "NAT 100", "Active", "Masquerade LAN traffic on WAN.", "/firewall", "Source", "203.0.113.42")}},
		}},
		"/routing": {Title: "Routing", Description: "BGP/BFD neighbor state and route inventory. Filter routes by prefix.", Source: "FRR routing snapshot", Sections: []Section{
			{"neighbors", "Neighbors", []string{"Peer", "ASN", "BFD"}, []Record{
				row("edge-a", "edge-a", "Established", "Peer is exchanging prefixes and BFD is up.", "/interfaces/eth0", "198.51.100.9", "64512", "Up"),
				row("edge-b", "edge-b", "Idle", "Backup neighbor has no active transport session.", "/interfaces/pppoe0", "198.51.100.10", "64513", "Down"),
			}},
			{"routes", "RIB / FIB", []string{"Next hop", "Protocol", "Interface"}, []Record{
				row("default", "0.0.0.0/0", "Installed", "Default IPv4 route selected from BGP.", "/interfaces/eth0", "203.0.113.1", "BGP", "eth0"),
				row("lan", "192.0.2.0/24", "Installed", "Connected route for LAN clients.", "/interfaces/br0", "On-link", "Connected", "br0"),
			}},
		}},
		"/vpn/wireguard": {Title: "WireGuard", Description: "Tunnel inventory, public identities, and handshake health.", Source: "WireGuard peer snapshot", Sections: []Section{
			{"peers", "Peers", []string{"Tunnel", "Public identity", "Endpoint", "Allowed IPs", "Handshake", "RX / TX", "Keepalive"}, []Record{
				row("laptop", "laptop-peer", "Healthy", "Example public identity only. No private key is present.", "/interfaces/wg0", "wg0", "laptop-peer-demo", "198.51.100.20:51820", "192.0.2.130/32", "25 sec ago", "1.2 GB / 340 MB", "25 s"),
				row("phone", "phone-peer", "Stale", "Example public identity only. Warning after 3 minutes; no private key is present.", "/health", "wg0", "phone-peer-demo", "198.51.100.21:51820", "192.0.2.131/32", "7 min ago", "280 MB / 88 MB", "25 s"),
			}},
			{"tunnels", "Tunnels", []string{"Listen port", "Peers"}, []Record{row("wg0", "wg0", "Up", "Remote access tunnel; no private key is present.", "/interfaces/wg0", "51820", "2 illustrated")}},
		}},
		"/qos": {Title: "QoS", Description: "Queue delay, CAKE tins, and scheduled probe comparisons.", Source: "Queue and probe snapshots", Sections: []Section{
			{"queues", "CAKE tins", []string{"Backlog", "Drops", "Peak delay"}, []Record{
				row("best-effort", "Best effort", "Healthy", "WAN egress queue, best effort tin.", "/interfaces/eth0", "24 KiB", "15", "8 ms"),
				row("bulk", "Bulk", "Warning", "Bulk traffic added queueing delay during the sample.", "/diagnostics", "61 KiB", "72", "19 ms"),
			}},
			{"probes", "Probe comparison", []string{"Idle", "Loaded", "Grade"}, []Record{row("wan-latency", "WAN latency", "Healthy", "Example bufferbloat grade is based on the difference between idle and loaded RTT.", "/diagnostics", "8.4 ms", "16.2 ms", "A")}},
		}},
		"/dataplane": {Title: "Dataplane (ASK2)", Description: "Simulated optional acceleration module; no hardware is attached.", Source: "ASK2 capability snapshot", Sections: []Section{
			{"offload", "Offload", []string{"Flows", "Share"}, []Record{
				row("hardware", "Hardware path", "Active", "Example traffic handled by the accelerated path.", "/conntrack", "9.9k", "93.8%"),
				row("software", "Software path", "Active", "Example flows handled by the kernel.", "/conntrack", "654", "6.2%"),
			}},
			{"queues", "Queues & pools", []string{"Used", "Capacity"}, []Record{
				row("ceetm", "CEETM queue 0", "Healthy", "Illustrative egress scheduling queue.", "/qos", "18", "128"),
				row("ceetm-1", "CEETM queue 1", "Healthy", "Illustrative bulk egress scheduling queue.", "/qos", "36", "128"),
				row("bman", "BMan buffer pool", "Healthy", "Example buffer allocation.", "", "112", "512"),
				row("shunt", "Suricata shunt", "Active", "Example inspected-versus-shunted dataplane snapshot.", "", "160", "1,024"),
			}},
		}},
		"/system": {Title: "System", Description: "Example CPU, memory, thermal, storage and boot evidence.", Source: "System resource snapshots", Sections: []Section{
			{"resources", "Resources", []string{"Reading", "Detail"}, []Record{
				row("cpu", "CPU", "Healthy", "Example core utilization: 12%, 17%, 15%, 12%.", "", "14%", "4 × Cortex-A72"),
				row("memory", "Memory", "Healthy", "Example memory usage from the router's system snapshot.", "", "1.9 GiB", "8 GiB total"),
				row("thermal", "SoC temperature", "Healthy", "Critical threshold is 95 °C.", "", "58 °C", "Fan 1,650 rpm"),
			}},
			{"storage", "Storage & images", []string{"Reading", "Detail"}, []Record{
				row("emmc", "eMMC wear", "Healthy", "Example EXT_CSD wear indicator.", "", "0x03", "20–30% used"),
				row("boot", "Active boot image", "Running", "Example prior boot: VyOS 1.5 rolling, revision r40.", "/logs", "VyOS 1.5 rolling", "12 d 4 h uptime"),
			}},
		}},
		"/config": {Title: "Configuration", Description: "Read-only fixture tree, set commands, redacted diff, and history.", Source: "Redacted configuration archive", Sections: []Section{
			{"tree", "Tree", []string{"Value"}, []Record{
				row("wan", "interfaces / ethernet / eth0", "Configured", "Primary WAN address and interface description.", "/interfaces/eth0", "203.0.113.42/24"),
				row("auth", "system / login / authentication", "Redacted", "No original secret exists in the demo data.", "", "[REDACTED]"),
			}},
			{"set", "Set commands", []string{"Command"}, []Record{row("wan", "WAN address", "Configured", "This example displays a command as text only.", "/interfaces/eth0", "set interfaces ethernet eth0 address '203.0.113.42/24'")}},
			{"diff", "Diff", []string{"Before", "After"}, []Record{
				row("ntp", "NTP source", "Changed", "Compare r40 with r41: example time source changed.", "/services", "198.51.100.123", "192.0.2.123"),
				row("auth", "Authentication value", "Redacted", "Change detected without revealing the old or new value.", "", "[REDACTED]", "[REDACTED]"),
			}},
			{"history", "History", []string{"When", "Summary"}, []Record{
				row("r41", "Revision r41", "Current", "Example commit by demo-admin.", "/logs", "15:12", "Update NTP source"),
				row("r40", "Revision r40", "Previous", "Example prior revision.", "", "Yesterday", "Set backup WAN"),
			}},
		}},
		"/logs": {Title: "Logs & Events", Description: "Filter the captured example journal or browse correlated events. No live stream.", Source: "Journal and event fixtures", Sections: []Section{
			{"journal", "Journal snapshot", []string{"Time", "Unit", "Message"}, []Record{
				row("wg-stale", "WireGuard peer stale", "Warning", "Example handshake warning for phone-peer.", "/vpn/wireguard", "18:41", "wireguard", "Handshake older than threshold"),
				row("dhcp-pool", "DHCP pool threshold", "Warning", "Example lease pool utilization exceeded 90%.", "/services", "17:02", "kea-dhcp4", "91% allocated"),
				row("wan-link", "WAN link recovered", "Info", "Example link returned after 4 seconds.", "/interfaces/eth0", "14:58", "kernel", "eth0 carrier restored"),
			}},
			{"events", "Timeline", []string{"Time", "Category"}, []Record{
				row("r41", "Commit r41", "Info", "A fixture configuration revision was recorded.", "/config", "15:12", "Configuration"),
				row("link", "WAN link flap", "Recovered", "The primary WAN recovered without intervention.", "/interfaces/eth0", "14:58", "Link"),
			}},
		}},
		"/topology": {Title: "Topology", Description: "Example network links and peers. Select a node to inspect its evidence.", Source: "Link, LLDP and BGP snapshots", Sections: []Section{
			{"nodes", "Nodes", []string{"Via", "Utilization"}, []Record{
				row("wan", "WAN uplink", "Up", "Primary physical WAN connection.", "/interfaces/eth0", "eth0 · 10G", "4.2%"),
				row("lan", "LAN bridge", "Up", "Local bridge connected to the example switch.", "/interfaces/br0", "br0", "6.9%"),
				row("vpn", "Remote peers", "Warning", "Logical WireGuard tunnel with a stale peer.", "/vpn/wireguard", "wg0 · tunnel", "2 peers"),
				row("bgp", "BGP upstream", "Established", "BGP neighbor 198.51.100.9.", "/routing", "eth0", "BFD up"),
			}},
		}},
		"/services": {Title: "Services", Description: "Read-only service health and evidence.", Source: "Service status snapshots", Sections: []Section{
			{"status", "Status", []string{"Reading"}, []Record{
				row("dhcp", "DHCP", "Warning", "Example LAN pool exceeds its warning threshold.", "/health", "233 / 256 leases"),
				row("dns", "DNS", "Healthy", "Example resolver is responding.", "/logs", "2.4 ms median"),
				row("ntp", "NTP", "Healthy", "Example synchronized clock; upstream 192.0.2.123.", "/config", "0.5 ms offset"),
			}},
		}},
		"/diagnostics": {Title: "Diagnostics", Description: "Recorded example probes only; no commands or probes can be triggered.", Source: "Scheduled probe results", Sections: []Section{
			{"results", "Probe results", []string{"When", "Result"}, []Record{
				row("latency", "WAN latency", "Healthy", "Read-only result from an example scheduled probe.", "/qos", "18:00", "8.4 ms RTT"),
				row("dns", "DNS resolution", "Healthy", "Example DNS probe result.", "/services", "18:01", "2.4 ms"),
			}},
		}},
		"/settings": {Title: "Settings", Description: "Appearance and fixture scenario are URL preferences for this demo.", Source: "Local demo options", Sections: []Section{
			{"display", "Display", nil, nil},
		}},
	}
}

package demo

import (
	"fmt"
	"strings"
)

func firewallZones(p *Page) {
	p.ZoneNames = []string{"LAN", "IoT", "Guest", "WAN", "VPN", "LOCAL"}
	type example struct {
		from, to, action string
		exceptions       int
	}
	examples := []example{
		{"LAN", "IoT", "accept", 0}, {"LAN", "Guest", "drop", 0}, {"LAN", "WAN", "accept", 0}, {"LAN", "VPN", "accept", 0}, {"LAN", "LOCAL", "mixed", 6},
		{"IoT", "LAN", "drop", 2}, {"IoT", "Guest", "drop", 0}, {"IoT", "WAN", "mixed", 4}, {"IoT", "VPN", "drop", 0}, {"IoT", "LOCAL", "mixed", 3},
		{"Guest", "LAN", "reject", 1}, {"Guest", "IoT", "drop", 0}, {"Guest", "WAN", "accept", 0}, {"Guest", "VPN", "drop", 0}, {"Guest", "LOCAL", "mixed", 2},
		{"WAN", "LAN", "mixed", 2}, {"WAN", "IoT", "drop", 0}, {"WAN", "Guest", "drop", 0}, {"WAN", "VPN", "drop", 0}, {"WAN", "LOCAL", "mixed", 5},
		{"VPN", "LAN", "accept", 1}, {"VPN", "IoT", "mixed", 2}, {"VPN", "Guest", "drop", 0}, {"VPN", "WAN", "accept", 0}, {"VPN", "LOCAL", "mixed", 3},
	}
	for _, e := range examples {
		policy := ZonePolicy{From: e.from, To: e.to, Action: e.action}
		for i := 0; i <= e.exceptions; i++ {
			id := fmt.Sprintf("zone-%s-%s-%d", strings.ToLower(e.from), strings.ToLower(e.to), i)
			name := e.from + " → " + e.to + " · default"
			match := "default policy"
			action := e.action
			if action == "mixed" {
				action = "drop"
			}
			if i > 0 {
				name = fmt.Sprintf("%s → %s · exception %d", e.from, e.to, i)
				match = "illustrative exception"
				if action == "accept" {
					action = "drop"
				} else {
					action = "accept"
				}
			}
			policy.RuleIDs = append(policy.RuleIDs, id)
			record := row(id, name, "Example", "Placeholder zone policy evidence; no match details or router configuration were read.", "", e.from+" → "+e.to, match, action, "—", "—")
			record.Category = "zone-policy"
			p.Sections[0].Records = append(p.Sections[0].Records, record)
		}
		p.ZonePolicies = append(p.ZonePolicies, policy)
	}
}

// enhancePage supplies illustrative snapshots. Sources are descriptions only;
// no commands are run by the demo.
func enhancePage(p *Page, path, rangeID string) {
	switch path {
	case "/clients":
		p.Summaries = map[string][]KPI{"talkers": {
			{"Clients", "47", "31 wired · 16 Wi-Fi", "ok"},
			{"Segments", "4", "LAN, IoT, Guest, Lab", "ok"},
			{"Accounting", "Enabled", "Example conntrack counters", "ok"},
		}}
		p.Evidence = map[string][]EvidencePanel{"talkers": {
			{Title: "Client identity evidence", Source: "Example DHCP, neighbor and LLDP snapshots", Columns: []string{"Address", "Segment", "Identity source"}, Rows: []EvidenceRow{
				{"nas-01", "/clients?tab=talkers&sel=nas", []string{"192.0.2.20", "LAN", "Static lease"}},
				{"tv-livingroom", "/clients?tab=talkers&sel=tv", []string{"192.0.2.41", "IoT", "DHCP hostname"}},
				{"phone", "/clients?tab=talkers&sel=phone", []string{"192.0.2.12", "LAN", "Randomized MAC · private"}},
			}},
			{Title: "Traffic paths", Source: "Example flow accounting", Columns: []string{"Destination", "Volume"}, Rows: []EvidenceRow{
				{"WAN eth0", "/clients?tab=paths&sel=wan-lan", []string{"LAN br0", "72 GB"}},
				{"LAN br0", "/clients?tab=paths&sel=lan-wan", []string{"WAN eth0", "9 GB"}},
			}},
		}}
		p.Series = map[string]*Series{"talkers": newSeries("Example aggregate client throughput", "Mb/s", rangeID, []int{160, 210, 250, 420, 690, 540, 420, 720, 180})}
	case "/firewall":
		firewallZones(p)
		p.Summaries = map[string][]KPI{"chains": {
			{"Chains", "3", "Example rulesets", "ok"},
			{"Policy", "Drop", "Default WAN input", "ok"},
			{"Dropped packets", "407", "Example snapshot", "warning"},
		}}
		p.Evidence = map[string][]EvidencePanel{"chains": {
			{Title: "Chains and named rulesets", Source: "Example firewall snapshot", Columns: []string{"Default action", "Rules"}, Rows: []EvidenceRow{
				{"WAN-IN", "/firewall?tab=chains&sel=wan-90", []string{"drop", "2 illustrated"}},
				{"LAN-OUT", "/firewall?tab=chains&sel=lan-20", []string{"accept", "1 illustrated"}},
			}},
		}}
	case "/conntrack":
		p.Summaries = map[string][]KPI{"flows": {
			{"Entries", "12.4k", "of 262k · 4.7%", "ok"},
			{"New flows", "184/s", "Example snapshot", "ok"},
			{"Offloaded", "9.9k", "Hardware path", "ok"},
			{"SNAT", "11.9k", "Masquerade eth0", "ok"},
			{"Drops", "0", "Example early drops", "ok"},
		}}
		p.Evidence = map[string][]EvidencePanel{"flows": {
			{Title: "Related flow evidence", Source: "Example flow snapshot", Columns: []string{"Evidence"}, Rows: []EvidenceRow{
				{"Client", "/clients", []string{"Client inventory"}},
				{"Firewall", "/firewall", []string{"Rule counters and zone policy"}},
				{"Route", "/routing?tab=routes", []string{"RIB / FIB"}},
				{"QoS", "/qos", []string{"Queue statistics"}},
			}},
		}}
	case "/routing":
		p.Summaries = map[string][]KPI{"neighbors": {
			{"IPv4 RIB", "1,042", "Example routes", "ok"},
			{"IPv6 RIB", "214", "Example routes", "ok"},
			{"BGP peers", "1 / 2", "1 idle", "warning"},
			{"BFD", "1 up", "1 down", "warning"},
		}}
		p.Series = map[string]*Series{"neighbors": newSeries("Example RIB size", "routes", rangeID, []int{980, 1000, 1010, 1030, 1040, 1035, 1042, 1050, 990})}
		p.Evidence = map[string][]EvidencePanel{"neighbors": {
			{Title: "Neighbor diagnosis", Source: "Example BGP notification", Columns: []string{"State", "Evidence"}, Rows: []EvidenceRow{
				{"edge-b", "/routing?tab=neighbors&sel=edge-b", []string{"Idle", "Example backup transport unavailable"}},
			}},
		}}
	case "/vpn/wireguard":
		p.Summaries = map[string][]KPI{"peers": {
			{"Interfaces", "1", "wg0", "ok"},
			{"Peers", "1 / 2", "1 stale handshake", "warning"},
			{"Acceleration", "Simulated", "Optional ASK2 capability", "ok"},
		}}
		p.Evidence = map[string][]EvidencePanel{"peers": {
			{Title: "Tunnel summary", Source: "Example WireGuard snapshot", Columns: []string{"Listen port", "Peer status", "Interface"}, Rows: []EvidenceRow{
				{"wg0", "/vpn/wireguard?tab=tunnels&sel=wg0", []string{"51820/UDP", "1 healthy · 1 stale", "wg0"}},
			}},
		}}
	case "/qos":
		p.Summaries = map[string][]KPI{"queues": {
			{"Shaper rate", "950 Mb/s", "Example eth0 egress", "ok"},
			{"Backlog", "85 KiB", "Illustrated tins", "ok"},
			{"Peak delay", "19 ms", "Bulk tin", "warning"},
			{"Bufferbloat", "A", "Example scheduled probe", "ok"},
		}}
		p.Evidence = map[string][]EvidencePanel{"queues": {
			{Title: "Latency under load", Source: "Example probe comparison", Columns: []string{"Idle", "Loaded", "Grade"}, Rows: []EvidenceRow{
				{"WAN latency", "/qos?tab=probes&sel=wan-latency", []string{"8.4 ms", "16.2 ms", "A"}},
			}},
		}}
	case "/dataplane":
		p.Summaries = map[string][]KPI{"offload": {
			{"Hardware flows", "9.9k", "Illustrative ASK2", "ok"},
			{"Software flows", "654", "Illustrative kernel path", "ok"},
			{"Offload hit", "93.8%", "Example snapshot", "ok"},
			{"Capability", "Simulated", "No hardware attached", "ok"},
		}}
		p.Series = map[string]*Series{"offload": newSeries("Example hardware flow count", "flows", rangeID, []int{6500, 7200, 8000, 9400, 11000, 10400, 9900, 12000, 6900})}
		p.Evidence = map[string][]EvidencePanel{"offload": {
			{Title: "Queues, buffers and shunt", Source: "Illustrative ASK2 capability", Columns: []string{"Usage", "Capacity"}, Rows: []EvidenceRow{
				{"CEETM queue 0", "/dataplane?tab=queues&sel=ceetm", []string{"18", "128"}},
				{"BMan buffer pool", "/dataplane?tab=queues&sel=bman", []string{"112", "512"}},
				{"Suricata shunt", "/dataplane?tab=queues&sel=shunt", []string{"160", "1,024"}},
			}},
		}}
	case "/system":
		p.Summaries = map[string][]KPI{"resources": {
			{"CPU", "14%", "4 × Cortex-A72", "ok"},
			{"Memory", "1.9 / 8 GiB", "24% used", "ok"},
			{"SoC temperature", "58 °C", "Critical 95 °C", "ok"},
			{"eMMC life", "20–30%", "Example estimate", "ok"},
			{"Uptime", "12 d 4 h", "Example boot", "ok"},
		}}
		p.Evidence = map[string][]EvidencePanel{"resources": {
			{Title: "Storage and boot history", Source: "Example system snapshot", Columns: []string{"Reading", "Status"}, Rows: []EvidenceRow{
				{"eMMC wear", "/system?tab=storage&sel=emmc", []string{"0x03 · 20–30% used", "Healthy"}},
				{"Active boot image", "/system?tab=storage&sel=boot", []string{"VyOS 1.5 rolling", "Running"}},
			}},
			{Title: "Services", Source: "Example systemd snapshot", Columns: []string{"State", "Evidence"}, Rows: []EvidenceRow{
				{"DHCP", "/services?sel=dhcp", []string{"Warning", "Pool 91% allocated"}},
				{"DNS", "/services?sel=dns", []string{"Healthy", "2.4 ms median"}},
				{"NTP", "/services?sel=ntp", []string{"Healthy", "0.5 ms offset"}},
			}},
		}}
	case "/config":
		p.Evidence = map[string][]EvidencePanel{"diff": {
			{Title: "Revision comparison", Source: "Example redacted archive", Columns: []string{"Revision", "Summary"}, Rows: []EvidenceRow{
				{"Before", "/config?tab=history&sel=r40", []string{"r40", "Prior revision"}},
				{"After", "/config?tab=history&sel=r41", []string{"r41", "Update NTP source"}},
			}},
		}}
	case "/logs":
		p.Evidence = map[string][]EvidencePanel{"journal": {
			{Title: "Event timeline", Source: "Example event detector", Columns: []string{"Time", "Category"}, Rows: []EvidenceRow{
				{"Commit r41", "/logs?tab=events&sel=r41", []string{"15:12", "Configuration"}},
				{"WAN link flap", "/logs?tab=events&sel=link", []string{"14:58", "Link"}},
			}},
		}}
	}
}

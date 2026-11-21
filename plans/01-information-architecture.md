# Information architecture (§4; PDF pp. 11–13)

The rail groups NETWORK, PLATFORM and INSPECT. Desktop rail is 232 px, collapsible to 64 px; below 640 px use five-destination bottom navigation and a More sheet. The top bar has breadcrumbs, global search, time range, live/pause, health, theme and user menu. Page header includes tabs, copy link, pin, export and source commands. Detail drawers use `sel=` and become full-screen sheets on phones.

| Section | Key views |
|---|---|
| Overview; Health & Alerts | Summary, kiosk; active, history, built-in checks |
| Interfaces | List and type-specific details: ethernet, bridge, VLAN, bond, PPPoE, WireGuard, tunnels, wireless, loopback/dummy |
| Traffic & Clients | Top talkers, inventory, client detail, protocols, conntrack, Sankey |
| Routing | RIB/FIB, static, BGP, OSPF, BFD, policy, VRFs |
| Firewall & NAT | Rules, zone matrix, groups, translations, conntrack, flowtables |
| VPN | WireGuard, IPsec, OpenVPN and optional Tailscale/ZeroTier |
| Services | DHCP, DNS, NTP, SSH, LLDP, mDNS, RA, DDNS, UPnP, SNMP, WAN failover |
| QoS; Dataplane | Policies and queues; optional ASK2 offload |
| System | Resources, sensors, storage, processes, units, kernel, images and IRQs |
| Configuration; Logs & Events | Tree, set, JSON, history, diff; journal and timeline |
| Diagnostics; Topology; Settings | Scheduled results; L2/L3 map; local browser preferences |

Paths identify entities; `range`, `from`, `to`, `filter`, `sort`, `tab`, `sel` and `cmp` capture view state in the URL. Theme and density remain in localStorage. The server renders the URL's complete initial state; signals mirror it and a history effect updates it. See Appendix E in [appendices](appendices.md) and the [complete route map](brief-extracted.txt).

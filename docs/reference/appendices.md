# Appendices & Reference Material

> Summarizes the updated URL/route map, data-source catalogue, domain glossary, and external references. Complete verbatim tables and code sketches from the original brief remain preserved in [brief-extracted.txt](brief-extracted.txt) (PDF pp. 68–76).

---

## Route Map

### Data-Plane Routes (`GET` / `HEAD` Only)

| Route | View / Endpoint |
| --- | --- |
| `/` | Overview (and `?kiosk=1`) |
| `/health` | Health & Alerts + Read-Only Security Check (`?tab=security`) |
| `/interfaces`, `/interfaces/{name}` | Interfaces 3-layer list & per-interface detail (`?tab=overview\|counters\|optics\|addresses\|queues\|config`) |
| `/clients`, `/clients/{id}`, `/conntrack` | Traffic & Clients, client/OT asset detail, virtualized Conntrack & NAT table |
| `/routing`, `/routing/{subview}` | Routing summary, per-VRF RIB/FIB (`?vrf=`), static, BGP, OSPF, BFD, PBR |
| `/firewall`, `/firewall/{subview}`, `/nat` | Zone policy matrix & rule table (`?zone=`), Firewall Security Check, groups, NAT/NAT66, flowtables |
| `/vpn`, `/vpn/{subview}` | WireGuard, IPsec, OpenVPN, optional Tailscale/ZeroTier |
| `/qos`, `/qos/{policy}` | Ingress & egress CAKE/FQ-CoDel shapers & bufferbloat history |
| `/topology` | Layered L2/L3 auto-drawn network map |
| `/redundancy`, `/redundancy/{subview}` | Capability-detected failover hub (`vrrp`, `conntrack-sync`, `wan`, `mrp`) |
| `/services`, `/services/{service}` | Podman containers (`containers`), DHCP, DNS, NTP/PTP, SSH, LLDP, mDNS, RA, DDNS, UPnP, SNMP |
| `/acceleration`, `/acceleration/{group}/{module}` | Capability-detected Dataplane / Acceleration tree (`summary`, `forwarding/ask2\|vpp\|flowtable\|af_xdp`, `crypto/caam\|qat`, `other`) |
| `/system`, `/system/{subview}` | Resources, thermals, Storage & images, processes, units, kernel/IRQs, self-health |
| `/config`, `/config/history` | Redacted config tree, `set` block, JSON, two-revision diff, commit history |
| `/logs`, `/events` | Severity histogram, journal tail, curated 30 d event timeline |
| `/diagnostics` | Step-by-step WAN connectivity chain & scheduled probe trends |
| `/settings` | Local UI preferences, units, timezone, retention & MCP read-only status |
| `/search`, `/export/{widget}` | Global command palette index & CSV/JSON/PNG widget export |
| `/sse/{page}`, `/a/{hash}.{ext}` | Per-page Datastar SSE stream & immutable pre-compressed embedded assets |
| `/metrics`, `/healthz` | Authenticated Prometheus/OpenMetrics endpoint (telemetry + health/alerts) & liveness check |

### Explicit Non-Data-Plane Exceptions
- **`/login`, `/logout`** (`GET` / `POST`): Authentication session endpoints protected by CSRF tokens.
- **Opt-in MCP endpoint** *(disabled by default)*: Dedicated authenticated interface for AI assistant read queries, gated diagnostics, and audit-logged `set → commit-confirm → save` writes (see [Security & MCP Model](../architecture/security-and-mcp.md#5-opt-in-mcp-interface--write-safety-model)).

---

## Data-Source Catalogue & Implementation Sketches

- **Data-source catalogue:** Canonical live mechanisms, privileges, and cadences are maintained in [System Architecture §3](../architecture/system-architecture.md#3-stage-1--data-gathering-pipeline) and [Performance & Testing §2](../architecture/performance-and-testing.md#2-per-collector-cpu-budget-breakdown-planning-estimates); original Appendix A tables are in [brief-extracted.txt](brief-extracted.txt) (pp. 68–69).
- **Datastar, UnoCSS & Icon sketches:** Original Appendices B–D (pp. 70–73) are preserved in [brief-extracted.txt](brief-extracted.txt); active icon rules use Font Awesome Pro 7.3.1 Sharp Light (`-tags fapro`) with Free 7.x fallback (see [Design System §4](../ui/design-system.md#4-iconography)).

---

## Glossary

| Term | Meaning |
| --- | --- |
| **AQM / CAKE** | Active Queue Management / Common Applications Kept Enhanced — fair-queuing shaper that fights bufferbloat using ECN marks and controlled drops across diffserv tins. |
| **ASK2 / DPAA1** | NXP LS1046A Application Solutions Kit hardware-offload stack built on Frame Manager (**FMan**), Queue Manager (**QMan**), Buffer Manager (**BMan**), and **CEETM** hardware QoS. |
| **CAAM / QAT** | Hardware cryptographic offload engines (NXP Cryptographic Acceleration & Assurance Module; Intel QuickAssist Technology). |
| **Conntrack / Flowtable** | Linux netfilter connection tracking table and fast-path bypass (`nf_flowtable`) for established flows. |
| **DOM** | Digital Optical Monitoring (SFF-8472 / SFF-8636) reporting SFP/SFP+ transceiver TX/RX optical power, temperature, bias current, and voltage. |
| **EXT_CSD** | eMMC Extended Card-Specific Data registers (`life_time` Type A/B wear estimates and `pre_eol_info` reserved-block alarm), exposed in Linux sysfs. |
| **FIB / RIB** | Forwarding Information Base (kernel routing table) vs. Routing Information Base (FRR control-plane table). |
| **MRP (IEC 62439-2)** | Media Redundancy Protocol — industrial Ethernet ring failover protocol (MRM manager / MRC client, blocked link). |
| **Morphing / Signals** | Datastar DOM diff-patching by element ID and reactive client/server state channels (`_chart.*` private signals). |
| **PSI** | Linux Pressure Stall Information (`/proc/pressure/*`) measuring CPU, memory, and I/O contention. |
| **Two-Plane Model** | Reconciling hardware/accelerator counters (Plane A) against kernel CPU counters (Plane B) to measure true `offload_ratio` and source-split drops. |
| **WAF** | Flash Write Amplification Factor — ratio of physical NAND writes to host OS writes (`~3×` typical on eMMC). |

---

## External References

- **Datastar:** https://data-star.dev/docs · https://github.com/starfederation/datastar
- **VyOS Docs:** https://docs.vyos.io/en/rolling/automation/vyos-api.html · https://docs.vyos.io/en/rolling/configuration/service/https.html
- **UI Libraries:** https://unocss.dev · https://fontawesome.com · https://github.com/leeoniya/uPlot · https://echarts.apache.org
- **Standards:** WCAG 2.2 (https://www.w3.org/TR/WCAG22/) · RFC 5737 & RFC 3849 (documentation IP ranges) · IEC 62439-2 (MRP)

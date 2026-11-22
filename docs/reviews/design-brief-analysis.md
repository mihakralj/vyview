# Design Brief v1.1 — Engineering Audit

> Historical review of *VyOS View-Only Dashboard Design Brief v1.1 (FA Pro, 17 Nov 2026)* documenting what was adopted, what was refined for this repository's two-file footprint, and where each requirement now lives in the canonical specs.

---

## 1. Summary of Refinements Applied Over the Original Brief

1. **Two-file router footprint (`vyview.ring`):** Consolidated the brief's separate per-tier RRD files, separate event log file, and optional `T1` shutdown dump into **one pre-allocated, 4 KiB-aligned single-file multi-ring datastore** with exact integer tier cascades (`200 ms → 5 s → 60 s → 5 min → 1 h`). See [Storage Format](../architecture/storage-format.md).
2. **Typography & geometry evolution:** Replaced the brief's original Inter / JetBrains Mono + rounded-card defaults with locally embedded **IBM Plex Sans** and **IBM Plex Mono** plus crisp flat square engineering panels. See [Design System](../ui/design-system.md).
3. **Font Awesome Pro Sharp Light dual-pack:** Standardised on Font Awesome Pro `7.3.1` **Sharp Light** (`-tags fapro`) with a 1:1 Font Awesome Free `7.x` fallback sharing identical SVG `<symbol>` IDs. See [Security & MCP §7](../architecture/security-and-mcp.md#7-font-awesome-pro-731-dual-pack-licensing-controls).
4. **Static binary (`CGO_ENABLED=0`) friction resolved upfront:** Replaced cgo `sd-journal` with one long-lived allowlisted `journalctl -o json -f` child process, and flagged `libpam` vs. pure-Go auth + narrow PAM helper for MVP Week 1 resolution. See [System Architecture §6](../architecture/system-architecture.md#6-static-binary-cgo_enabled0-trade-offs).
5. **Hardware & field feedback incorporated:** Added the Two-Plane Offload Model (preferring ASK2 Generic Netlink `ynl --family ask` over unstable `fe_flow` debugfs), unified **Dataplane / Acceleration** tree, dedicated **Redundancy** section (VRRP, `conntrack-sync`, WAN failover, MRP), Podman containers, Read-Only Security Checks, opt-in MCP (`set → commit-confirm → save`), and five binding UX principles.

---

## 2. Brief-to-Canonical-Spec Traceability

| Original Brief Section | Canonical Document in `docs/` |
| --- | --- |
| §1–3 Overview, Personas, Principles | [Product Overview](../architecture/overview.md) |
| §4 Information Architecture | [Information Architecture](../ui/information-architecture.md) |
| §5 Page Specifications | [Page Specifications](../ui/pages/README.md) (`01-overview.md` – `18-redundancy.md`) |
| §6 Visual Design System | [Design System](../ui/design-system.md) |
| §7–8 Dashboard Features & Mockups | [Dashboard Features & Mockups](../ui/dashboard-features.md) |
| §9 Architecture & Collectors | [System Architecture](../architecture/system-architecture.md) |
| §10 Storage & Retention | [Storage Format (`vyview.ring`)](../architecture/storage-format.md) |
| §11 Security, Redaction & Privacy | [Security & MCP Model](../architecture/security-and-mcp.md) |
| §12 Performance Budgets & §14 Testing | [Performance & Testing](../architecture/performance-and-testing.md) |
| §13 Packaging & Deployment | [Deployment & Packaging](../architecture/deployment-and-packaging.md) |
| §15 Roadmap, Risks & Decisions | [Roadmap, Decisions & Risks](../roadmap.md) |
| Appendices A–G & Raw PDF/Text | [Appendices](../reference/appendices.md), [PDF](../reference/VyOS-ViewOnly-Dashboard-Design-Brief.pdf), [Extraction](../reference/brief-extracted.txt) |

# vyview Documentation

Welcome to the engineering and product documentation for **vyview** — a single-router, LAN-only, real-time and historical dashboard for [VyOS](https://vyos.io).

The **web interface is strictly read-only by construction**. Alongside the web UI, an **opt-in MCP interface** allows AI assistants to answer operational questions and apply audit-logged configuration changes exclusively through VyOS's native `set → commit-confirm → save` workflow.

---

## Quick Reading Guides

- **New to vyview?** Start with [Product Overview](architecture/overview.md) and [Information Architecture](ui/information-architecture.md).
- **Building collectors, SSE, or storage?** Read [System Architecture](architecture/system-architecture.md), [Storage Format (`vyview.ring`)](architecture/storage-format.md), and [Performance & Testing](architecture/performance-and-testing.md).
- **Designing or implementing UI views?** Read [Design System](ui/design-system.md), [Dashboard Features & Mockups](ui/dashboard-features.md), and the [Page Specifications](ui/pages/README.md).
- **Reviewing security, deployment, or compliance?** Read [Security & MCP Model](architecture/security-and-mcp.md) and [Deployment & Packaging](architecture/deployment-and-packaging.md).

---

## Documentation Map

### Architecture & Engineering Specs (`docs/architecture/`)

| Document | Scope |
| --- | --- |
| [Product Overview](architecture/overview.md) | Vision, non-goals, five UX principles, personas (P1–P6 incl. OT), plain-language question workflows, and success metrics. |
| [System Architecture](architecture/system-architecture.md) | Runtime split (`vyview-web` + `vyview-probe`), Two-Plane Hardware Offload Model, collector catalogue, allowlist runner, Datastar/SSE transport, `CGO_ENABLED=0` trade-offs, and Local vs. Remote differences. |
| [Storage Format (`vyview.ring`)](architecture/storage-format.md) | Single-file, 4 KiB page-aligned multi-ring datastore (`200 ms`–`2 y` cascade), flash-wear math, cross-tier statistical rollups, bounded RAM tier warming, and crash recovery. |
| [Security & MCP Model](architecture/security-and-mcp.md) | Read-only web GUI enforcement, privilege split, authentication, CSP, server-side secret redaction, household privacy, Font Awesome Pro dual-pack licensing, and opt-in MCP `commit-confirm` safety. |
| [Deployment & Packaging](architecture/deployment-and-packaging.md) | Packaging sequence (`.deb`, container, upstream `service dashboard`), VyOS CLI & RBAC (`operator-group`), systemd units, USB `--data-path`, and delayed boot start. |
| [Performance & Testing](architecture/performance-and-testing.md) | LS1046A CPU, RSS, flash-write, payload, and OT/TSN latency budgets, per-collector cost estimates, and CI/soak/MCP write-safety gates. |

### User Interface & Screen Specifications (`docs/ui/`)

| Document | Scope |
| --- | --- |
| [Information Architecture](ui/information-architecture.md) | Collapsible left activity rail, breadcrumbs, `Ctrl/Cmd-K` palette, desktop drawers, 17-section inventory, and URL/state contract. |
| [Design System](ui/design-system.md) | Color tokens, typography (`IBM Plex Sans` / `IBM Plex Mono`), iconography (`Font Awesome Pro 7.3.1 Sharp Light` + Free fallback), flat square geometry, tables, units, and WCAG 2.2 AA motion. |
| [Dashboard Features & Mockups](ui/dashboard-features.md) | Cross-page widgets, `60 fps` `<vy-chart>` interpolation, time-range tier mapping, interactive behaviors, and visual mockup index (`M1`–`M16`). |
| [Page Specifications](ui/pages/README.md) | Standardized specifications for all 18 operational pages (`01-overview.md` through `18-redundancy.md`), including common loading/empty/stale/error rules. |

### Planning (`docs/`)

| Document | Scope |
| --- | --- |
| [Roadmap, Decisions & Risks](roadmap.md) | Milestone checklists (`v0.1` MVP through `v0.4` and post-`v0.4`), open architectural decisions, and risk mitigations. |

### Audits & Review Logs (`docs/reviews/`)

Non-normative historical logs tracing how external benchmarks and hardware field reviews shaped the canonical specs above:

| Document | Scope |
| --- | --- |
| [Design Brief Analysis](reviews/design-brief-analysis.md) | Engineering audit of the original *VyOS View-Only Dashboard Design Brief v1.1* against this repository's architecture. |
| [Monitoring Gap Analysis](reviews/monitoring-gap-analysis.md) | Benchmark against ntopng, btop++, Netdata, OPNsense Insight, OpenWrt LuCI, and MikroTik Torch on LS1046A hardware. |
| [Feedback Log — @Haico](reviews/feedback-haico.md) | Incorporation log for `@Haico`'s 25-point OT/prosumer review of the brief and `v0.1.0-demo`. |
| [Feedback Log — @Daegara](reviews/feedback-daegara.md) | Incorporation and cross-correlation log for `@Daegara`'s reviews of the tech stack, storage format, and `v0.1.0-demo`. |

### Reference Artifacts (`docs/reference/`)

| Artifact | Description |
| --- | --- |
| [Appendices](reference/appendices.md) | Updated URL/route map, data-source summary, glossary, and external references. |
| [Design Brief PDF](reference/VyOS-ViewOnly-Dashboard-Design-Brief.pdf) | Original immutable *VyOS View-Only Dashboard Design Brief v0.1.0* (76 pages). |
| [Extracted Brief Text](reference/brief-extracted.txt) | Full text extraction of the original brief (page markers correspond to PDF pages). |
| `docs/reference/mockups/` | Visual screen mockups `m01-overview.png` through `m16-topology.png` (indexed in [Dashboard Features & Mockups](ui/dashboard-features.md#mockups)). |

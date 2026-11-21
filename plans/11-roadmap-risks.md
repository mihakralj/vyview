# Roadmap, risks and decisions (§15; PDF pp. 66–67)

Estimates are person-weeks for one experienced developer; total to v0.4 is approximately 24–36.

## MVP v0.1 (6–8)

- [ ] Shell, PAM, Overview, Interfaces list/ethernet detail, System.
- [ ] Redacted configuration tree/set/diff/history and journal tail.
- [ ] Live rings, 10 s and 1 min tiers, SSE hub and privileged helper.
- [ ] Exit: LS1046A budgets met, golden tests for used commands, axe clean.

## v0.2 (5–7)

- [ ] Firewall/NAT counters, virtualized conntrack, DHCP/DNS/NTP and WireGuard/IPsec.
- [ ] Full flash retention, CSV/JSON exports.
- [ ] Exit: 50k conntrack load test and measured flash writes <50 MB/day.

## v0.3 (5–7)

- [ ] FRR routing, QoS/CAKE, topology, global search/palette, health/annotations and probes.
- [ ] Exit: palette <50 ms p95 and topology correct for reference and lab configs.

## v0.4 (4–6)

- [ ] ASK2 module once interfaces are defined; lazy Sankey/heatmaps, kiosk, PWA and mobile polish.
- [ ] Exit: graceful degradation on stock x86.

Later (4–8): optional notifications, risk-gated diagnostics actions, read-only federation and remote-write; separate threat model for each. If future interactive diagnostics are added: separate default-off role, fixed Go tools, validated restricted targets, rate limit/concurrency one and audit; no free-form arguments.

## Risks and open decisions

| Risk | Mitigation / owner |
|---|---|
| Op-mode structured gaps/release drift | Prefer kernel/FRR/Kea, release golden tests; Dev |
| Datastar CSP may require unsafe-eval | Verify the pinned version; static expressions, escaping, lint, track upstream; Dev/Sec |
| Flash wear | Time-major store, tmpfs, self-measure, lean profile; Arch |
| FRR JSON drift | Versioned tolerant parsers and goldens; Dev |
| CAP_NET_ADMIN | Split web/probe, netlink allowlist, fuzz RPC; Arch/Sec |
| Upstream acceptance | Custom image first, propose config node after v0.2; Owner |
| ASK2 contract undefined | Define capability/JSON contract, fake stub; Owner/Dev |
| BMan pool 512 | Surface free/depletion, alerts and tuning; ASK2 |
| Client privacy | Conservative defaults, settings and aggregate mode; Owner |
| Template engine | Prototype html/template and templ in week 1; prefer templ if tooling fits; Dev |

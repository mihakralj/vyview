# Redundancy

> **Primary Question:** *"If a router, WAN link, or ring cable fails, what takes over and will active sessions survive?"* · **Route:** `/redundancy`

---

## Capability-Detected Tree Structure

Dedicated top-level failover hub using the same auto-collapse rules as [Dataplane / Acceleration](10-dataplane.md): only configured/detected mechanisms appear; if only one exists, navigate straight to it without an intermediate tree; if none exist, hide the section.

```text
Redundancy
├── Router (VRRP)                 ← master/backup state, VRID, priority, preempt, failover timeline
├── State sync (conntrack-sync)   ← session sync health, peer reachability, dropped/failed syncs
├── WAN failover                  ← primary/backup WAN links, probe targets, active path, failover history
└── Rings (MRP / IEC 62439-2)     ← ring status (open/closed), MRM/MRC role, ring ports, blocked link, ring events
```

---

## Layout & Key Views

- **Router (VRRP):** Per-group VRID, virtual IPs, current state (`MASTER` / `BACKUP` / `FAULT`), priority, preempt status, peer IP, and failover event timeline.
- **State Sync (`conntrack-sync`):** Answers *"do active sessions survive a failover?"* — shows peer reachability, synced flow count, sync errors/drops, and last sync time.
- **WAN Failover:** Active vs. standby WAN uplinks, probe targets, trigger thresholds, and switchover history.
- **Rings (MRP / IEC 62439-2):** Shown once VyOS supports MRP — displays ring health (`Closed` / `Open`), MRM/MRC node role, ring ports, blocked port indicator, and sub-2-second ring transition timeline.

---

## Data Sources & Cadence

- **VRRP (`keepalived`):** Event-driven via `keepalived` D-Bus signals (zero polling) plus initial startup snapshot.
- **`conntrack-sync` (`10 s`):** Read-only `conntrackd` stats socket / allowlisted op-mode status.
- **WAN failover (`5 s`):** VyOS `wan-load-balance` / route-tracking status paired with built-in gateway probes.
- **MRP (IEC 62439-2):** Event-driven `bridge mrp` netlink events (`< 2 s` event-to-UI target, `100%` state-change capture in soak tests).

---

## UX & Consistency Constraints

- Every failover and ring-state transition feeds both **Health & Alerts** and the **Logs & Events** timeline automatically.
- Strictly read-only: never expose manual VRRP failover, preempt force, or `conntrack-sync` flush actions.

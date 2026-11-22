# Dataplane / Acceleration

> **Primary Question:** *"How much traffic is hardware-accelerated, are silicon pools/queues healthy, and why are remaining flows hitting the CPU?"* · **Route:** `/acceleration`

---

## Capability-Detected Tree Structure

Designed from day one for any single or multi-accelerator VyOS box (`LS1046A`, `x86_64`, `QAT`, `VPP`, `AF_XDP`) — never hard-coded to ASK2:

```text
Dataplane / Acceleration
├── Summary                  ← always first: per-port packet path + offload ratio (flows, packets, AND bytes)
├── Forwarding
│   ├── ASK2                 ← NXP LS1046A / Mono Gateway (FMan / QMan / BMan / CEETM / wgfp_rx)
│   ├── VPP                  ← supported NICs / x86 (VPP stats segment)
│   ├── Flowtable (SW / HW)  ← nftables flowtables when configured
│   └── AF_XDP               ← active XDP programs & AF_XDP sockets (marking AF_XDP ZC zero-copy)
├── Crypto
│   ├── CAAM                 ← e.g., LS1046A / Mono (IPsec/TLS HW crypto jobs vs. xfrm fallback)
│   └── QAT                  ← e.g., Intel Atom C3000 / Rangeley
└── Other                    ← unrecognized acceleration (ethtool -k, /proc/crypto, unexplained [HW_OFFLOAD])
```

- Empty groups disappear automatically. If only one accelerator is detected, navigate directly to its view without an intermediate tree. If none exist, hide the section or show `"Kernel only"`.

---

## Layout & Key Views

- **One Truth, Single Split Bar:** Never repeat the same flow counts across KPI tiles, "At a glance", and "Records". Use **one split HW/SW proportion bar** (not two mirrored bars) and always disambiguate whether an offload percentage counts **flows, packets, or bytes**.
- **HW vs. SW Time-Series Chart:** Stacked time-series chart splitting hardware-accelerated vs. software-path flows/bytes with Y-axis values and time labels.
- **Queues, Buffer Pools & Shunts:** Display CEETM queue heatmaps and BMan buffer-pool occupancy as `%` bars with clear thresholds (`< 30%` warn, `< 15%` critical), paired with a plain-language one-line explanation for `"Suricata shunt"`.
- **"Why Flows Aren't Offloaded" Breakdown:** Lists top software-path flows with concrete reasons (e.g., short-lived UDP below `hardware-offload-delay 500 ms`, unsupported L4/ALG helper, initial TCP handshake).
- **AF_XDP & OT/TSN Recognition:** Shows interfaces/queues running XDP or `AF_XDP ZC` zero-copy sockets and recognizes industrial real-time / TSN stacks and L2/IP protocols: **EtherCAT** (L2), **PROFINET (RT/IRT)** (L2/TSN), **OPC UA over TSN (FX/PubSub)**, **EtherNet/IP (CIP / CIP Sync)**, **CC-Link IE TSN**, **IEC/IEEE 60802**, and **Modbus TCP**.

---

## Data Sources & Cadence

- **ASK2 Generic Netlink first (`5 s`):** Read the ASK2 flow table via the build's **Generic Netlink interface** (`ynl --family ask` / `show offload flow`) inside `vyview-probe`, keeping `fe_flow` debugfs strictly as an optional guarded legacy fallback.
- **Two-Plane Reconciler:** Joins Plane A (`ethtool -S` FMan MAC/FIFO, ASK2 netlink, `fsl_dpa`/`qbman`, CAAM/QAT, `wgfp_rx`, VPP stats, `XDP_STATISTICS`) against Plane B (`netdev`, `conntrack` `[OFFLOAD]`/`[HW_OFFLOAD]`, `/proc/net/softnet_stat`, `/proc/net/xfrm_stat`).

---

## UX & Consistency Constraints

- Must degrade gracefully and pass all Playwright journeys on stock `x86_64` VyOS with zero offload hardware.
- Original field table: [brief-extracted.txt](../../reference/brief-extracted.txt) (pp. 20–21).

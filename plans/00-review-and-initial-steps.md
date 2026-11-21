# vyview — review and plan for initial steps

Source: *VyOS View-Only Dashboard — Design Brief* v0.1.0 (76 pages, 16 Nov 2026).

**Clarifications against the brief:** `/config/vyview` below is a proposed deviation, not the brief's preferred reference-hardware storage: §10.5 recommends a dedicated data partition and limits `/config` to the lean profile if used. The GET/HEAD-only rule applies to the data plane; §11.5 and Appendix E permit login/logout POST exceptions.

## 1. Review summary

**What the brief specifies.** vyview is a read-only, LAN-only dashboard that runs on the router. It is one static Go binary built with the Datastar Go SDK, server-rendered HTML, one SSE stream per page, UnoCSS, a Font Awesome subset, and uPlot. Collectors read from netlink, /proc and /sys, FRR, Kea, journald, and allowlisted op-mode commands. It has RAM rings, batched rollups to flash, and an optional ASK2 module. The MVP (v0.1) is estimated at 6–8 person-weeks.

**Strengths**
- The read-only guarantee is enforced by structure and backed by tests. There are no router or configuration mutation routes; authentication has the explicitly permitted POST exceptions. An exec allowlist with typed placeholders, no shell, an API client that can reach only allowlisted paths, and a split privileged probe reinforce the boundary.
- Budgets are measurable: first paint under 300 ms, under 3 % of one A72 core, under 40 MB RSS, and under 50 MB/day of flash writes. Each has a named measurement method.
- Development can start on a laptop. Fake collectors and golden-file parser tests mean most work doesn't need the router.
- The packaging path is pragmatic: ship first in this custom image (option a), then propose `service dashboard` upstream after v0.2.

**Gaps and risks to resolve early**

| # | Issue | Why it matters now | Proposal |
|---|-------|--------------------|----------|
| 1 | **Repository decision made.** The brief doesn't name a repository. | Implementation has a home. | Use `mihakralj/vyview`. `vyos-ls1046a-build` only packages the `.deb`. |
| 2 | **Datastar's CSP requirements depend on the pinned version.** The brief says its expressions require `unsafe-eval`. | If required, this weakens XSS containment. | In week 1, verify the pinned release's CSP requirements, prototype the policy, and add a lint that rejects dynamic `data-*` attributes. |
| 3 | **Templating engine is undecided** (`html/template` or `templ`). | It shapes every renderer. | Run a time-boxed spike in week 1. The brief prefers `templ`. |
| 4 | **Op-mode paths and output vary by release.** | Parsers will drift. | Capture golden outputs early from the LS1046A router running the current rolling build. Prefer netlink, /proc, and /sys in the MVP. |
| 5 | **The journald reader may need cgo.** | It conflicts with the goal of a static cross-compiled binary. | In the MVP, use one long-lived `journalctl -o json -f` process. |
| 6 | **The privileged probe is required in the MVP** (per Roadmap v0.1). | `CAP_NET_ADMIN` handling is security-sensitive. | The MVP pages barely need it: interfaces, system, config, and logs read without privileges. Build the probe skeleton and its RPC allowlist in the MVP, and add nft/conntrack/WireGuard calls in v0.2. |
| 7 | **Persistence across image upgrades** (§10.5). | Image-local `/var/lib` sits on the image overlay. | The brief recommends a dedicated persistent data partition for the reference hardware. As an alternative, propose `/config/vyview/` with a tmpfs hot tier for the lean profile; verify capacity, backup impact and persistence. |
| 8 | **The ASK2 adapter contract is undefined.** | It belongs to v0.4, but stubbing it is cheap. | Define the capability and JSON schema in a short spec. Don't implement it yet. |

## 2. Initial steps: a path to the MVP

Steps are ordered by dependency. Steps 1–3 can run on a laptop with fake data.

### Step 1 — Foundation and read-only skeleton (week 1)
- Create a new repository with a Go module and the layout `cmd/vyview`, `internal/{collect,parse,state,ring,hub,render,web,exec,probe}`, `web/` (assets), and `testdata/`.
- HTTP layer: allow GET and HEAD for the data plane and explicitly permit the brief's authentication POST exceptions for login/logout. Apply security headers and CSP and serve `go:embed` assets. Add a **route-table test** rejecting any other method or exception, and verify that no route mutates router state or configuration.
- Exec runner: an allowlist of argv templates with typed placeholders, per-command timeouts, output caps, and a minimal environment. Add a **static check** that fails if `os/exec` is used outside the runner.
- Assets: UnoCSS CLI build, Font Awesome sprite subset, and vendored uPlot and Datastar, all pre-compressed and embedded. CI only; the router needs no Node.js.
- Spike `templ` against `html/template` and decide.
- CI: build for arm64 and amd64, run `go test`, `govulncheck`, and a payload-size check.

### Step 2 — Live data core (weeks 2–3)
- Collector interface (`Collect(ctx)`, `Capabilities()`), a scheduler with jitter and backoff, and **fake collectors** with scenarios (link flap, high CPU).
- First real collectors:
  - netlink link stats and addresses
  - `/proc/{stat,meminfo,loadavg}`
  - `/sys` hwmon, thermal, and mmc (eMMC life)
- State model, 1-second ring buffers, and the SSE hub with per-page subscriptions, coalescing, backpressure, and 15-second keep-alive.
- Datastar patterns: morph by element id, and send chart data as private signals to a `<vy-chart>` uPlot web component.

### Step 3 — MVP pages (weeks 3–5)
- App shell: navigation, health pill, shared time range, and themes using the colour tokens from §6.
- **Overview**, **Interfaces** (list and ethernet detail with ethtool stats), and **System** (CPU, memory, thermal, storage, and fan from `fan-pid`).
- **Config viewer**: tree, `set`, diff, and history from `/config/archive`, with a **redaction** module and tests that cover every secret type.
- **Logs**: journal tail.
- Every widget gets a source chip, as the brief's source-transparency principle requires.
- Add 10-second and 1-minute rollup tiers.

### Step 4 — Router integration (weeks 5–7)
- Probe helper skeleton: a UNIX-socket RPC with a fixed method set, behind a separate unit with `CAP_NET_ADMIN`.
- PAM authentication with in-RAM sessions and the cookie rules from §11.5.
- Package the `.deb` with `vyview-web` and `vyview-probe` units, disabled by default. Wire it into **vyos-ls1046a-build** through `bin/ci-install-extra-packages.sh` or `ci-pick-packages.sh`, plus a new `data/hooks/*.chroot`. Per `AGENTS.md`, register any new hook in `bin/ci-setup-vyos-build.sh` in the **same commit**.
- On the device under test, capture golden outputs and measure CPU, RSS, and flash writes against §12.

### MVP exit criteria
The brief's v0.1 criteria: budgets met on the LS1046A, golden tests for every command used, and no serious axe violations.

## 3. Suggested first sessions

1. **vyview bootstrap**, in a new repository: Step 1 in full. One session, because everything in it shares the same files.
2. **Golden-output capture**, in `vyos-ls1046a-build` or the new repository: script the read-only capture of the allowlisted commands from the brief's Appendix A on the router. This is independent and can run in parallel. It needs SSH access to the device under test.
3. **ASK2 adapter contract**, an optional short spec in `we-are-mono/ASK` or the new repository: the capability and JSON schema only.

Steps 2–4 continue in session 1 after bootstrap. The packaging change in `vyos-ls1046a-build` starts once a `.deb` exists.

## 4. Decisions needed from you
1. Whether to accept the deviations proposed in §1: the `journalctl` fallback, a probe skeleton only in the MVP, and rollups under `/config/vyview`.
2. Whether sessions should have access to the device under test for golden captures and budget measurements.

The repository decision is made: `mihakralj/vyview`.

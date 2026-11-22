# Security, Privacy & Opt-In MCP Model

> Defines vyview's threat model, structural read-only enforcement on the web GUI, authentication and CSP rules, server-side secret redaction, privacy controls, Font Awesome Pro licensing boundaries, and the opt-in MCP write-safety model.

---

## 1. Security Posture & Privilege Boundaries

Browser input is treated as untrusted; OS access is confined to compile-time allowlisted read operations.

### Structural Read-Only Guarantees (Web GUI)
1. **HTTP route enforcement:** The entire data plane accepts only `GET` and `HEAD`. The **only** permitted `POST` exceptions are `/login` and `/logout` (protected by CSRF tokens). A CI route-table test fails if any other method or mutating route is registered.
2. **Exec runner confinement:** A static check fails CI if `os/exec` is imported outside `internal/exec`. The runner executes only compiled absolute `argv` templates (no `/bin/sh`) with typed entity placeholders, minimal `LANG=C`/fixed `PATH`, timeouts, and output caps.
3. **Netlink allowlist:** `vyview-probe` permits only read-only `*_GET*` and subscribe netlink message types (`DumpNft`, `DumpConntrack`, `SubscribeConntrack`, `WGDevices`, `EthtoolModule`, `DumpXfrm`, `Ask2Netlink`, `ReadHardwareStats`).
4. **Two-process split:** `vyview-web` holds **zero Linux capabilities** (`NoNewPrivileges=yes`, `ProtectSystem=strict`, `SystemCallFilter=@system-service`). Only `vyview-probe` holds `CAP_NET_ADMIN` (with `IPAddressDeny=any` and `RestrictAddressFamilies=AF_UNIX AF_NETLINK`).

---

## 2. Network Binding, TLS & Authentication

- **LAN / Management VRF only:** Bind explicitly to LAN or management VRF addresses (`SO_BINDTODEVICE` / `ip vrf exec mgmt`), never WAN. Wildcard binds (`0.0.0.0`/`::`) are refused unless verified input firewall rules protect the port.
- **HTTPS by default on LAN:** TLS 1.3 preferred (TLS 1.2 AEAD allowed), backed by VyOS PKI (`pki certificate <name>`) with automatic reload on rotation. Defaulting to HTTPS also guarantees browser Clipboard API (`Copy`) works across all LAN clients.
- **Authentication modes:**
  - **PAM / VyOS operator group** (default; align with `set system login operator-group vyview-ro command-policy allow 'show'`; validate PAM vs. `CGO_ENABLED=0` in MVP Week 1).
  - Optional local **`argon2id`** users (`m=64 MiB, t=3, p=1`), **mTLS** client certs (VyOS PKI), **trusted-proxy forward-auth**, and device-bound **kiosk tokens**.
  - Remote/off-router polling uses dedicated **SSH public-key-only** accounts (generated and tested on first run).
- **Sessions & rate limits:**
  - Random 256-bit session IDs stored in RAM (`12 h` idle, `7 d` absolute, rotated on login).
  - Cookies: `Secure; HttpOnly; SameSite=Strict`. Revalidate session on every SSE/fragment request.
  - Rate limits: login `5/min/IP` with lockout, search `10/s/session`, export `1/s/session`, new SSE streams `10/min/session`.

---

## 3. Content Security Policy (CSP) & XSS Containment

- **Static Datastar expressions only:** Datastar v1 expression evaluation requires `script-src 'self' 'unsafe-eval'` (verify against the exact pinned Datastar release in Week 1). To prevent XSS escalation, **every `data-*` attribute must be a static template literal** — never concatenated from dynamic or router-sourced input. A mandatory CI template lint enforces this alongside contextual auto-escaping.
- **Response headers:**
  - `Content-Security-Policy`: `default-src 'none'; script-src 'self' 'unsafe-eval' 'sha256-<prepaint-theme-hash>'; style-src 'self'; font-src 'self'; connect-src 'self'; manifest-src 'self'; worker-src 'self'; img-src 'self' data:; base-uri 'none'; form-action 'self'; frame-ancestors 'none'; object-src 'none'`
  - `X-Content-Type-Options: nosniff`, `Referrer-Policy: no-referrer`, restrictive `Permissions-Policy`, `Cross-Origin-Opener-Policy: same-origin`, and `Strict-Transport-Security`.

---

## 4. Server-Side Secret Redaction & Privacy

### Server-Side Redaction (19 Config Path Classes)
Secrets are redacted **on load** — before indexing, search, rendering, or diffing — with **no reveal control** and **no Copy button** on `[REDACTED]` fields:
- Covered paths: VyOS HTTP API keys, GraphQL JWT secrets, IPsec PSKs/keys, WireGuard private keys and PSKs (also dropped at the `wgctrl` collector boundary), login password hashes, SSH keys, RADIUS/TACACS shared secrets, PKI/OpenVPN private keys, BGP/OSPF/BFD auth keys, SNMP communities/v3 passphrases, PPPoE passwords, Wi-Fi PSKs, dynamic DNS tokens, VRRP auth secrets, and heuristic container environment variables (`*_SECRET`, `*_TOKEN`, `*_PASSWORD`, `*_KEY`).
- Diffs compare a **per-boot keyed HMAC** of the raw value so two-revision diffs display `•••••• (changed)` without ever exposing plaintext or persisting stable cross-boot hashes.

### Household & Network Privacy
- Configurable `privacy.per_client = off | aggregate | full`.
- Defaults to top-N clients only (`50 + other`), `30 d` client detail retention, and **zero DNS query logging**.
- Randomized client MAC addresses (LAA bit) are badged explicitly rather than misattributed as unknown vendors.

---

## 5. Opt-In MCP Interface & Write-Safety Model

While the web GUI stays 100% read-only, vyview provides an **opt-in Model Context Protocol (MCP)** interface (disabled by default) so AI assistants can assist operators safely:

| Capability | Default | Safety Controls |
| --- | --- | --- |
| **Read queries** | Opt-in | Answers every operational question in [Product Overview](overview.md#5-core-plain-language-question-workflows) using the same redacted snapshot and source-transparency evidence as the web GUI. |
| **On-demand diagnostics** (`ping`, `traceroute`) | Opt-in | Fixed Go tools / allowlisted invocations, validated target addresses (no free-form shell args), concurrency `1`, strict rate limits, and audit logging. |
| **Configuration writes** | **Explicit opt-in (off by default)** | **100% of writes go exclusively through VyOS's native config system (`set → commit-confirm → save`)** with automatic rollback timer protection. Every change is **100% audit-logged** (recording user identity, `MCP` source tag, commit comment, and `commit-confirm` status) and surfaced in **Configuration → Commit History**. |

- **Automated verification:** CI includes an MCP write-safety test suite asserting that no write bypasses `commit-confirm`, no raw shell execution is reachable, and every mutation emits an audit entry.

---

## 6. Read-Only Security Check Modules

vyview includes two built-in read-only audit checklists surfaced directly in the UI:
1. **Firewall Security Check (`Firewall & NAT → Security Check`, *"Is my firewall safe?"*):** Verifies WAN default-drop policy, checks for unintentionally WAN-exposed services, pairs port forwards with their filter rules, flags zones lacking policy, highlights rules never hit in `30 d`, and flags clients bypassing the router's DNS resolver.
2. **Router Security Posture (`Health & Alerts → Security`, *"Am I protected well enough?"*):** Aggregates the Firewall Security Check plus default password / factory SSH key detection, SSH management-interface-only verification, VyOS image age, certificate expiry (`< 21 d`), and unexpected listening ports (`NETLINK_SOCK_DIAG`).

---

## 7. Font Awesome Pro 7.3.1 Dual-Pack Licensing Controls

- **Pinned version:** Font Awesome `7.3.1` locked in build manifests; `icons.provenance.json` emitted per build.
- **Dual pack with identical SVG symbol IDs:**
  - **Private builds (`-tags fapro`):** Font Awesome Pro 7.3.1 **Sharp Light** (`fa-sharp fa-light`), generated locally or via private CI secret token into gitignored `internal/demo/pro_templates/`.
  - **Public / upstream VyOS builds (default):** Font Awesome Free 7.x fallback sprite mapping every Pro-only glyph 1:1 onto Free glyphs with required `NOTICE` attribution.
- **Leak guard:** `.gitignore` excludes Pro SVGs, `icons.provenance.json`, and `.npmrc`. CI fails any public build referencing Pro SVG path hashes, `@fortawesome/pro-*`, or `-tags fapro`.

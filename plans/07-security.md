# Security (§11; PDF pp. 56–59)

Browser input is untrusted; OS access is confined to allowlisted read operations. STRIDE mitigations: LAN/management binding, authentication and hardened cookies; no data-plane writes and validated signals; audit login and config views; server-side redaction and contextual escaping; stream/rate/output limits; capability-free web process and narrow helper. The complete threat table and illustrative systemd units are in [extraction](brief-extracted.txt), pp. 56–57.

Use unprivileged `vyview-web` plus `vyview-probe` with only `CAP_NET_ADMIN`, no IP listeners, fixed RPCs (DumpNft, DumpConntrack, WGDevices, EthtoolModule) and kernel GET-message allowlist. No shell. Test route and exec allowlists and netlink message types. **The brief contains a qualification:** data-plane routes are GET/HEAD only, but authentication permits login/logout GET/POST outside that plane; route tests must account for these explicitly.

PAM is the default; optional argon2id local users (m=64 MiB, t=3, p=1), mTLS, trusted-proxy forward auth and device-bound kiosk tokens. Sessions: random 256-bit IDs, RAM, 12 h idle, 7 d absolute, rotate on login; Secure/HttpOnly/SameSite=Strict cookie. Login CSRF token; revalidate each SSE/fragment. Limits: login 5/min/IP, search 10/s/session, export 1/s/session, streams 10/min/session.

The brief says Datastar expressions require `unsafe-eval`; verify that requirement against the pinned release before adopting it. Use static `data-*` expressions only, never concatenate dynamic input, lint templates and contextually escape all output. The brief's proposed CSP is default none, script self plus unsafe-eval, style/font/connect/manifest/worker self, image self/data, base none, form self, frame ancestors none, object none; hash the prepaint theme script. Add nosniff, no-referrer, restricted permissions, COOP same-origin and HSTS.

Bind explicitly to LAN/management VRF, never WAN; wildcard bind requires verified firewall. TLS 1.3 preferred, 1.2 AEAD allowed, reload VyOS PKI certificate changes.

Redact before indexing/rendering/diffing with constant marker plus per-boot HMAC to indicate changes; no reveal. Cover API/JWT, IPsec, WireGuard, login/SSH, RADIUS/TACACS, PKI/OpenVPN, BGP/OSPF, SNMP, PPPoE, Wi-Fi, dynamic DNS, VRRP and heuristic container environment secrets. Exact path patterns and table: [extraction](brief-extracted.txt), p. 59. Per-client history defaults top-N, detail 30 days, DNS query logging not collected; `privacy.per_client = off | aggregate | full`.

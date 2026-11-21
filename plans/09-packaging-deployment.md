# Packaging and deployment (§13; PDF pp. 63–64)

Options: (a) binary, `.deb`, systemd units and config in custom image, disabled by default; (b) VyOS container with read-only mounts and separate probe privileges, with host access complications; (c) upstream `service dashboard` config node, requiring acceptance. Start with (a); propose (c) after v0.2.

## Proposed CLI sketch (illustrative)

```text
set service dashboard listen-address 192.168.10.1
set service dashboard listen-address fd00:10::1
set service dashboard port 8443
set service dashboard vrf mgmt
set service dashboard certificate vyview-cert
set service dashboard authentication mode pam
set service dashboard authentication allow-group vyattaop
set service dashboard retention profile lean
set service dashboard retention path /var/lib/vyview
set service dashboard privacy per-client aggregate
set service dashboard modules ask2
```

The XML interface definition is a sketch, not an accepted upstream interface; see [extraction](brief-extracted.txt) p. 63. Its conf-mode script reads defaults, validates local/non-WAN listeners, PKI and port, generates `/run/vyview/vyview.toml`, then manages units.

Cross-compile Linux arm64/amd64 with `CGO_ENABLED=0`, `-trimpath`, `-ldflags='-s -w'`; cgo sd-journal requires a cross toolchain or use the long-lived journalctl fallback. CI builds UnoCSS, icon subset, content-hashed precompressed assets; router needs no Node. Pin Go/Node, publish CycloneDX SBOM and checksums. Tier headers carry format versions; upgrades migrate or preserve old files read-only to expiry. Config lives under `/config/vyview` for options a/b or is generated for c.

## Illustrative unit excerpts

```ini
# vyview-web.service
[Service]
User=vyview
Group=vyview
SupplementaryGroups=systemd-journal vyattacfg-ro
ExecStart=/usr/bin/vyview web --config /etc/vyview/vyview.toml
CapabilityBoundingSet=
AmbientCapabilities=
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
PrivateTmp=yes
ReadWritePaths=/var/lib/vyview /run/vyview
RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6 AF_NETLINK
SystemCallFilter=@system-service
SystemCallErrorNumber=EPERM
MemoryMax=96M
CPUQuota=25%
Environment=GOMEMLIMIT=40MiB GOMAXPROCS=2

# vyview-probe.service
[Service]
User=vyview-probe
ExecStart=/usr/bin/vyview probe --socket /run/vyview/probe.sock
AmbientCapabilities=CAP_NET_ADMIN
CapabilityBoundingSet=CAP_NET_ADMIN
RestrictAddressFamilies=AF_UNIX AF_NETLINK
IPAddressDeny=any
```

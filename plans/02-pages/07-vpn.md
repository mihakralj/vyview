# VPN (§5.7; PDF p. 19)

WireGuard interfaces and peers via wgctrl every 5 s, handshake age and rx/tx history; show only public keys and never read private keys. Expected-traffic peer stale thresholds: warn >3 min, critical >15 min. IPsec read-only VICI/swanctl every 10 s; OpenVPN status every 30 s; optional Tailscale/ZeroTier every 60 s and hidden when absent. Full table: [extraction](../brief-extracted.txt), page 19.

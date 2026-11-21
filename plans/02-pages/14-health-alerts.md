# Health & Alerts (§5.14; PDF p. 23)

Derived built-in checks evaluated every 5–60 s with evidence links and 90-day history. Checks include WAN, gateway loss >5% for 2 min, BGP down >2 min, disk >85%, temperature, certificate expiry <21 days, DHCP >90%, expected WireGuard stale >3 min, conntrack >75%/90%, failed unit, NTP >10 min and eMMC pre-EOL. Acknowledgment is browser-local; thresholds are configured outside the UI; no v1 outbound notifications. Full table: [extraction](../brief-extracted.txt), page 23.

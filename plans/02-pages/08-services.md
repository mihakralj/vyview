# Services (§5.8; PDF pp. 19–20)

Service status plus **VRRP / keepalived HA state timeline** (event-driven via `keepalived` D-Bus signals; master/backup transitions per group, zero polling), DHCP/Kea pool and leases (30 s; warn >90%), DNS (30 s), chrony (60 s; warn unsynchronized >10 min), SSH sessions (30 s), LLDP (60 s), mDNS, router advertisements, dynamic DNS (60 s; warn mismatch >15 min), UPnP (30 s), SNMP (60 s, secrets redacted), WAN load balancing (5 s). No session kill. Full fields, sources and retention: [extraction](../brief-extracted.txt), pages 19–20.

package system

import (
	"testing"
	"testing/fstest"
)

func TestCollect(t *testing.T) {
	root := fstest.MapFS{
		"proc/uptime":  {Data: []byte("90061.50 100.00\n")},
		"proc/loadavg": {Data: []byte("0.10 0.20 0.30 1/100 42\n")},
		"proc/meminfo": {Data: []byte("MemTotal: 1000 kB\nMemAvailable: 250 kB\n")},
	}
	snapshot := Collect(root)
	if snapshot.UptimeSeconds == nil || *snapshot.UptimeSeconds != 90061.5 || FormatUptime(*snapshot.UptimeSeconds) != "1d 1h 1m" {
		t.Fatalf("unexpected uptime: %v", snapshot.UptimeSeconds)
	}
	if snapshot.Load == nil || *snapshot.Load != [3]float64{0.1, 0.2, 0.3} {
		t.Fatalf("unexpected load: %v", snapshot.Load)
	}
	if snapshot.Memory == nil || snapshot.Memory.TotalBytes != 1024000 || snapshot.Memory.AvailableBytes != 256000 {
		t.Fatalf("unexpected memory: %+v", snapshot.Memory)
	}
}

func TestCollectUnavailableOrMalformed(t *testing.T) {
	root := fstest.MapFS{
		"proc/uptime":  {Data: []byte("-1 0\n")},
		"proc/loadavg": {Data: []byte("0.1 invalid 0.3\n")},
		"proc/meminfo": {Data: []byte("MemTotal: 100 kB\nMemAvailable: 101 kB\n")},
	}
	snapshot := Collect(root)
	if snapshot.UptimeSeconds != nil || snapshot.Load != nil || snapshot.Memory != nil {
		t.Fatalf("malformed data must be unavailable: %+v", snapshot)
	}
	if snapshot := Collect(fstest.MapFS{}); snapshot.UptimeSeconds != nil || snapshot.Load != nil || snapshot.Memory != nil {
		t.Fatalf("missing data must be unavailable: %+v", snapshot)
	}
}
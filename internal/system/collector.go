package system

import (
	"bufio"
	"fmt"
	"io/fs"
	"strconv"
	"strings"
)

type Snapshot struct {
	UptimeSeconds *float64
	Load          *[3]float64
	Memory        *Memory
}

type Memory struct {
	TotalBytes     uint64
	AvailableBytes uint64
}

// Collect reads a point-in-time snapshot. Individual fields remain nil when
// their source is missing or malformed, rather than reporting invented values.
func Collect(root fs.FS) Snapshot {
	var snapshot Snapshot
	if data, err := fs.ReadFile(root, "proc/uptime"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) > 0 {
			if value, err := strconv.ParseFloat(fields[0], 64); err == nil && value >= 0 {
				snapshot.UptimeSeconds = &value
			}
		}
	}
	if data, err := fs.ReadFile(root, "proc/loadavg"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) >= 3 {
			var load [3]float64
			valid := true
			for index := range load {
				value, err := strconv.ParseFloat(fields[index], 64)
				if err != nil || value < 0 {
					valid = false
					break
				}
				load[index] = value
			}
			if valid {
				snapshot.Load = &load
			}
		}
	}
	if data, err := fs.ReadFile(root, "proc/meminfo"); err == nil {
		values := make(map[string]uint64)
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) != 3 || fields[2] != "kB" || (fields[0] != "MemTotal:" && fields[0] != "MemAvailable:") {
				continue
			}
			value, err := strconv.ParseUint(fields[1], 10, 64)
			if err == nil && value <= ^uint64(0)/1024 {
				values[fields[0]] = value * 1024
			}
		}
		total, hasTotal := values["MemTotal:"]
		available, hasAvailable := values["MemAvailable:"]
		if scanner.Err() == nil && hasTotal && hasAvailable && total > 0 && available <= total {
			snapshot.Memory = &Memory{TotalBytes: total, AvailableBytes: available}
		}
	}
	return snapshot
}

func FormatUptime(seconds float64) string {
	minutes := uint64(seconds) / 60
	return fmt.Sprintf("%dd %dh %dm", minutes/1440, minutes%1440/60, minutes%60)
}
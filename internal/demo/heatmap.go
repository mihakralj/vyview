package demo

import (
	"fmt"
)

type Heatmap struct {
	Title, Unit string
	Capacity    int
	Buckets     []string
	Rows        []HeatmapRow
}

type HeatmapRow struct {
	ID, RecordID, Label string
	Samples             []int
}

type HeatmapCell struct {
	Value, Level int
	Label, Href  string
	Selected     bool
}

type HeatmapView struct {
	Title, Unit string
	Buckets     []string
	Rows        []HeatmapViewRow
	Selection   string
	Options     []NavItem
}

type HeatmapViewRow struct {
	Label string
	Cells []HeatmapCell
}

func renderHeatmap(data *Heatmap, visible map[string]bool, selectedID, sample string, href func(string, string) string) (*HeatmapView, error) {
	if data == nil || data.Capacity <= 0 {
		if sample != "" {
			return nil, ErrNotFound
		}
		return nil, nil
	}
	v := &HeatmapView{Title: data.Title, Unit: data.Unit, Buckets: data.Buckets}
	for _, row := range data.Rows {
		if !visible[row.RecordID] && row.RecordID != selectedID {
			continue
		}
		if len(row.Samples) != len(data.Buckets) {
			return nil, fmt.Errorf("heatmap row %q does not match bucket count", row.Label)
		}
		line := HeatmapViewRow{Label: row.Label}
		for i, value := range row.Samples {
			if value < 0 || value > data.Capacity {
				return nil, fmt.Errorf("heatmap value out of bounds in row %q", row.Label)
			}
			key := fmt.Sprintf("%s-%d", row.ID, i+1)
			reading := fmt.Sprintf("%d %s", value, data.Unit)
			if data.Unit == "%" {
				reading = fmt.Sprintf("%d%%", value)
			}
			label := fmt.Sprintf("%s, %s: %s", row.Label, data.Buckets[i], reading)
			active := row.RecordID == selectedID && key == sample
			line.Cells = append(line.Cells, HeatmapCell{value, int(float64(value) * 4 / float64(data.Capacity)), label, href(row.RecordID, key), active})
			if active {
				v.Selection = label
			}
		}
		v.Rows = append(v.Rows, line)
	}
	if sample != "" && v.Selection == "" {
		return nil, ErrNotFound
	}
	if len(v.Rows) == 0 {
		return nil, nil
	}
	return v, nil
}
